package domain

//HostNetworkInterfacesConfig is a struct for yaml configuration file
type HostNetworkInterfacesConfig struct {
	//Devices slice of HostNetworkDevice
	Devices []HostNetworkDevice
	//Vlans slice of HostNetworkVlan
	Vlans []HostNetworkVlan
}
