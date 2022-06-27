package tests

import (
	"os"
	"rol/app/services"
	"testing"
)

var (
	deviceTemplateService *services.DeviceTemplateService
	templatesCount        = 5
)

func Test_DeviceTemplateService_Prepare(t *testing.T) {
	err := createXTemplatesForTest(templatesCount)
	if err != nil {
		t.Errorf("creating dir failed: %s", err)
	}
	deviceTemplateService, err = services.NewDeviceTemplateService("testTemplates", nil)
	if err != nil {
		t.Errorf("creating service failed: %s", err.Error())
	}
}

func Test_DeviceTemplateService_GetByName(t *testing.T) {
	template, err := deviceTemplateService.GetByName(nil, "AutoTesting_3")
	if err != nil {
		t.Errorf("get by name failed: %s", err)
	}
	if template == nil {
		t.Error("template not found")
	}
	if template.Name != "AutoTesting_3" {
		t.Errorf("received wrong template: %v", template)
	}
}

func Test_DeviceTemplateService_GetList(t *testing.T) {
	templates, err := deviceTemplateService.GetList(nil, "", "", "", 1, templatesCount)
	if err != nil {
		t.Errorf("get list failed: %s", err)
	}
	if templates == nil {
		t.Error("templates are nil")
		return
	}
	if templates.Items == nil {
		t.Error("templates not found")
	}
	if len(*templates.Items) != templatesCount {
		t.Error("templates not found")
	}
}

func Test_DeviceTemplateService_Search(t *testing.T) {
	templates, err := deviceTemplateService.GetList(nil, "ValueForSearch", "", "", 1, templatesCount)
	if err != nil {
		t.Errorf("get list failed: %s", err)
	}
	if templates.Items == nil {
		t.Error("templates not found")
	}
	if len(*templates.Items) != 1 {
		t.Error("search failed")
	}
}

func Test_DeviceTemplateService_RemoveFiles(t *testing.T) {
	err := os.RemoveAll("../templates/testTemplates")
	if err != nil {
		t.Errorf("deleting dir failed: %s", err)
	}
}
