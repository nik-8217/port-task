package memory

import (
	"context"
	"fmt"
	"testing"

	"portservice/internal/domain"

	"github.com/stretchr/testify/assert"
)

func TestPortRepository_SaveAndGet(t *testing.T) {
	repo := NewPortRepository()
	ctx := context.Background()

	// Create a test port
	coords := []float64{55.5136433, 25.4052165}
	port, err := domain.NewPort("TEST1", "Test Port", "Test City", "Test Country", coords, "Test Province", "UTC", nil, "TEST")
	assert.NoError(t, err)

	// Test saving
	err = repo.SavePort(ctx, port)
	assert.NoError(t, err)

	// Test retrieval
	retrieved, err := repo.GetPort(ctx, "TEST1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, port.ID, retrieved.ID)
	assert.Equal(t, port.Name, retrieved.Name)

	// Test updating
	port.Name = "Updated Port"
	err = repo.SavePort(ctx, port)
	assert.NoError(t, err)

	// Test retrieving updated
	retrieved, err = repo.GetPort(ctx, "TEST1")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Port", retrieved.Name)

	// Test non-existent port
	retrieved, err = repo.GetPort(ctx, "NONEXISTENT")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestPortRepository_Statistics(t *testing.T) {
	repo := NewPortRepository()
	ctx := context.Background()

	// Create test ports
	coords := []float64{55.5136433, 25.4052165}
	port1, _ := domain.NewPort("TEST1", "Test Port 1", "Test City", "Test Country", coords, "", "", nil, "")
	port2, _ := domain.NewPort("TEST2", "Test Port 2", "Test City", "Test Country", coords, "", "", nil, "")

	// Initial stats should be zero
	stats := repo.GetStatistics()
	assert.Equal(t, int64(0), stats.TotalPorts)
	assert.Equal(t, int64(0), stats.TotalUpdates)

	// Save first port
	err := repo.SavePort(ctx, port1)
	assert.NoError(t, err)

	// Check stats after first save
	stats = repo.GetStatistics()
	assert.Equal(t, int64(1), stats.TotalPorts)
	assert.Equal(t, int64(1), stats.TotalUpdates)

	// Save second port
	err = repo.SavePort(ctx, port2)
	assert.NoError(t, err)

	// Check stats after second save
	stats = repo.GetStatistics()
	assert.Equal(t, int64(2), stats.TotalPorts)
	assert.Equal(t, int64(2), stats.TotalUpdates)

	// Update first port
	err = repo.SavePort(ctx, port1)
	assert.NoError(t, err)

	// Check stats after update
	stats = repo.GetStatistics()
	assert.Equal(t, int64(2), stats.TotalPorts)
	assert.Equal(t, int64(3), stats.TotalUpdates)
}

func TestPortRepository_ConcurrentAccess(t *testing.T) {
	repo := NewPortRepository()
	ctx := context.Background()
	coords := []float64{55.5136433, 25.4052165}

	// Create multiple ports
	ports := make([]*domain.Port, 100)
	for i := 0; i < 100; i++ {
		port, _ := domain.NewPort(
			fmt.Sprintf("TEST%d", i),
			fmt.Sprintf("Test Port %d", i),
			"Test City",
			"Test Country",
			coords,
			"",
			"",
			nil,
			"",
		)
		ports[i] = port
	}

	// Concurrently save ports
	done := make(chan bool)
	for _, p := range ports {
		go func(port *domain.Port) {
			err := repo.SavePort(ctx, port)
			assert.NoError(t, err)
			done <- true
		}(p)
	}

	// Wait for all goroutines to complete
	for i := 0; i < len(ports); i++ {
		<-done
	}

	// Verify all ports were saved
	stats := repo.GetStatistics()
	assert.Equal(t, int64(100), stats.TotalPorts)
	assert.Equal(t, int64(100), stats.TotalUpdates)

	// Verify each port can be retrieved
	for _, p := range ports {
		retrieved, err := repo.GetPort(ctx, p.ID)
		assert.NoError(t, err)
		assert.NotNil(t, retrieved)
		assert.Equal(t, p.ID, retrieved.ID)
	}
}

func TestPortRepository_ContextCancellation(t *testing.T) {
	repo := NewPortRepository()
	ctx, cancel := context.WithCancel(context.Background())
	coords := []float64{55.5136433, 25.4052165}
	port, _ := domain.NewPort("TEST1", "Test Port", "Test City", "Test Country", coords, "", "", nil, "")

	// Cancel context before operations
	cancel()

	// Attempt operations with canceled context
	err := repo.SavePort(ctx, port)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	_, err = repo.GetPort(ctx, "TEST1")
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)

	err = repo.Close(ctx)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestPortRepository_Close(t *testing.T) {
	repo := NewPortRepository()
	ctx := context.Background()
	coords := []float64{55.5136433, 25.4052165}
	port, _ := domain.NewPort("TEST1", "Test Port", "Test City", "Test Country", coords, "", "", nil, "")

	// Save a port
	err := repo.SavePort(ctx, port)
	assert.NoError(t, err)

	// Close repository
	err = repo.Close(ctx)
	assert.NoError(t, err)

	// Verify repository is empty after close
	retrieved, err := repo.GetPort(ctx, "TEST1")
	assert.NoError(t, err)
	assert.Nil(t, retrieved)
}
