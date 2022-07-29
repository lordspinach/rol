//go:build linux

package tests

import (
	"net"
	"os"
	"path/filepath"
	"rol/app/services"
	"rol/domain"
	"rol/dtos"
	"rol/infrastructure"
	"runtime"
	"strings"
	"testing"
)

var (
	vlanService            services.HostNetworkVlanService
	configServiceFilePath  string
	serviceMasterInterface string
	createdVlanName        string
)

func Test_HostNetworkVlanService_Prepare(t *testing.T) {
	_, filePath, _, _ := runtime.Caller(0)
	configServiceFilePath = filepath.Join(filepath.Dir(filePath), "testServiceNetworkConfig.yaml")
	networkManager = infrastructure.NewHostNetworkManager("testServiceNetworkConfig.yaml", domain.GlobalDIParameters{RootPath: filepath.Dir(configServiceFilePath)})
	vlanService = services.NewHostNetworkVlanService(networkManager)

	links, err := networkManager.GetList()
	if err != nil {
		t.Errorf("error getting list: %s", err.Error())
	}
	for _, link := range links {
		if link.GetName() != "lo" && link.GetType() != "vlan" {
			serviceMasterInterface = link.GetName()
			break
		}
	}
}

func Test_HostNetworkVlanService_CreateVlan(t *testing.T) {
	dto := dtos.HostNetworkVlanCreateDto{
		VlanID: 132,
		Master: serviceMasterInterface,
	}
	name, err := vlanService.Create(dto)
	if err != nil {
		t.Errorf("error creating vlan: %s", err.Error())
	}
	createdVlanName = name
	if !strings.Contains(name, "rol.") {
		t.Errorf("wrong vlan name: %s, expect rol.{%d}.{%s}", name, dto.VlanID, serviceMasterInterface)
	}
	vlan, err := vlanService.GetByName(name)
	if err != nil {
		t.Errorf("get vlan by name failed: %s", err.Error())
	}
	if vlan.Name != name {
		t.Errorf("wrong vlan name: %s, expect %s", vlan.Name, name)
	}
}

func Test_HostNetworkVlanService_SetAddr(t *testing.T) {
	ip, ipNet, err := net.ParseCIDR("123.123.123.123/24")
	if err != nil {
		t.Errorf("error parse CIDR: %s", err.Error())
	}
	ipNet.IP = ip
	err = vlanService.SetAddr(createdVlanName, *ipNet)
	if err != nil {
		t.Errorf("error set address: %s", err.Error())
	}
	vlan, err := vlanService.GetByName(createdVlanName)
	if err != nil {
		t.Errorf("get vlan by name failed: %s", err.Error())
	}
	addrFound := false
	for _, addr := range vlan.Addresses {
		if ipNet.String() == addr {
			addrFound = true
		}
	}
	if !addrFound {
		t.Error("ip address that was set for the vlan was not found")
	}
}

func Test_HostNetworkVlanService_GetByNameLo(t *testing.T) {
	expectedErr := "vlan not found"
	_, err := vlanService.GetByName("lo")
	if err.Error() != expectedErr {
		t.Errorf("unexpected behavior, expect '%s' error, got %s", expectedErr, err.Error())
	}
}

func Test_HostNetworkVlanService_GetByNameVlan(t *testing.T) {
	vlan, err := vlanService.GetByName(createdVlanName)
	if err != nil {
		t.Errorf("get vlan by name failed: %s", err.Error())
	}
	if vlan.Name != createdVlanName {
		t.Errorf("wrong vlan name: %s, expect: %s", vlan.Name, createdVlanName)
	}
}

func Test_HostNetworkVlanService_GetList(t *testing.T) {
	vlans, err := vlanService.GetList()
	if err != nil {
		t.Errorf("get list failed: %s", err.Error())
	}
	vlanFound := false
	for _, vlan := range vlans {
		if vlan.Name == "lo" {
			t.Error("got localhost through vlan service")
		}
		if vlan.Name == createdVlanName {
			vlanFound = true
		}
	}
	if !vlanFound {
		t.Error("created vlan was not found")
	}
}

func Test_HostNetworkVlanService_SaveChanges(t *testing.T) {
	err := vlanService.SaveChanges()
	if err != nil {
		t.Errorf("save changes failed: %s", err.Error())
	}
	config, err := infrastructure.ReadYamlFile[domain.HostNetworkInterfacesConfig](configServiceFilePath)
	if err != nil {
		t.Errorf("read yaml file failed: %s", err.Error())
	}
	vlanFound := false
	for _, vlan := range config.Vlans {

		if vlan.Name == createdVlanName {
			vlanFound = true
		}
	}
	if !vlanFound {
		t.Error("created vlan was not found in config file")
	}
}

func Test_HostNetworkVlanService_Delete(t *testing.T) {
	err := vlanService.Delete(createdVlanName)
	if err != nil {
		t.Errorf("delete vlan failed: %s", err.Error())
	}
	_, err = vlanService.GetByName(createdVlanName)
	if err == nil {
		t.Error("deleted vlan was received")
	}
}

func Test_HostNetworkVlanService_ResetChanges(t *testing.T) {
	err := vlanService.ResetChanges()
	if err != nil {
		t.Errorf("reset changes failed: %s", err.Error())
	}
	_, err = vlanService.GetByName(createdVlanName)
	if err != nil {
		t.Error("recovered vlan was not found")
	}
}

func Test_HostNetworkVlanService_CleaningAfterTests(t *testing.T) {
	err := vlanService.Delete(createdVlanName)
	if err != nil {
		t.Errorf("delete vlan failed: %s", err.Error())
	}
	_, err = vlanService.GetByName(createdVlanName)
	if err == nil {
		t.Error("deleted vlan was received")
	}
	err = os.Remove(configServiceFilePath)
	if err != nil {
		t.Errorf("remove network config file failed:  %q", err)
	}
}
