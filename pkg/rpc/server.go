// Package rpc provides a JSON-RPC style framework for building service-oriented
// APIs with PocketBase, featuring automatic service registration and method discovery.
//
// The RPC framework uses reflection to automatically discover and register
// service methods, providing a type-safe way to build APIs with minimal boilerplate.
// It supports both POST requests for method calls and GET requests for entity retrieval.
//
// Example usage:
//
//	import "github.com/sospartan/pb-toolkit/pkg/rpc"
//
//	// Create RPC server
//	rpcServer := rpc.NewServer()
//
//	// Register services
//	rpcServer.RegisterService("user", &UserService{})
//
//	// Bind to router
//	rpcServer.Bind(router.Group("/rpc"))
package rpc

import (
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
)

// RPCMethod represents a registered RPC method with its reflection information.
//
// RPCMethod contains the method's reflection data and parameter type information,
// which is used by the RPC server to validate and execute method calls.
type RPCMethod struct {
	Method     reflect.Method // The reflection method information
	Type       reflect.Type   // The parameter type for the method (nil if no parameters)
	HasParams  bool           // Whether the method has parameters
	HasResult  bool           // Whether the method returns a result value
	ResultType reflect.Type   // The result type (if HasResult is true)
}

// RPCService represents an RPC service with registered methods.
//
// RPCService contains a service instance and a map of its registered methods.
// Each method is validated to ensure it follows the required signature pattern.
type RPCService struct {
	service     interface{}           // The service instance
	methods     map[string]*RPCMethod // Map of method name to method info
	serviceName string                // The name of the service
}

// Server handles RPC requests and manages registered services.
//
// Server provides methods for registering services, binding to routers,
// and handling incoming RPC requests. It uses reflection to automatically
// discover and validate service methods.
type Server struct {
	services map[string]*RPCService // Map of service name to service info
}

// NewServer creates a new RPC server instance.
//
// Returns a new Server that can be used to register services and handle
// RPC requests.
//
// Example:
//
//	server := rpc.NewServer()
func NewServer() *Server {
	return &Server{
		services: make(map[string]*RPCService),
	}
}

// Bind binds the RPC server to a router group, setting up the necessary routes.
//
// This method creates two types of routes:
// - POST /{service}/{method} for method calls
// - GET /{service}/{entity}/{id} for entity retrieval
//
// The router group should be configured with any necessary middleware
// (e.g., authentication, CORS, etc.).
//
// Example:
//
//	g := se.Router.Group("/rpc")
//	g.Bind(apis.RequireAuth("users"))
//	server.Bind(g)
func (s *Server) Bind(g *router.RouterGroup[*core.RequestEvent]) {
	g.POST("/{service}/{method}", s.handle)
	g.GET("/{service}/{entity}/{id}", s.handleGet)
}

// handleGet handles incoming GET RPC requests with an ID parameter.
//
// This method processes GET requests in the format /{service}/{entity}/{id}
// and converts them to method calls like GetEntity(id). The entity name
// is converted from kebab-case to PascalCase for method name matching.
//
// Example URL: GET /rpc/user/user/123 → calls GetUser("123")
func (s *Server) handleGet(e *core.RequestEvent) error {
	serv := e.Request.PathValue("service")
	entity := e.Request.PathValue("entity")
	id := e.Request.PathValue("id")
	entity = kebabToPascal(entity)
	return s.handleRPCGet(e, serv, entity, id)
}

// handle handles incoming POST RPC requests for method calls.
//
// This method processes POST requests in the format /{service}/{method}
// and converts the kebab-case method name to PascalCase for method matching.
// The request body is automatically deserialized into the method's parameter type.
//
// Example URL: POST /rpc/user/create-user → calls CreateUser(requestBody)
func (s *Server) handle(e *core.RequestEvent) error {
	serv := e.Request.PathValue("service")
	method := e.Request.PathValue("method")
	// Convert kebab-case to PascalCase for method name
	method = kebabToPascal(method)
	return s.handleRPC(e, serv, method)
}

// RegisterService registers a service with the RPC server.
//
// This method uses reflection to automatically discover all exported methods
// in the service that follow the required signature patterns:
//   - Method must be exported (start with uppercase)
//   - Method must have either:
//     a) Exactly one parameter (the request struct)
//     b) No parameters (for parameterless methods)
//   - Method must return either:
//     a) Exactly two values: (result, error) - where the second return value must implement the error interface
//     b) Exactly one value: error - where the return value must implement the error interface
//
// The name parameter is used as the service identifier in URLs.
// The service parameter should be a pointer to a struct with methods.
// If service is nil, it will be registered but with no methods.
//
// This implementation is based on the Ethereum go-ethereum approach.
//
// Example:
//
//	type UserService struct{}
//
//	// Method with result and error
//	func (s *UserService) CreateUser(req CreateUserRequest) (CreateUserResponse, error) {
//	    // Implementation
//	    return response, nil
//	}
//
//	// Method with error only
//	func (s *UserService) DeleteUser(req DeleteUserRequest) error {
//	    // Implementation
//	    return nil // success
//	}
//
//	// Parameterless method with result and error
//	func (s *UserService) GetStats() (StatsResponse, error) {
//	    // Implementation
//	    return stats, nil
//	}
//
//	// Parameterless method with error only
//	func (s *UserService) RefreshCache() error {
//	    // Implementation
//	    return nil // success
//	}
//
//	err := server.RegisterService("user", &UserService{})
func (s *Server) RegisterService(name string, service interface{}) error {
	svc := &RPCService{
		service:     service,
		methods:     make(map[string]*RPCMethod),
		serviceName: name,
	}

	// Handle nil service gracefully
	if service == nil {
		s.services[name] = svc
		return nil
	}

	// Use reflection to get service methods
	serviceType := reflect.TypeOf(service)
	for i := 0; i < serviceType.NumMethod(); i++ {
		method := serviceType.Method(i)

		// Check if method is exported (starts with uppercase)
		if method.PkgPath != "" {
			continue
		}

		// Check if method has exactly one argument (the request parameter) or no arguments
		numIn := method.Type.NumIn()
		if numIn != 1 && numIn != 2 { // receiver only, or receiver + 1 argument
			continue
		}

		// Check method return values
		numOut := method.Type.NumOut()
		if numOut != 1 && numOut != 2 {
			continue
		}

		// Check if the last return type is error
		lastOut := method.Type.Out(numOut - 1)
		if !lastOut.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			continue
		}

		// Create method info
		methodInfo := &RPCMethod{
			Method: method,
		}

		// Set parameter information
		if numIn == 2 {
			// Method has one parameter (receiver + 1 argument)
			methodInfo.HasParams = true
			methodInfo.Type = method.Type.In(1) // The argument type
		} else {
			// Method has no parameters (only receiver)
			methodInfo.HasParams = false
			methodInfo.Type = nil
		}

		// Set return type information
		if numOut == 2 {
			// Method returns (result, error)
			methodInfo.HasResult = true
			methodInfo.ResultType = method.Type.Out(0)
		} else {
			// Method returns only error
			methodInfo.HasResult = false
		}

		// Register the method
		svc.methods[method.Name] = methodInfo
	}

	s.services[name] = svc

	// Print registered methods for debugging
	log.Printf("Registered service '%s' with methods:", name)
	for methodName := range svc.methods {
		log.Printf("  - %s", methodName)
	}
	return nil
}

// handleRPC handles incoming RPC requests (JSON-RPC style).
//
// This method processes POST requests by:
// 1. Finding the requested service
// 2. Finding the requested method within the service
// 3. If the method has parameters, deserializing the request body into the method's parameter type
// 4. Calling the method with the deserialized parameters (or no parameters for parameterless methods)
// 5. Returning the method's result as JSON
//
// If any step fails, an appropriate HTTP error response is returned.
//
// Example requests:
//
//	POST /rpc/user/create-user
//	Content-Type: application/json
//	{
//	    "name": "John Doe",
//	    "email": "john@example.com"
//	}
//
//	POST /rpc/user/get-stats
//	Content-Type: application/json
//	{} // Empty body for parameterless method
func (s *Server) handleRPC(e *core.RequestEvent, serviceName, methodName string) error {

	// Find service
	service, exists := s.services[serviceName]
	if !exists {
		return e.JSON(http.StatusNotFound, fmt.Errorf("Service '%s' not found", serviceName))
	}

	// Find method
	method, exists := service.methods[methodName]
	if !exists {
		return e.JSON(http.StatusNotFound, fmt.Errorf("Method '%s' not found in service '%s'", methodName, serviceName))
	}

	// Call the method
	serviceValue := reflect.ValueOf(service.service)
	methodValue := serviceValue.MethodByName(method.Method.Name)

	var results []reflect.Value

	if method.HasParams {
		// Method has parameters, create argument instance and bind body
		argType := method.Type
		arg := reflect.New(argType).Interface()

		// Decode parameters into argument
		if err := e.BindBody(arg); err != nil {
			return e.JSON(http.StatusBadRequest, fmt.Errorf("Invalid parameters: %v", err))
		}

		// Call the method with the argument
		results = methodValue.Call([]reflect.Value{reflect.ValueOf(arg).Elem()})
	} else {
		// Method has no parameters, call without arguments
		results = methodValue.Call([]reflect.Value{})
	}

	// Check for error
	if method.HasResult {
		// Method returns (result, error)
		if !results[1].IsNil() {
			err := results[1].Interface().(error)
			return e.JSON(http.StatusInternalServerError, err)
		}
		// Return the first result as the response
		response := results[0].Interface()
		return e.JSON(http.StatusOK, response)
	} else {
		// Method returns only error
		if !results[0].IsNil() {
			err := results[0].Interface().(error)
			return e.JSON(http.StatusInternalServerError, err)
		}
		// Return success status
		return e.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
}

// handleRPCGet handles incoming GET RPC requests with an ID parameter.
//
// This method processes GET requests for entity retrieval by:
// 1. Finding the requested service
// 2. Constructing the method name as "Get" + EntityName
// 3. Finding the method within the service
// 4. Validating that the method accepts a string parameter
// 5. Calling the method with the ID parameter
// 6. Returning the method's result as JSON
//
// Example URL: GET /rpc/user/user/123 → calls GetUser("123")
func (s *Server) handleRPCGet(e *core.RequestEvent, serviceName, entityName, id string) error {
	// Find service
	service, exists := s.services[serviceName]
	if !exists {
		return e.JSON(http.StatusNotFound, fmt.Errorf("Service '%s' not found", serviceName))
	}

	// Construct method name: Get + EntityName (e.g., GetOrder)
	methodName := "Get" + entityName

	// Find method
	method, exists := service.methods[methodName]
	if !exists {
		return e.JSON(http.StatusNotFound, fmt.Errorf("Method '%s' not found in service '%s'", methodName, serviceName))
	}

	// Check if the method accepts a string parameter
	argType := method.Type
	if argType.Kind() != reflect.String {
		return e.JSON(http.StatusBadRequest, fmt.Errorf("Method '%s' does not accept a string parameter", methodName))
	}

	// Call the method with the ID
	serviceValue := reflect.ValueOf(service.service)
	methodValue := serviceValue.MethodByName(method.Method.Name)

	results := methodValue.Call([]reflect.Value{reflect.ValueOf(id)})

	// Check for error
	if method.HasResult {
		// Method returns (result, error)
		if !results[1].IsNil() {
			err := results[1].Interface().(error)
			return e.JSON(http.StatusInternalServerError, err)
		}
		// Return the first result as the response
		response := results[0].Interface()
		return e.JSON(http.StatusOK, response)
	} else {
		// Method returns only error
		if !results[0].IsNil() {
			err := results[0].Interface().(error)
			return e.JSON(http.StatusInternalServerError, err)
		}
		// Return success status
		return e.JSON(http.StatusOK, map[string]string{"status": "ok"})
	}
}

// kebabToPascal converts kebab-case to PascalCase.
//
// This utility function converts URL-friendly kebab-case strings to
// PascalCase method names for reflection-based method lookup.
//
// Example:
//   - "create-user" → "CreateUser"
//   - "update-order" → "UpdateOrder"
//   - "get-user-profile" → "GetUserProfile"
func kebabToPascal(kebab string) string {
	parts := strings.Split(kebab, "-")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
		}
	}
	return strings.Join(parts, "")
}
