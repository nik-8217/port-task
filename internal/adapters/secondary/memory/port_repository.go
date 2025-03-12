package memory

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"portservice/internal/domain"
	"portservice/internal/ports/out"
)

// PortRepository implements out.PortRepository using an in-memory map
type PortRepository struct {
	ports map[string]*domain.Port
	mu    sync.RWMutex

	// Statistics
	totalPorts     atomic.Int64
	totalUpdates   atomic.Int64
	lastUpdateTime atomic.Int64
}

// NewPortRepository creates a new instance of PortRepository
func NewPortRepository() out.PortRepository {
	return &PortRepository{
		ports: make(map[string]*domain.Port),
	}
}

// SavePort saves or updates a port in the repository
func (r *PortRepository) SavePort(ctx context.Context, port *domain.Port) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		if err := port.Validate(); err != nil {
			return err
		}

		r.mu.Lock()
		defer r.mu.Unlock()

		_, exists := r.ports[port.ID]
		r.ports[port.ID] = port

		// Update statistics
		if !exists {
			r.totalPorts.Add(1)
		}
		r.totalUpdates.Add(1)
		r.lastUpdateTime.Store(time.Now().UnixNano())

		return nil
	}
}

// GetPort retrieves a port by its ID
func (r *PortRepository) GetPort(ctx context.Context, id string) (*domain.Port, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		r.mu.RLock()
		defer r.mu.RUnlock()

		if port, exists := r.ports[id]; exists {
			return port, nil
		}
		return nil, nil
	}
}

// GetStatistics returns current repository statistics
func (r *PortRepository) GetStatistics() out.RepositoryStats {
	lastUpdate := time.Unix(0, r.lastUpdateTime.Load())
	return out.RepositoryStats{
		TotalPorts:   r.totalPorts.Load(),
		TotalUpdates: r.totalUpdates.Load(),
		LastUpdate:   lastUpdate.Format(time.RFC3339),
	}
}

// Close implements the Close method for the repository
func (r *PortRepository) Close(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		r.mu.Lock()
		defer r.mu.Unlock()

		// Clear the map to free memory
		for k := range r.ports {
			delete(r.ports, k)
		}
		r.ports = nil
		return nil
	}
}
