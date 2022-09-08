//go:build linux

package tests

import (
	"os"
	"path/filepath"
	"rol/app/errors"
	"rol/app/services"
	"rol/domain"
	"rol/dtos"
	"rol/infrastructure"
	"runtime"
	"strings"
	"testing"
)

var (
	bridgeService                  *services.HostNetworkService
	bridgeConfigServiceFilePath    string
	bridgeSlaveVlanName            string
	createdBridgeName              string
	bridgeSlaveVlanMasterInterface string
)

func Test_HostNetworkBridgeService_Prepare(t *testing.T) {
	_, filePath, _, _ := runtime.Caller(0)
	bridgeConfigServiceFilePath = filepath.Join(filepath.Dir(filePath), "hostNetworkConfig.yaml")
	configStorage := infrastructure.NewYamlHostNetworkConfigStorage(domain.GlobalDIParameters{RootPath: filepath.Dir(bridgeConfigServiceFilePath)})
	networkManager, err := infrastructure.NewHostNetworkManager(configStorage)
	if err != nil {
		t.Error("error to create host network manager")
	}
	bridgeService = services.NewHostNetworkService(networkManager)

	links, err := networkManager.GetList()
	if err != nil {
		t.Errorf("error getting list: %s", err.Error())
	}
	for _, link := range links {
		if link.GetName() != "lo" && link.GetType() != "vlan" {
			bridgeSlaveVlanMasterInterface = link.GetName()
			break
		}
	}

	createDto := dtos.HostNetworkVlanCreateDto{
		VlanID: 132,
		Parent: bridgeSlaveVlanMasterInterface,
		Addresses: []string{
			"123.123.123.123/24",
			"123.123.124.124/24",
		},
	}
	dto, err := bridgeService.CreateVlan(createDto)
	if err != nil {
		t.Errorf("error creating vlan: %s", err.Error())
	}
	bridgeSlaveVlanName = dto.Name
}

func Test_HostNetworkBridgeService_CreateBridge(t *testing.T) {
	createDto := dtos.HostNetworkBridgeCreateDto{
		Name: "test",
		HostNetworkBridgeBaseDto: dtos.HostNetworkBridgeBaseDto{
			Addresses: []string{
				"123.123.123.123/24",
				"123.123.123.124/24",
			},
			Slaves: nil,
		},
	}
	dto, err := bridgeService.CreateBridge(createDto)
	if err != nil {
		t.Errorf("error creating bridge: %s", err.Error())
	}
	createdBridgeName = dto.Name
	if !strings.Contains(createdBridgeName, "rol.br.") {
		t.Errorf("wrong bridge name: %s, expect rol.br.{%s}", dto.Name, createDto.Name)
	}
}

func Test_HostNetworkBridgeService_CreateBridgeWithIncorrectName(t *testing.T) {
	createDto := dtos.HostNetworkBridgeCreateDto{
		Name: " incorrect ",
		HostNetworkBridgeBaseDto: dtos.HostNetworkBridgeBaseDto{
			Addresses: []string{
				"123.123.123.123/24",
			},
			Slaves: nil,
		},
	}
	dto, err := bridgeService.CreateBridge(createDto)
	if err != nil {
		if !errors.As(err, errors.Validation) {
			t.Errorf("expected error is not Validation error: %s", err.Error())
		}
	} else {
		_ = bridgeService.DeleteBridge(dto.Name)
		t.Error("successfully created vlan with incorrect master interface name")
	}
}

//func Test_HostNetworkVlanService_CreateBridgeWithNotExistedMasterInterface(t *testing.T) {
//	createDto := dtos.HostNetworkVlanCreateDto{
//		VlanID:    133,
//		Parent:    "notexisted",
//		Addresses: []string{},
//	}
//	dto, err := vlanService.CreateVlan(createDto)
//	if err != nil {
//		if !errors.As(err, errors.Validation) {
//			t.Errorf("expected error is not Validation error: %s", err.Error())
//		}
//	} else {
//		_ = vlanService.DeleteVlan(dto.Name)
//		t.Error("successfully created vlan with incorrect master interface name")
//	}
//}

func Test_HostNetworkBridgeService_UpdateBridge(t *testing.T) {
	updateDto := dtos.HostNetworkBridgeUpdateDto{
		HostNetworkBridgeBaseDto: dtos.HostNetworkBridgeBaseDto{
			Addresses: []string{
				"123.123.125.125/24",
			},
			Slaves: []string{
				bridgeSlaveVlanName,
			},
		},
	}
	dto, err := bridgeService.UpdateBridge(createdBridgeName, updateDto)
	if err != nil {
		t.Errorf("error creating vlan: %s", err.Error())
	}
	bridge, err := bridgeService.GetBridgeByName(createdBridgeName)
	if err != nil {
		t.Errorf("get bridge by name failed: %s", err.Error())
	}
	if len(bridge.Addresses) != 1 {
		t.Error("failed to update bridge addresses")
	}
	for _, addressStr := range dto.Addresses {
		if addressStr != "123.123.125.125/24" {
			t.Error("failed to update vlan addresses")
			return
		}
	}
	if len(bridge.Slaves) != 1 {
		t.Error("failed to update bridge slaves")
	}
	if bridge.Slaves[0] != bridgeSlaveVlanName {
		t.Error("failed to update bridge slaves")
	}
}

func Test_HostNetworkBridgeService_UpdateIncorrectAddress(t *testing.T) {
	updateDto := dtos.HostNetworkBridgeUpdateDto{
		HostNetworkBridgeBaseDto: dtos.HostNetworkBridgeBaseDto{
			Addresses: []string{
				"123.123.125.1252/24",
			},
		},
	}
	_, err := bridgeService.UpdateBridge(createdBridgeName, updateDto)
	if err != nil {
		if !errors.As(err, errors.Validation) {
			t.Errorf("expected error is not Validation error: %s", err.Error())
		}
		return
	}
	t.Errorf("error is nil")
}

func Test_HostNetworkBridgeService_GetByNameBridge(t *testing.T) {
	bridge, err := bridgeService.GetBridgeByName(createdBridgeName)
	if err != nil {
		t.Errorf("get bridge by name failed: %s", err.Error())
	}
	if bridge.Name != createdBridgeName {
		t.Errorf("wrong bridge name: %s, expect: %s", bridge.Name, createdBridgeName)
	}
}

func Test_HostNetworkBridgeService_GetList(t *testing.T) {
	bridges, err := bridgeService.GetBridgeList()
	if err != nil {
		t.Errorf("get list failed: %s", err.Error())
	}
	bridgeFound := false
	for _, bridge := range bridges {
		if bridge.Name == "lo" {
			t.Error("got lo interface through bridge service")
		}
		if bridge.Name == createdBridgeName {
			bridgeFound = true
		}
	}
	if !bridgeFound {
		t.Error("created bridge was not found")
	}
}

func Test_HostNetworkBridgeService_Delete(t *testing.T) {
	err := bridgeService.DeleteBridge(createdBridgeName)
	if err != nil {
		t.Errorf("delete bridge failed: %s", err.Error())
	}
	_, err = bridgeService.GetBridgeByName(createdBridgeName)
	if err == nil {
		t.Error("deleted bridge was received")
	}
}

func Test_HostNetworkBridgeService_CleaningAfterTests(t *testing.T) {
	err := bridgeService.DeleteVlan(bridgeSlaveVlanName)
	if err != nil {
		return
	}
	err = os.Remove(bridgeConfigServiceFilePath)
	if err != nil {
		t.Errorf("remove network config file failed:  %q", err)
	}
}
