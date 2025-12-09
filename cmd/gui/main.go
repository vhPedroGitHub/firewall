package main

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"github.com/vhPedroGitHub/firewall/internal/app"
	"github.com/vhPedroGitHub/firewall/internal/config"
	"github.com/vhPedroGitHub/firewall/internal/logging"
	"github.com/vhPedroGitHub/firewall/internal/monitor"
	"github.com/vhPedroGitHub/firewall/internal/profiles"
	"github.com/vhPedroGitHub/firewall/internal/rules"
	"github.com/vhPedroGitHub/firewall/internal/stats"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Load config
	cfg, err := config.Load("firewall.json")
	if err != nil {
		log.Printf("Failed to load config, using defaults: %v", err)
		cfg = config.Default()
	}

	// Initialize sqlite store
	db, err := sql.Open("sqlite3", cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ruleStore, err := rules.NewSQLiteStore(db)
	if err != nil {
		log.Fatal(err)
	}

	profileStore, err := profiles.NewSQLiteStore(db)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize logging
	if err := logging.Init(cfg.LogPath); err != nil {
		log.Fatal(err)
	}
	defer logging.Close()

	// Create app service
	svc := &AppService{
		Service:      app.Service{Store: ruleStore},
		profileStore: profileStore,
	}

	// Create Wails application
	err = wails.Run(&options.App{
		Title:  "Firewall Manager",
		Width:  cfg.GUI.Width,
		Height: cfg.GUI.Height,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        svc.startup,
		Bind: []interface{}{
			svc,
		},
	})

	if err != nil {
		log.Fatal(err)
	}
}

// AppService wraps the core service for Wails binding.
type AppService struct {
	ctx          context.Context
	Service      app.Service
	profileStore profiles.Store
	monitorSvc   *monitor.Service
}

func (a *AppService) startup(ctx context.Context) {
	a.ctx = ctx
}

// Wails-exported methods for frontend

func (a *AppService) ListRules() ([]rules.Rule, error) {
	return a.Service.ListRules()
}

func (a *AppService) AddRule(r rules.Rule) error {
	return a.Service.SaveRule(r)
}

func (a *AppService) RemoveRule(name string) error {
	return a.Service.DeleteRule(name)
}

func (a *AppService) ApplyRule(r rules.Rule) error {
	return a.Service.ApplyRule(r)
}

func (a *AppService) ListProfiles() ([]profiles.Profile, error) {
	return a.profileStore.ListProfiles()
}

func (a *AppService) CreateProfile(p profiles.Profile) error {
	return a.profileStore.SaveProfile(p)
}

func (a *AppService) ActivateProfile(name string) error {
	return a.profileStore.SetActiveProfile(name)
}

func (a *AppService) GetStats() (map[string]int64, error) {
	return stats.Snapshot()
}

func (a *AppService) GetStatsFiltered(filter stats.Filter) []stats.ConnectionStat {
	return stats.Query(filter)
}

func (a *AppService) GetLogs(filepath string) ([]logging.Event, error) {
	return logging.ReadEvents(filepath)
}

func (a *AppService) StartMonitoring() error {
	if a.monitorSvc == nil {
		var err error
		a.monitorSvc, err = monitor.NewService(a.Service.Store)
		if err != nil {
			return fmt.Errorf("failed to create monitor service: %w", err)
		}
	}
	return a.monitorSvc.Start()
}

func (a *AppService) StopMonitoring() error {
	if a.monitorSvc == nil {
		return fmt.Errorf("monitor service not initialized")
	}
	return a.monitorSvc.Stop()
}

func (a *AppService) GetMonitoringStatus() bool {
	if a.monitorSvc == nil {
		return false
	}
	return a.monitorSvc.IsRunning()
}

func (a *AppService) GetMonitoringEvents() []monitor.ConnectionEventLog {
	if a.monitorSvc == nil {
		return []monitor.ConnectionEventLog{}
	}
	return a.monitorSvc.GetRecentEvents()
}

func (a *AppService) ClearMonitoringEvents() error {
	if a.monitorSvc == nil {
		return fmt.Errorf("monitor service not initialized")
	}
	a.monitorSvc.ClearEvents()
	return nil
}

func (a *AppService) EnablePrompts() error {
	if a.monitorSvc == nil {
		return fmt.Errorf("monitor service not initialized")
	}
	a.monitorSvc.EnablePrompts()
	return nil
}

func (a *AppService) DisablePrompts() error {
	if a.monitorSvc == nil {
		return fmt.Errorf("monitor service not initialized")
	}
	a.monitorSvc.DisablePrompts()
	return nil
}

func (a *AppService) PromptsEnabled() bool {
	if a.monitorSvc == nil {
		return false
	}
	return a.monitorSvc.PromptsEnabled()
}

func (a *AppService) GetActiveProcesses() []monitor.ConnectionEvent {
	if a.monitorSvc == nil {
		return []monitor.ConnectionEvent{}
	}
	return a.monitorSvc.GetActiveProcesses()
}

func (a *AppService) ClearActiveProcesses() error {
	if a.monitorSvc == nil {
		return fmt.Errorf("monitor service not initialized")
	}
	a.monitorSvc.ClearActiveProcesses()
	return nil
}

func (a *AppService) GetProcessTraffic() []monitor.ProcessTraffic {
	if a.monitorSvc == nil {
		return []monitor.ProcessTraffic{}
	}
	return a.monitorSvc.GetProcessTraffic()
}

func (a *AppService) ClearProcessTraffic() error {
	if a.monitorSvc == nil {
		return fmt.Errorf("monitor service not initialized")
	}
	a.monitorSvc.ClearProcessTraffic()
	return nil
}

// GetTopApplications returns the top N applications by data transferred.
func (a *AppService) GetTopApplications(n int) []struct {
	App        string
	BytesSent  int64
	BytesRecv  int64
	TotalBytes int64
} {
	return stats.GetTopApplications(n)
}

// ClearStats clears all statistics.
func (a *AppService) ClearStats() error {
	stats.Clear()
	return nil
}
