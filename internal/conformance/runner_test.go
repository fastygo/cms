package conformance

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunnerReportsPassedFailedAndSkippedCases(t *testing.T) {
	runner := NewRunner(Options{
		ContractVersion: "go-codex.0.1",
		Implementation:  "GoCMS",
		Level:           LevelAdmin,
		Profiles:        []string{"graphql"},
		KnownDeviations: []string{"comments are not part of the current compatibility target"},
		Now: func() time.Time {
			return time.Date(2026, 5, 9, 10, 0, 0, 0, time.UTC)
		},
	},
		Case{ID: "core.content_visibility", Level: LevelCore, Run: func(context.Context) error { return nil }},
		Case{ID: "admin.login", Level: LevelAdmin, Run: func(context.Context) error { return errors.New("login failed") }},
		Case{ID: "extension.plugin", Level: LevelExtension, Run: func(context.Context) error { return nil }},
		Case{ID: "graphql.public_reads", Level: LevelAdmin, Profiles: []string{"graphql"}, Run: func(context.Context) error { return nil }},
		Case{ID: "playground.browser", Level: LevelAdmin, Profiles: []string{"playground"}, Run: func(context.Context) error { return nil }},
	)

	report := runner.Run(context.Background())

	if report.ContractVersion != "go-codex.0.1" || report.Implementation != "GoCMS" || report.Level != "admin" {
		t.Fatalf("unexpected report identity: %+v", report)
	}
	if len(report.Passed) != 2 {
		t.Fatalf("passed = %v, want core and graphql cases", report.Passed)
	}
	if len(report.Failed) != 1 || report.Failed[0].ID != "admin.login" {
		t.Fatalf("failed = %+v, want admin.login", report.Failed)
	}
	if len(report.Skipped) != 2 {
		t.Fatalf("skipped = %+v, want level and profile skips", report.Skipped)
	}
	if len(report.KnownDeviations) != 1 {
		t.Fatalf("known deviations not preserved: %+v", report.KnownDeviations)
	}
}
