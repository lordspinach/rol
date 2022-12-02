// Package tests contains project unit tests and all related structs and interfaces
package tests

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"reflect"
	"rol/domain"
	"rol/infrastructure"
)

//YamlSliceGenericRepositoryTester generic struct for slice repository tester
type YamlSliceGenericRepositoryTester[IDType comparable, EntityType ITestEntity[IDType]] struct {
	*GenericRepositoryTester[IDType, EntityType]
	directoryName string
}

//NewYamlSliceGenericRepositoryTester constructor for slice generic repository tester
func NewYamlSliceGenericRepositoryTester[IDType comparable, EntityType ITestEntity[IDType]]() *YamlSliceGenericRepositoryTester[IDType, EntityType] {
	filePath, _ := os.Executable()
	diParams := domain.GlobalDIParameters{
		RootPath: filepath.Dir(filePath),
	}

	dirName := fmt.Sprintf("/templates/%s/", reflect.TypeOf(*new(EntityType)).Name())
	fileContext, err := infrastructure.NewYamlManyFilesContext[IDType, EntityType](diParams, dirName)
	if err != nil {
		panic(err.Error())
	}
	repo := infrastructure.NewSliceGenericRepository[IDType, EntityType](fileContext, logrus.New())
	genericTester, _ := NewGenericRepositoryTester[IDType, EntityType](repo)
	genericTester.Implementation = "YAML/FILE"
	tester := &YamlSliceGenericRepositoryTester[IDType, EntityType]{
		GenericRepositoryTester: genericTester,
		directoryName:           dirName,
	}
	return tester
}
