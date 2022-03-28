package mappers

import (
	"rol/domain"
	"rol/dtos"
)

//MapAppLogEntityToDto writes app log entity fields in the dto
func MapAppLogEntityToDto(entity domain.AppLog, dto *dtos.AppLogDto) {
	dto.BaseDto.Id = entity.ID
	dto.CreatedAt = entity.CreatedAt
	dto.ActionID = entity.ActionID
	dto.Level = entity.Level
	dto.Source = entity.Source
	dto.Message = entity.Message
}

//MapHTTPLogEntityToDto  writes http log entity fields in the dto
func MapHTTPLogEntityToDto(entity domain.HttpLog, dto *dtos.HttpLogDto) {
	dto.BaseDto.Id = entity.ID
	dto.CreatedAt = entity.CreatedAt
	dto.HttpMethod = entity.HttpMethod
	dto.Domain = entity.Domain
	dto.RelativePath = entity.RelativePath
	dto.QueryParams = entity.QueryParams
	dto.ClientIP = entity.ClientIP
	dto.Latency = entity.Latency
	dto.RequestBody = entity.RequestBody
	dto.ResponseBody = entity.ResponseBody
	dto.RequestHeaders = entity.RequestHeaders
	dto.CustomRequestHeaders = entity.CustomRequestHeaders
}
