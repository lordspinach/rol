package infrastructure

import (
	"github.com/sirupsen/logrus"
	"rol/domain"
)

//DeviceTemplateStorage storage for domain.DeviceTemplate
type DeviceTemplateStorage struct {
	*YamlGenericTemplateStorage[domain.DeviceTemplate]
}

//NewDeviceTemplateStorage constructor for DeviceTemplateStorage
func NewDeviceTemplateStorage(dirName string, log *logrus.Logger) *DeviceTemplateStorage {
	return &DeviceTemplateStorage{
		NewYamlGenericTemplateStorage[domain.DeviceTemplate](dirName, log),
	}
}
