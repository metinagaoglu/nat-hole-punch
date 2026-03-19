package repositories

import (
	"testing"
)

func TestGetRepository(t *testing.T) {
	// Reset repository to nil for testing
	repository = nil

	repo := GetRepository()

	if repo == nil {
		t.Fatal("GetRepository() returned nil")
	}

	// Test singleton pattern - should return the same instance
	repo2 := GetRepository()

	if repo != repo2 {
		t.Error("GetRepository() should return the same instance (singleton pattern)")
	}
}

func TestGetRepository_ReturnsIRepository(t *testing.T) {
	repository = nil

	repo := GetRepository()

	// Type assertion to verify it implements IRepository
	_, ok := repo.(IRepository)
	if !ok {
		t.Error("GetRepository() should return an object implementing IRepository")
	}
}
