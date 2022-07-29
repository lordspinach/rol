package main

import (
	"path/filepath"
	"rol/app/services"
	"rol/domain"
	"rol/infrastructure"
	"rol/webapi"
	"rol/webapi/controllers"
	"runtime"

	_ "rol/webapi/swagger"

	"go.uber.org/fx"
)

func SetRootPath() domain.GlobalDIParameters {
	_, filePath, _, _ := runtime.Caller(0)
	return domain.GlobalDIParameters{RootPath: filepath.Dir(filePath)}
}

// @title Rack of labs API
// @version version(1.0)
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
	app := fx.New(
		fx.Provide(
			// Domains
			SetRootPath,
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
			infrastructure.NewHostNetworkManager,
			// Application logic
			services.NewEthernetSwitchService,
			services.NewHTTPLogService,
			services.NewAppLogService,
			services.NewEthernetSwitchPortService,
			services.NewDeviceTemplateService,
			services.NewHostNetworkVlanService,
			// WEB API -> Server
			webapi.NewGinHTTPServer,
			// WEB API -> Controllers
			controllers.NewEthernetSwitchGinController,
			controllers.NewHTTPLogGinController,
			controllers.NewAppLogGinController,
			controllers.NewEthernetSwitchPortGinController,
			controllers.NewDeviceTemplateController,
			controllers.NewHostNetworkVlanController,
		),
		fx.Invoke(
			infrastructure.RegisterLogHooks,
			controllers.RegisterEthernetSwitchController,
			controllers.RegisterHTTPLogController,
			controllers.RegisterAppLogController,
			controllers.RegisterEthernetSwitchPortController,
			controllers.RegisterDeviceTemplateController,
			controllers.RegisterHostNetworkVlanController,
			webapi.StartHTTPServer,
		),
	)
	app.Run()
}
