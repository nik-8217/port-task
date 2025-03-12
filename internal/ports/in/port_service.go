package in

import (
	"context"

	"portservice/internal/domain"
)

// PortService defines the primary port (input) for port operations
type PortService interface {
	// CreateOrUpdatePort creates a new port or updates an existing one
	CreateOrUpdatePort(ctx context.Context, port *domain.Port) error

	// GetPort retrieves a port by its ID
	GetPort(ctx context.Context, id string) (*domain.Port, error)

	// ProcessPortsFile processes a JSON file containing port data
	ProcessPortsFile(ctx context.Context, filePath string) error
}
