package domain

import (
	"github.com/google/uuid"
)

//HttpLog http log entity
type HttpLog struct {
	//	Entity - nested base entity
	Entity
	//	HttpMethod - http method
	HttpMethod string
	//	Domain - domain that processed the request
	Domain string
	//	RelativePath - path to the endpoint
	RelativePath string `gorm:"index"`
	//	QueryParams - query params passed
	QueryParams string
	//	QueryParamsInd - query params passed with indexing
	QueryParamsInd string `gorm:"index"`
	//	ClientIP - client IP address
	ClientIP string `gorm:"index"`
	//	Latency - latency in milliseconds
	Latency int
	//	RequestBody - body of the request
	RequestBody string
	//	ResponseBody - body of the response
	ResponseBody string
	//	RequestHeaders - headers of the request
	RequestHeaders string
	//	RequestHeaders - headers of the response
	ResponseHeaders string
	//	CustomRequestHeaders - custom headers of the request
	CustomRequestHeaders string
	//	CustomRequestHeadersInd - custom headers of the request with indexing
	CustomRequestHeadersInd string `gorm:"index"`
}

//GetID gets the id of the entity
func (log HttpLog) GetID() uuid.UUID {
	return log.ID
}
