package interfaces

import "rol/domain"

//IEthernetSwitchManagerGetter is the interface is needed to get ethernet switch manager
type IEthernetSwitchManagerGetter interface {
	//Get ethernet switch manager
	Get(ethernetSwitch domain.EthernetSwitch) IEthernetSwitchManager
}
