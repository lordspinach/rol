package tests

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"rol/domain"
	"rol/infrastructure"
	"testing"
)

var testerTemplateStorage *GenericYamlStorageTest[domain.DeviceTemplate]

func Test_DeviceTemplateStorage_Prepare(t *testing.T) {
	dirName := "testTemplates"
	storage, err := infrastructure.NewYamlGenericTemplateStorage(dirName, logrus.New())
	if err != nil {
		t.Errorf("creating templates storage failed: %s", err.Error())
	}
	testerTemplateStorage = NewGenericYamlStorageTest[domain.DeviceTemplate](storage, dirName, 30)
	err = createXTemplatesForTest(testerTemplateStorage.TemplatesCount)
	if err != nil {
		t.Errorf("creating templates failed: %s", err)
	}
}

func Test_DeviceTemplateStorage_GetByName(t *testing.T) {
	err := testerTemplateStorage.GetByNameTest(fmt.Sprintf("AutoTesting_%d", testerTemplateStorage.TemplatesCount/2))
	if err != nil {
		t.Errorf("get by name failed: %s", err)
	}
}

func Test_DeviceTemplateStorage_GetList(t *testing.T) {
	err := testerTemplateStorage.GetListTest()
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
	err := testerTemplateStorage.PaginationTest(1, pageSize)
	if err != nil {
		t.Errorf("get list pagination failed: %s", err)
	}
}

func Test_DeviceTemplateStorage_Sort(t *testing.T) {
	err := testerTemplateStorage.SortTest()
	if err != nil {
		t.Errorf("get list sort failed: %s", err)
	}
}

func Test_DeviceTemplateStorage_Filter(t *testing.T) {
	err := testerTemplateStorage.FilterTest()
	if err != nil {
		t.Errorf("get list filter failed: %s", err)
	}
}

func Test_DeviceTemplateStorage_DeleteTemplates(t *testing.T) {
	err := os.RemoveAll("../templates/testTemplates")
	if err != nil {
		t.Errorf("deleting dir failed: %s", err)
	}
}
