# DSL Package

A fluent Domain Specific Language (DSL) for convenient querying and manipulating data in PocketBase using Go.

## Features

- **Chainable API**: Fluent interface for building complex queries
- **Type Safety**: Full type safety with Go's strong typing
- **Parameter Binding**: Safe parameter binding to prevent SQL injection
- **Pagination Support**: Built-in pagination with page and per-page controls
- **Sorting**: Flexible sorting with multiple field support
- **Expansion**: Automatic relation expansion
- **CRUD Operations**: Complete Create, Read, Update, Delete operations

## Installation

```bash
go get github.com/sospartan/pb-toolkit/pkg/dsl
```

## Quick Start

```go
import "github.com/sospartan/pb-toolkit/pkg/dsl"

// Create a query builder
query := dsl.Query("status = 'active'").Page(1, 10).Sort("-created")

// Use with PocketBase app
records, err := dsl.Collection(app, "users").List(query)
```

## API Reference

### Query Builder

#### Creating Queries

```go
// Basic query
query := dsl.Query("status = 'active'")

// With parameters
query := dsl.Query("status = {:status} AND value > {:min_value}")
    .Params(dbx.Params{"status": "active", "min_value": 100})
```

#### Pagination

```go
// Set page and per-page
query := dsl.Query("status = 'active'").Page(1, 20)

// Page 1 with 10 items per page
query := dsl.Query("").Page(1, 10)
```

#### Sorting

```go
// Sort by single field
query := dsl.Query("").Sort("created")

// Sort by multiple fields (descending)
query := dsl.Query("").Sort("-created,-updated")

// Sort by multiple fields (mixed)
query := dsl.Query("").Sort("name,-created")
```

#### Expansion

```go
// Expand single relation
query := dsl.Query("").Expand("profile")

// Expand multiple relations
query := dsl.Query("").Expand("profile,posts,comments")
```

#### Chaining

```go
// Chain multiple operations
query := dsl.Query("status = 'active'")
    .Page(1, 10)
    .Sort("-created")
    .Expand("profile,posts")
    .Params(dbx.Params{"status": "active"})
```

### Collection Operations

#### List Records

```go
// Get all records
records, err := dsl.Collection(app, "users").List(dsl.Query(""))

// Get filtered records
query := dsl.Query("status = 'active'").Page(1, 10)
records, err := dsl.Collection(app, "users").List(query)
```

#### Get Single Record

```go
// Get by ID
record, err := dsl.Collection(app, "users").One("user123")

// Get first matching record
query := dsl.Query("email = {:email}")
record, err := dsl.Collection(app, "users").First(query, dbx.Params{"email": "user@example.com"})
```

#### Create Record

```go
// Create new record
data := map[string]any{
    "name":  "John Doe",
    "email": "john@example.com",
    "age":   30,
}
record, err := dsl.Collection(app, "users").Create(data)
```

#### Update Record

```go
// Update existing record
data := map[string]any{
    "name": "John Smith",
    "age":  31,
}
record, err := dsl.Collection(app, "users").Update("user123", data)
```

#### Delete Record

```go
// Delete record by ID
err := dsl.Collection(app, "users").Delete("user123")
```

#### Count Records

```go
// Count all records
count, err := dsl.Collection(app, "users").Count("")

// Count filtered records
count, err := dsl.Collection(app, "users").Count("status = 'active'")

// Count with parameters
count, err := dsl.Collection(app, "users").Count("status = {:status}", dbx.Params{"status": "active"})
```

## Examples

### Basic CRUD Operations

```go
package main

import (
    "log"
    "github.com/sospartan/pb-toolkit/pkg/dsl"
    "github.com/pocketbase/pocketbase"
)

func main() {
    app := pocketbase.New()
    
    // Create a user
    userData := map[string]any{
        "name":  "Alice Johnson",
        "email": "alice@example.com",
        "age":   25,
    }
    user, err := dsl.Collection(app, "users").Create(userData)
    if err != nil {
        log.Fatal(err)
    }
    
    // Get user by ID
    user, err = dsl.Collection(app, "users").One(user.Id)
    if err != nil {
        log.Fatal(err)
    }
    
    // Update user
    updateData := map[string]any{"age": 26}
    user, err = dsl.Collection(app, "users").Update(user.Id, updateData)
    if err != nil {
        log.Fatal(err)
    }
    
    // List active users
    query := dsl.Query("status = 'active'").Page(1, 10).Sort("-created")
    users, err := dsl.Collection(app, "users").List(query)
    if err != nil {
        log.Fatal(err)
    }
    
    // Delete user
    err = dsl.Collection(app, "users").Delete(user.Id)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Complex Queries

```go
// Complex query with multiple conditions
query := dsl.Query(`
    status = {:status} 
    AND age >= {:min_age} 
    AND age <= {:max_age}
    AND created >= {:start_date}
`).Params(dbx.Params{
    "status":     "active",
    "min_age":    18,
    "max_age":    65,
    "start_date": "2024-01-01",
}).Page(1, 20).Sort("name,-created").Expand("profile,posts")

users, err := dsl.Collection(app, "users").List(query)
```

### Error Handling

```go
// Handle different types of errors
record, err := dsl.Collection(app, "users").One("nonexistent-id")
if err != nil {
    if strings.Contains(err.Error(), "not found") {
        log.Println("User not found")
    } else {
        log.Printf("Database error: %v", err)
    }
}
```

## Best Practices

1. **Use Parameter Binding**: Always use parameter binding for user input to prevent SQL injection
2. **Handle Errors**: Always check and handle errors appropriately
3. **Use Meaningful Filters**: Write clear and specific filter conditions
4. **Limit Results**: Use pagination for large datasets
5. **Expand Relations Carefully**: Only expand relations you actually need

## Error Types

The DSL package may return various types of errors:

- **Not Found**: When a record or collection doesn't exist
- **Validation Errors**: When input data is invalid
- **Database Errors**: When underlying database operations fail
- **Permission Errors**: When access is denied

## Performance Considerations

- Use specific filters to limit the result set
- Avoid expanding large relations unnecessarily
- Use pagination for large datasets
- Consider indexing frequently queried fields 