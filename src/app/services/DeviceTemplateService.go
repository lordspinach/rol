package services

import (
	"context"
	"github.com/sirupsen/logrus"
	"rol/app/interfaces"
	"rol/domain"
	"rol/infrastructure"
)

type DeviceTemplateService struct {
	storage interfaces.IGenericTemplateStorage[domain.DeviceTemplate]
}

func NewDeviceTemplateService(dirName string, log *logrus.Logger) *DeviceTemplateService {
	return &DeviceTemplateService{
		storage: infrastructure.NewDeviceTemplateStorage(dirName, log),
	}
}

func (d *DeviceTemplateService) GetList(ctx context.Context, search, orderBy, orderDirection string, page, pageSize int) (*[]domain.DeviceTemplate, error) {
	return d.storage.GetList(ctx, search, orderBy, orderDirection, page, pageSize, nil)
}

func (d *DeviceTemplateService) GetByName(ctx context.Context, templateName string) (*domain.DeviceTemplate, error) {
	return d.storage.GetByName(ctx, templateName)
}
