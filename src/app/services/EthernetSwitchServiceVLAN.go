package services

import (
	"context"
	"github.com/google/uuid"
	"rol/app/errors"
	"rol/app/mappers"
	"rol/app/utils"
	"rol/app/validators"
	"rol/domain"
	"rol/dtos"
)

const (
	//ErrorGetSwitch repository failed to get ethernet switch
	ErrorGetSwitch = "repository failed to get ethernet switch"
	//ErrorSwitchExistence error when checking the existence of the switch
	ErrorSwitchExistence = "error when checking the existence of the switch"
	//ErrorPortExistence error when checking the existence of the switch port
	ErrorPortExistence = "error when checking the existence of the switch port"
	//ErrorVlanOnNonexistentPort error when checking the existence of the switch port
	ErrorVlanOnNonexistentPort = "can't create a vlan on a port that doesn't exist"
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

func (e *EthernetSwitchService) existenceOfRelatedEntitiesCheck(ctx context.Context, switchID uuid.UUID, dto dtos.EthernetSwitchVLANBaseDto) error {
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return errors.Internal.Wrap(err, ErrorSwitchExistence)
	}
	if !switchExist {
		return errors.NotFound.New(ErrorSwitchNotFound)
	}
	portsExist, err := e.dtoPortsExist(ctx, switchID, dto)
	if err != nil {
		return err //we already wrap error
	}
	if !portsExist {
		return err //we already wrap error
	}
	return nil
}

func (e *EthernetSwitchService) createVlansOnSwitch(ctx context.Context, switchID uuid.UUID, dto dtos.EthernetSwitchVLANCreateDto) error {
	ethernetSwitch, err := e.switchRepo.GetByID(ctx, switchID)
	if err != nil {
		return errors.Internal.Wrap(err, ErrorGetSwitch)
	}
	switchManager := GetEthernetSwitchManager(ethernetSwitch)
	if switchManager == nil {
		return nil
	}
	err = switchManager.CreateVLAN(dto.VlanID)
	if err != nil {
		return errors.Internal.Wrap(err, "create VLAN on switch failed")
	}
	for _, taggedPortID := range dto.TaggedPorts {
		switchPort, err := e.portRepo.GetByID(ctx, taggedPortID)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		err = switchManager.AddTaggedVLANOnPort(switchPort.Name, dto.VlanID)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorAddTaggedVLAN)
		}
	}
	for _, untaggedPortID := range dto.UntaggedPorts {
		switchPort, err := e.portRepo.GetByID(ctx, untaggedPortID)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		err = switchManager.AddUntaggedVLANOnPort(switchPort.Name, dto.VlanID)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorAddTaggedVLAN)
		}
	}
	err = switchManager.SaveConfig()
	if err != nil {
		return errors.Internal.Wrap(err, "save switch config failed")
	}
	return nil
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
func (e *EthernetSwitchService) GetVLANByID(ctx context.Context, switchID uuid.UUID, vlanID int) (dtos.EthernetSwitchVLANDto, error) {
	dto := dtos.EthernetSwitchVLANDto{}
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, ErrorSwitchExistence)
	}
	if !switchExist {
		return dto, errors.NotFound.New(ErrorSwitchNotFound)
	}
	queryBuilder := e.vlanRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchID", "==", switchID).Where("VlanID", "==", vlanID)
	vlanSlice, err := GetListExtended[dtos.EthernetSwitchVLANDto](ctx, e.vlanRepo, queryBuilder, "", "", 1, 1)
	if len(vlanSlice.Items) == 0 {
		return dto, errors.NotFound.New("vlan not found")
	}
	if err != nil {
		return dto, errors.Internal.Wrap(err, "failed to get list")
	}
	return vlanSlice.Items[0], nil
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
	return GetListExtended[dtos.EthernetSwitchVLANDto](ctx, e.vlanRepo, queryBuilder, orderBy, orderDirection, page, pageSize)
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
	err = e.existenceOfRelatedEntitiesCheck(ctx, switchID, createDto.EthernetSwitchVLANBaseDto)
	if err != nil {
		return dto, err
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
	err = e.createVlansOnSwitch(ctx, switchID, createDto)
	if err != nil {
		return dto, err
	}
	outDto := dtos.EthernetSwitchVLANDto{}
	err = mappers.MapEntityToDto(newVLAN, &outDto)
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
func (e *EthernetSwitchService) UpdateVLAN(ctx context.Context, switchID uuid.UUID, vlanID int, updateDto dtos.EthernetSwitchVLANUpdateDto) (dtos.EthernetSwitchVLANDto, error) {
	dto := dtos.EthernetSwitchVLANDto{}
	err := validators.ValidateEthernetSwitchVLANUpdateDto(updateDto)
	if err != nil {
		return dto, err //we already wrap error in validators
	}
	err = e.existenceOfRelatedEntitiesCheck(ctx, switchID, updateDto.EthernetSwitchVLANBaseDto)
	if err != nil {
		return dto, err
	}
	VLAN, err := e.GetVLANByID(ctx, switchID, vlanID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "get VLAN by id failed")
	}
	err = e.updateVLANsOnPort(ctx, VLAN, updateDto, switchID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "update VLANs on port failed")
	}
	queryBuilder := e.vlanRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchID", "==", switchID)
	return Update[dtos.EthernetSwitchVLANDto](ctx, e.vlanRepo, updateDto, VLAN.ID, queryBuilder)
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
	return e.vlanRepo.Delete(ctx, id)
}

func (e *EthernetSwitchService) updateVLANsOnPort(ctx context.Context, VLAN dtos.EthernetSwitchVLANDto, updateDto dtos.EthernetSwitchVLANUpdateDto, switchID uuid.UUID) error {
	ethernetSwitch, err := e.switchRepo.GetByIDExtended(ctx, switchID, nil)
	if err != nil {
		return errors.Internal.Wrap(err, ErrorGetSwitch)
	}
	switchManager := GetEthernetSwitchManager(ethernetSwitch)
	if switchManager == nil {
		return nil
	}
	diffTaggedToRemove, diffTaggedToAdd := utils.SliceDiffElements[uuid.UUID](VLAN.TaggedPorts, updateDto.TaggedPorts)
	for _, id := range diffTaggedToRemove {
		switchPort, err := e.portRepo.GetByID(ctx, id)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		err = switchManager.RemoveVLANFromPort(switchPort.Name, VLAN.VlanID)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorRemoveVLAN)
		}
	}
	for _, id := range diffTaggedToAdd {
		switchPort, err := e.portRepo.GetByID(ctx, id)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		err = switchManager.AddTaggedVLANOnPort(switchPort.Name, VLAN.VlanID)
		if err != nil {
			return errors.Internal.Wrap(err, "failed to add tagged VLAN on port")
		}
	}
	diffUntaggedToRemove, diffUntaggedToAdd := utils.SliceDiffElements[uuid.UUID](VLAN.UntaggedPorts, updateDto.UntaggedPorts)
	for _, id := range diffUntaggedToRemove {
		switchPort, err := e.portRepo.GetByID(ctx, id)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		err = switchManager.RemoveVLANFromPort(switchPort.Name, VLAN.VlanID)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorRemoveVLAN)
		}
	}
	for _, id := range diffUntaggedToAdd {
		switchPort, err := e.portRepo.GetByID(ctx, id)
		if err != nil {
			return errors.Internal.Wrap(err, ErrorGetPortByID)
		}
		err = switchManager.AddUntaggedVLANOnPort(switchPort.Name, VLAN.VlanID)
		if err != nil {
			return errors.Internal.Wrap(err, "failed to add untagged VLAN on port")
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
