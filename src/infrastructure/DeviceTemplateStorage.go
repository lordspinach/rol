package infrastructure

import (
	"github.com/sirupsen/logrus"
	"rol/domain"
)

type DeviceTemplateStorage struct {
	*YamlGenericTemplateStorage[domain.DeviceTemplate]
}

func NewDeviceTemplateStorage(dirName string, log *logrus.Logger) *DeviceTemplateStorage {
	return &DeviceTemplateStorage{
		NewYamlGenericTemplateStorage[domain.DeviceTemplate](dirName, log),
	}
}
