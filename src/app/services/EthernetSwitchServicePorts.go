package services

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"rol/app/errors"
	"rol/app/mappers"
	"rol/app/utils"
	"rol/app/validators"
	"rol/domain"
	"rol/dtos"
)

func (e *EthernetSwitchService) portNameIsUniqueWithinTheSwitch(ctx context.Context, name string, switchID, id uuid.UUID) (bool, error) {
	uniqueNameQueryBuilder := e.portRepo.NewQueryBuilder(ctx)
	uniqueNameQueryBuilder.Where("Name", "==", name)
	uniqueNameQueryBuilder.Where("EthernetSwitchID", "==", switchID)
	if [16]byte{} != id {
		uniqueNameQueryBuilder.Where("ID", "!=", id)
	}
	ethSwitchPortsList, err := e.portRepo.GetList(ctx, "", "asc", 1, 1, uniqueNameQueryBuilder)
	if err != nil {
		return false, errors.Internal.Wrap(err, "service failed to get list of switch ports")
	}
	if len(ethSwitchPortsList) > 0 {
		return false, nil
	}
	return true, nil
}

//GetPortByID Get ethernet switch port by switch ID and port ID
//Params
//	ctx - context is used only for logging|
//  switchID - Ethernet switch ID
//	id - entity id
//Return
//	*dtos.EthernetSwitchPortDto - point to ethernet switch port dto, if existed, otherwise nil
//	error - if an error occurs, otherwise nil
func (e *EthernetSwitchService) GetPortByID(ctx context.Context, switchID, id uuid.UUID) (dtos.EthernetSwitchPortDto, error) {
	dto := dtos.EthernetSwitchPortDto{}
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "error when checking the existence of the switch")
	}
	if !switchExist {
		return dto, errors.NotFound.New("switch is not found")
	}
	queryBuilder := e.portRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchID", "==", switchID)
	return GetByID[dtos.EthernetSwitchPortDto](ctx, e.portRepo, id, queryBuilder)
}

//CreatePort Create ethernet switch port by EthernetSwitchPortCreateDto
//Params
//	ctx - context is used only for logging|
//  switchID - Ethernet switch ID
//	createDto - EthernetSwitchPortCreateDto
//Return
//	dtos.EthernetSwitchPortDto - created switch port
//	error - if an error occurs, otherwise nil
func (e *EthernetSwitchService) CreatePort(ctx context.Context, switchID uuid.UUID, createDto dtos.EthernetSwitchPortCreateDto) (dtos.EthernetSwitchPortDto, error) {
	dto := dtos.EthernetSwitchPortDto{}
	err := validators.ValidateEthernetSwitchPortCreateDto(createDto)
	if err != nil {
		return dto, err //we already wrap error in validators
	}
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "error when checking the existence of the switch")
	}
	if !switchExist {
		return dto, errors.NotFound.New("ethernet switch is not found")
	}
	uniqName, err := e.portNameIsUniqueWithinTheSwitch(ctx, createDto.Name, switchID, [16]byte{})
	if err != nil {
		return dto, errors.Internal.Wrap(err, "name uniqueness check error")
	}
	if !uniqName {
		err = errors.Validation.New(errors.ValidationErrorMessage)
		return dto, errors.AddErrorContext(err, "Name", "port with this name already exist")
	}
	entity := new(domain.EthernetSwitchPort)
	entity.EthernetSwitchID = switchID
	err = mappers.MapDtoToEntity(createDto, entity)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "error map dto to entity")
	}
	createdEntity, err := e.portRepo.Insert(ctx, *entity)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "create switch port in repository failed")
	}
	err = mappers.MapEntityToDto(createdEntity, &dto)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "failed to map entity to dto")
	}
	return dto, nil
}

//UpdatePort Update ethernet switch port
//Params
//	ctx - context is used only for logging|
//  switchID - Ethernet switch ID
//	id - entity id
//  updateDto - dtos.EthernetSwitchPortUpdateDto DTO for updating entity
//Return
//	dtos.EthernetSwitchPortDto - updated switch port
//	error - if an error occurs, otherwise nil
func (e *EthernetSwitchService) UpdatePort(ctx context.Context, switchID, id uuid.UUID, updateDto dtos.EthernetSwitchPortUpdateDto) (dtos.EthernetSwitchPortDto, error) {
	dto := dtos.EthernetSwitchPortDto{}
	err := validators.ValidateEthernetSwitchPortUpdateDto(updateDto)
	if err != nil {
		return dto, err // we already wrap error in validators
	}
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "error when checking the existence of the switch")
	}
	if !switchExist {
		return dto, errors.NotFound.New("switch is not found")
	}
	uniqName, err := e.portNameIsUniqueWithinTheSwitch(ctx, updateDto.Name, switchID, id)
	if err != nil {
		return dto, errors.Internal.Wrap(err, "name uniqueness check error")
	}
	if !uniqName {
		err = errors.Validation.New(errors.ValidationErrorMessage)
		return dto, errors.AddErrorContext(err, "Name", "port with this name already exist")
	}
	queryBuilder := e.portRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchId", "==", switchID)
	return Update[dtos.EthernetSwitchPortDto](ctx, e.portRepo, updateDto, id, queryBuilder)
}

//GetPorts Get list of ethernet switch ports with filtering and pagination
//Params
//	ctx - context is used only for logging
//  switchID - uuid of the ethernet switch
//	search - string for search in ethernet switch port string fields
//	orderBy - order by ethernet switch port field name
//	orderDirection - ascending or descending order
//	page - page number
//	pageSize - page size
//Return
//	*dtos.PaginatedItemsDto[dtos.EthernetSwitchPortDto] - pointer to paginated list of ethernet switches
//	error - if an error occurs, otherwise nil
func (e *EthernetSwitchService) GetPorts(ctx context.Context, switchID uuid.UUID, search, orderBy, orderDirection string,
	page, pageSize int) (dtos.PaginatedItemsDto[dtos.EthernetSwitchPortDto], error) {
	paginatedItemsDto := dtos.NewEmptyPaginatedItemsDto[dtos.EthernetSwitchPortDto]()
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return paginatedItemsDto, errors.Internal.Wrap(err, "error when checking the existence of the switch")
	}
	if !switchExist {
		return paginatedItemsDto, errors.NotFound.New("switch is not found")
	}
	queryBuilder := e.portRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchId", "==", switchID)
	if len(search) > 3 {
		AddSearchInAllFields(search, e.portRepo, queryBuilder)
	}
	return GetListExtended[dtos.EthernetSwitchPortDto](ctx, e.portRepo, queryBuilder, orderBy, orderDirection, page, pageSize)
}

//DeletePort mark ethernet switch port as deleted
//Params
//	ctx - context is used only for logging
//	switchID - ethernet switch id
//Return
//	error - if an error occurs, otherwise nil
func (e *EthernetSwitchService) DeletePort(ctx context.Context, switchID, id uuid.UUID) error {
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return errors.Internal.Wrap(err, "error when checking the existence of the switch")
	}
	if !switchExist {
		return errors.NotFound.New("switch is not found")
	}
	queryBuilder := e.portRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchID", "==", switchID)
	_, err = e.portRepo.GetByIDExtended(ctx, id, queryBuilder)
	if err != nil {
		return err
	}
	err = e.removePortFromVLANs(ctx, switchID, id)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to remove port from VLANs")
	}
	err = e.portRepo.Delete(ctx, id)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to delete port")
	}
	return nil
}

func (e *EthernetSwitchService) removePortFromVLANs(ctx context.Context, switchID, portID uuid.UUID) error {
	queryBuilder := e.vlanRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchID", "==", switchID)
	count, err := e.vlanRepo.Count(ctx, queryBuilder)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to count VLANs")
	}
	VLANs, err := e.GetVLANs(ctx, switchID, portID.String(), "", "", 1, int(count))
	if err != nil {
		return errors.Internal.Wrap(err, "failed to get vlan's list")
	}
	for _, vlan := range VLANs.Items {
		tPorts := utils.RemoveElementFromSlice[uuid.UUID](vlan.TaggedPorts, portID)
		uPorts := utils.RemoveElementFromSlice[uuid.UUID](vlan.UntaggedPorts, portID)
		if len(tPorts) == 0 {
			tPorts = nil
		}
		if len(uPorts) == 0 {
			uPorts = nil
		}
		updDto := dtos.EthernetSwitchVLANUpdateDto{EthernetSwitchVLANBaseDto: dtos.EthernetSwitchVLANBaseDto{
			UntaggedPorts: uPorts,
			TaggedPorts:   tPorts,
		}}
		_, err = Update[dtos.EthernetSwitchVLANDto](ctx, e.vlanRepo, updDto, vlan.ID, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *EthernetSwitchService) deleteAllPortsBySwitchID(ctx context.Context, switchID uuid.UUID) error {
	switchExist, err := e.switchIsExist(ctx, switchID)
	if err != nil {
		return errors.Internal.Wrap(err, "error when checking the existence of the switch")
	}
	if !switchExist {
		return errors.NotFound.New("switch is not found")
	}
	queryBuilder := e.portRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("EthernetSwitchID", "==", switchID)
	portsCount, err := e.portRepo.Count(ctx, queryBuilder)
	if err != nil {
		return errors.Internal.Wrap(err, "ports counting failed")
	}
	ports, err := e.portRepo.GetList(ctx, "ID", "asc", 1, int(portsCount), queryBuilder)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to get ports")
	}
	for _, port := range ports {
		err = e.portRepo.Delete(ctx, port.ID)
		if err != nil {
			return errors.Internal.Wrap(err, "failed to remove port by id in repository")
		}
	}
	return nil
}

func (e *EthernetSwitchService) portIsExist(ctx context.Context, switchID, portID uuid.UUID) (bool, error) {
	_, err := e.GetPortByID(ctx, switchID, portID)
	if err != nil {
		if !errors.As(err, errors.NotFound) {
			return false, errors.Internal.Wrap(err, "repository failed to get ethernet switch")
		}
		return false, nil
	}
	return true, nil
}

func (e *EthernetSwitchService) getNonexistentPorts(ctx context.Context, switchID uuid.UUID, portsToCheck []uuid.UUID) ([]uuid.UUID, error) {
	queryBuilder := e.portRepo.NewQueryBuilder(ctx)
	queryBuilder.Where("SwitchID", "==", switchID)
	for _, port := range portsToCheck {
		queryBuilder.Or("ID", "==", port.ID())
	}
	count, err := e.portRepo.Count(ctx, queryBuilder)
	if err != nil {
		return []uuid.UUID{}, errors.Internal.Wrap(err, "failed to count ports")
	}
	ports, err := e.portRepo.GetList(ctx, "", "", 1, int(count), queryBuilder)
	if err != nil {
		return []uuid.UUID{}, errors.Internal.Wrap(err, "failed to get ports list")
	}
	if len(ports) == len(portsToCheck) {
		return []uuid.UUID{}, nil
	}
	var existentPorts []uuid.UUID
	for _, port := range ports {
		existentPorts = append(existentPorts, port.ID)
	}
	nonexistentPorts, _ := utils.SliceDiffElements(portsToCheck, existentPorts)
	return nonexistentPorts, nil
}

func (e *EthernetSwitchService) dtoPortsExist(ctx context.Context, switchID uuid.UUID, dto dtos.EthernetSwitchVLANBaseDto) (bool, error) {
	nonexistentTaggedPorts, err := e.getNonexistentPorts(ctx, switchID, dto.TaggedPorts)
	if err != nil {
		return false, errors.Internal.Wrap(err, "failed to get nonexistent tagged ports")
	}
	if len(nonexistentTaggedPorts) > 0 {
		err = errors.Validation.New(errors.ValidationErrorMessage)
		for _, port := range nonexistentTaggedPorts {
			err = errors.AddErrorContext(err, "TaggedPorts", fmt.Sprintf("port %s doesn't exist", port.String()))
		}
		return false, err
	}

	nonexistentUntaggedPorts, err := e.getNonexistentPorts(ctx, switchID, dto.UntaggedPorts)
	if err != nil {
		return false, errors.Internal.Wrap(err, "failed to get nonexistent untagged ports")
	}
	if len(nonexistentUntaggedPorts) > 0 {
		err = errors.Validation.New(errors.ValidationErrorMessage)
		for _, port := range nonexistentUntaggedPorts {
			err = errors.AddErrorContext(err, "UntaggedPorts", fmt.Sprintf("port %s doesn't exist", port.String()))
		}
		return false, err
	}
	return true, nil
}
