package infrastructure

import (
	"fmt"
	"rol/app/interfaces"
	"strings"
)

//SliceQueryBuilder query builder struct for yaml
type SliceQueryBuilder struct {
	QueryString string
	Values      []interface{}
}

//NewSliceQueryBuilder is a constructor for SliceQueryBuilder
func NewSliceQueryBuilder() *SliceQueryBuilder {
	return &SliceQueryBuilder{}
}

func (b *SliceQueryBuilder) addQuery(condition, fieldName, comparator string, value interface{}) interfaces.IQueryBuilder {
	if len(b.QueryString) > 0 {
		b.QueryString += fmt.Sprintf(" %s ", condition)
	}
	b.QueryString += fmt.Sprintf("%s %s ?", fieldName, comparator)
	b.Values = append(b.Values, value)
	return b
}

func (b *SliceQueryBuilder) addQueryBuilder(condition string, builder interfaces.IQueryBuilder) interfaces.IQueryBuilder {
	if len(b.QueryString) > 0 {
		b.QueryString += fmt.Sprintf(" %s ", condition)
	}
	argsInterface, err := builder.Build()
	if err != nil {
		return b
	}
	argsArrInterface := argsInterface.([]interface{})
	switch v := argsArrInterface[0].(type) {
	case string:
		if len(argsArrInterface[0].(string)) < 1 {
			return b
		}
		b.QueryString += fmt.Sprintf("(%s)", strings.ReplaceAll(v, "WHERE ", ""))
	default:
		panic("[SliceQueryBuilder] can't add passed query builder to current builder, check what you pass SliceQueryBuilder")
	}
	for i := 1; i < len(argsArrInterface); i++ {
		b.Values = append(b.Values, argsArrInterface[i])
	}
	return b
}

//Where add new AND condition to the query
//Params
// fieldName - name of the field
// comparator - logical comparison operator
// value - value of the field
//Return
// updated query
func (b *SliceQueryBuilder) Where(fieldName, comparator string, value interface{}) interfaces.IQueryBuilder {
	return b.addQuery("AND", fieldName, comparator, value)
}

//WhereQuery add new complicated AND condition to the query based on another query
//Params
// builder - query builder
//Return
// updated query
func (b *SliceQueryBuilder) WhereQuery(builder interfaces.IQueryBuilder) interfaces.IQueryBuilder {
	return b.addQueryBuilder("AND", builder)
}

//Or add new OR condition to the query
//Params
// fieldName - name of the field
// comparator - logical comparison operator
// value - value of the field
//Return
// updated query
func (b *SliceQueryBuilder) Or(fieldName, comparator string, value interface{}) interfaces.IQueryBuilder {
	return b.addQuery("OR", fieldName, comparator, value)
}

//OrQuery add new complicated OR condition to the query based on another query
//Params
// builder - query builder
//Return
// updated query
func (b *SliceQueryBuilder) OrQuery(builder interfaces.IQueryBuilder) interfaces.IQueryBuilder {
	return b.addQueryBuilder("OR", builder)
}

//Build a slice of query arguments
//Return
// slice of query arguments
// error - if an error occurred, otherwise nil
func (b *SliceQueryBuilder) Build() (interface{}, error) {
	arr := make([]interface{}, 0)
	if len(b.QueryString) < 1 {
		return arr, nil
	}
	arr = append(arr, b.QueryString)
	arr = append(arr, b.Values...)
	return arr, nil
}
