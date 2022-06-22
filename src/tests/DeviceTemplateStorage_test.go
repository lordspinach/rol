package tests

import (
	"fmt"
	"os"
	"rol/domain"
	"rol/infrastructure"
	"testing"
)

var testerTemplateStorage *GenericYamlStorageTest[domain.DeviceTemplate]

func Test_DeviceTemplateStorage_Prepare(t *testing.T) {
	dirName := "tests/testTemplates"
	storage := infrastructure.NewYamlGenericTemplateStorage(dirName, nil)
	testerTemplateStorage = NewGenericYamlStorageTest[domain.DeviceTemplate](storage, dirName, 30)
	err := createXTemplatesForTest(testerTemplateStorage.TemplatesCount)
	if err != nil {
		t.Errorf("creating templates failed: %s", err)
	}
}

func Test_DeviceTemplateStorage_GetByName(t *testing.T) {
	err := testerTemplateStorage.GenericYamlStorage_GetByName(fmt.Sprintf("AutoTesting_%d.yml", testerTemplateStorage.TemplatesCount/2))
	if err != nil {
		t.Errorf("get by name failed: %s", err)
	}
}

func Test_DeviceTemplateStorage_GetList(t *testing.T) {
	err := testerTemplateStorage.GenericYamlStorage_GetList()
	if err != nil {
		t.Errorf("get list failed: %s", err)
	}
}

func Test_DeviceTemplateStorage_Pagination(t *testing.T) {
	var pageSize int
	if testerTemplateStorage.TemplatesCount > 10 {
		pageSize = 10
	} else {
		pageSize = testerTemplateStorage.TemplatesCount / 2
	}
	err := testerTemplateStorage.GenericYamlStorage_Pagination(1, pageSize)
	if err != nil {
		t.Errorf("get list pagination failed: %s", err)
	}
}

func Test_DeviceTemplateStorage_Sort(t *testing.T) {
	err := testerTemplateStorage.GenericYamlStorage_Sort()
	if err != nil {
		t.Errorf("get list sort failed: %s", err)
	}
}

func Test_DeviceTemplateStorage_Filter(t *testing.T) {
	err := testerTemplateStorage.GenericYamlStorage_Filter()
	if err != nil {
		t.Errorf("get list filter failed: %s", err)
	}
}

func Test_DeviceTemplateStorage_DeleteTemplates(t *testing.T) {
	err := os.RemoveAll("testTemplates")
	if err != nil {
		t.Errorf("deleting dir failed: %s", err)
	}
}
