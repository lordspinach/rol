package infrastructure

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"rol/app/interfaces"
	"rol/domain"
	"runtime"
	"sort"
	"strconv"
	"strings"
)

//YamlGenericTemplateStorage is a storage for yaml templates
type YamlGenericTemplateStorage[TemplateType domain.DeviceTemplate] struct {
	//TemplatesDirectory is a directory where the templates are located
	TemplatesDirectory string
	logger             *logrus.Logger
	logSourceName      string
}

//NewYamlGenericTemplateStorage is a constructor for YamlGenericTemplateStorage
//
//Params:
//	dirName - directory name where the templates are located
//	log - logrus.Logger
func NewYamlGenericTemplateStorage[TemplateType domain.DeviceTemplate](dirName string, log *logrus.Logger) *YamlGenericTemplateStorage[TemplateType] {
	model := new(TemplateType)
	_, b, _, _ := runtime.Caller(0)
	templatesDirectory := path.Join(filepath.Dir(b), dirName)
	return &YamlGenericTemplateStorage[TemplateType]{
		TemplatesDirectory: templatesDirectory,
		logger:             log,
		logSourceName:      fmt.Sprintf("YamlGenericTemplateStorage<%s>", reflect.TypeOf(*model).Name()),
	}
}

func (y *YamlGenericTemplateStorage[TemplateType]) getTemplateObjFromYaml(templateName string) (*TemplateType, error) {
	template := new(TemplateType)
	templateFilePath := path.Join(y.TemplatesDirectory, fmt.Sprintf(templateName))
	f, err := os.Open(templateFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(template)
	if err != nil {
		return nil, err
	}
	return template, nil
}

func (y *YamlGenericTemplateStorage[TemplateType]) sortTemplatesSlice(templates *[]TemplateType, orderBy, orderDirection string) {
	sort.Slice(*templates, func(i, j int) bool {
		firstElem := (*templates)[i]
		secondElem := (*templates)[j]
		firstReflect := reflect.ValueOf(firstElem).FieldByName(orderBy)
		secondReflect := reflect.ValueOf(secondElem).FieldByName(orderBy)
		switch firstReflect.Kind() {
		case reflect.String:
			if strings.ToLower(orderDirection) == "desc" || strings.ToLower(orderDirection) == "descending" {
				return reflect.Indirect(firstReflect).String() > reflect.Indirect(secondReflect).String()
			} else {
				return reflect.Indirect(firstReflect).String() < reflect.Indirect(secondReflect).String()
			}
		case reflect.Int:
			if strings.ToLower(orderDirection) == "desc" || strings.ToLower(orderDirection) == "descending" {
				return reflect.Indirect(firstReflect).Int() > reflect.Indirect(secondReflect).Int()
			} else {
				return reflect.Indirect(firstReflect).Int() < reflect.Indirect(secondReflect).Int()
			}
		default:
			return false
		}
	})
}

//GetByName gets template by name
//Params:
//	ctx - context is used only for logging
//	templateName - name of template
//Return:
//	*TemplateType - pointer to template
//	error - if an error occurs, otherwise nil
func (y *YamlGenericTemplateStorage[TemplateType]) GetByName(ctx context.Context, templateName string) (*TemplateType, error) {
	return y.getTemplateObjFromYaml(templateName)
}

//GetList gets list of templates with filtering and pagination
//
//Params:
//	ctx - context is used only for logging
//	search - word for search in templates
//	orderBy - order by string parameter
//	orderDirection - ascending or descending order
//	page - page number
//	size - page size
//	queryBuilder - query builder for filtering
//Return:
//	*[]TemplateType - pointer to array of templates
//	error - if an error occurs, otherwise nil
func (y *YamlGenericTemplateStorage[TemplateType]) GetList(ctx context.Context, orderBy, orderDirection string, page, pageSize int, queryBuilder interfaces.IQueryBuilder) (*[]TemplateType, error) {
	var templatesSlice []TemplateType
	files, err := ioutil.ReadDir(y.TemplatesDirectory)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		template, err := y.getTemplateObjFromYaml(f.Name())
		if err != nil {
			return nil, err
		}
		templatesSlice = append(templatesSlice, *template)
	}
	queryBuild, err := queryBuilder.Build()
	if err != nil {
		return nil, err
	}
	y.find(&templatesSlice, queryBuild)
	y.sortTemplatesSlice(&templatesSlice, orderBy, orderDirection)
	return &templatesSlice, nil
}

//Count gets total count of templates with current query
//Params
//	ctx - context is used only for logging
//	queryBuilder - query for entities to count
//Return
//	int64 - number of entities
//	error - if an error occurs, otherwise nil
func (y *YamlGenericTemplateStorage[TemplateType]) Count(ctx context.Context, queryBuilder interfaces.IQueryBuilder) (int64, error) {
	var templatesSlice []TemplateType
	files, err := ioutil.ReadDir(y.TemplatesDirectory)
	if err != nil {
		return 0, err
	}
	for _, f := range files {
		template, err := y.getTemplateObjFromYaml(f.Name())
		if err != nil {
			return 0, err
		}
		templatesSlice = append(templatesSlice, *template)
	}
	y.find(&templatesSlice, queryBuilder)
	return int64(len(templatesSlice)), nil
}

//NewQueryBuilder gets new query builder
//Params
//	ctx - context is used only for logging
//Return
//	interfaces.IQueryBuilder - new query builder
func (y *YamlGenericTemplateStorage[TemplateType]) NewQueryBuilder(ctx context.Context) interfaces.IQueryBuilder {
	return NewYamlQueryBuilder()
}

func getSuitabilityOfValuesForConditions(condition Condition, templateReflect reflect.Value) bool {
	var valueMeetsCondition bool
	var conditions []bool
	for key, values := range condition.ConditionsMap {
		keySlice := strings.Split(key, " ")
		fieldName := keySlice[0]
		comparator := keySlice[1]

		for _, value := range values {
			var templateField any
			fieldReflect := templateReflect.FieldByName(fieldName)
			switch fieldReflect.Kind() {
			case reflect.String:
				templateField = templateReflect.FieldByName(fieldName).String()
			case reflect.Int:
				templateField = templateReflect.FieldByName(fieldName).Int()
			}

			switch comparator {
			case "==":
				conditions = append(conditions, templateField == value)
			case "!=":
				conditions = append(conditions, templateField != value)
			case ">":
				conditions = append(conditions, isBigger(templateField, value))
			case "<":
				conditions = append(conditions, isLesser(templateField, value))
			case ">=":
				conditions = append(conditions, isBiggerOrEqual(templateField, value))
			case "<=":
				conditions = append(conditions, isLesserOrEqual(templateField, value))
			}
		}
		if len(conditions) == 1 {
			valueMeetsCondition = conditions[0]
		} else if len(conditions) > 1 {
			for i := 0; i < len(conditions); i++ {
				if condition.Type == AND {
					valueMeetsCondition = valueMeetsCondition && conditions[i]
				} else {
					valueMeetsCondition = valueMeetsCondition || conditions[i]
				}
			}
		}
	}
	return valueMeetsCondition
}

func mergeSuitabilities(first, second map[int]bool, mergeType LogicType) map[int]bool {
	var out = make(map[int]bool)
	if len(first) > 0 {
		for i, firstVal := range first {
			if second[i] {
				if mergeType == OR {
					out[i] = firstVal || second[i]
				} else {
					out[i] = firstVal && second[i]
				}
			} else {
				out[i] = firstVal
			}
		}
	} else {
		for i, secondVal := range second {
			out[i] = secondVal
		}
	}
	return out
}

func (y *YamlGenericTemplateStorage[TemplateType]) find(templatesSlice *[]TemplateType, queryMaps interface{}) {
	suiteOfConditions := queryMaps.(*[]SuiteOfConditions)

	for templateIndex, template := range *templatesSlice {
		templateReflect := reflect.ValueOf(template)
		for conditionIndex, condition := range *suiteOfConditions {
			condition.AndMap.Suitability[templateIndex] = getSuitabilityOfValuesForConditions(condition.AndMap, templateReflect)
			condition.OrMap.Suitability[templateIndex] = getSuitabilityOfValuesForConditions(condition.OrMap, templateReflect)
			merged := mergeSuitabilities(condition.AndMap.Suitability, condition.OrMap.Suitability, OR)
			(*suiteOfConditions)[conditionIndex].MergedSuitability = merged
		}
	}
	templatesInterface := reflect.ValueOf(templatesSlice).Interface()

	var globalSuitability = make(map[int]bool)
	if len(*suiteOfConditions) > 1 {
		for i := 0; i < len(*suiteOfConditions); i++ {
			if i <= len(*suiteOfConditions)-2 {
				globalSuitability = mergeSuitabilities((*suiteOfConditions)[i].MergedSuitability, (*suiteOfConditions)[i+1].MergedSuitability, (*suiteOfConditions)[i+1].SuiteType)
			}
		}
	} else {
		globalSuitability = (*suiteOfConditions)[0].MergedSuitability
	}

	descMapKeys := getDescendingMapKeys(globalSuitability)
	for _, i := range descMapKeys {
		meetsCondition := globalSuitability[i]
		if !meetsCondition {
			remove(templatesInterface.(*[]domain.DeviceTemplate), i)
		}
	}

}

func remove(slice *[]domain.DeviceTemplate, i int) {
	s := *slice
	if len(s) > 1 {
		s = append((*slice)[:i], (*slice)[i+1:]...)
		*slice = s
	} else {
		*slice = make([]domain.DeviceTemplate, 0)
	}
}

func isBigger(first, second any) bool {
	switch first.(type) {
	case string:
		return first.(string) > second.(string)
	case int64:
		intVal, err := strconv.ParseInt(second.(string), 0, 64)
		if err != nil {
			panic("failed type assertion")
		}
		return first.(int64) > intVal
	default:
		panic("wrong type")
	}
}

func isBiggerOrEqual(first, second any) bool {
	switch first.(type) {
	case string:
		return first.(string) >= second.(string)
	case int64:
		intVal, err := strconv.ParseInt(second.(string), 0, 64)
		if err != nil {
			panic("failed type assertion")
		}
		return first.(int64) >= intVal
	default:
		panic("wrong type")
	}
}

func isLesser(first, second any) bool {
	switch first.(type) {
	case string:
		return first.(string) < second.(string)
	case int64:
		intVal, err := strconv.ParseInt(second.(string), 0, 64)
		if err != nil {
			panic("failed type assertion")
		}
		return first.(int64) < intVal
	default:
		panic("wrong type")
	}
}

func isLesserOrEqual(first, second any) bool {
	switch first.(type) {
	case string:
		return first.(string) <= second.(string)
	case int64:
		intVal, err := strconv.ParseInt(second.(string), 0, 64)
		if err != nil {
			panic("failed type assertion")
		}
		return first.(int64) <= intVal
	default:
		panic("wrong type")
	}
}

func getDescendingMapKeys(m map[int]bool) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(keys)))
	return keys
}
