package services

import (
	"context"
	"github.com/sirupsen/logrus"
	"rol/app/interfaces"
	"rol/app/mappers"
	"rol/app/utils"
	"rol/domain"
	"rol/dtos"
	"rol/infrastructure"
	"strings"
)

type DeviceTemplateService struct {
	storage interfaces.IGenericTemplateStorage[domain.DeviceTemplate]
}

func NewDeviceTemplateService(dirName string, log *logrus.Logger) *DeviceTemplateService {
	return &DeviceTemplateService{
		storage: infrastructure.NewDeviceTemplateStorage(dirName, log),
	}
}

func (d *DeviceTemplateService) GetList(ctx context.Context, search, orderBy, orderDirection string, page, pageSize int) (*dtos.PaginatedListDto[dtos.DeviceTemplateDto], error) {
	pageFinal := page
	pageSizeFinal := pageSize
	if page < 1 {
		pageFinal = 1
	}
	if pageSize < 1 {
		pageSizeFinal = 10
	}
	queryBuilder := d.storage.NewQueryBuilder(ctx)
	if len(search) > 3 {
		queryBuilder = d.addSearchInAllFields(search, queryBuilder)
	}
	templatesArr, err := d.storage.GetList(ctx, orderBy, orderDirection, pageFinal, pageSizeFinal, queryBuilder)
	if err != nil {
		return nil, err
	}
	count, err := d.storage.Count(ctx, queryBuilder)
	if err != nil {
		return nil, err
	}
	dtosArr := make([]dtos.DeviceTemplateDto, len(*templatesArr))
	for i := range *templatesArr {
		err := mappers.MapEntityToDto((*templatesArr)[i], &dtosArr[i])
		if err != nil {
			return nil, err
		}
	}
	paginatedDto := new(dtos.PaginatedListDto[dtos.DeviceTemplateDto])
	paginatedDto.Page = pageFinal
	paginatedDto.Size = pageSizeFinal
	paginatedDto.Total = count
	paginatedDto.Items = &dtosArr
	return paginatedDto, nil
}

func (d *DeviceTemplateService) GetByName(ctx context.Context, templateName string) (*dtos.DeviceTemplateDto, error) {
	template, err := d.storage.GetByName(ctx, templateName)
	if err != nil {
		return nil, err
	}
	dto := new(dtos.DeviceTemplateDto)
	mappers.MapDeviceTemplateToDto(*template, dto)
	return dto, nil
}

func (d *DeviceTemplateService) addSearchInAllFields(search string, queryBuilder interfaces.IQueryBuilder) interfaces.IQueryBuilder {
	template := new(domain.DeviceTemplate)
	stringFieldNames := &[]string{}
	utils.GetStringFieldsNames(template, stringFieldNames)
	queryGroup := d.storage.NewQueryBuilder(nil)
	for i := 0; i < len(*stringFieldNames); i++ {
		fieldName := (*stringFieldNames)[i]
		containPass := strings.Contains(strings.ToLower(fieldName), "pass")
		containKey := strings.Contains(strings.ToLower(fieldName), "key")
		if containPass || containKey {
			continue
		}
		queryGroup.Or(fieldName, "LIKE", search)
	}
	return queryBuilder.WhereQuery(queryGroup)
}
