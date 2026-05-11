package runtimeprofile

import "testing"

func TestRuntimeProfileValidation(t *testing.T) {
	if err := ValidateRuntimeProfile(string(RuntimeProfileHeadless)); err != nil {
		t.Fatalf("valid runtime profile should pass: %v", err)
	}
	if err := ValidateRuntimeProfile("bad"); err == nil {
		t.Fatalf("invalid runtime profile should fail")
	}
	if !IsRuntimeProfile(string(RuntimeProfilePlayground)) {
		t.Fatalf("expected playground to be recognized")
	}
	if IsRuntimeProfile("bad") {
		t.Fatalf("expected bad runtime profile to be rejected")
	}
}

func TestStorageProfileValidation(t *testing.T) {
	if err := ValidateStorageProfile(string(StorageProfileSQLite)); err != nil {
		t.Fatalf("valid storage profile should pass: %v", err)
	}
	if err := ValidateStorageProfile("bad"); err == nil {
		t.Fatalf("invalid storage profile should fail")
	}
	if !IsStorageProfile(string(StorageProfileBrowserIndexedDB)) {
		t.Fatalf("expected browser-indexeddb to be recognized")
	}
	if IsStorageProfile("bad") {
		t.Fatalf("expected bad storage profile to be rejected")
	}
}

func TestDeploymentProfileValidation(t *testing.T) {
	if err := ValidateDeploymentProfile(string(DeploymentProfileServerless)); err != nil {
		t.Fatalf("valid deployment profile should pass: %v", err)
	}
	if err := ValidateDeploymentProfile("bad"); err == nil {
		t.Fatalf("invalid deployment profile should fail")
	}
	if !IsDeploymentProfile(string(DeploymentProfileBrowser)) {
		t.Fatalf("expected browser to be recognized")
	}
	if IsDeploymentProfile("bad") {
		t.Fatalf("expected bad deployment profile to be rejected")
	}
}
