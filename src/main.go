package main

import (
	"go.uber.org/fx"
	"rol/app/services"
	_ "rol/docs"
	"rol/infrastructure"
	"rol/webapi"
	"rol/webapi/controllers"
)

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
	TFTPServer := infrastructure.NewTFTPServer("192.168.88.254", "69", nil)
	TFTPServer.AddNewPathToTFTPRatio("MAC123123/test.txt", "files/test.txt")
	TFTPServer.Start()
	TFTPServer.AddNewPathToTFTPRatio("MAC322/test.txt", "files/test.txt")

	app := fx.New(
		fx.Provide(
			// Realizations
			infrastructure.NewYmlConfig,
			infrastructure.NewGormEntityDb,
			infrastructure.NewGormLogDb,
			infrastructure.NewEthernetSwitchRepository,
			infrastructure.NewHTTPLogRepository,
			infrastructure.NewAppLogRepository,
			infrastructure.NewLogrusLogger,
			// Application logic
			services.NewEthernetSwitchService,
			services.NewHTTPLogService,
			services.NewAppLogService,
			// WEB API -> Server
			webapi.NewGinHTTPServer,
			// WEB API -> Controllers
			controllers.NewEthernetSwitchGinController,
			controllers.NewHTTPLogGinController,
			controllers.NewAppLogGinController,
		),
		fx.Invoke(
			infrastructure.RegisterLogHooks,
			controllers.RegisterSwitchController,
			controllers.RegisterHTTPLogController,
			controllers.RegisterAppLogController,
			webapi.StartHTTPServer,
		),
	)
	app.Run()
}
