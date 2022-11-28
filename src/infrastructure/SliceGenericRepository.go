package infrastructure

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"reflect"
	"rol/app/errors"
	"rol/app/interfaces"
	"sort"
	"strconv"
	"strings"
	"time"
)

//SliceGenericRepository repository that implements interfaces.IGenericRepository and use
// interfaces.IDataContext as entity storage instead of DB
type SliceGenericRepository[EntityIDType comparable, EntityType interfaces.IEntityModel[EntityIDType]] struct {
	//sliceCtx - slice data context
	sliceCtx interfaces.IDataContext[EntityIDType, EntityType]
	//logger - logrus logger
	logger *logrus.Logger
	//logSourceName - logger recording source
	logSourceName string
}

//QueryUnit represents bracketed expression that is part of query string
type QueryUnit struct {
	FieldName  string
	Comparator string
	ValueIndex int
}

//NewSliceGenericRepository constructor for SliceGenericRepository
//
//Params
//	interfaces.IDataContext[EntityIDType, EntityType] - file context
//	*logrus.Logger - logrus logger
//Return
//	*SliceGenericRepository[EntityIDType, EntityType] - repository for instantiated entity
func NewSliceGenericRepository[EntityIDType comparable, EntityType interfaces.IEntityModel[EntityIDType]](fileContext interfaces.IDataContext[EntityIDType, EntityType],
	log *logrus.Logger) *SliceGenericRepository[EntityIDType, EntityType] {
	model := new(EntityType)
	return &SliceGenericRepository[EntityIDType, EntityType]{
		sliceCtx:      fileContext,
		logger:        log,
		logSourceName: fmt.Sprintf("SliceGenericRepository<%s>", reflect.TypeOf(*model).Name()),
	}
}

func (r *SliceGenericRepository[EntityIDType, EntityType]) log(ctx context.Context, level, message string) {
	if ctx != nil {
		actionID := uuid.UUID{}
		if ctx.Value("requestID") != nil {
			actionID = ctx.Value("requestID").(uuid.UUID)
		}

		entry := r.logger.WithFields(logrus.Fields{
			"actionID": actionID,
			"source":   r.logSourceName,
		})
		switch level {
		case "err":
			entry.Error(message)
		case "info":
			entry.Info(message)
		case "warn":
			entry.Warn(message)
		case "debug":
			entry.Debug(message)
		}
	}
}

func (r *SliceGenericRepository[EntityIDType, EntityType]) getEntitiesSlice() ([]EntityType, error) {
	entitiesMap, err := r.sliceCtx.Get()
	if err != nil {
		return *new([]EntityType), errors.Internal.Wrap(err, "failed to get entities map")
	}
	slice := make([]EntityType, 0, len(entitiesMap))
	for _, value := range entitiesMap {
		slice = append(slice, value)
	}
	return slice, nil
}

func (r *SliceGenericRepository[EntityIDType, EntityType]) getPaginatedSlice(templates []EntityType, offset, limit int) ([]EntityType, error) {
	limit += offset
	if offset > len(templates) {
		return nil, errors.Internal.Newf("paginated slice offset bounds out of range [%d:] with length %d", offset, len(templates))
	}
	if limit > len(templates) {
		return templates[offset:], nil
	}
	return templates[offset:limit], nil
}

func (r *SliceGenericRepository[EntityIDType, EntityType]) sortSlice(templates *[]EntityType, orderBy, orderDirection string) error {
	if len(*templates) < 1 {
		return nil
	}
	if !isFieldExist((*templates)[0], orderBy) && orderBy != "" {
		return errors.Internal.Newf("there is no field with name '%s' at template", orderBy)
	}
	sort.Slice(*templates, func(i, j int) bool {
		firstElem := (*templates)[i]
		secondElem := (*templates)[j]
		firstReflect := reflect.ValueOf(firstElem).FieldByName(orderBy)
		secondReflect := reflect.ValueOf(secondElem).FieldByName(orderBy)
		switch firstReflect.Kind() {
		case reflect.String:
			if strings.ToLower(orderDirection) == "desc" || strings.ToLower(orderDirection) == "descending" {
				return reflect.Indirect(firstReflect).String() > reflect.Indirect(secondReflect).String()
			}
			return reflect.Indirect(firstReflect).String() < reflect.Indirect(secondReflect).String()

		case reflect.Int:
			if strings.ToLower(orderDirection) == "desc" || strings.ToLower(orderDirection) == "descending" {
				return reflect.Indirect(firstReflect).Int() > reflect.Indirect(secondReflect).Int()
			}
			return reflect.Indirect(firstReflect).Int() < reflect.Indirect(secondReflect).Int()

		case reflect.Struct:
			if firstReflect.Type().String() == "time.Time" {
				fTime := firstReflect.Interface().(time.Time)
				sTime := secondReflect.Interface().(time.Time)
				if strings.ToLower(orderDirection) == "desc" || strings.ToLower(orderDirection) == "descending" {
					return fTime.After(sTime)
				}
				return fTime.Before(sTime)
			}
			return false

		default:
			return false
		}
	})
	return nil
}

//GetList of elements with filtering and pagination
//
//Params
//	ctx - context is used only for logging
//	orderBy - order by string parameter
//	orderDirection - ascending or descending order
//	page - page number
//	size - page size
//	queryBuilder - query builder for filtering
//Return
//	*[]EntityType - pointer to array of entities
//	error - if an error occurs, otherwise nil
func (r *SliceGenericRepository[EntityIDType, EntityType]) GetList(ctx context.Context, orderBy string, orderDirection string, page int, size int, queryBuilder interfaces.IQueryBuilder) ([]EntityType, error) {
	r.log(ctx, "debug", "Call method GetList")
	var (
		foundEntities []EntityType
		queryArr      []interface{}
		err           error
	)
	slice, err := r.getEntitiesSlice()
	if err != nil {
		return nil, errors.Internal.Wrap(err, "failed to get entities slice")
	}
	offset := (page - 1) * size
	if queryBuilder != nil {
		query, err := queryBuilder.Build()
		if err != nil {
			return *new([]EntityType), errors.Internal.Wrap(err, "query building error")
		}
		queryArr = query.([]interface{})
	}
	if len(queryArr) > 1 {
		foundEntities, err = r.handleQuery(slice, queryArr...)
		if err != nil {
			return *new([]EntityType), errors.Internal.Wrap(err, "query handling error")
		}
	} else {
		foundEntities = slice
	}
	err = r.sortSlice(&foundEntities, orderBy, orderDirection)
	if err != nil {
		return *new([]EntityType), errors.Internal.Wrap(err, "templates sorting error")
	}
	return r.getPaginatedSlice(foundEntities, offset, size)
}

//Count gets total count of entities with current query
//
//Params
//	ctx - context
//	queryBuilder - query builder with conditions
//Return
//	int - number of entities
//	error - if an error occurs, otherwise nil
func (r *SliceGenericRepository[EntityIDType, EntityType]) Count(ctx context.Context, queryBuilder interfaces.IQueryBuilder) (int, error) {
	r.log(ctx, "debug", "Call method Count")
	slice, err := r.getEntitiesSlice()
	if err != nil {
		return -1, errors.Internal.Wrap(err, "failed to get entities slice")
	}

	if queryBuilder == nil {
		queryBuilder = r.NewQueryBuilder(ctx)
	}
	queryStr, err := queryBuilder.Build()
	if err != nil {
		return -1, errors.Internal.Wrap(err, "query building error")
	}
	queryArr := queryStr.([]interface{})
	foundEntities, err := r.handleQuery(slice, queryArr...)
	if err != nil {
		return -1, errors.Internal.Wrap(err, "failed to handle query")
	}
	return len(foundEntities), nil
}

//NewQueryBuilder gets new query builder
//
//Params
//	ctx - context is used only for logging
//Return
//	interfaces.IQueryBuilder - new query builder
func (r *SliceGenericRepository[EntityIDType, EntityType]) NewQueryBuilder(ctx context.Context) interfaces.IQueryBuilder {
	r.log(ctx, "debug", "Call method NewQueryBuilder")
	return NewSliceQueryBuilder()
}

//GetByID gets entity by ID from repository
//
//Params
//	ctx - context
//	id - entity id
//Return
//	EntityType - point to entity
//	error - if an error occurs, otherwise nil
func (r *SliceGenericRepository[EntityIDType, EntityType]) GetByID(ctx context.Context, id EntityIDType) (EntityType, error) {
	r.log(ctx, "debug", "Call method GetByID")
	return r.sliceCtx.GetByID(id)
}

//GetByIDExtended Get entity by ID and query from repository
//
//Params
//	ctx - context
//	id - entity id
//	queryBuilder - extended query conditions
//Return
//	*EntityType - point to entity
//	error - if an error occurs, otherwise nil
func (r *SliceGenericRepository[EntityIDType, EntityType]) GetByIDExtended(ctx context.Context, id EntityIDType, queryBuilder interfaces.IQueryBuilder) (EntityType, error) {
	r.log(ctx, "debug", "Call method GetByIDExtended")
	slice, err := r.getEntitiesSlice()
	if err != nil {
		return *new(EntityType), errors.Internal.Wrap(err, "failed to get entities slice")
	}

	if queryBuilder == nil {
		queryBuilder = r.NewQueryBuilder(ctx)
	}
	queryBuilder.Where("ID", "==", id)
	queryStr, err := queryBuilder.Build()
	if err != nil {
		return *new(EntityType), errors.Internal.Wrap(err, "query building error")
	}
	queryArr := queryStr.([]interface{})
	foundEntities, err := r.handleQuery(slice, queryArr...)
	if err != nil {
		return *new(EntityType), errors.Internal.Wrap(err, "failed to handle query")
	}
	if len(foundEntities) > 1 {
		return *new(EntityType), errors.Internal.New("more than one object found with given id")
	}
	if len(foundEntities) == 0 {
		return *new(EntityType), errors.NotFound.New("entity with given id was not found")
	}
	return foundEntities[0], nil
}

//Update save the changes to the existing entity in the repository
//
//Params
//	ctx - context
//	entity - updated entity to save
//Return
//	EntityType - updated entity
//	error - if an error occurs, otherwise nil
func (r *SliceGenericRepository[EntityIDType, EntityType]) Update(ctx context.Context, entity EntityType) (EntityType, error) {
	r.log(ctx, "debug", "Call method Update")
	err := r.sliceCtx.Update(entity.GetID(), entity)
	return entity, errors.Internal.Wrap(err, "failed to update entity")
}

//Insert entity to the repository
//
//Params
//	ctx - context
//	entity - entity to save
//Return
//	EntityType - created entity
//	error - if an error occurs, otherwise nil
func (r *SliceGenericRepository[EntityIDType, EntityType]) Insert(ctx context.Context, entity EntityType) (EntityType, error) {
	r.log(ctx, "debug", "Call method Insert")
	return r.sliceCtx.Add(entity)
}

//Delete entity from the repository
//
//Params
//	ctx - context
//	id - entity id
//Return
//	error - if an error occurs, otherwise nil
func (r *SliceGenericRepository[EntityIDType, EntityType]) Delete(ctx context.Context, id EntityIDType) error {
	r.log(ctx, "debug", "Call method Delete")
	return r.sliceCtx.Delete(id)
}

//Dispose releases all resources
//
//Return
//	error - if an error occurred, otherwise nil
func (r *SliceGenericRepository[EntityIDType, EntityType]) Dispose() error {
	return nil
}

//DeleteAll entities matching the condition
//
//Params
//	ctx - context
//	queryBuilder - query builder with conditions
//Return
//	error - if an error occurs, otherwise nil
func (r *SliceGenericRepository[EntityIDType, EntityType]) DeleteAll(ctx context.Context, queryBuilder interfaces.IQueryBuilder) error {
	r.log(ctx, "debug", "Call method DeleteAll")
	slice, err := r.getEntitiesSlice()
	if err != nil {
		return errors.Internal.Wrap(err, "failed to get entities slice")
	}
	if queryBuilder == nil {
		queryBuilder = r.NewQueryBuilder(ctx)
	}
	queryStr, err := queryBuilder.Build()
	if err != nil {
		return errors.Internal.Wrap(err, "query building error")
	}
	queryArr := queryStr.([]interface{})
	foundEntities, err := r.handleQuery(slice, queryArr...)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to handle query")
	}

	for _, entity := range foundEntities {
		err = r.sliceCtx.Delete(entity.GetID())
		if err != nil {
			return errors.Internal.Wrap(err, "failed to delete entity")
		}
	}
	return nil
}

//IsExist checks that entity is existed in repository
//
//Params
//	ctx - context is used only for logging
//	id - id of the entity
//	queryBuilder - query builder with addition conditions, can be nil
//Return
//	bool - true if existed, otherwise false
//	error - if an error occurs, otherwise nil
func (r *SliceGenericRepository[EntityIDType, EntityType]) IsExist(ctx context.Context, id EntityIDType, queryBuilder interfaces.IQueryBuilder) (bool, error) {
	r.log(ctx, "debug", "Call method IsExist")
	slice, err := r.getEntitiesSlice()
	if err != nil {
		return false, errors.Internal.Wrap(err, "failed to get entities slice")
	}
	if queryBuilder == nil {
		queryBuilder = r.NewQueryBuilder(ctx)
	}
	queryBuilder.Where("ID", "==", id)
	queryStr, err := queryBuilder.Build()
	if err != nil {
		return false, errors.Internal.Wrap(err, "query building error")
	}
	queryArr := queryStr.([]interface{})
	foundEntities, err := r.handleQuery(slice, queryArr...)
	if err != nil {
		return false, errors.Internal.Wrap(err, "failed to handle query")
	}
	if len(foundEntities) > 1 {
		return false, errors.Internal.New("more than one object found with given id")
	} else if len(foundEntities) == 1 {
		return true, nil
	}
	return false, nil
}

func (r *SliceGenericRepository[EntityIDType, EntityType]) handleQuery(templatesSlice []EntityType, args ...interface{}) ([]EntityType, error) {
	if len(args) < 1 {
		return templatesSlice, nil
	}
	query := replaceQuestionsToIndexes(args[0].(string))
	queryValues := args[1:]
	finalSlice := []EntityType{}
	for _, template := range templatesSlice {
		queryForTemplate := query
		startIndex, endIndex := findLowerQueryIndexes(queryForTemplate)
		for {
			if startIndex == -1 && endIndex == -1 {
				break
			}
			result, err := handleSimpleQuery(template, queryForTemplate[startIndex+1:endIndex-1], queryValues)
			if err != nil {
				return nil, errors.Internal.Wrap(err, "simple query handling error")
			}
			if result {
				queryForTemplate = replaceWithFakeTrueQuery(queryForTemplate, startIndex, endIndex)
			} else {
				queryForTemplate = replaceWithFakeFalseQuery(queryForTemplate, startIndex, endIndex)
			}
			startIndex, endIndex = findLowerQueryIndexes(queryForTemplate)
		}
		result, err := handleSimpleQuery(template, queryForTemplate, queryValues)
		if err != nil {
			return nil, errors.Internal.Wrap(err, "simple query handling error")
		}
		if result {
			finalSlice = append(finalSlice, template)
		}
	}
	return finalSlice, nil
}

func handleSimpleQuery(template interface{}, query string, queryValues []interface{}) (bool, error) {
	condition := ""
	queryUnitString, lastParsedIndex := getQueryUnitString(query, 0)
	result := false
	for {
		if len(queryUnitString) < 3 {
			break
		}
		queryUnit, err := parseQueryUnitString(strings.Trim(queryUnitString, " "))
		if err != nil {
			return false, errors.Internal.Wrap(err, "error parsing query unit")
		}
		if !isFieldExist(template, queryUnit.FieldName) && queryUnit.FieldName != "FakeFalse" && queryUnit.FieldName != "FakeTrue" {
			return false, errors.Internal.Newf("there is no field with name '%s' at template", queryUnit.FieldName)
		}
		value := queryValues[queryUnit.ValueIndex]
		if queryUnit.Comparator == "LIKE" {
			value = strings.Replace(value.(string), "%", "", -1)
		}
		interimResult, err := getResultOfQueryUnit(template, queryUnit, value)
		if err != nil {
			return false, errors.Internal.Wrap(err, "error getting result of query unit")
		}
		if condition == "" {
			result = interimResult
		} else if condition == "AND" {
			result = result && interimResult
		} else if condition == "OR" {
			result = result || interimResult
		}
		// Get condition if exist for the next iteration
		condition, lastParsedIndex = getConditionString(query, lastParsedIndex)
		queryUnitString, lastParsedIndex = getQueryUnitString(query, lastParsedIndex)
	}
	return result, nil
}

func replaceQuestionsToIndexes(query string) string {
	count := strings.Count(query, "?")
	for i := 0; i < count; i++ {
		query = strings.Replace(query, "?", strconv.Itoa(i), 1)
	}
	return query
}

func findLowerQueryIndexes(query string) (int, int) {
	endIndexOfQueryGroup := strings.Index(query, ")")
	if endIndexOfQueryGroup < 1 {
		return -1, -1
	}
	endIndexOfQueryGroup = endIndexOfQueryGroup + 1
	startIndexOfQueryGroup := strings.LastIndex(query[0:endIndexOfQueryGroup], "(")
	return startIndexOfQueryGroup, endIndexOfQueryGroup
}

func findConditionIndexAndLen(query string, searchStartIndex int) (int, int) {
	searchAbleQuery := query[searchStartIndex:]
	andIndex := strings.Index(searchAbleQuery, " AND ")
	orIndex := strings.Index(searchAbleQuery, " OR ")
	if orIndex != -1 && andIndex != -1 {
		if andIndex < orIndex {
			return searchStartIndex + andIndex + 1, 3
		}
		return searchStartIndex + orIndex + 1, 2
	}
	if orIndex != -1 {
		return searchStartIndex + orIndex + 1, 2
	}
	if andIndex != -1 {
		return searchStartIndex + andIndex + 1, 3
	}
	return -1, -1
}

func getFieldValue(template interface{}, fieldName string) (interface{}, error) {
	valueOfTemplate := reflect.ValueOf(template)
	if valueOfTemplate.Kind() == reflect.Ptr {
		valueOfTemplate = valueOfTemplate.Elem()
	}
	fieldReflect := valueOfTemplate.FieldByName(fieldName)
	var fieldValue interface{}
	kind := fieldReflect.Kind()
	switch kind {
	case reflect.String:
		fieldValue = valueOfTemplate.FieldByName(fieldName).String()
	case reflect.Int:
		fieldValue = int(valueOfTemplate.FieldByName(fieldName).Int())
	case reflect.TypeOf(time.Time{}).Kind():
		fieldValue = valueOfTemplate.FieldByName(fieldName).Interface().(time.Time)
	case reflect.TypeOf(uuid.UUID{}).Kind():
		fieldValue = valueOfTemplate.FieldByName(fieldName).Interface().(uuid.UUID)
	case reflect.Pointer:
		if fieldReflect.IsZero() {
			return nil, nil
		}
		if fieldReflect.Elem().Kind() == reflect.TypeOf(time.Time{}).Kind() {
			fieldValue = valueOfTemplate.FieldByName(fieldName).Interface().(*time.Time)
			break
		}
		return nil, errors.Internal.New("wrong field type")
	default:
		return nil, errors.Internal.New("wrong field type")
	}
	return fieldValue, nil
}

func isFieldExist(template interface{}, fieldName string) bool {
	return reflect.ValueOf(template).FieldByName(fieldName).IsValid()
}

func parseQueryUnitString(queryUnit string) (QueryUnit, error) {
	queryUnitSlice := strings.Split(queryUnit, " ")
	fieldName, comparator := queryUnitSlice[0], queryUnitSlice[1]
	valueIndex, err := strconv.Atoi(queryUnitSlice[2])
	if err != nil {
		return QueryUnit{}, errors.Internal.Wrap(err, "error converted to type int")
	}
	return QueryUnit{
		FieldName:  fieldName,
		Comparator: comparator,
		ValueIndex: valueIndex,
	}, nil
}

func isBigger(first, second any) (bool, error) {
	switch first.(type) {
	case string:
		return first.(string) > second.(string), nil
	case int:
		return first.(int) > second.(int), nil
	case time.Time:
		fTime := first.(time.Time)
		sTime := second.(time.Time)
		return fTime.After(sTime), nil
	default:
		return false, errors.Internal.New("wrong type")
	}
}

func isBiggerOrEqual(first, second any) (bool, error) {
	switch first.(type) {
	case string:
		return first.(string) >= second.(string), nil
	case int:
		return first.(int) >= second.(int), nil
	case time.Time:
		fTime := first.(time.Time)
		sTime := second.(time.Time)
		return fTime.After(sTime) || fTime.Equal(sTime), nil
	default:
		return false, errors.Internal.New("wrong type")
	}
}

func isLesser(first, second any) (bool, error) {
	switch first.(type) {
	case string:
		return first.(string) < second.(string), nil
	case int:
		return first.(int) < second.(int), nil
	case time.Time:
		fTime := first.(time.Time)
		sTime := second.(time.Time)
		return fTime.Before(sTime), nil
	default:
		return false, errors.Internal.New("wrong type")
	}
}

func isLesserOrEqual(first, second any) (bool, error) {
	switch first.(type) {
	case string:
		return first.(string) <= second.(string), nil
	case int:
		return first.(int) <= second.(int), nil
	case time.Time:
		fTime := first.(time.Time)
		sTime := second.(time.Time)
		return fTime.Before(sTime) || fTime.Equal(sTime), nil
	default:
		return false, errors.Internal.New("wrong type")
	}
}

func getResultOfQueryUnit(template interface{}, queryUnit QueryUnit, value interface{}) (bool, error) {
	// This is a hack
	if queryUnit.FieldName == "FakeTrue" {
		return true, nil
	}
	if queryUnit.FieldName == "FakeFalse" {
		return false, nil
	}

	fieldValue, err := getFieldValue(template, queryUnit.FieldName)
	if err != nil {
		return false, errors.Internal.Wrap(err, "error getting a field value")
	}
	switch queryUnit.Comparator {
	case "==":
		return fieldValue == value, nil
	case "!=":
		return fieldValue != value, nil
	case ">":
		return isBigger(fieldValue, value)
	case "<":
		return isLesser(fieldValue, value)
	case ">=":
		return isBiggerOrEqual(fieldValue, value)
	case "<=":
		return isLesserOrEqual(fieldValue, value)
	case "LIKE":
		return strings.Contains(fieldValue.(string), value.(string)), nil
	default:
		return false, errors.Internal.New("invalid comparator")
	}
}

func replaceWithFakeTrueQuery(query string, start, end int) string {
	return query[:start] + "FakeTrue == 0" + query[end:]
}

func replaceWithFakeFalseQuery(query string, start, end int) string {
	return query[:start] + "FakeFalse == 0" + query[end:]
}

func getQueryUnitString(query string, lastParsedIndex int) (string, int) {
	condIndex, _ := findConditionIndexAndLen(query, lastParsedIndex)
	if condIndex != -1 {
		return query[lastParsedIndex : condIndex-1], condIndex - 1
	}
	return query[lastParsedIndex:], len(query)
}

func getConditionString(query string, lastParsedIndex int) (string, int) {
	condIndex, condLength := findConditionIndexAndLen(query, lastParsedIndex)
	if condIndex != -1 {
		return query[condIndex : condIndex+condLength], condIndex + condLength
	}
	return "", lastParsedIndex
}
