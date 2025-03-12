package out

import (
	"context"

	"portservice/internal/domain"
)

// RepositoryStats holds statistics about the repository
type RepositoryStats struct {
	TotalPorts   int64
	TotalUpdates int64
	LastUpdate   string
}

// PortRepository defines the secondary port (output) for port persistence
type PortRepository interface {
	// SavePort saves or updates a port in the repository
	SavePort(ctx context.Context, port *domain.Port) error

	// GetPort retrieves a port by its ID
	GetPort(ctx context.Context, id string) (*domain.Port, error)

	// Close closes the repository and frees any resources
	Close(ctx context.Context) error

	// GetStatistics returns the repository statistics
	GetStatistics() RepositoryStats
}
