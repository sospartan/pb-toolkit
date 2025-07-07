# RPC Framework

A JSON-RPC style framework for building service-oriented APIs with PocketBase, featuring automatic service registration and method discovery.

## Why Use This Library?

PocketBase does provide complete CRUD APIs, right? Yes, but in my personal experience, I don't prefer using API rules to control all data security and validation. When business logic becomes more complex, it becomes difficult to maintain for me. For example, with a collection where "some fields are user-modifiable while others are not allowed" - this kind of scenario. I personally prefer to keep all collection APIs open only to admin users, and handle other parts through custom backend code.

## Features

- **Automatic Service Registration**: Register services using reflection
- **Method Signature Validation**: Automatic validation of method signatures
- **JSON Request/Response**: Native JSON handling for requests and responses
- **RESTful Endpoints**: Automatic generation of RESTful endpoints
- **Type Safety**: Full type safety with Go's reflection system
- **Error Handling**: Comprehensive error handling and reporting

## Installation

```bash
go get github.com/sospartan/pb-toolkit/pkg/rpc
```

## Quick Start

```go
import "github.com/sospartan/pb-toolkit/pkg/rpc"

// Create RPC server
rpcServer := rpc.NewServer()

// Register services
rpcServer.RegisterService("user", &UserService{})

// Bind to router
rpcServer.Bind(router.Group("/rpc"))
```

## API Reference

### Server

#### Creating a Server

```go
// Create new RPC server
server := rpc.NewServer()
```

#### Registering Services

```go
// Register a service
type UserService struct{}

func (s *UserService) CreateUser(req CreateUserRequest) (CreateUserResponse, error) {
    // Implementation
    return response, nil
}

func (s *UserService) UpdateUser(req UpdateUserRequest) (UpdateUserResponse, error) {
    // Implementation
    return response, nil
}

// Register the service
err := server.RegisterService("user", &UserService{})
```

#### Binding to Router

```go
// Bind to PocketBase router
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    g := se.Router.Group("/rpc")
    g.Bind(apis.RequireAuth("users"))
    server.Bind(g)
    return se.Next()
})
```

### Service Method Requirements

Service methods must follow these rules:

1. **Exported Methods**: Methods must start with uppercase letter
2. **Single Parameter**: Exactly one parameter (the request struct)
3. **Two Return Values**: Must return exactly two values: `(result, error)`
4. **Error Type**: Second return value must implement the `error` interface

```go
// ✅ Correct method signature
func (s *UserService) CreateUser(req CreateUserRequest) (CreateUserResponse, error) {
    return response, nil
}

// ❌ Wrong - no error return
func (s *UserService) CreateUser(req CreateUserRequest) CreateUserResponse {
    return response
}

// ❌ Wrong - multiple parameters
func (s *UserService) CreateUser(req CreateUserRequest, extra string) (CreateUserResponse, error) {
    return response, nil
}
```

### Request/Response Types

Define request and response types for your methods:

```go
// Request type
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}

// Response type
type CreateUserResponse struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    CreatedAt string `json:"created_at"`
}
```

## Usage Examples

### Basic Service Implementation

```go
package main

import (
    "log"
    "github.com/sospartan/pb-toolkit/pkg/rpc"
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/apis"
    "github.com/pocketbase/pocketbase/core"
)

// UserService handles user operations
type UserService struct {
    App *pocketbase.PocketBase
}

// CreateUserRequest represents a request to create a user
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

// CreateUserResponse represents the response from creating a user
type CreateUserResponse struct {
    ID    string `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

// CreateUser creates a new user
func (s *UserService) CreateUser(req CreateUserRequest) (CreateUserResponse, error) {
    // Implementation here
    response := CreateUserResponse{
        ID:    "user_123",
        Name:  req.Name,
        Email: req.Email,
    }
    return response, nil
}

func main() {
    app := pocketbase.New()
    
    // Create RPC server
    rpcServer := rpc.NewServer()
    
    // Register services
    userService := &UserService{App: app}
    if err := rpcServer.RegisterService("user", userService); err != nil {
        log.Fatal("Failed to register user service:", err)
    }
    
    // Add RPC routes to PocketBase
    app.OnServe().BindFunc(func(se *core.ServeEvent) error {
        g := se.Router.Group("/rpc")
        g.Bind(apis.RequireAuth("users"))
        rpcServer.Bind(g)
        return se.Next()
    })
    
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### Making RPC Calls

#### POST Request (Method Call)

```bash
# Call CreateUser method
curl -X POST http://localhost:8090/rpc/user/create-user \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com"
  }'
```

#### GET Request (Entity Retrieval)

```bash
# Get user by ID
curl -X GET http://localhost:8090/rpc/user/user/123
```

### Advanced Service Example

```go
// OrderService handles order operations
type OrderService struct {
    App *pocketbase.PocketBase
}

type CreateOrderRequest struct {
    UserID    string   `json:"user_id"`
    Items     []string `json:"items"`
    Total     float64  `json:"total"`
}

type CreateOrderResponse struct {
    ID        string  `json:"id"`
    UserID    string  `json:"user_id"`
    Items     []string `json:"items"`
    Total     float64  `json:"total"`
    Status    string  `json:"status"`
    CreatedAt string  `json:"created_at"`
}

type UpdateOrderRequest struct {
    ID     string  `json:"id"`
    Status string  `json:"status"`
    Total  float64 `json:"total"`
}

type UpdateOrderResponse struct {
    ID        string  `json:"id"`
    Status    string  `json:"status"`
    Total     float64 `json:"total"`
    UpdatedAt string  `json:"updated_at"`
}

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(req CreateOrderRequest) (CreateOrderResponse, error) {
    // Implementation
    return CreateOrderResponse{
        ID:        "order_123",
        UserID:    req.UserID,
        Items:     req.Items,
        Total:     req.Total,
        Status:    "pending",
        CreatedAt: "2024-01-01T00:00:00Z",
    }, nil
}

// UpdateOrder updates an existing order
func (s *OrderService) UpdateOrder(req UpdateOrderRequest) (UpdateOrderResponse, error) {
    // Implementation
    return UpdateOrderResponse{
        ID:        req.ID,
        Status:    req.Status,
        Total:     req.Total,
        UpdatedAt: "2024-01-01T00:00:00Z",
    }, nil
}

// GetOrder retrieves an order by ID
func (s *OrderService) GetOrder(id string) (CreateOrderResponse, error) {
    // Implementation
    return CreateOrderResponse{
        ID:        id,
        UserID:    "user_123",
        Items:     []string{"item1", "item2"},
        Total:     99.99,
        Status:    "completed",
        CreatedAt: "2024-01-01T00:00:00Z",
    }, nil
}

// Register the service
orderService := &OrderService{App: app}
rpcServer.RegisterService("order", orderService)
```

## Endpoint Structure

The RPC framework creates the following endpoints:

### POST Endpoints (Method Calls)
- `POST /rpc/{service}/{method}` - Call service methods

### GET Endpoints (Entity Retrieval)
- `GET /rpc/{service}/{entity}/{id}` - Get entity by ID

### URL Conversion

The framework automatically converts kebab-case URLs to PascalCase method names:

- `POST /rpc/user/create-user` → `CreateUser` method
- `POST /rpc/order/update-order` → `UpdateOrder` method
- `GET /rpc/user/user/123` → `GetUser("123")` method

## Error Handling

The framework provides comprehensive error handling:

```go
// Service method error handling
func (s *UserService) CreateUser(req CreateUserRequest) (CreateUserResponse, error) {
    if req.Name == "" {
        return CreateUserResponse{}, fmt.Errorf("name is required")
    }
    
    if req.Email == "" {
        return CreateUserResponse{}, fmt.Errorf("email is required")
    }
    
    // Implementation
    return response, nil
}
```

### HTTP Status Codes

- `200 OK` - Successful operation
- `400 Bad Request` - Invalid parameters
- `404 Not Found` - Service or method not found
- `500 Internal Server Error` - Service method error

## Best Practices

1. **Use Descriptive Method Names**: Make method names clear and descriptive
2. **Validate Input**: Always validate request parameters
3. **Return Meaningful Errors**: Provide clear error messages
4. **Use Consistent Naming**: Follow consistent naming conventions
5. **Handle Edge Cases**: Consider all possible error scenarios
6. **Document Your Services**: Provide clear documentation for your services

## Performance Considerations

- Service registration happens at startup, so it doesn't affect runtime performance
- Method calls use reflection but are cached for efficiency
- Consider the overhead of JSON serialization/deserialization
- Use appropriate HTTP status codes for different error types 