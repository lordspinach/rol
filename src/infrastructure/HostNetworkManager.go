//go:build linux

package infrastructure

import (
	"fmt"
	"github.com/vishvananda/netlink"
	"net"
	"os"
	"path/filepath"
	"rol/app/errors"
	"rol/app/interfaces"
	"rol/domain"
	"strings"
)

//HostNetworkManager is a struct for network manager
type HostNetworkManager struct {
	configFilePath string
	networkConfig  domain.HostNetworkInterfacesConfig
}

//NewHostNetworkManager constructor for HostNetworkManager
func NewHostNetworkManager(parameters domain.GlobalDIParameters) interfaces.IHostNetworkManager {
	return HostNetworkManager{configFilePath: filepath.Join(parameters.RootPath, "hostNetworkConfig.yaml")}
}

func (h HostNetworkManager) parseLinkAddr(link netlink.Link) ([]net.IPNet, error) {
	addrList, err := netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		return nil, errors.Internal.Wrap(err, "get addresses error")
	}

	var out []net.IPNet
	for _, addr := range addrList {
		out = append(out, *addr.IPNet)
	}
	return out, nil
}

func (h HostNetworkManager) getMasterName(link netlink.Link) (string, error) {
	parent, err := netlink.LinkByIndex(link.Attrs().ParentIndex)
	if err != nil {
		return "", errors.Internal.Wrap(err, "get host interface by index failed")
	}
	return parent.Attrs().Name, nil
}

func (h HostNetworkManager) mapLink(link netlink.Link) (interfaces.IHostNetworkLink, error) {
	if link.Type() == "device" {
		addresses, err := h.parseLinkAddr(link)
		if err != nil {
			return nil, errors.Internal.Wrap(err, "error parsing link addresses")
		}
		device := domain.HostNetworkDevice{HostNetworkLink: domain.HostNetworkLink{
			Name:      link.Attrs().Name,
			Type:      link.Type(),
			Addresses: addresses,
		}}
		return device, nil

	} else if link.Type() == "vlan" {
		addresses, err := h.parseLinkAddr(link)
		if err != nil {
			return nil, errors.Internal.Wrap(err, "error parsing link addresses")
		}
		master, err := h.getMasterName(link)
		if err != nil {
			return nil, errors.Internal.Wrap(err, "error getting parent name")
		}
		vlan := domain.HostNetworkVlan{
			HostNetworkLink: domain.HostNetworkLink{
				Name:      link.Attrs().Name,
				Type:      link.Type(),
				Addresses: addresses,
			},
			VlanID: link.(*netlink.Vlan).VlanId,
			Master: master,
		}
		return vlan, nil
	}
	return domain.HostNetworkLink{Name: link.Attrs().Name, Type: "none", Addresses: []net.IPNet{}}, nil
}

//GetList gets list of host network interfaces
//
//Return:
//	[]interfaces.IHostNetworkLink - list of interfaces
//	error - if an error occurs, otherwise nil
func (h HostNetworkManager) GetList() ([]interfaces.IHostNetworkLink, error) {
	var out []interfaces.IHostNetworkLink
	links, err := netlink.LinkList()
	if err != nil {
		return nil, errors.Internal.Wrap(err, "error getting a list of link devices")
	}
	for _, link := range links {
		networkInterface, err := h.mapLink(link)
		if err != nil {
			return nil, errors.Internal.Wrap(err, "failed to map device link to HostNetworkLink")
		}
		out = append(out, networkInterface)
	}

	return out, nil
}

//GetByName gets host network interface by its name
//
//Params:
//	name - interface name
//Return:
//	interfaces.IHostNetworkLink - interfaces
//	error - if an error occurs, otherwise nil
func (h HostNetworkManager) GetByName(name string) (interfaces.IHostNetworkLink, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, errors.Internal.Wrap(err, "failed to map device link to HostNetworkLink")
	}
	out, err := h.mapLink(link)
	if err != nil {
		return nil, errors.Internal.Wrap(err, "failed to map device link to HostNetworkLink")
	}
	return out, nil
}

//CreateVlan creates vlan on host
//
//Params:
//	master - name of the master network interface
//	vlanID - ID to be set for vlan
//Return:
//	string - new vlan name that will be rol.{master}.{vlanID}
//	error - if an error occurs, otherwise nil
func (h HostNetworkManager) CreateVlan(master string, vlanID int) (string, error) {
	parent, err := netlink.LinkByName(master)
	if err != nil {
		return "", errors.Internal.Wrap(err, "getting device link by name failed")
	}
	la := netlink.NewLinkAttrs()

	vlanName := fmt.Sprintf("rol.%s.%d", master, vlanID)
	la.Name = vlanName
	la.ParentIndex = parent.Attrs().Index
	vlan := &netlink.Vlan{
		LinkAttrs:    la,
		VlanId:       vlanID,
		VlanProtocol: netlink.VLAN_PROTOCOL_8021Q,
	}
	err = netlink.LinkAdd(vlan)
	if err != nil {
		return "", errors.Internal.Wrap(err, "failed to add vlan link")
	}

	err = netlink.LinkSetUp(vlan)
	if err != nil {
		return "", errors.Internal.Wrap(err, "vlan link set up failed")
	}

	return vlanName, nil
}

//DeleteLinkByName deletes interface on host by its name
//
//Params:
//	name - interface name
//Return
//	error - if an error occurs, otherwise nil
func (h HostNetworkManager) DeleteLinkByName(name string) error {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return errors.Internal.Wrap(err, "getting link by name failed")
	}
	err = netlink.LinkDel(link)
	if err != nil {
		return errors.Internal.Wrap(err, "deleting link failed")
	}
	return nil
}

//SetAddr sets new ip address for network interface
//
//Params:
//	linkName - name of the interface
//	addr - ip address with mask net.IPNet
//Return:
//	error - if an error occurs, otherwise nil
func (h HostNetworkManager) SetAddr(linkName string, addr net.IPNet) error {
	link, err := netlink.LinkByName(linkName)
	if err != nil {
		return errors.Internal.Wrap(err, "getting link by name failed")
	}
	cidr := addr.String()
	linkAddr, err := netlink.ParseAddr(cidr)
	if err != nil {
		return errors.Internal.Wrap(err, "parse cidr address failed")
	}
	err = netlink.AddrAdd(link, linkAddr)
	if err != nil {
		return errors.Internal.Wrap(err, "error adding address to link")
	}
	return nil
}

//SaveConfiguration save current host network configuration to the config file
//Save previous config file to .back file
//
//Return:
//	error - if an error occurs, otherwise nil
func (h HostNetworkManager) SaveConfiguration() error {
	h.networkConfig = domain.HostNetworkInterfacesConfig{}
	networkInterfaces, err := h.GetList()
	if err != nil {
		return errors.Internal.Wrap(err, "failed to get list of host network interfaces")
	}

	for _, inter := range networkInterfaces {
		if inter.GetType() == "vlan" {
			h.networkConfig.Vlans = append(h.networkConfig.Vlans, inter.(domain.HostNetworkVlan))
		} else if inter.GetType() == "device" {
			h.networkConfig.Devices = append(h.networkConfig.Devices, inter.(domain.HostNetworkDevice))
		}
	}

	if _, err := os.Stat(h.configFilePath); err == nil {
		// Create backup file
		backupFilePath := fmt.Sprintf(h.configFilePath + ".back")
		err = os.Rename(h.configFilePath, backupFilePath)
		if err != nil {
			return errors.Internal.Wrap(err, "error when creating backup file")
		}
	}

	err = SaveYamlFile[domain.HostNetworkInterfacesConfig](h.networkConfig, h.configFilePath)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to save host network configuration")
	}
	return nil
}

//RestoreConfiguration restore host network configuration from .back file
//
//Return:
//	error - if an error occurs, otherwise nil
func (h HostNetworkManager) RestoreConfiguration() error {
	backupPath := h.configFilePath + ".back"
	_, err := os.Stat(backupPath)
	if err != nil {
		return errors.Internal.Wrap(err, "backup config file not found")
	}
	err = os.Remove(h.configFilePath)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to remove config file")
	}
	err = os.Rename(backupPath, h.configFilePath)
	if err != nil {
		return errors.Internal.Wrap(err, "failed to restore config file")
	}
	err = h.LoadConfiguration()
	if err != nil {
		return errors.Internal.Wrap(err, "load configuration failed")
	}
	return nil
}

//LoadConfiguration Load host network configuration from config file
//
//Return:
//	error - if an error occurs, otherwise nil
func (h HostNetworkManager) LoadConfiguration() error {
	var err error
	h.networkConfig, err = ReadYamlFile[domain.HostNetworkInterfacesConfig](h.configFilePath)
	if err != nil {
		return errors.Internal.Wrap(err, "error reading configuration from file")
	}
	err = h.loadVlanConfiguration()
	if err != nil {
		return errors.Internal.Wrap(err, "error loading vlan configuration")
	}
	return nil
}

func (h HostNetworkManager) loadVlanConfiguration() error {
	networkInterfaces, err := h.GetList()
	if err != nil {
		return errors.Internal.Wrap(err, "failed to get list of host network interfaces")
	}
	for _, vlan := range h.networkConfig.Vlans {
		vlanExist, err := h.vlanExistOnHost(vlan.Name)
		if err != nil {
			return errors.Internal.Wrap(err, "vlan existence on host check error")
		}
		if !vlanExist {
			vlanName, err := h.CreateVlan(vlan.Master, vlan.VlanID)
			if err != nil {
				return errors.Internal.Wrap(err, "error when creating a vlan")
			}
			for _, addr := range vlan.Addresses {
				err = h.SetAddr(vlanName, addr)
				if err != nil {
					return errors.Internal.Wrap(err, "failed set address to vlan")
				}
			}
		}
	}
	for _, inter := range networkInterfaces {
		if inter.GetType() != "vlan" {
			continue
		}
		if !h.vlanExistInConfig(inter.GetName()) && strings.Contains(inter.GetName(), "rol.") {
			err := h.DeleteLinkByName(inter.GetName())
			if err != nil {
				return errors.Internal.Wrap(err, "delete link by name error")
			}
		}
	}
	return nil
}

func (h HostNetworkManager) vlanExistOnHost(vlanName string) (bool, error) {
	networkInterfaces, err := h.GetList()
	if err != nil {
		return false, errors.Internal.Wrap(err, "failed to get list of host network interfaces")
	}
	for _, inter := range networkInterfaces {
		if inter.GetType() == "vlan" && inter.GetName() == vlanName {
			return true, nil
		}
	}
	return false, nil
}

func (h HostNetworkManager) vlanExistInConfig(vlanName string) bool {
	for _, vlan := range h.networkConfig.Vlans {
		if vlan.GetName() == vlanName {
			return true
		}
	}
	return false
}
