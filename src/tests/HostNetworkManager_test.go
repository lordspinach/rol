//go:build linux

package tests

import (
	"net"
	"os"
	"path/filepath"
	"rol/app/interfaces"
	"rol/domain"
	"rol/infrastructure"
	"runtime"
	"testing"
)

var (
	vlanName              string
	vlanID                int
	networkManager        interfaces.IHostNetworkManager
	configManagerFilePath string
)

func Test_HostNetworkManager_Prepare(_ *testing.T) {
	_, filePath, _, _ := runtime.Caller(0)
	configManagerFilePath = filepath.Join(filepath.Dir(filePath), "testNetworkConfig.yaml")
	networkManager = infrastructure.NewHostNetworkManager("testNetworkConfig.yaml", domain.GlobalDIParameters{RootPath: filepath.Dir(configManagerFilePath)})
}

func Test_HostNetworkManager_GetList(t *testing.T) {
	links, err := networkManager.GetList()
	if err != nil {
		t.Errorf("error getting list: %s", err.Error())
	}
	loExist := false
	for _, link := range links {
		if link.GetName() == "lo" {
			loExist = true
		}
	}
	if !loExist {
		t.Errorf("localhost not found")
	}
	err = networkManager.SaveConfiguration()
	if err != nil {
		t.Errorf("error saving configuration: %s", err.Error())
	}
}

func Test_HostNetworkManager_CreateVlan(t *testing.T) {
	var master string
	links, err := networkManager.GetList()
	if err != nil {
		t.Errorf("error getting list: %s", err.Error())
	}
	for _, link := range links {
		if link.GetName() != "lo" && link.GetType() != "vlan" {
			master = link.GetName()
			break
		}
	}
	vlanID = 146
	vlanName, err = networkManager.CreateVlan(master, vlanID)
	if err != nil {
		t.Errorf("error creating vlan: %s", err.Error())
	}
	links, err = networkManager.GetList()
	vlanFound := false
	for _, link := range links {
		if link.GetName() == vlanName {
			vlanFound = true
		}
	}
	if !vlanFound {
		t.Error("created vlan not found")
	}
}

func Test_HostNetworkManager_SetVlanAddr(t *testing.T) {
	ip, ipNet, err := net.ParseCIDR("192.111.111.111/24")
	if err != nil {
		t.Errorf("parse CIDR failed: %s", err.Error())
	}
	ipNet.IP = ip
	err = networkManager.SetAddr(vlanName, *ipNet)
	if err != nil {
		t.Errorf("failed to set the ip address: %s", err.Error())
	}
	vlan, err := networkManager.GetByName(vlanName)
	if err != nil {
		t.Errorf("error getting by name: %s", err.Error())
	}
	addresses := vlan.GetAddresses()
	addrFound := false
	for _, addr := range addresses {
		if addr.IP.Equal(ip) {
			addrFound = true
		}
	}
	if !addrFound {
		t.Error("the address that was added was not found")
	}
}

func Test_HostNetworkManager_SaveConfiguration(t *testing.T) {
	err := networkManager.SaveConfiguration()
	if err != nil {
		t.Errorf("failed saving configuration: %s", err.Error())
	}
	conf, err := infrastructure.ReadYamlFile[domain.HostNetworkInterfacesConfig](configManagerFilePath)
	vlanFound := false
	for _, vlan := range conf.Vlans {
		if vlan.Name == vlanName {
			vlanFound = true
		}
	}
	if !vlanFound {
		t.Error("created vlan not found in configuration file")
	}
}

func Test_HostNetworkManager_Delete(t *testing.T) {
	err := networkManager.DeleteLinkByName(vlanName)
	if err != nil {
		t.Errorf("failed deleting vlan: %s", err.Error())
	}
	links, err := networkManager.GetList()
	vlanFound := false
	for _, link := range links {
		if link.GetName() == vlanName {
			vlanFound = true
		}
	}
	if vlanFound {
		t.Error("deleted vlan was found")
	}
}

func Test_HostNetworkManager_LoadConfiguration(t *testing.T) {
	err := networkManager.LoadConfiguration()
	if err != nil {
		t.Errorf("failed loading configuration: %s", err.Error())
	}
	links, err := networkManager.GetList()
	vlanFound := false
	for _, link := range links {
		if link.GetName() == vlanName {
			vlanFound = true
		}
	}
	if !vlanFound {
		t.Error("vlan not found at host")
	}
}

func Test_HostNetworkManager_RestoreConfiguration(t *testing.T) {
	err := networkManager.RestoreConfiguration()
	if err != nil {
		t.Errorf("failed restore configuration: %s", err.Error())
	}
	conf, err := infrastructure.ReadYamlFile[domain.HostNetworkInterfacesConfig](configManagerFilePath)
	vlanFoundConf := false
	for _, vlan := range conf.Vlans {
		if vlan.Name == vlanName {
			vlanFoundConf = true
		}
	}
	if vlanFoundConf {
		t.Error("vlan was found in restored configuration file")
	}

	links, err := networkManager.GetList()
	vlanFoundHost := false
	for _, link := range links {
		if link.GetType() == "vlan" && link.GetName() == vlanName {
			vlanFoundHost = true
		}
	}
	if vlanFoundHost {
		t.Error("vlan was found on host")
	}
}

func Test_HostNetworkManager_CleaningAfterTests(t *testing.T) {
	err := os.Remove(configManagerFilePath)
	if err != nil {
		t.Errorf("remove network config file failed:  %q", err)
	}
}
