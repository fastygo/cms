// Package loginbrute holds a small, dependency-free brute-force lockout rule
// used by the authentication layer (failed password attempts before cooldown).
package loginbrute

import "time"

// Attempt is a minimal login outcome record for lockout evaluation.
type Attempt struct {
	Success   bool
	CreatedAt time.Time
}

// Policy configures how failures are counted and how long access stays blocked.
type Policy struct {
	MaxAttempts   int
	AttemptWindow time.Duration
	LockoutWindow time.Duration
}

// DefaultPolicy is the baseline site policy: 3 failures within the attempt window,
// then block until lockout window elapses from the most recent failure.
func DefaultPolicy() Policy {
	return Policy{
		MaxAttempts:   3,
		AttemptWindow: 24 * time.Hour,
		LockoutWindow: 24 * time.Hour,
	}
}

// LockedOut reports whether login should be rejected until the lockout expires.
// Only attempts with CreatedAt >= windowStart are considered (callers typically
// pass windowStart = now.Add(-policy.AttemptWindow)).
func LockedOut(now time.Time, policy Policy, windowStart time.Time, attempts []Attempt) bool {
	if policy.MaxAttempts <= 0 || policy.LockoutWindow <= 0 {
		return false
	}
	failures := 0
	var latestFail time.Time
	for _, item := range attempts {
		if item.CreatedAt.Before(windowStart) {
			continue
		}
		if item.Success {
			continue
		}
		failures++
		if item.CreatedAt.After(latestFail) {
			latestFail = item.CreatedAt
		}
	}
	if failures < policy.MaxAttempts || latestFail.IsZero() {
		return false
	}
	return latestFail.Add(policy.LockoutWindow).After(now)
}
