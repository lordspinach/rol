package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net"
	"net/http"
	"rol/app/errors"
	"rol/app/services"
	"rol/dtos"
	"rol/webapi"
)

//HostNetworkVlanController host network vlan API controller
type HostNetworkVlanController struct {
	service *services.HostNetworkVlanService
	logger  *logrus.Logger
}

//NewHostNetworkVlanController host network vlan controller constructor. Parameters pass through DI
//Params
//	vlanService - vlan service
//	log - logrus logger
//Return
//	*HostNetworkVlanController - instance of host network vlan controller
func NewHostNetworkVlanController(vlanService *services.HostNetworkVlanService, log *logrus.Logger) *HostNetworkVlanController {
	return &HostNetworkVlanController{
		service: vlanService,
		logger:  log,
	}
}

//RegisterHostNetworkVlanController registers controller for getting host network vlans via api
func RegisterHostNetworkVlanController(controller *HostNetworkVlanController, server *webapi.GinHTTPServer) {
	groupRoute := server.Engine.Group("/api/v1")

	groupRoute.GET("/vlan/", controller.GetList)
	groupRoute.GET("/vlan/:name", controller.GetByName)
	groupRoute.GET("/vlan/save-changes", controller.SaveChanges)
	groupRoute.GET("/vlan/reset-changes", controller.ResetChanges)
	groupRoute.POST("/vlan/:name", controller.SetAddr)
	groupRoute.POST("/vlan/", controller.Create)
	groupRoute.DELETE("/vlan/:name", controller.Delete)
}

//GetList get list of host network vlans with 'rol.' prefix
//	Params
//	ctx - gin context
// @Summary Get list of host network vlans
// @version 1.0
// @Tags vlan
// @Accept  json
// @Produce json
// @Success 200 {object} dtos.ResponseDataDto{data=dtos.PaginatedListDto{items=[]dtos.HostNetworkVlanDto}}
// @router /vlan/ [get]
func (h *HostNetworkVlanController) GetList(ctx *gin.Context) {
	vlanList, err := h.service.GetList()
	if err != nil {
		controllerErr := ctx.AbortWithError(http.StatusBadRequest, err)
		if controllerErr != nil {
			h.logger.Errorf("%s : %s", err, controllerErr)
		}
	}
	responseDto := &dtos.ResponseDataDto{
		Status: dtos.ResponseStatusDto{
			Code:    0,
			Message: "",
		},
		Data: vlanList,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

//GetByName get vlan port by name
//Params
//	ctx - gin context
// @Summary Gets vlan port by name
// @version 1.0
// @Tags vlan
// @Accept  json
// @Produce  json
// @param	 name	    path	string		true	"Vlan name"
// @Success 200 {object} dtos.ResponseDataDto{data=dtos.HostNetworkVlanDto}
// @router /vlan/{name} [get]
func (h *HostNetworkVlanController) GetByName(ctx *gin.Context) {
	name := ctx.Param("name")
	vlan, err := h.service.GetByName(name)
	if err != nil {
		if errors.As(err, errors.NotFound) {
			controllerErr := ctx.AbortWithError(http.StatusNotFound, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		} else if errors.As(err, errors.Internal) {
			controllerErr := ctx.AbortWithError(http.StatusInternalServerError, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		} else {
			controllerErr := ctx.AbortWithError(http.StatusBadRequest, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		}
	}
	responseDto := &dtos.ResponseDataDto{
		Status: dtos.ResponseStatusDto{
			Code:    0,
			Message: "",
		},
		Data: vlan,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

//Create new host vlan
//	Params
//	ctx - gin context
// @Summary Create new host vlan
// @version 1.0
// @Tags vlan
// @Accept  json
// @Produce  json
// @Param request body dtos.HostNetworkVlanCreateDto true "Host vlan fields"
// @Success 200 {object} dtos.ResponseDataDto{data=string}
// @router /vlan/ [post]
func (h *HostNetworkVlanController) Create(ctx *gin.Context) {
	reqDto := new(dtos.HostNetworkVlanCreateDto)
	err := ctx.ShouldBindJSON(&reqDto)
	if err != nil {
		if errors.As(err, errors.Internal) {
			controllerErr := ctx.AbortWithError(http.StatusInternalServerError, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		} else {
			controllerErr := ctx.AbortWithError(http.StatusBadRequest, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		}
	}

	// Restoring body in gin.Context for logging it later in middleware
	err = RestoreBody(reqDto, ctx)
	if err != nil {
		ctx.AbortWithStatus(http.StatusInternalServerError)
	}

	name, err := h.service.Create(*reqDto)
	if err != nil {
		if errors.As(err, errors.Internal) {
			controllerErr := ctx.AbortWithError(http.StatusInternalServerError, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		} else {
			controllerErr := ctx.AbortWithError(http.StatusBadRequest, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		}
	}
	responseDto := &dtos.ResponseDataDto{
		Status: dtos.ResponseStatusDto{
			Code:    0,
			Message: "",
		},
		Data: name,
	}
	ctx.JSON(http.StatusOK, responseDto)
}

//SetAddr set address to host network vlan
//	Params
//	ctx - gin context
// @Summary Sets address to host network vlan
// @version 1.0
// @Tags vlan
// @Accept  json
// @Produce  json
// @param	 name	     path	string	false	"Vlan name"
// @param	 addr	     path	string	false	"Address"
// @Success 200 {object} dtos.ResponseDto
// @router /vlan/{name} [post]
func (h *HostNetworkVlanController) SetAddr(ctx *gin.Context) {
	name := ctx.Param("name")
	//addr := ctx.Param("addr")
	addr := ctx.DefaultQuery("addr", "33.33.33.33/24")
	ip, ipNet, err := net.ParseCIDR(addr)
	if err != nil {
		controllerErr := ctx.AbortWithError(http.StatusBadRequest, err)
		if controllerErr != nil {
			h.logger.Errorf("%s : %s", err, controllerErr)
		}
	}
	ipNet.IP = ip
	err = h.service.SetAddr(name, *ipNet)
	if err != nil {
		if errors.As(err, errors.Internal) {
			controllerErr := ctx.AbortWithError(http.StatusInternalServerError, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		} else {
			controllerErr := ctx.AbortWithError(http.StatusBadRequest, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		}
	}
	responseDto := &dtos.ResponseDto{
		Status: dtos.ResponseStatusDto{
			Code:    0,
			Message: "",
		},
	}
	ctx.JSON(http.StatusOK, responseDto)
}

//Delete host network vlan
//	Params
//	ctx - gin context
// @Summary Delete host network vlan by name
// @version 1.0
// @Tags vlan
// @Accept  json
// @Produce  json
// @param	 name	path	string		true	"Vlan name"
// @Success 200 {object} dtos.ResponseDto
// @router /vlan/{name} [delete]
func (h *HostNetworkVlanController) Delete(ctx *gin.Context) {
	name := ctx.Param("name")
	err := h.service.Delete(name)
	if err != nil {
		if errors.As(err, errors.NotFound) {
			controllerErr := ctx.AbortWithError(http.StatusNotFound, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		} else if errors.As(err, errors.Internal) {
			controllerErr := ctx.AbortWithError(http.StatusInternalServerError, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		} else {
			controllerErr := ctx.AbortWithError(http.StatusBadRequest, err)
			if controllerErr != nil {
				h.logger.Errorf("%s : %s", err, controllerErr)
			}
		}
	}
	responseDto := &dtos.ResponseDto{
		Status: dtos.ResponseStatusDto{
			Code:    0,
			Message: "",
		},
	}
	ctx.JSON(http.StatusOK, responseDto)
}

//SaveChanges save changes to config file
//	Params
//	ctx - gin context
// @Summary SaveChanges saves changes to config file
// @version 1.0
// @Tags vlan
// @Accept  json
// @Produce  json
// @Success 200 {object} dtos.ResponseDto
// @router /vlan/save-changes [get]
func (h *HostNetworkVlanController) SaveChanges(ctx *gin.Context) {
	err := h.service.SaveChanges()
	if err != nil {
		controllerErr := ctx.AbortWithError(http.StatusInternalServerError, err)
		if controllerErr != nil {
			h.logger.Errorf("%s : %s", err, controllerErr)
		}
	}
	responseDto := &dtos.ResponseDto{
		Status: dtos.ResponseStatusDto{
			Code:    0,
			Message: "",
		},
	}
	ctx.JSON(http.StatusOK, responseDto)
}

//ResetChanges reset all unsaved changes
//	Params
//	ctx - gin context
// @Summary ResetChanges reset all unsaved changes
// @version 1.0
// @Tags vlan
// @Accept  json
// @Produce  json
// @Success 200 {object} dtos.ResponseDto
// @router /vlan/reset-changes [get]
func (h *HostNetworkVlanController) ResetChanges(ctx *gin.Context) {
	err := h.service.ResetChanges()
	if err != nil {
		controllerErr := ctx.AbortWithError(http.StatusInternalServerError, err)
		if controllerErr != nil {
			h.logger.Errorf("%s : %s", err, controllerErr)
		}
	}
	responseDto := &dtos.ResponseDto{
		Status: dtos.ResponseStatusDto{
			Code:    0,
			Message: "",
		},
	}
	ctx.JSON(http.StatusOK, responseDto)
}
