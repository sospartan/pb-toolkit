package rpc

import (
	"testing"
)

// TestService represents a test service for testing
type TestService struct{}

// TestRequest represents a test request
type TestRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

// TestResponse represents a test response
type TestResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

// StatsResponse represents a stats response for parameterless methods
type StatsResponse struct {
	TotalUsers int    `json:"total_users"`
	Status     string `json:"status"`
}

// CreateUser is a valid method that follows the required signature
func (s *TestService) CreateUser(req TestRequest) (TestResponse, error) {
	return TestResponse{
		ID:    "user_123",
		Name:  req.Name,
		Email: req.Email,
	}, nil
}

// UpdateUser is another valid method
func (s *TestService) UpdateUser(req TestRequest) (TestResponse, error) {
	return TestResponse{
		ID:    req.Name, // Using name as ID for test
		Name:  req.Name,
		Email: req.Email,
	}, nil
}

// GetUser is a valid method for entity retrieval
func (s *TestService) GetUser(id string) (TestResponse, error) {
	return TestResponse{
		ID:    id,
		Name:  "Test User",
		Email: "test@example.com",
	}, nil
}

// GetStats is a valid parameterless method with result and error
func (s *TestService) GetStats() (StatsResponse, error) {
	return StatsResponse{
		TotalUsers: 100,
		Status:     "active",
	}, nil
}

// RefreshCache is a valid parameterless method with error only
func (s *TestService) RefreshCache() error {
	return nil // success
}

// InvalidMethod1 has wrong return signature (no error)
func (s *TestService) InvalidMethod1(req TestRequest) TestResponse {
	return TestResponse{}
}

// InvalidMethod2 has wrong parameter count (multiple parameters)
func (s *TestService) InvalidMethod2(req TestRequest, extra string) (TestResponse, error) {
	return TestResponse{}, nil
}

// TestNewServer tests the NewServer function
func TestNewServer(t *testing.T) {
	server := NewServer()

	if server == nil {
		t.Fatal("NewServer() returned nil")
	}

	if server.services == nil {
		t.Fatal("Server services map is nil")
	}

	if len(server.services) != 0 {
		t.Fatalf("Expected empty services map, got %d services", len(server.services))
	}
}

// TestRegisterService_ValidMethods tests registering a service with valid methods
func TestRegisterService_ValidMethods(t *testing.T) {
	server := NewServer()
	testService := &TestService{}

	err := server.RegisterService("test", testService)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	// Check if service was registered
	service, exists := server.services["test"]
	if !exists {
		t.Fatal("Service was not registered")
	}

	if service.service != testService {
		t.Fatal("Service instance mismatch")
	}

	if service.serviceName != "test" {
		t.Fatalf("Expected service name 'test', got '%s'", service.serviceName)
	}

	// Check if valid methods were registered
	expectedMethods := []string{"CreateUser", "UpdateUser", "GetUser", "GetStats", "RefreshCache"}
	for _, methodName := range expectedMethods {
		method, exists := service.methods[methodName]
		if !exists {
			t.Fatalf("Method '%s' was not registered", methodName)
		}

		if method.Method.Name != methodName {
			t.Fatalf("Expected method name '%s', got '%s'", methodName, method.Method.Name)
		}
	}

	// Check that invalid methods were not registered
	invalidMethods := []string{"InvalidMethod1", "InvalidMethod2"}
	for _, methodName := range invalidMethods {
		if _, exists := service.methods[methodName]; exists {
			t.Fatalf("Invalid method '%s' was incorrectly registered", methodName)
		}
	}
}

// TestRegisterService_InvalidMethods tests that invalid methods are not registered
func TestRegisterService_InvalidMethods(t *testing.T) {
	server := NewServer()
	testService := &TestService{}

	err := server.RegisterService("test", testService)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	service := server.services["test"]

	// Verify that invalid methods are not registered
	invalidMethods := []string{
		"InvalidMethod1", // Wrong return signature
		"InvalidMethod2", // Wrong parameter count (multiple params)
	}

	for _, methodName := range invalidMethods {
		if _, exists := service.methods[methodName]; exists {
			t.Errorf("Invalid method '%s' was incorrectly registered", methodName)
		}
	}
}

// TestRegisterService_DuplicateService tests registering the same service name twice
func TestRegisterService_DuplicateService(t *testing.T) {
	server := NewServer()
	testService1 := &TestService{}
	testService2 := &TestService{}

	// Register first service
	err := server.RegisterService("test", testService1)
	if err != nil {
		t.Fatalf("Failed to register first service: %v", err)
	}

	// Register second service with same name (should overwrite)
	err = server.RegisterService("test", testService2)
	if err != nil {
		t.Fatalf("Failed to register second service: %v", err)
	}

	// Check that the second service is now registered
	service := server.services["test"]
	if service.service != testService2 {
		t.Fatal("Second service was not registered")
	}
}

// TestRegisterService_EmptyServiceName tests registering with empty service name
func TestRegisterService_EmptyServiceName(t *testing.T) {
	server := NewServer()
	testService := &TestService{}

	err := server.RegisterService("", testService)
	if err != nil {
		t.Fatalf("Failed to register service with empty name: %v", err)
	}

	// Check if service was registered with empty name
	_, exists := server.services[""]
	if !exists {
		t.Fatal("Service with empty name was not registered")
	}
}

// TestRegisterService_NilService tests registering a nil service
func TestRegisterService_NilService(t *testing.T) {
	server := NewServer()

	// RegisterService should handle nil service gracefully
	err := server.RegisterService("test", nil)
	if err != nil {
		t.Fatalf("Failed to register nil service: %v", err)
	}

	// Check if nil service was registered
	service, exists := server.services["test"]
	if !exists {
		t.Fatal("Nil service was not registered")
	}

	if service.service != nil {
		t.Fatal("Service instance is not nil")
	}

	// Nil service should have no methods
	if len(service.methods) != 0 {
		t.Fatalf("Expected 0 methods for nil service, got %d", len(service.methods))
	}
}

// TestKebabToPascal tests the kebabToPascal function
func TestKebabToPascal(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"create-user", "CreateUser"},
		{"update-order", "UpdateOrder"},
		{"get-user-profile", "GetUserProfile"},
		{"", ""},
		{"single", "Single"},
		{"UPPER-CASE", "UpperCase"},
		{"mixed-Case", "MixedCase"},
		{"multiple-words-test", "MultipleWordsTest"},
	}

	for _, tc := range testCases {
		result := kebabToPascal(tc.input)
		if result != tc.expected {
			t.Errorf("kebabToPascal(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

// TestServerServicesCount tests that services are properly counted
func TestServerServicesCount(t *testing.T) {
	server := NewServer()
	testService := &TestService{}

	// Register multiple services
	services := []string{"user", "order", "product"}

	for _, serviceName := range services {
		err := server.RegisterService(serviceName, testService)
		if err != nil {
			t.Fatalf("Failed to register service '%s': %v", serviceName, err)
		}
	}

	// Check total count
	if len(server.services) != len(services) {
		t.Fatalf("Expected %d services, got %d", len(services), len(server.services))
	}

	// Check each service exists
	for _, serviceName := range services {
		if _, exists := server.services[serviceName]; !exists {
			t.Fatalf("Service '%s' was not registered", serviceName)
		}
	}
}

// TestServiceMethodsCount tests that methods are properly counted for a service
func TestServiceMethodsCount(t *testing.T) {
	server := NewServer()
	testService := &TestService{}

	err := server.RegisterService("test", testService)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	service := server.services["test"]
	expectedMethodCount := 5 // CreateUser, UpdateUser, GetUser, GetStats, RefreshCache

	if len(service.methods) != expectedMethodCount {
		t.Fatalf("Expected %d methods, got %d", expectedMethodCount, len(service.methods))
	}

	// Verify method names
	expectedMethods := map[string]bool{
		"CreateUser":   true,
		"UpdateUser":   true,
		"GetUser":      true,
		"GetStats":     true,
		"RefreshCache": true,
	}

	for methodName := range service.methods {
		if !expectedMethods[methodName] {
			t.Errorf("Unexpected method registered: %s", methodName)
		}
	}
}

// TestParameterlessMethods tests that parameterless methods are properly registered
func TestParameterlessMethods(t *testing.T) {
	server := NewServer()
	testService := &TestService{}

	err := server.RegisterService("test", testService)
	if err != nil {
		t.Fatalf("Failed to register service: %v", err)
	}

	service := server.services["test"]

	// Test GetStats method (parameterless with result and error)
	getStatsMethod, exists := service.methods["GetStats"]
	if !exists {
		t.Fatal("GetStats method was not registered")
	}

	if getStatsMethod.HasParams {
		t.Error("GetStats method should not have parameters")
	}

	if getStatsMethod.Type != nil {
		t.Error("GetStats method should have nil Type")
	}

	if !getStatsMethod.HasResult {
		t.Error("GetStats method should have a result")
	}

	// Test RefreshCache method (parameterless with error only)
	refreshCacheMethod, exists := service.methods["RefreshCache"]
	if !exists {
		t.Fatal("RefreshCache method was not registered")
	}

	if refreshCacheMethod.HasParams {
		t.Error("RefreshCache method should not have parameters")
	}

	if refreshCacheMethod.Type != nil {
		t.Error("RefreshCache method should have nil Type")
	}

	if refreshCacheMethod.HasResult {
		t.Error("RefreshCache method should not have a result")
	}

	// Test CreateUser method (with parameters)
	createUserMethod, exists := service.methods["CreateUser"]
	if !exists {
		t.Fatal("CreateUser method was not registered")
	}

	if !createUserMethod.HasParams {
		t.Error("CreateUser method should have parameters")
	}

	if createUserMethod.Type == nil {
		t.Error("CreateUser method should have a non-nil Type")
	}

	if !createUserMethod.HasResult {
		t.Error("CreateUser method should have a result")
	}
}
