package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCoordinate(t *testing.T) {
	tests := []struct {
		name      string
		longitude float64
		latitude  float64
		wantErr   bool
	}{
		{
			name:      "valid coordinate",
			longitude: 55.5136433,
			latitude:  25.4052165,
			wantErr:   false,
		},
		{
			name:      "invalid longitude",
			longitude: 181.0,
			latitude:  25.4052165,
			wantErr:   true,
		},
		{
			name:      "invalid latitude",
			longitude: 55.5136433,
			latitude:  91.0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			coord, err := NewCoordinate(tt.longitude, tt.latitude)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, coord)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, coord)
				assert.Equal(t, tt.longitude, coord.Longitude)
				assert.Equal(t, tt.latitude, coord.Latitude)
			}
		})
	}
}

func TestNewPort(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		portName string
		city     string
		country  string
		coords   []float64
		wantErr  bool
	}{
		{
			name:     "valid port",
			id:       "AEAJM",
			portName: "Ajman",
			city:     "Ajman",
			country:  "United Arab Emirates",
			coords:   []float64{55.5136433, 25.4052165},
			wantErr:  false,
		},
		{
			name:     "empty id",
			id:       "",
			portName: "Ajman",
			city:     "Ajman",
			country:  "United Arab Emirates",
			coords:   []float64{55.5136433, 25.4052165},
			wantErr:  true,
		},
		{
			name:     "empty name",
			id:       "AEAJM",
			portName: "",
			city:     "Ajman",
			country:  "United Arab Emirates",
			coords:   []float64{55.5136433, 25.4052165},
			wantErr:  true,
		},
		{
			name:     "invalid coordinates",
			id:       "AEAJM",
			portName: "Ajman",
			city:     "Ajman",
			country:  "United Arab Emirates",
			coords:   []float64{181.0, 25.4052165},
			wantErr:  true,
		},
		{
			name:     "missing coordinates",
			id:       "AEAJM",
			portName: "Ajman",
			city:     "Ajman",
			country:  "United Arab Emirates",
			coords:   []float64{},
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			port, err := NewPort(tt.id, tt.portName, tt.city, tt.country, tt.coords, "", "", nil, "")
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, port)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, port)
				assert.Equal(t, tt.id, port.ID)
				assert.Equal(t, tt.portName, port.Name)
				assert.Equal(t, tt.city, port.City)
				assert.Equal(t, tt.country, port.Country)
				assert.NotNil(t, port.Coordinates)
				assert.Equal(t, tt.coords[0], port.Coordinates.Longitude)
				assert.Equal(t, tt.coords[1], port.Coordinates.Latitude)
			}
		})
	}
}

func TestPort_Validate(t *testing.T) {
	validCoord, _ := NewCoordinate(55.5136433, 25.4052165)

	tests := []struct {
		name    string
		port    *Port
		wantErr bool
	}{
		{
			name: "valid port",
			port: &Port{
				ID:          "AEAJM",
				Name:        "Ajman",
				City:        "Ajman",
				Country:     "United Arab Emirates",
				Coordinates: validCoord,
			},
			wantErr: false,
		},
		{
			name: "empty id",
			port: &Port{
				ID:          "",
				Name:        "Ajman",
				City:        "Ajman",
				Country:     "United Arab Emirates",
				Coordinates: validCoord,
			},
			wantErr: true,
		},
		{
			name: "empty name",
			port: &Port{
				ID:          "AEAJM",
				Name:        "",
				City:        "Ajman",
				Country:     "United Arab Emirates",
				Coordinates: validCoord,
			},
			wantErr: true,
		},
		{
			name: "nil coordinates",
			port: &Port{
				ID:          "AEAJM",
				Name:        "Ajman",
				City:        "Ajman",
				Country:     "United Arab Emirates",
				Coordinates: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.port.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestPort_String(t *testing.T) {
	coords := []float64{55.5136433, 25.4052165}
	port, err := NewPort("TEST1", "Test Port", "Test City", "Test Country", coords, "Test Province", "UTC", []string{"TEST1"}, "TEST")
	assert.NoError(t, err)

	str := port.String()
	assert.Contains(t, str, "TEST1")
	assert.Contains(t, str, "Test Port")
	assert.Contains(t, str, "Test City")
	assert.Contains(t, str, "55.513643")
	assert.Contains(t, str, "25.405217")
}

func TestCoordinate_String(t *testing.T) {
	coord := &Coordinate{
		Longitude: 55.5136433,
		Latitude:  25.4052165,
	}

	str := coord.String()
	assert.Contains(t, str, "55.513643")
	assert.Contains(t, str, "25.405217")
}

func TestPort_Equal(t *testing.T) {
	coords1 := []float64{55.5136433, 25.4052165}
	port1, err := NewPort("TEST1", "Test Port", "Test City", "Test Country", coords1, "Test Province", "UTC", []string{"TEST1"}, "TEST")
	assert.NoError(t, err)

	// Same port with different object
	coords2 := []float64{55.5136433, 25.4052165}
	port2, err := NewPort("TEST1", "Test Port", "Test City", "Test Country", coords2, "Test Province", "UTC", []string{"TEST1"}, "TEST")
	assert.NoError(t, err)

	// Same port with updated name
	coords3 := []float64{55.5136433, 25.4052165}
	port3, err := NewPort("TEST1", "Updated Port", "Test City", "Test Country", coords3, "Test Province", "UTC", []string{"TEST1"}, "TEST")
	assert.NoError(t, err)

	// Different port
	coords4 := []float64{54.37, 24.47}
	port4, err := NewPort("TEST2", "Other Port", "Other City", "Other Country", coords4, "Other Province", "UTC", []string{"TEST2"}, "TEST")
	assert.NoError(t, err)

	assert.Equal(t, port1.ID, port2.ID)
	assert.Equal(t, port1.Coordinates, port2.Coordinates)
	assert.Equal(t, port1.Name, port2.Name)

	assert.Equal(t, port1.ID, port3.ID)
	assert.NotEqual(t, port1.Name, port3.Name)

	assert.NotEqual(t, port1.ID, port4.ID)
	assert.NotEqual(t, port1.Coordinates, port4.Coordinates)
}
