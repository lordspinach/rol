package services

import (
	"rol/app/interfaces"
	"rol/domain"
	"rol/infrastructure"
)

//GetEthernetSwitchManager get instance of ethernet switch manager depending on the model of the switch
func GetEthernetSwitchManager(ethernetSwitch domain.EthernetSwitch) interfaces.IEthernetSwitchManager {
	switch ethernetSwitch.SwitchModel {
	case "tl-sg2210mp":
		return infrastructure.NewTPLinkEthernetSwitchManager(ethernetSwitch.Address+":23", ethernetSwitch.Username, ethernetSwitch.Password)
	}
	return nil
}
