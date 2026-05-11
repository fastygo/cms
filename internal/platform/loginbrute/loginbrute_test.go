package loginbrute

import (
	"testing"
	"time"
)

func TestLockedOut_requiresMaxFailuresWithinWindow(t *testing.T) {
	now := time.Date(2026, 5, 10, 12, 0, 0, 0, time.UTC)
	policy := Policy{MaxAttempts: 3, AttemptWindow: time.Hour, LockoutWindow: time.Hour}
	start := now.Add(-policy.AttemptWindow)

	attempts := []Attempt{
		{Success: false, CreatedAt: now.Add(-10 * time.Minute)},
		{Success: false, CreatedAt: now.Add(-9 * time.Minute)},
	}
	if LockedOut(now, policy, start, attempts) {
		t.Fatal("expected not locked after 2 failures")
	}

	attempts = append(attempts, Attempt{Success: false, CreatedAt: now.Add(-8 * time.Minute)})
	if !LockedOut(now, policy, start, attempts) {
		t.Fatal("expected locked after 3 failures")
	}
}

func TestLockedOut_expiresAfterLockoutWindow(t *testing.T) {
	now := time.Date(2026, 5, 11, 13, 0, 0, 0, time.UTC)
	policy := Policy{MaxAttempts: 3, AttemptWindow: 48 * time.Hour, LockoutWindow: 24 * time.Hour}
	start := now.Add(-policy.AttemptWindow)
	attempts := []Attempt{
		{Success: false, CreatedAt: now.Add(-50 * time.Hour)},
		{Success: false, CreatedAt: now.Add(-49 * time.Hour)},
		{Success: false, CreatedAt: now.Add(-25 * time.Hour)}, // latest fail + lockout ends before now
	}
	if LockedOut(now, policy, start, attempts) {
		t.Fatal("expected lockout to have expired")
	}
}

func TestDefaultPolicy(t *testing.T) {
	p := DefaultPolicy()
	if p.MaxAttempts != 3 || p.AttemptWindow != 24*time.Hour || p.LockoutWindow != 24*time.Hour {
		t.Fatalf("DefaultPolicy() = %+v", p)
	}
}
