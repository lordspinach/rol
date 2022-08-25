package services

import (
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"rol/app/interfaces"
	"rol/domain"
	"rol/dtos"
)

//AppLogService service structure for AppLog entity
type AppLogService struct {
	repository interfaces.IGenericRepository[domain.AppLog]
}

//NewAppLogService preparing domain.AppLog repository for passing it in DI
//Params
//	repo - generic repository with app log instantiated
//	log - logrus logger
//Return
//	New app log service
func NewAppLogService(repo interfaces.IGenericRepository[domain.AppLog], log *logrus.Logger) (*AppLogService, error) {
	return &AppLogService{
		repository: repo,
	}, nil
}

//GetList Get list of logs dto elements with search in all fields and pagination
//
//Params:
//	ctx - context is used only for logging
//	search - string for search in entity string fields
//	orderBy - order by entity field name
//	orderDirection - ascending or descending order
//	page - page number
//	pageSize - page size
//Return:
//	dtos.PaginatedItemsDto[dtos.AppLogDto] - pointer to paginated list
//	error - if an error occurs, otherwise nil
func (a *AppLogService) GetList(ctx context.Context, search string, orderBy string, orderDirection string, page int, pageSize int) (dtos.PaginatedItemsDto[dtos.AppLogDto], error) {
	return GetList[dtos.AppLogDto, domain.AppLog](ctx, a.repository, search, orderBy, orderDirection, page, pageSize)
}

//GetByID Get log dto by ID
//
//Params:
//	ctx - context
//	id - entity id
//Return:
//	dtos.AppLogDto - pointer to dto
//	error - if an error occurs, otherwise nil
func (a *AppLogService) GetByID(ctx context.Context, id uuid.UUID) (dtos.AppLogDto, error) {
	return GetByID[dtos.AppLogDto, domain.AppLog](ctx, a.repository, id, nil)
}
