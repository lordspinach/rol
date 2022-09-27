package tests

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"path"
	customErrors "rol/app/errors"
	"rol/app/services"
	"rol/domain"
	"rol/dtos"
	"rol/infrastructure"
	"runtime"
	"testing"
)

type switchVLANServiceTester struct {
	service  *services.EthernetSwitchService
	dbPath   string
	portID   uuid.UUID
	switchID uuid.UUID
	vlanID   uuid.UUID
}

var vlanServiceTester *switchVLANServiceTester

func Test_EthernetSwitchVLANService_Prepare(t *testing.T) {
	vlanServiceTester = &switchVLANServiceTester{}
	vlanServiceTester.dbPath = "ethernetSwitchVlanService_test.db"
	dbConnection := sqlite.Open(vlanServiceTester.dbPath)
	testGenDb, err := gorm.Open(dbConnection, &gorm.Config{})
	if err != nil {
		t.Errorf("creating db failed: %v", err)
	}
	err = testGenDb.AutoMigrate(
		new(domain.EthernetSwitch),
		new(domain.EthernetSwitchPort),
		new(domain.EthernetSwitchVLAN),
	)
	if err != nil {
		t.Errorf("migration failed: %v", err)
	}

	logger := logrus.New()
	switchRepo := infrastructure.NewEthernetSwitchRepository(testGenDb, logger)
	portRepo := infrastructure.NewEthernetSwitchPortRepository(testGenDb, logger)
	vlanRepo := infrastructure.NewEthernetSwitchVLANRepository(testGenDb, logger)
	getter := infrastructure.NewEthernetSwitchManagerGetter()

	service, _ := services.NewEthernetSwitchService(switchRepo, portRepo, vlanRepo, getter)
	vlanServiceTester.service = service
	err = services.EthernetSwitchServiceInit(vlanServiceTester.service)
	if err != nil {
		t.Errorf("init service failed:  %q", err)
	}

	_, filename, _, _ := runtime.Caller(1)
	if _, err := os.Stat(path.Join(path.Dir(filename), vlanServiceTester.dbPath)); errors.Is(err, os.ErrNotExist) {
		return
	}
	err = os.Remove(vlanServiceTester.dbPath)
	if err != nil {
		t.Errorf("remove db failed:  %q", err)
	}
}

func Test_EthernetSwitchVLANService_CreateRelatedEntities(t *testing.T) {
	ethSwitch := dtos.EthernetSwitchCreateDto{
		EthernetSwitchBaseDto: dtos.EthernetSwitchBaseDto{
			Name:        "TestSwitch",
			Serial:      "serial",
			SwitchModel: "unifi_switch_us-24-250w",
			Address:     "1.1.1.1",
			Username:    "Test",
		},
		Password: "TestTest",
	}
	createdSwitch, err := vlanServiceTester.service.Create(context.Background(), ethSwitch)
	if err != nil {
		t.Errorf("create switch failed:  %q", err)
	}
	switchPort := dtos.EthernetSwitchPortCreateDto{EthernetSwitchPortBaseDto: dtos.EthernetSwitchPortBaseDto{
		POEType:    "poe",
		Name:       "Gi",
		POEEnabled: false,
	}}
	createdPort, err := vlanServiceTester.service.CreatePort(context.Background(), createdSwitch.ID, switchPort)
	if err != nil {
		t.Errorf("create switch port failed:  %q", err)
	}
	vlanServiceTester.switchID = createdSwitch.ID
	vlanServiceTester.portID = createdPort.ID
}

func Test_EthernetSwitchVLANService_CreateFailByWrongID(t *testing.T) {
	vlanDto := dtos.EthernetSwitchVLANCreateDto{
		EthernetSwitchVLANBaseDto: dtos.EthernetSwitchVLANBaseDto{
			UntaggedPorts: nil,
			TaggedPorts:   nil,
		},
		VlanID: -1,
	}
	_, err := vlanServiceTester.service.CreateVLAN(context.Background(), vlanServiceTester.switchID, vlanDto)
	if err == nil {
		t.Error("nil error acquired")
	}
	if !customErrors.As(err, customErrors.Validation) {
		t.Error("wrong error type acquired")
	}
}

func Test_EthernetSwitchVLANService_CreateFailBySamePortsIDs(t *testing.T) {
	vlanDto := dtos.EthernetSwitchVLANCreateDto{
		EthernetSwitchVLANBaseDto: dtos.EthernetSwitchVLANBaseDto{
			UntaggedPorts: []uuid.UUID{vlanServiceTester.portID},
			TaggedPorts:   []uuid.UUID{vlanServiceTester.portID},
		},
		VlanID: 2,
	}
	_, err := vlanServiceTester.service.CreateVLAN(context.Background(), vlanServiceTester.switchID, vlanDto)
	if err == nil {
		t.Error("nil error acquired")
	}
	if !customErrors.As(err, customErrors.Validation) {
		t.Error("wrong error type acquired")
	}
}

func Test_EthernetSwitchVLANSService_CreateFailByNonExistentSwitch(t *testing.T) {
	vlanDto := dtos.EthernetSwitchVLANCreateDto{
		EthernetSwitchVLANBaseDto: dtos.EthernetSwitchVLANBaseDto{
			UntaggedPorts: []uuid.UUID{vlanServiceTester.portID},
			TaggedPorts:   nil,
		},
		VlanID: 2,
	}
	_, err := vlanServiceTester.service.CreateVLAN(context.Background(), uuid.New(), vlanDto)
	if err == nil {
		t.Error("nil error acquired")
	}
	if !customErrors.As(err, customErrors.NotFound) {
		t.Error("wrong error type acquired")
	}
}

func Test_EthernetSwitchVLANSService_CreateFailByNonExistentPort(t *testing.T) {
	vlanDto := dtos.EthernetSwitchVLANCreateDto{
		EthernetSwitchVLANBaseDto: dtos.EthernetSwitchVLANBaseDto{
			UntaggedPorts: []uuid.UUID{uuid.New()},
			TaggedPorts:   nil,
		},
		VlanID: 2,
	}
	_, err := vlanServiceTester.service.CreateVLAN(context.Background(), vlanServiceTester.switchID, vlanDto)
	if err == nil {
		t.Error("nil error acquired")
	}
	if !customErrors.As(err, customErrors.Validation) {
		t.Error("wrong error type acquired")
	}
}

func Test_EthernetSwitchVLANSService_CreateOK(t *testing.T) {
	vlanDto := dtos.EthernetSwitchVLANCreateDto{
		EthernetSwitchVLANBaseDto: dtos.EthernetSwitchVLANBaseDto{
			UntaggedPorts: []uuid.UUID{vlanServiceTester.portID},
			TaggedPorts:   nil,
		},
		VlanID: 2,
	}
	vlan, err := vlanServiceTester.service.CreateVLAN(context.Background(), vlanServiceTester.switchID, vlanDto)
	if err != nil {
		t.Errorf("failed to create switch vlan: %q", err)
	}
	vlanServiceTester.vlanID = vlan.ID
}

func Test_EthernetSwitchVLANSService_CreateFailByVLANIDUniqueness(t *testing.T) {
	vlanDto := dtos.EthernetSwitchVLANCreateDto{
		EthernetSwitchVLANBaseDto: dtos.EthernetSwitchVLANBaseDto{
			UntaggedPorts: []uuid.UUID{vlanServiceTester.portID},
			TaggedPorts:   nil,
		},
		VlanID: 2,
	}
	_, err := vlanServiceTester.service.CreateVLAN(context.Background(), vlanServiceTester.switchID, vlanDto)
	if err == nil {
		t.Error("nil error acquired")
	}
	if !customErrors.As(err, customErrors.Validation) {
		t.Error("wrong error type acquired")
	}
}

func Test_EthernetSwitchVLANSService_GetByIDFailByNonExistentSwitch(t *testing.T) {
	_, err := vlanServiceTester.service.GetVLANByID(context.Background(), uuid.New(), vlanServiceTester.vlanID)
	if err == nil {
		t.Error("nil error acquired")
	}
	if !customErrors.As(err, customErrors.NotFound) {
		t.Error("wrong error type acquired")
	}
}

func Test_EthernetSwitchVLANSService_GetByIDOK(t *testing.T) {
	vlan, err := vlanServiceTester.service.GetVLANByID(context.Background(), vlanServiceTester.switchID, vlanServiceTester.vlanID)
	if err != nil {
		t.Error("get vlan by id failed")
	}
	if vlan.UntaggedPorts[0] != vlanServiceTester.portID {
		t.Error("wrong untagged port acquired")
	}
}

func Test_EthernetSwitchVLANSService_Update(t *testing.T) {
	updDto := dtos.EthernetSwitchVLANUpdateDto{EthernetSwitchVLANBaseDto: dtos.EthernetSwitchVLANBaseDto{
		UntaggedPorts: nil,
		TaggedPorts:   []uuid.UUID{vlanServiceTester.portID},
	}}
	vlan, err := vlanServiceTester.service.UpdateVLAN(context.Background(), vlanServiceTester.switchID, vlanServiceTester.vlanID, updDto)
	if err != nil {
		t.Errorf("failed to update switch vlan: %q", err)
	}
	if len(vlan.UntaggedPorts) != 0 || vlan.TaggedPorts[0] != vlanServiceTester.portID {
		t.Error("vlan update failed: wrong ports acquired")
	}
}

func Test_EthernetSwitchVLANSService_Delete(t *testing.T) {
	err := vlanServiceTester.service.DeleteVLAN(context.Background(), vlanServiceTester.switchID, vlanServiceTester.vlanID)
	if err != nil {
		t.Errorf("failed to delete switch vlan: %q", err)
	}
	_, err = vlanServiceTester.service.GetVLANByID(context.Background(), vlanServiceTester.switchID, vlanServiceTester.vlanID)
	if err == nil {
		t.Error("successfully get removed vlan")
	}
}

func Test_EthernetSwitchVLANSService_Create20(t *testing.T) {
	for i := 1; i <= 20; i++ {
		dto := dtos.EthernetSwitchVLANCreateDto{
			EthernetSwitchVLANBaseDto: dtos.EthernetSwitchVLANBaseDto{
				UntaggedPorts: []uuid.UUID{vlanServiceTester.portID},
				TaggedPorts:   nil,
			},
			VlanID: i,
		}
		_, err := vlanServiceTester.service.CreateVLAN(context.Background(), vlanServiceTester.switchID, dto)
		if err != nil {
			t.Errorf("failed to create switch vlan: %q", err)
		}
	}
}

func Test_EthernetSwitchVLANSService_GetList(t *testing.T) {
	vlans, err := vlanServiceTester.service.GetVLANs(context.Background(), vlanServiceTester.switchID, "", "VlanID", "desc", 1, 20)
	if err != nil {
		t.Errorf("failed to get switch vlan list: %q", err)
	}
	if len(vlans.Items) != 20 {
		t.Error("wrong vlan's count")
	}
	if vlans.Pagination.Size != 20 {
		t.Error("unexpected page size")
	}
	if vlans.Items[0].VlanID != 20 {
		t.Error("order by failed")
	}
}

func Test_EthernetSwitchVLANSService_RemoveDb(t *testing.T) {
	if err := os.Remove(vlanServiceTester.dbPath); err != nil {
		t.Errorf("remove db failed:  %s", err)
	}
}
