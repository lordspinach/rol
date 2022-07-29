package interfaces

import "net"

//IHostNetworkManager is an interface for network manager
type IHostNetworkManager interface {
	//GetList gets list of host network interfaces
	//
	//Return:
	//	[]interfaces.IHostNetworkLink - list of interfaces
	//	error - if an error occurs, otherwise nil
	GetList() ([]IHostNetworkLink, error)
	//GetByName gets host network interface by its name
	//
	//Params:
	//	name - interface name
	//Return:
	//	interfaces.IHostNetworkLink - interfaces
	//	error - if an error occurs, otherwise nil
	GetByName(name string) (IHostNetworkLink, error)
	//CreateVlan creates vlan on host
	//
	//Params:
	//	master - name of the master network interface
	//	vlanID - ID to be set for vlan
	//Return:
	//	string - new vlan name that will be {master}.{vlanID}
	//	error - if an error occurs, otherwise nil
	CreateVlan(master string, vlanID int) (string, error)
	//DeleteLinkByName deletes interface on host by its name
	//
	//Params:
	//	name - interface name
	//Return
	//	error - if an error occurs, otherwise nil
	DeleteLinkByName(name string) error
	//SetAddr sets new ip address for network interface
	//
	//Params:
	//	linkName - name of the interface
	//	addr - ip address with mask net.IPNet
	//Return:
	//	error - if an error occurs, otherwise nil
	SetAddr(linkName string, addr net.IPNet) error
	//SaveConfiguration save current host network configuration to the config file
	//Save previous config file to .back file
	//
	//Return:
	//	error - if an error occurs, otherwise nil
	SaveConfiguration() error
	//RestoreConfiguration restore host network configuration from .back file
	//
	//Return:
	//	error - if an error occurs, otherwise nil
	RestoreConfiguration() error
	//LoadConfiguration Load host network configuration from config file
	//
	//Return:
	//	error - if an error occurs, otherwise nil
	LoadConfiguration() error
}
