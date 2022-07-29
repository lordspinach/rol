package dtos

type HostNetworkVlanDto struct {
	Name      string
	Addresses []string
	VlanID    int
	Master    string
}
