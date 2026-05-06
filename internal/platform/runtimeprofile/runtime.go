package runtimeprofile

import "fmt"

type RuntimeProfile string

type StorageProfile string

const (
	RuntimeProfileHeadless    RuntimeProfile = "headless"
	RuntimeProfileAdmin       RuntimeProfile = "admin"
	RuntimeProfilePlayground  RuntimeProfile = "playground"
	RuntimeProfileFull        RuntimeProfile = "full"
	RuntimeProfileConformance RuntimeProfile = "conformance"

	DefaultRuntimeProfile = RuntimeProfileFull

	StorageProfileBrowserIndexedDB StorageProfile = "browser-indexeddb"
	StorageProfileMemory           StorageProfile = "memory"
	StorageProfileJSONFixtures     StorageProfile = "json-fixtures"
	StorageProfileBbolt            StorageProfile = "bbolt"
	StorageProfileSQLite           StorageProfile = "sqlite"
	StorageProfileMySQL            StorageProfile = "mysql"
	StorageProfilePostgres         StorageProfile = "postgres"

	DefaultStorageProfile = StorageProfileSQLite
)

func IsRuntimeProfile(value string) bool {
	return parseRuntimeProfile(value) != ""
}

func IsStorageProfile(value string) bool {
	return parseStorageProfile(value) != ""
}

func parseRuntimeProfile(value string) RuntimeProfile {
	switch RuntimeProfile(value) {
	case RuntimeProfileHeadless, RuntimeProfileAdmin, RuntimeProfilePlayground, RuntimeProfileFull, RuntimeProfileConformance:
		return RuntimeProfile(value)
	default:
		return ""
	}
}

func parseStorageProfile(value string) StorageProfile {
	switch StorageProfile(value) {
	case StorageProfileBrowserIndexedDB, StorageProfileMemory, StorageProfileJSONFixtures, StorageProfileBbolt, StorageProfileSQLite, StorageProfileMySQL, StorageProfilePostgres:
		return StorageProfile(value)
	default:
		return ""
	}
}

func ValidateRuntimeProfile(value string) error {
	if value == "" {
		return nil
	}
	if parseRuntimeProfile(value) == "" {
		return fmt.Errorf("unsupported runtime profile: %q", value)
	}
	return nil
}

func ValidateStorageProfile(value string) error {
	if value == "" {
		return nil
	}
	if parseStorageProfile(value) == "" {
		return fmt.Errorf("unsupported storage profile: %q", value)
	}
	return nil
}
