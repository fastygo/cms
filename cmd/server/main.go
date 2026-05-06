package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fastygo/cms/internal/infra/features/cms"
	"github.com/fastygo/cms/internal/infra/features/system"
	platformconfig "github.com/fastygo/cms/internal/platform/config"
	"github.com/fastygo/cms/internal/platform/logging"
	"github.com/fastygo/framework/pkg/app"
	"github.com/fastygo/framework/pkg/web/locale"
)

func main() {
	cfg, err := platformconfig.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	logger := logging.Configure(cfg.Framework.LogLevel, cfg.Framework.LogFormat)
	application, err := buildApp(cfg, logger)
	if err != nil {
		logger.Error("app build failed", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := application.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped", "error", err)
		os.Exit(1)
	}
}

func buildApp(cfg platformconfig.Config, logger *slog.Logger) (*app.App, error) {
	cmsModule, err := cms.NewWithOptions(cms.Options{
		DataSource:       cfg.Framework.DataSource,
		SessionKey:       cfg.Framework.SessionKey,
		SeedFixtures:     cfg.SeedFixtures,
		RuntimeProfile:   cfg.RuntimeProfile,
		StorageProfile:   cfg.StorageProfile,
		ActivePlugins:    cfg.ActivePlugins,
		SitePackageDir:   cfg.SitePackageDir,
		PlaygroundAuth:   cfg.PlaygroundAuth,
		BrowserStateless: cfg.BrowserStateless,
		EnableDevBearer:  cfg.EnableDevBearer,
		LoginPolicy:      cfg.LoginPolicy,
		AdminPolicy:      cfg.AdminPolicy,
		Preset:           cfg.Preset,
	})
	if err != nil {
		return nil, err
	}
	builder := app.New(cfg.Framework).
		WithLogger(logger).
		WithFeature(cmsModule).
		WithFeature(system.New()).
		WithHealthEndpoints(cfg.Framework.HealthLivePath, cfg.Framework.HealthReadyPath)

	builder = builder.WithLocales(app.LocalesConfig{
		Strategy: &locale.PathPrefixStrategy{
			Available:       cfg.Framework.AvailableLocales,
			Default:         cfg.Framework.DefaultLocale,
			RedirectMissing: false,
		},
		Cookie: locale.CookieOptions{
			Enabled: true,
			Name:    "lang",
		},
		SPA: true,
	})

	if cfg.Framework.MetricsPath != "" {
		builder = builder.WithMetricsEndpoint(cfg.Framework.MetricsPath)
	}

	return builder.Build(), nil
}
