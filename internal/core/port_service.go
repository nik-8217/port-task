package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"

	"portservice/internal/domain"
	"portservice/internal/ports/in"
	"portservice/internal/ports/out"
)

// portService implements in.PortService
type portService struct {
	repository out.PortRepository
}

// NewPortService creates a new instance of portService
func NewPortService(repository out.PortRepository) in.PortService {
	return &portService{
		repository: repository,
	}
}

// CreateOrUpdatePort creates a new port or updates an existing one
func (s *portService) CreateOrUpdatePort(ctx context.Context, port *domain.Port) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	if port == nil {
		return fmt.Errorf("nil port")
	}
	if err := port.Validate(); err != nil {
		return fmt.Errorf("invalid port: %w", err)
	}
	return s.repository.SavePort(ctx, port)
}

// GetPort retrieves a port by its ID
func (s *portService) GetPort(ctx context.Context, id string) (*domain.Port, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if id == "" {
		return nil, fmt.Errorf("empty port ID")
	}
	return s.repository.GetPort(ctx, id)
}

// ProcessPortsFile processes a JSON file containing port data
func (s *portService) ProcessPortsFile(ctx context.Context, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)

	// Read opening brace
	if _, err := decoder.Token(); err != nil {
		return fmt.Errorf("failed to read JSON start: %w", err)
	}

	// Read key-value pairs
	for decoder.More() {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Read port ID (key)
			portID, err := decoder.Token()
			if err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed to read port ID: %w", err)
			}

			// Read port data into map
			var portData map[string]interface{}
			if err := decoder.Decode(&portData); err != nil {
				return fmt.Errorf("failed to decode port data: %w", err)
			}

			// Extract and validate required fields
			name, ok := portData["name"].(string)
			if !ok {
				log.Printf("Warning: port %v has invalid or missing name", portID)
				continue
			}

			// Extract optional fields with defaults
			city := getStringOrDefault(portData, "city", "")
			country := getStringOrDefault(portData, "country", "")
			province := getStringOrDefault(portData, "province", "")
			timezone := getStringOrDefault(portData, "timezone", "")
			code := getStringOrDefault(portData, "code", "")

			// Extract and validate coordinates
			var coordinates []float64
			if coords, ok := portData["coordinates"].([]interface{}); ok && len(coords) == 2 {
				lon, lonOk := coords[0].(float64)
				lat, latOk := coords[1].(float64)
				if !lonOk || !latOk {
					log.Printf("Warning: port %v has invalid coordinate types", portID)
					continue
				}
				if lon < -180 || lon > 180 {
					log.Printf("Warning: port %v has invalid longitude: %v", portID, lon)
					continue
				}
				if lat < -90 || lat > 90 {
					log.Printf("Warning: port %v has invalid latitude: %v", portID, lat)
					continue
				}
				coordinates = []float64{lon, lat}
			} else {
				log.Printf("Warning: port %v has invalid coordinates format", portID)
				continue
			}

			// Extract unlocs
			unlocs := make([]string, 0)
			if unlocsRaw, ok := portData["unlocs"].([]interface{}); ok {
				for _, u := range unlocsRaw {
					if s, ok := u.(string); ok {
						unlocs = append(unlocs, s)
					}
				}
			}

			// Create port entity
			port, err := domain.NewPort(
				portID.(string),
				name,
				city,
				country,
				coordinates,
				province,
				timezone,
				unlocs,
				code,
			)
			if err != nil {
				log.Printf("Warning: failed to create port entity for %v: %v", portID, err)
				continue
			}

			// Save port
			if err := s.repository.SavePort(ctx, port); err != nil {
				return fmt.Errorf("failed to save port %v: %w", portID, err)
			}
		}
	}

	return nil
}

// getStringOrDefault safely extracts a string value from a map with a default value
func getStringOrDefault(data map[string]interface{}, key, defaultValue string) string {
	if val, ok := data[key].(string); ok {
		return val
	}
	return defaultValue
}
