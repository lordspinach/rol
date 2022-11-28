package infrastructure

import (
	"github.com/sirupsen/logrus"
	"rol/app/errors"
	"rol/app/interfaces"
	"rol/domain"
)

//SliceDeviceTemplateRepository repository for DeviceTemplate entity
type SliceDeviceTemplateRepository struct {
	*SliceGenericRepository[string, domain.DeviceTemplate]
}

//NewSliceDeviceTemplateRepository constructor for domain.DeviceTemplate slice generic repository
//Params
//	diParams - global dependency injection parameters
//	log - logrus logger
//Return
//	interfaces.IGenericRepository[string, domain.DeviceTemplate] - new device template repository
func NewSliceDeviceTemplateRepository(diParams domain.GlobalDIParameters, log *logrus.Logger) (interfaces.IGenericRepository[string, domain.DeviceTemplate], error) {
	fileContext, err := NewYamlManyFilesContext[string, domain.DeviceTemplate](diParams, "/templates/devices/")
	if err != nil {
		return nil, errors.Internal.Wrap(err, "failed to create new yaml many files context")
	}
	repo := NewSliceGenericRepository[string, domain.DeviceTemplate](fileContext, log)
	return &SliceDeviceTemplateRepository{
		repo,
	}, nil
}
