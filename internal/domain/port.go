package domain

import (
	"errors"
	"fmt"
)

// Coordinate represents a geographical coordinate
type Coordinate struct {
	Longitude float64
	Latitude  float64
}

// NewCoordinate creates a new coordinate with validation
func NewCoordinate(longitude, latitude float64) (*Coordinate, error) {
	if longitude < -180 || longitude > 180 {
		return nil, errors.New("longitude must be between -180 and 180")
	}
	if latitude < -90 || latitude > 90 {
		return nil, errors.New("latitude must be between -90 and 90")
	}
	return &Coordinate{
		Longitude: longitude,
		Latitude:  latitude,
	}, nil
}

// String returns a string representation of the coordinate
func (c *Coordinate) String() string {
	return fmt.Sprintf("[%.6f, %.6f]", c.Longitude, c.Latitude)
}

// Port represents a port entity in the domain
type Port struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	City        string      `json:"city"`
	Country     string      `json:"country"`
	Coordinates *Coordinate `json:"-"`
	Province    string      `json:"province"`
	Timezone    string      `json:"timezone"`
	Unlocs      []string    `json:"unlocs"`
	Code        string      `json:"code"`
}

// NewPort creates a new Port with validation
func NewPort(id, name, city, country string, coords []float64, province, timezone string, unlocs []string, code string) (*Port, error) {
	if id == "" {
		return nil, errors.New("port ID cannot be empty")
	}
	if name == "" {
		return nil, errors.New("port name cannot be empty")
	}
	if len(coords) != 2 {
		return nil, errors.New("coordinates must contain exactly longitude and latitude")
	}

	coordinate, err := NewCoordinate(coords[0], coords[1])
	if err != nil {
		return nil, fmt.Errorf("invalid coordinates: %w", err)
	}

	return &Port{
		ID:          id,
		Name:        name,
		City:        city,
		Country:     country,
		Coordinates: coordinate,
		Province:    province,
		Timezone:    timezone,
		Unlocs:      unlocs,
		Code:        code,
	}, nil
}

// Validate performs domain validation on the port
func (p *Port) Validate() error {
	if p.ID == "" {
		return errors.New("port ID cannot be empty")
	}
	if p.Name == "" {
		return errors.New("port name cannot be empty")
	}
	if p.Coordinates == nil {
		return errors.New("port must have coordinates")
	}
	return nil
}

// String returns a string representation of the port
func (p *Port) String() string {
	return fmt.Sprintf("Port{ID: %s, Name: %s, Location: %s, %s}",
		p.ID, p.Name, p.City, p.Coordinates)
}

// PortRepository defines the interface for port data persistence
type PortRepository interface {
	// SavePort saves or updates a port in the repository
	SavePort(port *Port) error

	// GetPort retrieves a port by its ID
	GetPort(id string) (*Port, error)
}

// PortService defines the interface for port business operations
type PortService interface {
	// CreateOrUpdatePort creates a new port or updates an existing one
	CreateOrUpdatePort(port *Port) error

	// GetPort retrieves a port by its ID
	GetPort(id string) (*Port, error)

	// ProcessPortsFile processes a JSON file containing port data
	ProcessPortsFile(filePath string) error
}
