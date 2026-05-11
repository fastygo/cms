package conformance

import (
	"context"
	"fmt"
	"time"
)

type Level int

const (
	LevelCore Level = iota
	LevelREST
	LevelAdmin
	LevelExtension
	LevelFull
)

type Case struct {
	ID       string
	Level    Level
	Profiles []string
	Run      func(context.Context) error
}

type Options struct {
	ContractVersion string
	Implementation  string
	Level           Level
	Profiles        []string
	KnownDeviations []string
	Now             func() time.Time
}

type Runner struct {
	options Options
	cases   []Case
}

type Report struct {
	ContractVersion string    `json:"contract_version"`
	Implementation  string    `json:"implementation"`
	Level           string    `json:"level"`
	Profiles        []string  `json:"profiles"`
	Passed          []string  `json:"passed"`
	Failed          []Failure `json:"failed"`
	Skipped         []Skipped `json:"skipped"`
	Warnings        []string  `json:"warnings"`
	KnownDeviations []string  `json:"known_deviations"`
	GeneratedAt     time.Time `json:"generated_at"`
}

type Failure struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type Skipped struct {
	ID     string `json:"id"`
	Reason string `json:"reason"`
}

func NewRunner(options Options, cases ...Case) Runner {
	if options.Now == nil {
		options.Now = time.Now
	}
	return Runner{
		options: options,
		cases:   append([]Case(nil), cases...),
	}
}

func (r Runner) Run(ctx context.Context) Report {
	report := Report{
		ContractVersion: defaultString(r.options.ContractVersion, "go-codex.0"),
		Implementation:  defaultString(r.options.Implementation, "GoCMS"),
		Level:           r.options.Level.String(),
		Profiles:        append([]string(nil), r.options.Profiles...),
		KnownDeviations: append([]string(nil), r.options.KnownDeviations...),
		GeneratedAt:     r.options.Now().UTC(),
	}
	for _, testCase := range r.cases {
		if testCase.ID == "" {
			report.Warnings = append(report.Warnings, "skipped unnamed conformance case")
			continue
		}
		if testCase.Level > r.options.Level {
			report.Skipped = append(report.Skipped, Skipped{ID: testCase.ID, Reason: "above declared compatibility level"})
			continue
		}
		if missing := missingProfiles(testCase.Profiles, r.options.Profiles); len(missing) > 0 {
			report.Skipped = append(report.Skipped, Skipped{ID: testCase.ID, Reason: fmt.Sprintf("missing profiles: %v", missing)})
			continue
		}
		if testCase.Run == nil {
			report.Failed = append(report.Failed, Failure{ID: testCase.ID, Message: "case has no runner"})
			continue
		}
		if err := testCase.Run(ctx); err != nil {
			report.Failed = append(report.Failed, Failure{ID: testCase.ID, Message: err.Error()})
			continue
		}
		report.Passed = append(report.Passed, testCase.ID)
	}
	return report
}

func (l Level) String() string {
	switch l {
	case LevelCore:
		return "core"
	case LevelREST:
		return "rest"
	case LevelAdmin:
		return "admin"
	case LevelExtension:
		return "extension"
	case LevelFull:
		return "full"
	default:
		return "unknown"
	}
}

func missingProfiles(required []string, enabled []string) []string {
	enabledSet := make(map[string]struct{}, len(enabled))
	for _, profile := range enabled {
		enabledSet[profile] = struct{}{}
	}
	missing := make([]string, 0)
	for _, profile := range required {
		if _, ok := enabledSet[profile]; !ok {
			missing = append(missing, profile)
		}
	}
	return missing
}

func defaultString(value string, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}
