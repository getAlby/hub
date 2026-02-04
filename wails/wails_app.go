package wails

import (
	"context"
	"embed"
	"fmt"

	"github.com/getAlby/hub/api"
	"github.com/getAlby/hub/apps"
	"github.com/getAlby/hub/logger"
	"github.com/getAlby/hub/service"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/mac"
	"github.com/wailsapp/wails/v2/pkg/runtime"
	"gorm.io/gorm"
)

type WailsApp struct {
	ctx     context.Context
	svc     service.Service
	api     api.API
	db      *gorm.DB
	appsSvc apps.AppsService
}

func NewApp(svc service.Service) *WailsApp {
	return &WailsApp{
		svc:     svc,
		api:     api.NewAPI(svc, svc.GetDB(), svc.GetConfig(), svc.GetKeys(), svc.GetAlbySvc(), svc.GetAlbyOAuthSvc(), svc.GetEventPublisher()),
		db:      svc.GetDB(),
		appsSvc: apps.NewAppsService(svc.GetDB(), svc.GetEventPublisher(), svc.GetKeys(), svc.GetConfig()),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (app *WailsApp) startup(ctx context.Context) {
	app.ctx = ctx
}

func (app *WailsApp) onBeforeClose(ctx context.Context) bool {
	response, err := runtime.MessageDialog(ctx, runtime.MessageDialogOptions{
		Type:          runtime.QuestionDialog,
		Title:         "Confirm Exit",
		Message:       "Are you sure you want to shut down Alby Hub? Alby Hub needs to stay online to send and receive transactions. Channels may be closed if your hub stays offline for an extended period of time.",
		Buttons:       []string{"Yes", "No"},
		DefaultButton: "No",
	})
	if err != nil {
		logger.Logger.WithError(err).Error("failed to show confirmation dialog")
		return false
	}
	return response != "Yes"
}

func LaunchWailsApp(app *WailsApp, assets embed.FS, appIcon []byte) {
	err := wails.Run(&options.App{
		Title:  "Alby Hub",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		Logger: NewWailsLogger(),
		// HideWindowOnClose: true, // with this on, there is no way to close the app - wait for v3

		//BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: app.startup,
		OnBeforeClose: func(ctx context.Context) bool {
			return app.onBeforeClose(ctx)
		},
		Bind: []interface{}{
			app,
		},
		Mac: &mac.Options{
			About: &mac.AboutInfo{
				Title: "Alby Hub",
				Icon:  appIcon,
			},
		},
		Linux: &linux.Options{
			Icon: appIcon,
		},
	})

	if err != nil {
		logger.Logger.WithError(err).Error("failed to run Wails app")
	}
}

func (app *WailsApp) CheckShutdown() error {
    if app.svc.IsShuttingDown() {
        return fmt.Errorf("node is shutting down, please wait")
    }
    return nil
}

func NewWailsLogger() WailsLogger {
	return WailsLogger{}
}

type WailsLogger struct {
}

func (wailsLogger WailsLogger) Print(message string) {
	logger.Logger.WithField("wails", true).Print(message)
}

func (wailsLogger WailsLogger) Trace(message string) {
	logger.Logger.WithField("wails", true).Trace(message)
}

func (wailsLogger WailsLogger) Debug(message string) {
	logger.Logger.WithField("wails", true).Debug(message)
}

func (wailsLogger WailsLogger) Info(message string) {
	logger.Logger.WithField("wails", true).Info(message)
}

func (wailsLogger WailsLogger) Warning(message string) {
	logger.Logger.WithField("wails", true).Warning(message)
}

func (wailsLogger WailsLogger) Error(message string) {
	logger.Logger.WithField("wails", true).Error(message)
}

func (wailsLogger WailsLogger) Fatal(message string) {
	logger.Logger.WithField("wails", true).Fatal(message)
}
