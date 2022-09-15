package infrastructure

import (
	"github.com/google/uuid"
	"rol/app/interfaces"
	"rol/domain"
)

//EthernetSwitchManagerGetter struct for switch manager getter
type EthernetSwitchManagerGetter struct {
	managers map[uuid.UUID]interfaces.IEthernetSwitchManager
}

//NewEthernetSwitchManagerGetter constructor for EthernetSwitchManagerGetter
func NewEthernetSwitchManagerGetter() interfaces.IEthernetSwitchManagerGetter {
	return &EthernetSwitchManagerGetter{
		managers: make(map[uuid.UUID]interfaces.IEthernetSwitchManager),
	}
}

//Get ethernet switch manager
//
//Params:
//	ethernetSwitch - switch entity
//Return:
//	interfaces.IEthernetSwitchManager - switch manager interface
func (e *EthernetSwitchManagerGetter) Get(ethernetSwitch domain.EthernetSwitch) interfaces.IEthernetSwitchManager {
	if e.managers[ethernetSwitch.ID] == nil {
		switch ethernetSwitch.SwitchModel {
		case "tl-sg2210mp":
			e.managers[ethernetSwitch.ID] = NewTPLinkEthernetSwitchManager(ethernetSwitch.Address+":23", ethernetSwitch.Username, ethernetSwitch.Password)
			return e.managers[ethernetSwitch.ID]
		}
		return nil
	}
	return e.managers[ethernetSwitch.ID]
}
