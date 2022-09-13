package services

import (
	"context"
	"github.com/Azure/go-asynctask"
	"github.com/google/uuid"
	"rol/app/errors"
	"rol/app/interfaces"
	"rol/app/mappers"
	"rol/domain"
	"rol/dtos"
	"rol/infrastructure"
	"time"
)

//TFTPServerService service structure for TFTP server
type TFTPServerService struct {
	configsRepo interfaces.IGenericRepository[domain.TFTPConfig]
	pathsRepo   interfaces.IGenericRepository[domain.TFTPPathRatio]
	manager     *infrastructure.TFTPServerManager
}

//NewTFTPServerService constructor for TFTP server
//Params
//	configsRepo - generic repository with domain.TFTPConfig entity
//	pathsRepo - generic repository with domain.TFTPPathRatio entity
//	manager - tftp server manager
//Return
//	New TFTP server service
func NewTFTPServerService(configsRepo interfaces.IGenericRepository[domain.TFTPConfig],
	pathsRepo interfaces.IGenericRepository[domain.TFTPPathRatio],
	manager *infrastructure.TFTPServerManager) *TFTPServerService {
	return &TFTPServerService{
		configsRepo: configsRepo,
		pathsRepo:   pathsRepo,
		manager:     manager,
	}
}

func (t *TFTPServerService) getServerPaths(ctx context.Context, id uuid.UUID) ([]domain.TFTPPathRatio, error) {
	paths := []domain.TFTPPathRatio{}
	queryBuilder := t.pathsRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("TFTPServerID", "==", id)
	pathsCount, err := t.pathsRepo.Count(ctx, queryBuilder)
	if err != nil {
		return paths, errors.Internal.Wrap(err, "failed to count tftp paths")
	}
	return t.pathsRepo.GetList(ctx, "", "", 1, int(pathsCount), queryBuilder)
}

func (t *TFTPServerService) startServerAndCheckStatus(ctx context.Context, config domain.TFTPConfig) error {
	t.manager.CreateTFTPServer(config)
	paths, err := t.getServerPaths(ctx, config.ID)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to get tftp server paths")
	}
	t.manager.UpdatePaths(config.ID, paths)
	if config.Enabled {
		t.manager.StartTFTPServer(config.ID)
		time.Sleep(time.Second / 2)
		if !t.manager.ServerIsRunning(config.ID) {
			config.Enabled = false
			_, err := t.configsRepo.Update(ctx, config)
			if err != nil {
				return errors.Internal.Wrap(err, "failed to update tftp server config")
			}
		}
	}
	return nil
}

func (t *TFTPServerService) asyncStartServerAndCheckStatus(s *TFTPServerService, config domain.TFTPConfig) asynctask.AsyncFunc[int] {
	return func(ctx context.Context) (*int, error) {
		out := 0
		s.manager.CreateTFTPServer(config)
		paths, err := t.getServerPaths(ctx, config.ID)
		if err != nil {
			return &out, errors.Internal.Wrap(err, "failed to get tftp server paths")
		}
		s.manager.UpdatePaths(config.ID, paths)
		if config.Enabled {
			s.manager.StartTFTPServer(config.ID)
			time.Sleep(time.Second / 2)
			if !s.manager.ServerIsRunning(config.ID) {
				config.Enabled = false
				_, err := s.configsRepo.Update(ctx, config)
				if err != nil {
					return &out, errors.Internal.Wrap(err, "failed to update tftp server config")
				}
			}
		}
		return &out, nil
	}
}

//TFTPServerServiceInitialize initialize TFTP service through DI
func TFTPServerServiceInitialize(s *TFTPServerService) error {
	ctx := context.Background()
	serversCount, err := s.configsRepo.Count(ctx, nil)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to count tftp server configs")
	}
	serverConfigs, err := s.configsRepo.GetList(ctx, "", "", 1, int(serversCount), nil)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to get tftp server configs list")
	}
	for _, config := range serverConfigs {
		err = s.startServerAndCheckStatus(ctx, config)
		if err != nil {
			return errors.Internal.Wrap(err, "failed to start server and check status")
		}
	}
	return nil
}

//GetByID get TFTP server by ID
//Params
//	ctx - context is used only for logging
//	id - TFTP server id
//Return
//	dtos.TFTPConfigDto - TFTP server dto
//	error - if an error occurs, otherwise nil
func (t *TFTPServerService) GetByID(ctx context.Context, id uuid.UUID) (dtos.TFTPConfigDto, error) {
	dto, err := GetByID[dtos.TFTPConfigDto](ctx, t.configsRepo, id, nil)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "failed to get tftp server config")
	}
	dto.Enabled = t.manager.ServerIsRunning(id)
	return dto, err
}

//GetList get list of TFTP servers with filtering and pagination
//Params
//	ctx - context is used only for logging
//	search - string for search in entity string fields
//	orderBy - order by entity field name
//	orderDirection - ascending or descending order
//	page - page number
//	pageSize - page size
//Return
//	dtos.PaginatedItemsDto[dtos.TFTPConfigDto] - paginated list of TFTP servers
//	error - if an error occurs, otherwise nil
func (t *TFTPServerService) GetList(ctx context.Context, search, orderBy, orderDirection string, page, pageSize int) (dtos.PaginatedItemsDto[dtos.TFTPConfigDto], error) {
	servers, err := GetList[dtos.TFTPConfigDto](ctx, t.configsRepo, search, orderBy, orderDirection, page, pageSize)
	if err != nil {
		return servers, errors.Internal.Wrap(err, "failed to get tftp server configs list")
	}
	for _, server := range servers.Items {
		server.Enabled = t.manager.ServerIsRunning(server.ID)
	}
	return servers, nil
}

//Create add new TFTP server
//Params
//	ctx - context
//	createDto - TFTP server create dto
//Return
//	dtos.TFTPConfigDto - created TFTP server
//	error - if an error occurs, otherwise nil
func (t *TFTPServerService) Create(ctx context.Context, createDto dtos.TFTPConfigCreateDto) (dtos.TFTPConfigDto, error) {
	dto := dtos.TFTPConfigDto{}
	entity := new(domain.TFTPConfig)
	err := mappers.MapDtoToEntity(createDto, entity)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "error map entity to dto")
	}
	newServer, err := t.configsRepo.Insert(ctx, *entity)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "create entity error")
	}
	t.manager.CreateTFTPServer(newServer)
	if createDto.Enabled {
		t.manager.StartTFTPServer(newServer.ID)
	}
	err = mappers.MapEntityToDto(newServer, &dto)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "error map dto to entity")
	}

	return dto, nil
}

//Update save the changes to the existing TFTP server
//Params
//	ctx - context is used only for logging
//	updateDto - TFTP server update dto
//	id - TFTP server id
//Return
//	dtos.TFTPConfigDto - updated TFTP server
//	error - if an error occurs, otherwise nil
func (t *TFTPServerService) Update(ctx context.Context, updateDto dtos.TFTPConfigUpdateDto, id uuid.UUID) (dtos.TFTPConfigDto, error) {
	dto := dtos.TFTPConfigDto{}
	updatedConfig, err := Update[dtos.TFTPConfigDto](ctx, t.configsRepo, updateDto, id, nil)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "failed to get update tftp server config")
	}
	if updateDto.Enabled {
		t.manager.StartTFTPServer(id)
	} else {
		t.manager.StopTFTPServer(id)
	}
	return updatedConfig, nil
}

//Delete mark TFTP server as deleted
//Params
//	ctx - context is used only for logging
//	id - TFTP server id
//Return
//	error - if an error occurs, otherwise nil
func (t *TFTPServerService) Delete(ctx context.Context, id uuid.UUID) error {
	err := t.configsRepo.Delete(ctx, id)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to delete tftp server config")
	}
	paths, err := t.getServerPaths(ctx, id)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to get tftp server paths")
	}
	for _, path := range paths {
		err = t.pathsRepo.Delete(ctx, path.ID)
		if err != nil {
			return errors.Internal.Wrap(err, "failed to delete tftp server paths")
		}
	}
	return nil
}

//GetPaths get list of TFTP servers paths with filtering and pagination
//Params
//	ctx - context is used only for logging
//	configID - tftp config id
//	orderBy - order by entity field name
//	orderDirection - ascending or descending order
//	page - page number
//	pageSize - page size
//Return
//	dtos.PaginatedItemsDto[dtos.TFTPPathDto] - paginated list of TFTP servers paths
//	error - if an error occurs, otherwise nil
func (t *TFTPServerService) GetPaths(ctx context.Context, configID uuid.UUID, orderBy, orderDirection string, page, pageSize int) (dtos.PaginatedItemsDto[dtos.TFTPPathDto], error) {
	queryBuilder := t.pathsRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("TFTPServerID", "==", configID)
	return GetListExtended[dtos.TFTPPathDto](ctx, t.pathsRepo, queryBuilder, orderBy, orderDirection, page, pageSize)
}

//CreatePath add new TFTP server path
//Params
//	ctx - context
//	configID - tftp config id
//	createDto - TFTP server create dto
//Return
//	dtos.TFTPPathDto - created TFTP server
//	error - if an error occurs, otherwise nil
func (t *TFTPServerService) CreatePath(ctx context.Context, configID uuid.UUID, createDto dtos.TFTPPathCreateDto) (dtos.TFTPPathDto, error) {
	entity := new(domain.TFTPPathRatio)
	outDto := new(dtos.TFTPPathDto)
	err := mappers.MapDtoToEntity(createDto, entity)
	if err != nil {
		return *outDto, errors.Internal.Wrap(err, "error map entity to dto")
	}
	entity.TFTPServerID = configID
	newEntity, err := t.pathsRepo.Insert(ctx, *entity)
	if err != nil {
		return *outDto, errors.Internal.Wrap(err, "create entity error")
	}
	err = mappers.MapEntityToDto(newEntity, outDto)
	if err != nil {
		return *outDto, errors.Internal.Wrap(err, "error map dto to entity")
	}
	return *outDto, nil
}

//DeletePath mark TFTP server path as deleted
//Params
//	ctx - context is used only for logging
//	configID - tftp config id
//	id - TFTP server path id
//Return
//	error - if an error occurs, otherwise nil
func (t *TFTPServerService) DeletePath(ctx context.Context, configID uuid.UUID, id uuid.UUID) error {
	config, err := t.GetByID(ctx, configID)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to get tftp server config")
	}
	if config.ID != configID {
		return errors.NotFound.New("server with given id not found")
	}
	return t.pathsRepo.Delete(ctx, id)
}
