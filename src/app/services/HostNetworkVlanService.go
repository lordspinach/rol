package services

import (
	"net"
	"rol/app/errors"
	"rol/app/interfaces"
	"rol/app/mappers"
	"rol/dtos"
	"strings"
)

//HostNetworkVlanService is a struct for host network vlan service
type HostNetworkVlanService struct {
	manager            interfaces.IHostNetworkManager
	haveUnsavedChanges bool
}

//NewHostNetworkVlanService is a constructor for HostNetworkVlanService
//
//Params:
//	manager - host network manager
//Return:
//	HostNetworkVlanService - instance of network vlan service
func NewHostNetworkVlanService(manager interfaces.IHostNetworkManager) *HostNetworkVlanService {
	return &HostNetworkVlanService{
		manager:            manager,
		haveUnsavedChanges: false,
	}
}

//GetList gets list of host vlans
//
//Return:
//	[]dtos.HostNetworkVlanDto - slice of vlan dtos
//	error - if an error occurs, otherwise nil
func (h HostNetworkVlanService) GetList() ([]dtos.HostNetworkVlanDto, error) {
	out := []dtos.HostNetworkVlanDto{}
	links, err := h.manager.GetList()
	if err != nil {
		return nil, errors.Internal.Wrap(err, "error getting link list")
	}
	for _, link := range links {
		if link.GetType() == "vlan" && strings.Contains(link.GetName(), "rol.") {
			var dto dtos.HostNetworkVlanDto
			err = mappers.MapEntityToDto(link, &dto)
			if err != nil {
				return nil, errors.Internal.Wrap(err, "error mapping vlan")
			}
			out = append(out, dto)
		}
	}
	return out, nil
}

//GetByName gets vlan by name
//
//Params:
//	vlanName - name of the vlan
//Return:
//	dtos.HostNetworkVlanDto - vlan dto
//	error - if an error occurs, otherwise nil
func (h HostNetworkVlanService) GetByName(vlanName string) (dtos.HostNetworkVlanDto, error) {
	link, err := h.manager.GetByName(vlanName)
	if err != nil {
		return dtos.HostNetworkVlanDto{}, errors.Internal.Wrap(err, "error getting vlan by name")
	}
	if link == nil || (link.GetType() != "vlan" && strings.Contains(link.GetName(), "rol.")) {
		return dtos.HostNetworkVlanDto{}, errors.NotFound.New("vlan not found")
	}
	var dto dtos.HostNetworkVlanDto
	err = mappers.MapEntityToDto(link, &dto)
	if err != nil {
		return dtos.HostNetworkVlanDto{}, errors.Internal.Wrap(err, "error mapping vlan")
	}
	return dto, nil
}

//Create new vlan on host
//
//Params:
//	vlan - vlan create dto
//Return:
//	string - new vlan name that will be rol.{master}.{vlanID}
//	error - if an error occurs, otherwise nil
func (h *HostNetworkVlanService) Create(vlan dtos.HostNetworkVlanCreateDto) (string, error) {
	vlanName, err := h.manager.CreateVlan(vlan.Master, vlan.VlanID)
	if err != nil {
		return "", errors.Internal.Wrap(err, "error creating vlan")
	}
	h.haveUnsavedChanges = true
	return vlanName, nil
}

//SetAddr sets new ip address for vlan
//
//Params:
//	vlanName - vlan name
//	addr - ip address with mask net.IPNet
//Return:
//	error - if an error occurs, otherwise nil
func (h *HostNetworkVlanService) SetAddr(vlanName string, addr net.IPNet) error {
	link, err := h.manager.GetByName(vlanName)
	if err != nil {
		return errors.Internal.Wrap(err, "error getting vlan by name")
	}
	if link == nil || (link.GetType() != "vlan" && strings.Contains(link.GetName(), "rol.")) {
		return errors.NotFound.New("vlan not found")
	}
	err = h.manager.SetAddr(vlanName, addr)
	if err != nil {
		return errors.Internal.Wrap(err, "set address failed")
	}
	h.haveUnsavedChanges = true
	return nil
}

//Delete deletes vlan on host by its name
//
//Params:
//	vlanName - vlan name
//Return
//	error - if an error occurs, otherwise nil
func (h *HostNetworkVlanService) Delete(vlanName string) error {
	link, err := h.manager.GetByName(vlanName)
	if err != nil {
		return errors.Internal.Wrap(err, "error getting vlan by name")
	}
	if link == nil || (link.GetType() != "vlan" && strings.Contains(link.GetName(), "rol.")) {
		return errors.NotFound.New("vlan not found")
	}
	err = h.manager.DeleteLinkByName(vlanName)
	if err != nil {
		return errors.Internal.Wrap(err, "delete vlan failed")
	}
	h.haveUnsavedChanges = true
	return nil
}

//SaveChanges saves all changes on host to configuration file so as not to lose changes during system reboot
//
//Return
//	error - if an error occurs, otherwise nil
func (h *HostNetworkVlanService) SaveChanges() error {
	if h.haveUnsavedChanges {
		err := h.manager.SaveConfiguration()
		if err != nil {
			return errors.Internal.Wrap(err, "failed to save configuration")
		}
		h.haveUnsavedChanges = false
	}
	return nil
}

//ResetChanges rolls back all changes on the host to the previous state of the config
//
//Return
//	error - if an error occurs, otherwise nil
func (h *HostNetworkVlanService) ResetChanges() error {
	if h.haveUnsavedChanges {
		err := h.manager.LoadConfiguration()
		if err != nil {
			return errors.Internal.Wrap(err, "failed to restore configuration")
		}
		h.haveUnsavedChanges = false
	}
	return nil
}
