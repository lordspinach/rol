package main

import (
	"os"
	"path/filepath"
	"rol/app/services"
	"rol/domain"
	"rol/infrastructure"
	"rol/webapi"
	"rol/webapi/controllers"
	_ "rol/webapi/swagger"

	"go.uber.org/fx"
)

//GetGlobalDIParameters get global parameters for DI
func GetGlobalDIParameters() domain.GlobalDIParameters {
	filePath, _ := os.Executable()
	return domain.GlobalDIParameters{
		RootPath: filepath.Dir(filePath),
	}
}

// @title Rack of labs API
// @version 0.1.0
// @description Description of specifications
// @Precautions when using termsOfService specifications

// @contact.name API supporter
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name license(Mandatory)
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api/v1/
func main() {
	//str := infrastructure.NewYamlHostNetworkConfigStorage(GetGlobalDIParameters())
	//mn, _ := infrastructure.NewHostNetworkManager(str)
	//name, _ := mn.CreateVlan("enp0s3", 5)
	//ip, ipNet, _ := net.ParseCIDR("1.1.1.5/24")
	//ipNet.IP = ip
	//mn.AddrAdd(name, *ipNet)
	//mn.SaveConfiguration()

	//f := func(addr net.UDPAddr, virtualPath, actualPath string) error {
	//	fmt.Println("im in callback")
	//	fmt.Println(addr.String())
	//	fmt.Println("virtual path: " + virtualPath)
	//	fmt.Println("actual path: " + actualPath)
	//	return nil
	//}
	//TFTPServer := infrastructure.NewTFTPServer("1.1.1.5", "69", nil)
	//TFTPServer.AddCallbackFunc(f)
	//TFTPServer.AddNewPathToTFTPRatio("MAC123123/test.txt", "files/test.txt")
	//TFTPServer.Start()
	//TFTPServer.AddNewPathToTFTPRatio("MAC322/test.txt", "files/test.txt")
	//id := uuid.New()

	//cfg1 := domain.TFTPConfig{
	//	Entity: domain.Entity{
	//		ID:        id,
	//		CreatedAt: time.Time{},
	//		UpdatedAt: time.Time{},
	//		DeletedAt: gorm.DeletedAt{},
	//	},
	//	Address: "192.168.1.2",
	//	Port:    "69",
	//	Enabled: true,
	//}
	//cfg2 := domain.TFTPConfig{
	//	Entity: domain.Entity{
	//		ID:        id,
	//		CreatedAt: time.Time{},
	//		UpdatedAt: time.Time{},
	//		DeletedAt: gorm.DeletedAt{},
	//	},
	//	Address: "192.168.10.2",
	//	Port:    "69",
	//	Enabled: true,
	//}
	//paths := []domain.TFTPPathRatio{{
	//	TFTPServerID: id,
	//	ActualPath:   "files/test.txt",
	//	VirtualPath:  "MAC123123/test.txt",
	//}, {
	//	TFTPServerID: id,
	//	ActualPath:   "files/xtx.txt",
	//	VirtualPath:  "MAC322/test.txt",
	//}}
	//mng := infrastructure.NewTFTPServerManager(nil)
	//mng.CreateTFTPServer(cfg)
	//mng.StartTFTPServer(id)
	//mng.UpdatePaths(id, paths)
	//time.Sleep(time.Second / 2)
	//e := mng.ServerIsRunning(id)

	//fmt.Print(e)

	app := fx.New(
		fx.Provide(
			// Core
			GetGlobalDIParameters,
			// Realizations
			infrastructure.NewYmlConfig,
			infrastructure.NewGormEntityDb,
			infrastructure.NewGormLogDb,
			infrastructure.NewEthernetSwitchRepository,
			infrastructure.NewHTTPLogRepository,
			infrastructure.NewAppLogRepository,
			infrastructure.NewLogrusLogger,
			infrastructure.NewEthernetSwitchPortRepository,
			infrastructure.NewDeviceTemplateStorage,
			infrastructure.NewYamlHostNetworkConfigStorage,
			infrastructure.NewHostNetworkManager,
			infrastructure.NewTFTPServerManager,
			infrastructure.NewTFTPConfigRepository,
			infrastructure.NewTFTPPathRatioRepository,
			// Application logic
			services.NewEthernetSwitchService,
			services.NewHTTPLogService,
			services.NewAppLogService,
			services.NewDeviceTemplateService,
			services.NewHostNetworkVlanService,
			services.NewTFTPServerService,
			// WEB API -> GIN Server
			webapi.NewGinHTTPServer,
			// WEB API -> GIN Controllers
			controllers.NewEthernetSwitchGinController,
			controllers.NewHTTPLogGinController,
			controllers.NewAppLogGinController,
			controllers.NewEthernetSwitchPortGinController,
			controllers.NewDeviceTemplateController,
			controllers.NewHostNetworkVlanController,
			controllers.NewTFTPServerGinController,
		),
		fx.Invoke(
			//Register logrus hooks
			infrastructure.RegisterLogHooks,
			//Services initialization
			services.EthernetSwitchServiceInit,
			services.TFTPServerServiceInitialize,
			//GIN Controllers registration
			controllers.RegisterEthernetSwitchController,
			controllers.RegisterHTTPLogController,
			controllers.RegisterAppLogController,
			controllers.RegisterEthernetSwitchPortController,
			controllers.RegisterDeviceTemplateController,
			controllers.RegisterHostNetworkVlanController,
			controllers.RegisterTFTPServerGinController,
			//Start GIN http server
			webapi.StartHTTPServer,
		),
	)
	app.Run()
}
