package mappers

import (
	"rol/domain"
	"rol/dtos"
)

//MapTFTPConfigToDto writes TFTP config entity to dto
//Params
//	entity - ethernet switch port entity
//	dto - dest ethernet switch port dto
func MapTFTPConfigToDto(entity domain.TFTPConfig, dto *dtos.TFTPConfigDto) {
	dto.Address = entity.Address
	dto.Port = entity.Port
	dto.ID = entity.ID
	dto.UpdatedAt = entity.UpdatedAt
	dto.CreatedAt = entity.CreatedAt
	dto.Enabled = entity.Enabled
}

//MapTFTPConfigCreateDtoToEntity writes TFTP config create dto fields to entity
//Params
//	dto - TFTP config create dto
//	entity - dest TFTP config entity
func MapTFTPConfigCreateDtoToEntity(dto dtos.TFTPConfigCreateDto, entity *domain.TFTPConfig) {
	entity.Port = dto.Port
	entity.Address = dto.Address
	entity.Enabled = dto.Enabled
}

//MapTFTPConfigUpdateDtoToEntity writes TFTP config update dto fields to entity
//Params
//	dto - TFTP config update dto
//	entity - dest TFTP config entity
func MapTFTPConfigUpdateDtoToEntity(dto dtos.TFTPConfigUpdateDto, entity *domain.TFTPConfig) {
	entity.Port = dto.Port
	entity.Address = dto.Address
	entity.Enabled = dto.Enabled
}
