package infrastructure

import (
	"fmt"
	"rol/app/interfaces"
)

//YamlQueryBuilder query builder struct for yaml
type YamlQueryBuilder struct {
	//QueryConditions is a pointer to a slice of SuiteOfConditions
	QueryConditions *[]SuiteOfConditions
}

//SuiteOfConditions is a set of Condition's and their MergedSuitability.
//Also, its have own LogicType and bool value that indicates whether the condition was added from another query
type SuiteOfConditions struct {
	//AndMap a Condition with AND LogicType
	AndMap Condition
	//AndMap a Condition with OR LogicType
	OrMap Condition
	//MergedSuitability is a suitability of templates based on both AndMap and OrMap
	MergedSuitability map[int]bool
	//FinishedQuery indicates whether the condition was added from another query
	FinishedQuery bool
	//SuiteType is a LogicType of all suite which used for merging suitabilities between suits
	SuiteType LogicType
}

//Condition is a struct which contains set of rules, suitability for templates and its LogicType
type Condition struct {
	//ConditionsMap map which store conditions. Key has field name and comparator separated by space and interface{} contains field value
	//
	//For example:
	//	"Name ==": "MyName"
	//	"CPUCount >": 2
	ConditionsMap map[string][]interface{}
	//Suitability shows if template fits ConditionsMap conditions
	Suitability map[int]bool
	//Type is LogicType of Condition which will determine which logical operation will be used to ConditionsMap
	Type LogicType
}

//LogicType is set of AND and OR values
type LogicType uint

const (
	AND LogicType = iota
	OR
)

//NewYamlQueryBuilder is a constructor for YamlQueryBuilder
func NewYamlQueryBuilder() *YamlQueryBuilder {
	return &YamlQueryBuilder{
		QueryConditions: &[]SuiteOfConditions{},
	}
}

//Where
//	Add new AND condition to the query
//Params
// fieldName - name of the field
// comparator - logical comparison operator
// value - value of the field
//Return
// updated query
func (y *YamlQueryBuilder) Where(fieldName, comparator string, value interface{}) interfaces.IQueryBuilder {
	for _, suite := range *y.QueryConditions {
		if !suite.FinishedQuery {
			suite.AndMap.ConditionsMap[fieldName+" "+comparator] = append(suite.AndMap.ConditionsMap[fieldName+" "+comparator], value)
			return y
		}
	}
	var conditionMap = make(map[string][]interface{})
	conditionMap[fieldName+" "+comparator] = append(conditionMap[fieldName+" "+comparator], value)
	*y.QueryConditions = append(*y.QueryConditions, SuiteOfConditions{
		AndMap: Condition{
			ConditionsMap: conditionMap,
			Suitability:   make(map[int]bool),
			Type:          AND,
		},
		OrMap: Condition{
			ConditionsMap: make(map[string][]interface{}),
			Suitability:   make(map[int]bool),
			Type:          OR,
		},
		MergedSuitability: make(map[int]bool),
		FinishedQuery:     false,
		SuiteType:         AND,
	})
	return y
}

//WhereQuery
//	Add new complicated AND condition to the query based on another query
//Params
// builder - query builder
//Return
// updated query
func (y *YamlQueryBuilder) WhereQuery(builder interfaces.IQueryBuilder) interfaces.IQueryBuilder {
	build, _ := builder.Build()
	suite := build.(*[]SuiteOfConditions)
	for _, b := range *suite {
		b.SuiteType = AND
		*y.QueryConditions = append(*y.QueryConditions, b)
	}
	return y
}

//Or
//	Add new OR condition to the query
//Params
// fieldName - name of the field
// comparator - logical comparison operator
// value - value of the field
//Return
// updated query
func (y *YamlQueryBuilder) Or(fieldName, comparator string, value interface{}) interfaces.IQueryBuilder {
	for _, suite := range *y.QueryConditions {
		if !suite.FinishedQuery {
			suite.OrMap.ConditionsMap[fieldName+" "+comparator] = append(suite.OrMap.ConditionsMap[fieldName+" "+comparator], value)
			return y
		}
	}
	var conditionMap = make(map[string][]interface{})
	conditionMap[fieldName+" "+comparator] = append(conditionMap[fieldName+" "+comparator], value)
	*y.QueryConditions = append(*y.QueryConditions, SuiteOfConditions{
		AndMap: Condition{
			ConditionsMap: conditionMap,
			Suitability:   make(map[int]bool),
			Type:          AND,
		},
		OrMap: Condition{
			ConditionsMap: make(map[string][]interface{}),
			Suitability:   make(map[int]bool),
			Type:          OR,
		},
		MergedSuitability: make(map[int]bool),
		FinishedQuery:     false,
		SuiteType:         AND,
	})
	return y
}

//OrQuery
//	Add new complicated OR condition to the query based on another query
//Params
// builder - query builder
//Return
// updated query
func (y *YamlQueryBuilder) OrQuery(builder interfaces.IQueryBuilder) interfaces.IQueryBuilder {
	build, _ := builder.Build()
	suite := build.(*[]SuiteOfConditions)
	for _, b := range *suite {
		b.SuiteType = OR
		*y.QueryConditions = append(*y.QueryConditions, b)
	}
	return y
}

//Build
//	Build a slice of SuiteOfConditions
//Return
// slice of query arguments
// error - if an error occurred, otherwise nil
func (y *YamlQueryBuilder) Build() (interface{}, error) {
	if len(*y.QueryConditions) < 1 {
		return nil, fmt.Errorf("queryBuilder is empty")
	}
	return y.QueryConditions, nil
}
