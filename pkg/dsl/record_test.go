package dsl

import (
	"testing"
)

func TestQueryBuilderChaining(t *testing.T) {
	// Test Query builder chaining
	query := Query("status = 'active'").
		Page(1, 10).
		Expand("user,profile").
		Sort("-created")

	if query.filter != "status = 'active'" {
		t.Errorf("Expected filter 'status = 'active'', got '%s'", query.filter)
	}

	if query.page != 1 {
		t.Errorf("Expected page 1, got %d", query.page)
	}

	if query.perPage != 10 {
		t.Errorf("Expected perPage 10, got %d", query.perPage)
	}

	if query.expand != "user,profile" {
		t.Errorf("Expected expand 'user,profile', got '%s'", query.expand)
	}

	if query.sort != "-created" {
		t.Errorf("Expected sort '-created', got '%s'", query.sort)
	}

}

func TestQueryBuilderMethods(t *testing.T) {
	// Test individual methods
	query := Query("name = 'test'")

	if query.filter != "name = 'test'" {
		t.Errorf("Expected filter 'name = 'test'', got '%s'", query.filter)
	}

	// Test Page method
	query.Page(2, 20)
	if query.page != 2 {
		t.Errorf("Expected page 2, got %d", query.page)
	}
	if query.perPage != 20 {
		t.Errorf("Expected perPage 20, got %d", query.perPage)
	}

	// Test Expand method
	query.Expand("user")
	if query.expand != "user" {
		t.Errorf("Expected expand 'user', got '%s'", query.expand)
	}

	// Test Sort method
	query.Sort("created")
	if query.sort != "created" {
		t.Errorf("Expected sort 'created', got '%s'", query.sort)
	}

}

func TestQueryBuilderDefaultValues(t *testing.T) {
	// Test default values
	query := Query("test")

	if query.page != 0 {
		t.Errorf("Expected default page 0, got %d", query.page)
	}

	if query.perPage != 0 {
		t.Errorf("Expected default perPage 0, got %d", query.perPage)
	}

	if query.expand != "" {
		t.Errorf("Expected default expand '', got '%s'", query.expand)
	}

	if query.sort != "" {
		t.Errorf("Expected default sort '', got '%s'", query.sort)
	}

}

func TestQueryBuilderEmptyFilter(t *testing.T) {
	// Test empty filter
	query := Query("")

	if query.filter != "" {
		t.Errorf("Expected empty filter, got '%s'", query.filter)
	}
}
