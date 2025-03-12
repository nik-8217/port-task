package core

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"

	"portservice/internal/domain"
	"portservice/internal/ports/out"

	"github.com/stretchr/testify/assert"
)

func TestPortService_CreateAndGet(t *testing.T) {
	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// Create a test port
	coords := []float64{55.5136433, 25.4052165}
	port, err := domain.NewPort("TEST1", "Test Port", "Test City", "Test Country", coords, "Test Province", "UTC", nil, "TEST")
	assert.NoError(t, err)

	// Test creating
	err = service.CreateOrUpdatePort(ctx, port)
	assert.NoError(t, err)

	// Test retrieval
	retrieved, err := service.GetPort(ctx, "TEST1")
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, port.ID, retrieved.ID)
	assert.Equal(t, port.Name, retrieved.Name)

	// Test updating
	port.Name = "Updated Port"
	err = service.CreateOrUpdatePort(ctx, port)
	assert.NoError(t, err)

	// Test retrieving updated
	retrieved, err = service.GetPort(ctx, "TEST1")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Port", retrieved.Name)
}

func TestPortService_ProcessFile(t *testing.T) {
	// Create a temporary test file
	content := `{
		"AEAJM": {
			"name": "Ajman",
			"city": "Ajman",
			"country": "United Arab Emirates",
			"coordinates": [55.5136433, 25.4052165],
			"province": "Ajman",
			"timezone": "Asia/Dubai",
			"unlocs": ["AEAJM"],
			"code": "52000"
		},
		"AEAUH": {
			"name": "Abu Dhabi",
			"coordinates": [54.37, 24.47],
			"city": "Abu Dhabi",
			"province": "Abu Dhabi",
			"country": "United Arab Emirates",
			"timezone": "Asia/Dubai",
			"unlocs": ["AEAUH"],
			"code": "52001"
		}
	}`

	tmpfile, err := os.CreateTemp("", "ports*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// Test processing the file
	err = service.ProcessPortsFile(ctx, tmpfile.Name())
	assert.NoError(t, err)

	// Verify first port was saved
	port1, err := service.GetPort(ctx, "AEAJM")
	assert.NoError(t, err)
	assert.NotNil(t, port1)
	assert.Equal(t, "Ajman", port1.Name)

	// Verify second port was saved
	port2, err := service.GetPort(ctx, "AEAUH")
	assert.NoError(t, err)
	assert.NotNil(t, port2)
	assert.Equal(t, "Abu Dhabi", port2.Name)
}

func TestPortService_ProcessFile_Errors(t *testing.T) {
	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// Test non-existent file
	err := service.ProcessPortsFile(ctx, "nonexistent.json")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file")

	// Test invalid JSON content
	tmpfile, err := os.CreateTemp("", "invalid*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(`{"invalid": json`)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	err = service.ProcessPortsFile(ctx, tmpfile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid")

	// Test context cancellation
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	err = service.ProcessPortsFile(cancelCtx, tmpfile.Name())
	assert.Equal(t, context.Canceled, err)
}

func TestPortService_CreateOrUpdatePort_Errors(t *testing.T) {
	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// Test nil port
	err := service.CreateOrUpdatePort(ctx, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "nil port")

	// Test invalid port
	invalidPort := &domain.Port{
		ID:   "",
		Name: "",
	}
	err = service.CreateOrUpdatePort(ctx, invalidPort)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "port ID cannot be empty")

	// Test context cancellation
	coords := []float64{55.5136433, 25.4052165}
	validPort, _ := domain.NewPort("TEST1", "Test Port", "Test City", "Test Country", coords, "", "", nil, "")
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	err = service.CreateOrUpdatePort(cancelCtx, validPort)
	assert.Equal(t, context.Canceled, err)
}

func TestPortService_GetPort_Errors(t *testing.T) {
	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// Test empty ID
	port, err := service.GetPort(ctx, "")
	assert.Error(t, err)
	assert.Nil(t, port)
	assert.Contains(t, err.Error(), "empty port ID")

	// Test context cancellation
	cancelCtx, cancel := context.WithCancel(context.Background())
	cancel()
	port, err = service.GetPort(cancelCtx, "TEST1")
	assert.Equal(t, context.Canceled, err)
	assert.Nil(t, port)
}

func TestPortService_ProcessFile_MalformedData(t *testing.T) {
	content := `{
		"INVALID1": {
			"name": "Invalid Port 1",
			"coordinates": "not-an-array",
			"city": "Test City"
		},
		"INVALID2": {
			"name": "Invalid Port 2",
			"coordinates": [181.0, 91.0],
			"city": "Test City"
		},
		"VALID1": {
			"name": "Valid Port",
			"coordinates": [55.5136433, 25.4052165],
			"city": "Test City",
			"country": "Test Country"
		}
	}`

	tmpfile, err := os.CreateTemp("", "malformed*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// Process file should not fail on malformed data
	err = service.ProcessPortsFile(ctx, tmpfile.Name())
	assert.NoError(t, err)

	// Invalid ports should be skipped
	port1, err := service.GetPort(ctx, "INVALID1")
	assert.NoError(t, err)
	assert.Nil(t, port1)

	port2, err := service.GetPort(ctx, "INVALID2")
	assert.NoError(t, err)
	assert.Nil(t, port2)

	// Valid port should be saved
	port3, err := service.GetPort(ctx, "VALID1")
	assert.NoError(t, err)
	assert.NotNil(t, port3)
	assert.Equal(t, "Valid Port", port3.Name)
}

func TestPortService_ProcessFile_LargeFile(t *testing.T) {
	// Generate a large JSON file with 1000 ports
	var content strings.Builder
	content.WriteString("{\n")
	for i := 0; i < 1000; i++ {
		if i > 0 {
			content.WriteString(",\n")
		}
		portID := fmt.Sprintf("PORT%d", i)
		content.WriteString(fmt.Sprintf(`
			"%s": {
				"name": "Port %d",
				"coordinates": [%.6f, %.6f],
				"city": "City %d",
				"country": "Country %d",
				"timezone": "UTC"
			}`, portID, i,
			-180.0+float64(i)*0.36, // Spread ports across longitudes
			-90.0+float64(i)*0.18,  // Spread ports across latitudes
			i, i))
	}
	content.WriteString("\n}")

	tmpfile, err := os.CreateTemp("", "large*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content.String())); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// Process file
	err = service.ProcessPortsFile(ctx, tmpfile.Name())
	assert.NoError(t, err)

	// Verify random samples
	for _, i := range []int{0, 249, 499, 749, 999} {
		portID := fmt.Sprintf("PORT%d", i)
		port, err := service.GetPort(ctx, portID)
		assert.NoError(t, err)
		assert.NotNil(t, port)
		assert.Equal(t, fmt.Sprintf("Port %d", i), port.Name)
	}

	// Check repository statistics
	stats := repo.GetStatistics()
	assert.Equal(t, int64(1000), stats.TotalPorts)
}

func TestPortService_ProcessFile_EOF(t *testing.T) {
	// Create a file with incomplete JSON
	content := `{
		"AEAJM": {
			"name": "Ajman",
			"city": "Ajman",
			"country": "United Arab Emirates",
			"coordinates": [55.5136433, 25.4052165]
		},
		"AEAUH": {`

	tmpfile, err := os.CreateTemp("", "eof*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// Test processing the file
	err = service.ProcessPortsFile(ctx, tmpfile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to decode port data")

	// Verify first port was saved
	port, err := service.GetPort(ctx, "AEAJM")
	assert.NoError(t, err)
	assert.NotNil(t, port)
	assert.Equal(t, "Ajman", port.Name)

	// Verify second port was not saved
	port, err = service.GetPort(ctx, "AEAUH")
	assert.NoError(t, err)
	assert.Nil(t, port)
}

func TestPortService_ProcessFile_EmptyFile(t *testing.T) {
	// Create an empty file
	tmpfile, err := os.CreateTemp("", "empty*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// Test processing the empty file
	err = service.ProcessPortsFile(ctx, tmpfile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read JSON start")

	// Verify repository is empty
	stats := repo.GetStatistics()
	assert.Equal(t, int64(0), stats.TotalPorts)
}

func TestPortService_ProcessFile_LargePortData(t *testing.T) {
	// Generate a JSON file with a port containing large data fields
	var content strings.Builder
	content.WriteString("{\n")
	content.WriteString(`"LARGE": {
		"name": "`)

	// Generate a large name (100KB)
	for i := 0; i < 100*1024; i++ {
		content.WriteString("a")
	}

	content.WriteString(`",
		"coordinates": [55.5136433, 25.4052165],
		"city": "Test City",
		"country": "Test Country",
		"province": "Test Province",
		"timezone": "UTC",
		"unlocs": [`)

	// Generate many unlocs
	for i := 0; i < 1000; i++ {
		if i > 0 {
			content.WriteString(",")
		}
		content.WriteString(fmt.Sprintf(`"UNLOC%d"`, i))
	}

	content.WriteString(`],
		"code": "TEST"
	}
}`)

	tmpfile, err := os.CreateTemp("", "largeport*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content.String())); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// Process file
	err = service.ProcessPortsFile(ctx, tmpfile.Name())
	assert.NoError(t, err)

	// Verify port was saved
	port, err := service.GetPort(ctx, "LARGE")
	assert.NoError(t, err)
	assert.NotNil(t, port)
	assert.Equal(t, 100*1024, len(port.Name))
	assert.Equal(t, 1000, len(port.Unlocs))
}

func TestPortService_ProcessFile_DuplicateIDs(t *testing.T) {
	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// First file with initial port
	content1 := `{
		"AEAJM": {
			"name": "Ajman",
			"city": "Ajman",
			"country": "United Arab Emirates",
			"coordinates": [55.5136433, 25.4052165],
			"province": "Ajman",
			"timezone": "Asia/Dubai",
			"unlocs": ["AEAJM"],
			"code": "52000"
		}
	}`

	tmpfile1, err := os.CreateTemp("", "duplicate1*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile1.Name())

	if _, err := tmpfile1.Write([]byte(content1)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile1.Close(); err != nil {
		t.Fatal(err)
	}

	// Process first file
	err = service.ProcessPortsFile(ctx, tmpfile1.Name())
	assert.NoError(t, err)

	// Second file with updated port
	content2 := `{
		"AEAJM": {
			"name": "Ajman Updated",
			"city": "Ajman",
			"country": "United Arab Emirates",
			"coordinates": [55.5136433, 25.4052165],
			"province": "Ajman",
			"timezone": "Asia/Dubai",
			"unlocs": ["AEAJM"],
			"code": "52000"
		}
	}`

	tmpfile2, err := os.CreateTemp("", "duplicate2*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile2.Name())

	if _, err := tmpfile2.Write([]byte(content2)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile2.Close(); err != nil {
		t.Fatal(err)
	}

	// Process second file
	err = service.ProcessPortsFile(ctx, tmpfile2.Name())
	assert.NoError(t, err)

	// Verify the last version of the port was saved
	port, err := service.GetPort(ctx, "AEAJM")
	assert.NoError(t, err)
	assert.NotNil(t, port)
	assert.Equal(t, "Ajman Updated", port.Name)

	// Verify only one port was saved with two updates
	stats := repo.GetStatistics()
	assert.Equal(t, int64(1), stats.TotalPorts)
	assert.Equal(t, int64(2), stats.TotalUpdates)
}

// mockRepository is a mock implementation of PortRepository for testing
type mockRepository struct {
	mu           sync.RWMutex
	ports        map[string]*domain.Port
	totalUpdates int64
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		ports: make(map[string]*domain.Port),
	}
}

func (m *mockRepository) SavePort(ctx context.Context, port *domain.Port) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ports[port.ID] = port
	m.totalUpdates++
	return nil
}

func (m *mockRepository) GetPort(ctx context.Context, id string) (*domain.Port, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.ports[id], nil
}

func (m *mockRepository) Close(ctx context.Context) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ports = make(map[string]*domain.Port)
	m.totalUpdates = 0
	return nil
}

func (m *mockRepository) GetStatistics() out.RepositoryStats {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return out.RepositoryStats{
		TotalPorts:   int64(len(m.ports)),
		TotalUpdates: m.totalUpdates,
		LastUpdate:   "",
	}
}

// errorRepository is a mock repository that always returns errors
type errorRepository struct{}

func (e *errorRepository) SavePort(ctx context.Context, port *domain.Port) error {
	return fmt.Errorf("mock save error")
}

func (e *errorRepository) GetPort(ctx context.Context, id string) (*domain.Port, error) {
	return nil, fmt.Errorf("mock get error")
}

func (e *errorRepository) Close(ctx context.Context) error {
	return fmt.Errorf("mock close error")
}

func (e *errorRepository) GetStatistics() out.RepositoryStats {
	return out.RepositoryStats{}
}

func TestPortService_RepositoryErrors(t *testing.T) {
	repo := &errorRepository{}
	service := NewPortService(repo)
	ctx := context.Background()

	// Test save error
	coords := []float64{55.5136433, 25.4052165}
	port, err := domain.NewPort("TEST1", "Test Port", "Test City", "Test Country", coords, "", "", nil, "")
	assert.NoError(t, err)

	err = service.CreateOrUpdatePort(ctx, port)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "mock save error")

	// Test get error
	port, err = service.GetPort(ctx, "TEST1")
	assert.Error(t, err)
	assert.Nil(t, port)
	assert.Contains(t, err.Error(), "mock get error")

	// Test file processing with repository error
	content := `{
		"AEAJM": {
			"name": "Ajman",
			"coordinates": [55.5136433, 25.4052165],
			"city": "Ajman"
		}
	}`

	tmpfile, err := os.CreateTemp("", "repoerror*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	err = service.ProcessPortsFile(ctx, tmpfile.Name())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to save port")
	assert.Contains(t, err.Error(), "mock save error")
}

func TestPortService_ConcurrentFileProcessing(t *testing.T) {
	// Create multiple test files
	const numFiles = 5
	files := make([]string, numFiles)

	for i := 0; i < numFiles; i++ {
		content := fmt.Sprintf(`{
			"PORT%d": {
				"name": "Port %d",
				"coordinates": [%.6f, %.6f],
				"city": "City %d",
				"country": "Country %d",
				"timezone": "UTC"
			}
		}`, i, i,
			-180.0+float64(i)*36.0,
			-90.0+float64(i)*18.0,
			i, i)

		tmpfile, err := os.CreateTemp("", fmt.Sprintf("concurrent%d*.json", i))
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmpfile.Name())

		if _, err := tmpfile.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
		if err := tmpfile.Close(); err != nil {
			t.Fatal(err)
		}
		files[i] = tmpfile.Name()
	}

	repo := newMockRepository()
	service := NewPortService(repo)
	ctx := context.Background()

	// Process files concurrently
	var wg sync.WaitGroup
	wg.Add(numFiles)
	errs := make(chan error, numFiles)

	for i := 0; i < numFiles; i++ {
		go func(file string) {
			defer wg.Done()
			if err := service.ProcessPortsFile(ctx, file); err != nil {
				errs <- err
			}
		}(files[i])
	}

	wg.Wait()
	close(errs)

	// Check for errors
	for err := range errs {
		assert.NoError(t, err)
	}

	// Verify all ports were saved
	for i := 0; i < numFiles; i++ {
		portID := fmt.Sprintf("PORT%d", i)
		port, err := service.GetPort(ctx, portID)
		assert.NoError(t, err)
		assert.NotNil(t, port)
		assert.Equal(t, fmt.Sprintf("Port %d", i), port.Name)
	}

	// Verify repository statistics
	stats := repo.GetStatistics()
	assert.Equal(t, int64(numFiles), stats.TotalPorts)
}
