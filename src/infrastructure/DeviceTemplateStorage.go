package infrastructure

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"rol/domain"
)

//DeviceTemplateStorage storage for domain.DeviceTemplate
type DeviceTemplateStorage struct {
	*YamlGenericTemplateStorage[domain.DeviceTemplate]
}

//NewDeviceTemplateStorage constructor for DeviceTemplateStorage
func NewDeviceTemplateStorage(config *domain.AppConfig, log *logrus.Logger) (*DeviceTemplateStorage, error) {
	storage, err := NewYamlGenericTemplateStorage[domain.DeviceTemplate](config.Templates.DeviceTemplateDir, log)
	if err != nil {
		return nil, fmt.Errorf("device templates storage creating error: %s", err.Error())
	}
	return &DeviceTemplateStorage{
		storage,
	}, nil
}
