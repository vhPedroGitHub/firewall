package main

import (
	"context"
	"database/sql"
	"embed"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"firewall/internal/app"
	"firewall/internal/config"
	"firewall/internal/logging"
	"firewall/internal/profiles"
	"firewall/internal/rules"
	"firewall/internal/stats"
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
