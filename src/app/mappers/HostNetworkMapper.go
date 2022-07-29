package mappers

import (
	"rol/domain"
	"rol/dtos"
)

func MapHostNetworkVlanToDto(vlan domain.HostNetworkVlan, dto *dtos.HostNetworkVlanDto) {
	dto.Name = vlan.Name
	dto.VlanID = vlan.VlanID
	for _, addr := range vlan.Addresses {
		dto.Addresses = append(dto.Addresses, addr.String())
	}
	dto.Master = vlan.Master
}
