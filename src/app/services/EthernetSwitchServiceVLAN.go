package services

import (
	"context"
	"github.com/google/uuid"
	"rol/app/errors"
	"rol/app/interfaces"
	"rol/app/mappers"
	"rol/app/validators"
	"rol/domain"
	"rol/dtos"
)

const (
	//ErrorGetSwitch repository failed to get ethernet switch
	ErrorGetSwitch = "repository failed to get ethernet switch"
	//ErrorSwitchExistence error when checking the existence of the switch
	ErrorSwitchExistence = "error when checking the existence of the switch"
	//ErrorSwitchNotFound switch is not found error
	ErrorSwitchNotFound = "switch is not found"
	//ErrorGetPortByID get port by id failed
	ErrorGetPortByID = "get port by id failed"
	//ErrorAddTaggedVLAN add tagged VLAN on port failed
	ErrorAddTaggedVLAN = "add tagged VLAN on port failed"
	//ErrorRemoveVLAN failed to remove VLAN from port
	ErrorRemoveVLAN = "failed to remove VLAN from port"
)

func (e *EthernetSwitchService) isVLANIdUnique(ctx context.Context, vlanID int, switchID uuid.UUID) (bool, error) {
	uniqueIDQueryBuilder := e.vlanRepo.NewQueryBuilder(ctx)
	uniqueIDQueryBuilder.Where("VlanID", "==", vlanID)
	uniqueIDQueryBuilder.Where("EthernetSwitchID", "==", switchID)
	switchVLANsList, err := e.vlanRepo.GetList(ctx, "", "asc", 1, 1, uniqueIDQueryBuilder)
	if err != nil {
		return false, errors.Internal.Wrap(err, "service failed to get list of switch ports")
	}
	if len(switchVLANsList) > 0 {
		return false, nil
	}
	return true, nil
}

//GetVLANByID Get ethernet switch VLAN by switch ID and VLAN ID
//
//Params
//	ctx - context is used only for logging|
//  switchID - Ethernet switch ID
//	id - VLAN id
//Return
//	*dtos.EthernetSwitchVLANDto - point to ethernet switch VLAN dto, if existed, otherwise nil
//	error - if an error occurs, otherwise nil
func (e *EthernetSwitchService) GetVLANByID(ctx context.Context, switchID, id uuid.UUID) (dtos.EthernetSwitchVLANDto, error) {
	dto := dtos.EthernetSwitchVLANDto{}
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, ErrorSwitchExistence)
	}
	if !switchExist {
		return dto, errors.NotFound.New(ErrorSwitchNotFound)
	}
	queryBuilder := e.vlanRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchID", "==", switchID)
	return GetByID[dtos.EthernetSwitchVLANDto, domain.EthernetSwitchVLAN](ctx, e.vlanRepo, id, queryBuilder)
}

//GetVLANs Get list of ethernet switch VLANs with filtering and pagination
//
//Params
//	ctx - context is used only for logging
//  switchID - uuid of the ethernet switch
//	search - string for search in ethernet switch port string fields
//	orderBy - order by ethernet switch port field name
//	orderDirection - ascending or descending order
//	page - page number
//	pageSize - page size
//Return
//	*dtos.PaginatedListDto[dtos.EthernetSwitchVLANDto] - pointer to paginated list of ethernet switch VLANs
//	error - if an error occurs, otherwise nil
func (e *EthernetSwitchService) GetVLANs(ctx context.Context, switchID uuid.UUID, search, orderBy, orderDirection string, page, pageSize int) (dtos.PaginatedItemsDto[dtos.EthernetSwitchVLANDto], error) {
	dto := dtos.PaginatedItemsDto[dtos.EthernetSwitchVLANDto]{}
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, ErrorSwitchExistence)
	}
	if !switchExist {
		return dto, errors.NotFound.New(ErrorSwitchNotFound)
	}
	queryBuilder := e.vlanRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchID", "==", switchID)
	if len(search) > 3 {
		AddSearchInAllFields(search, e.vlanRepo, queryBuilder)
	}
	return GetListExtended[dtos.EthernetSwitchVLANDto, domain.EthernetSwitchVLAN](ctx, e.vlanRepo, queryBuilder, orderBy, orderDirection, page, pageSize)
}

//CreateVLAN Create ethernet switch VLAN by EthernetSwitchVLANCreateDto
//Params
//	ctx - context is used only for logging
//  switchID - Ethernet switch ID
//	createDto - EthernetSwitchVLANCreateDto
//Return
//	uuid.UUID - created VLAN ID
//	error - if an error occurs, otherwise nil
func (e *EthernetSwitchService) CreateVLAN(ctx context.Context, switchID uuid.UUID, createDto dtos.EthernetSwitchVLANCreateDto) (dtos.EthernetSwitchVLANDto, error) {
	dto := dtos.EthernetSwitchVLANDto{}
	err := validators.ValidateEthernetSwitchVLANCreateDto(createDto)
	if err != nil {
		return dto, err //we already wrap error in validators
	}
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, ErrorSwitchExistence)
	}
	if !switchExist {
		return dto, errors.NotFound.New(ErrorSwitchNotFound)
	}
	uniqVLANId, err := e.isVLANIdUnique(ctx, createDto.VlanID, switchID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "VLAN ID uniqueness check error")
	}
	if !uniqVLANId {
		err = errors.Validation.New(errors.ValidationErrorMessage)
		return dto, errors.AddErrorContext(err, "VlanID", "vlan with this id already exist")
	}
	entity := new(domain.EthernetSwitchVLAN)
	err = mappers.MapDtoToEntity(createDto, entity)
	entity.EthernetSwitchID = switchID
	if err != nil {
		return dto, errors.Internal.Wrap(err, "failed to map ethernet switch port dto to entity")
	}
	newVLAN, err := e.vlanRepo.Insert(ctx, *entity)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "repository failed to insert VLAN")
	}
	ethernetSwitch, err := e.switchRepo.GetByID(ctx, switchID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, ErrorGetSwitch)
	}
	outDto := dtos.EthernetSwitchVLANDto{}
	err = mappers.MapEntityToDto(newVLAN, &outDto)
	switchManager := GetEthernetSwitchManager(ethernetSwitch)
	if switchManager == nil {
		return outDto, nil
	}
	err = switchManager.CreateVLAN(createDto.VlanID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "create VLAN on switch failed")
	}
	for _, taggedPortID := range createDto.TaggedPorts {
		switchPort, err := e.portRepo.GetByID(ctx, taggedPortID)
		if err != nil {
			return dto, errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		err = switchManager.AddTaggedVLANOnPort(switchPort.Name, createDto.VlanID)
		if err != nil {
			return dto, errors.Internal.Wrap(err, ErrorAddTaggedVLAN)
		}
	}
	for _, untaggedPortID := range createDto.UntaggedPorts {
		switchPort, err := e.portRepo.GetByID(ctx, untaggedPortID)
		if err != nil {
			return dto, errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		err = switchManager.AddUntaggedVLANOnPort(switchPort.Name, createDto.VlanID)
		if err != nil {
			return dto, errors.Internal.Wrap(err, ErrorAddTaggedVLAN)
		}
	}
	err = switchManager.SaveConfig()
	if err != nil {
		return dto, errors.Internal.Wrap(err, "save switch config failed")
	}
	return outDto, nil
}

//UpdateVLAN Update ethernet switch VLAN
//
//Params
//	ctx - context is used only for logging
//  switchID - Ethernet switch ID
//	id - VLAN id
//  updateDto - dtos.EthernetSwitchVLANUpdateDto DTO for updating entity
//Return
//	error - if an error occurs, otherwise nil
func (e *EthernetSwitchService) UpdateVLAN(ctx context.Context, switchID, id uuid.UUID, updateDto dtos.EthernetSwitchVLANUpdateDto) (dtos.EthernetSwitchVLANDto, error) {
	dto := dtos.EthernetSwitchVLANDto{}
	err := validators.ValidateEthernetSwitchVLANUpdateDto(updateDto)
	if err != nil {
		return dto, err //we already wrap error in validators
	}
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, ErrorSwitchExistence)
	}
	if !switchExist {
		return dto, errors.NotFound.New(ErrorSwitchNotFound)
	}

	VLAN, err := e.GetVLANByID(ctx, switchID, id)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "get VLAN by id failed")
	}

	ethernetSwitch, err := e.switchRepo.GetByIDExtended(ctx, switchID, nil)
	if err != nil {
		return dto, errors.Internal.Wrap(err, ErrorGetSwitch)
	}
	switchManager := GetEthernetSwitchManager(ethernetSwitch)
	//if switchManager != nil {
	//TODO: warning message log that switch manager is nil
	err = e.updateVLANsOnPort(ctx, VLAN, updateDto, switchManager)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "update VLANs on port failed")
	}
	//}

	queryBuilder := e.vlanRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchID", "==", switchID)
	return Update[dtos.EthernetSwitchVLANDto, dtos.EthernetSwitchVLANUpdateDto, domain.EthernetSwitchVLAN](ctx, e.vlanRepo, updateDto, VLAN.ID, queryBuilder)
}

//DeleteVLAN mark ethernet switch VLAN as deleted
//Params
//	ctx - context is used only for logging
//	switchID - ethernet switch id
//	id - VLAN ID
//Return
//	error - if an error occurs, otherwise nil
func (e *EthernetSwitchService) DeleteVLAN(ctx context.Context, switchID, id uuid.UUID) error {
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return errors.Internal.Wrap(err, ErrorSwitchExistence)
	}
	if !switchExist {
		return errors.NotFound.New(ErrorSwitchNotFound)
	}
	queryBuilder := e.vlanRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchID", "==", switchID)
	//entity, err := e.repository.GetByIDExtended(ctx, id, queryBuilder)
	//if err != nil {
	//	return errors.Internal.Wrap(err, "failed to get by id")
	//}
	//if entity == nil {
	//	return errors.NotFound.New("ethernet switch port is not exist")
	//}
	return e.vlanRepo.Delete(ctx, id)
}

func (e *EthernetSwitchService) updateVLANsOnPort(ctx context.Context, VLAN dtos.EthernetSwitchVLANDto, updateDto dtos.EthernetSwitchVLANUpdateDto, switchManager interfaces.IEthernetSwitchManager) error {
	diffTaggedToRemove := e.getDifference(VLAN.TaggedPorts, updateDto.TaggedPorts)
	if len(updateDto.TaggedPorts) == 0 {
		for _, tPort := range VLAN.TaggedPorts {
			diffTaggedToRemove = append(diffTaggedToRemove, tPort)
		}
	}
	for _, id := range diffTaggedToRemove {
		switchPort, err := e.portRepo.GetByID(ctx, id)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		if switchManager != nil {
			err = switchManager.RemoveVLANFromPort(switchPort.Name, VLAN.VlanID)
			if err != nil {
				return errors.Internal.Wrap(err, ErrorRemoveVLAN)
			}
		}
	}
	diffTaggedToAdd := e.getDifference(updateDto.TaggedPorts, VLAN.TaggedPorts)
	for _, id := range diffTaggedToAdd {
		switchPort, err := e.portRepo.GetByID(ctx, id)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		if switchManager != nil {
			err = switchManager.AddTaggedVLANOnPort(switchPort.Name, VLAN.VlanID)
			if err != nil {
				return errors.Internal.Wrap(err, "failed to add tagged VLAN on port")
			}
		}
	}
	diffUntaggedToRemove := e.getDifference(VLAN.UntaggedPorts, updateDto.UntaggedPorts)
	if len(updateDto.UntaggedPorts) == 0 {
		for _, uPort := range VLAN.UntaggedPorts {
			diffTaggedToRemove = append(diffTaggedToRemove, uPort)
		}
	}
	for _, id := range diffUntaggedToRemove {
		switchPort, err := e.portRepo.GetByID(ctx, id)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		if switchManager != nil {
			err = switchManager.RemoveVLANFromPort(switchPort.Name, VLAN.VlanID)
			if err != nil {
				return errors.Internal.Wrap(err, ErrorRemoveVLAN)
			}
		}
	}
	diffUntaggedToAdd := e.getDifference(VLAN.UntaggedPorts, updateDto.UntaggedPorts)
	for _, id := range diffUntaggedToAdd {
		switchPort, err := e.portRepo.GetByID(ctx, id)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		if switchManager != nil {
			err = switchManager.AddUntaggedVLANOnPort(switchPort.Name, VLAN.VlanID)
			if err != nil {
				return errors.Internal.Wrap(err, "failed to add untagged VLAN on port")
			}
		}
	}
	return nil
}

func (e *EthernetSwitchService) deleteAllVLANsBySwitchID(ctx context.Context, switchID uuid.UUID) error {
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return errors.Internal.Wrap(err, "error when checking the existence of the switch")
	}
	if !switchExist {
		return errors.NotFound.New("switch is not found")
	}
	queryBuilder := e.vlanRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchID", "==", switchID)
	vlansCount, err := e.vlanRepo.Count(ctx, queryBuilder)
	if err != nil {
		return errors.Internal.Wrap(err, "VLANs counting failed")
	}
	vlans, err := e.vlanRepo.GetList(ctx, "ID", "asc", 1, int(vlansCount), queryBuilder)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to get VLANs")
	}
	for _, vlan := range vlans {
		err = e.vlanRepo.Delete(ctx, vlan.ID)
		if err != nil {
			return errors.Internal.Wrap(err, "failed to remove vlan by id in repository")
		}
	}
	return nil
}

func (e *EthernetSwitchService) getDifference(a, b []uuid.UUID) []uuid.UUID {
	mb := make(map[uuid.UUID]struct{}, len(b))
	for _, x := range b {
		mb[x] = struct{}{}
	}
	var diff []uuid.UUID
	for _, x := range a {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}