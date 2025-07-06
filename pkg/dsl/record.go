// Package dsl provides a fluent Domain Specific Language (DSL) for convenient
// querying and manipulating data in PocketBase using Go.
//
// The DSL package offers a chainable API that makes it easy to build complex
// queries with type safety and parameter binding. It supports all common
// operations including filtering, pagination, sorting, expansion, and CRUD
// operations.
//
// Example usage:
//
//	import "github.com/sospartan/pb-toolkit/pkg/dsl"
//
//	// Create a query
//	query := dsl.Query("status = 'active'").Page(1, 10).Sort("-created")
//
//	// Use with PocketBase
//	records, err := dsl.Collection(app, "users").List(query)
package dsl

import (
	"fmt"
	"strings"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/core"
)

// QueryBuilder represents a query configuration for building complex
// database queries with support for filtering, pagination, sorting,
// and relation expansion.
//
// QueryBuilder provides a fluent interface that allows chaining multiple
// operations together to create sophisticated queries.
type QueryBuilder struct {
	filter  string       // The filter expression (e.g., "status = 'active'")
	page    int          // Current page number (1-based)
	perPage int          // Number of items per page
	expand  string       // Comma-separated list of relations to expand
	sort    string       // Sort expression (e.g., "-created,name")
	params  []dbx.Params // Parameters for parameterized queries
}

// Query creates a new QueryBuilder with the specified filter expression.
//
// The filter parameter should be a valid PocketBase filter expression.
// For parameterized queries, use placeholders like {:param_name}.
//
// Example:
//
//	// Simple filter
//	query := dsl.Query("status = 'active'")
//
//	// Parameterized filter
//	query := dsl.Query("status = {:status} AND age > {:min_age}")
func Query(filter string) *QueryBuilder {
	return &QueryBuilder{
		filter: filter,
	}
}

// Page sets the pagination parameters for the query.
//
// page specifies the page number (1-based), and perPage specifies
// the number of items per page. Both values must be positive integers.
//
// Example:
//
//	query := dsl.Query("status = 'active'").Page(1, 20)
func (q *QueryBuilder) Page(page int, perPage int) *QueryBuilder {
	q.page = page
	q.perPage = perPage
	return q
}

// Expand specifies which relations should be expanded in the query results.
//
// The expand parameter should be a comma-separated list of relation names.
// Each relation will be automatically loaded and included in the response.
//
// Example:
//
//	// Expand single relation
//	query := dsl.Query("").Expand("profile")
//
//	// Expand multiple relations
//	query := dsl.Query("").Expand("profile,posts,comments")
func (q *QueryBuilder) Expand(expand string) *QueryBuilder {
	q.expand = expand
	return q
}

// Sort sets the sorting order for the query results.
//
// The sort parameter should be a comma-separated list of field names.
// Prefix a field with "-" for descending order, or use "+" or no prefix
// for ascending order.
//
// Example:
//
//	// Sort by single field (ascending)
//	query := dsl.Query("").Sort("created")
//
//	// Sort by single field (descending)
//	query := dsl.Query("").Sort("-created")
//
//	// Sort by multiple fields
//	query := dsl.Query("").Sort("name,-created,updated")
func (q *QueryBuilder) Sort(sort string) *QueryBuilder {
	q.sort = sort
	return q
}

// Params adds parameters for parameterized queries.
//
// This method supports multiple dbx.Params arguments, which will be
// merged together. Parameters are used to safely bind values to
// placeholders in the filter expression.
//
// Example:
//
//	query := dsl.Query("status = {:status} AND age > {:min_age}")
//		.Params(dbx.Params{"status": "active"}, dbx.Params{"min_age": 18})
func (q *QueryBuilder) Params(params ...dbx.Params) *QueryBuilder {
	q.params = params
	return q
}

// CollectionQueryBuilder represents a collection query builder that provides
// methods for performing CRUD operations on a specific PocketBase collection.
//
// CollectionQueryBuilder is created by calling Collection() and provides
// methods like One(), First(), List(), Create(), Update(), and Delete().
type CollectionQueryBuilder struct {
	app        core.App // The PocketBase app instance
	collection string   // The collection name or ID
}

// One retrieves a single record by ID from the collection.
//
// Returns the record if found, or nil if not found. An error is returned
// if the collection doesn't exist or if there's a database error.
//
// Example:
//
//	record, err := dsl.Collection(app, "users").One("user123")
//	if err != nil {
//	    // Handle error
//	}
func (c *CollectionQueryBuilder) One(id string) (*core.Record, error) {
	return c.app.FindRecordById(c.collection, id)
}

// First retrieves the first record matching the query criteria.
//
// This method is useful when you expect only one result or want to get
// the first record from a filtered set. It automatically limits the
// query to 1 result for efficiency.
//
// The query parameter specifies the filter and other query options.
// Additional parameters can be passed for parameterized queries.
//
// Example:
//
//	query := dsl.Query("email = {:email}").Sort("-created")
//	record, err := dsl.Collection(app, "users").First(query, dbx.Params{"email": "user@example.com"})
func (c *CollectionQueryBuilder) First(query QueryBuilder, params ...dbx.Params) (*core.Record, error) {
	records, err := c.app.FindRecordsByFilter(
		c.collection,
		query.filter,
		query.sort,
		1, // limit to 1
		0, // offset 0
		params...,
	)
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, fmt.Errorf("no records found")
	}
	record := records[0]
	if query.expand != "" {
		expands := strings.Split(query.expand, ",")
		for i, expand := range expands {
			expands[i] = strings.TrimSpace(expand)
		}
		errs := c.app.ExpandRecord(record, expands, nil)
		if len(errs) > 0 {
			return nil, fmt.Errorf("failed to expand relations: %v", errs)
		}
	}
	return record, nil
}

// List retrieves multiple records based on the query criteria.
//
// This method supports pagination, filtering, sorting, and relation expansion.
// The query parameter specifies all query options, and additional parameters
// can be passed for parameterized queries.
//
// Example:
//
//	query := dsl.Query("status = 'active'").Page(1, 10).Sort("-created").Expand("profile")
//	records, err := dsl.Collection(app, "users").List(query)
func (c *CollectionQueryBuilder) List(query QueryBuilder, params ...dbx.Params) ([]*core.Record, error) {
	offset := (query.page - 1) * query.perPage
	if offset < 0 {
		offset = 0
	}
	records, err := c.app.FindRecordsByFilter(
		c.collection,
		query.filter,
		query.sort,
		query.perPage,
		offset,
		params...,
	)
	if err != nil {
		return nil, err
	}
	if query.expand != "" {
		expands := strings.Split(query.expand, ",")
		for i, expand := range expands {
			expands[i] = strings.TrimSpace(expand)
		}
		for _, record := range records {
			errs := c.app.ExpandRecord(record, expands, nil)
			if len(errs) > 0 {
				return nil, fmt.Errorf("failed to expand relations: %v", errs)
			}
		}
	}
	return records, nil
}

// Create creates a new record in the collection with the provided data.
//
// The recordMap parameter should contain the field values for the new record.
// The method returns the created record with its generated ID and timestamps.
//
// Example:
//
//	data := map[string]any{
//	    "name":  "John Doe",
//	    "email": "john@example.com",
//	    "age":   30,
//	}
//	record, err := dsl.Collection(app, "users").Create(data)
func (c *CollectionQueryBuilder) Create(recordMap map[string]any) (*core.Record, error) {
	collection, err := c.app.FindCollectionByNameOrId(c.collection)
	if err != nil {
		return nil, fmt.Errorf("collection not found: %v", err)
	}

	record := core.NewRecord(collection)
	record.Load(recordMap)
	if err := c.app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

// Update updates an existing record by ID with the provided data.
//
// The id parameter specifies the record to update, and recordMap contains
// the new field values. Only the specified fields will be updated.
// The method returns the updated record.
//
// Example:
//
//	data := map[string]any{
//	    "name": "John Smith",
//	    "age":  31,
//	}
//	record, err := dsl.Collection(app, "users").Update("user123", data)
func (c *CollectionQueryBuilder) Update(id string, recordMap map[string]any) (*core.Record, error) {
	record, err := c.app.FindRecordById(c.collection, id)
	if err != nil {
		return nil, err
	}
	record.Load(recordMap)
	if err := c.app.Save(record); err != nil {
		return nil, err
	}
	return record, nil
}

// Delete deletes a record by ID from the collection.
//
// The id parameter specifies the record to delete. The method returns
// an error if the record doesn't exist or if deletion fails.
//
// Example:
//
//	err := dsl.Collection(app, "users").Delete("user123")
func (c *CollectionQueryBuilder) Delete(id string) error {
	record, err := c.app.FindRecordById(c.collection, id)
	if err != nil {
		return err
	}
	return c.app.Delete(record)
}

// Count returns the total number of records matching the filter criteria.
//
// The filter parameter specifies the filter expression, and additional
// parameters can be passed for parameterized queries.
//
// Example:
//
//	// Count all records
//	count, err := dsl.Collection(app, "users").Count("")
//
//	// Count filtered records
//	count, err := dsl.Collection(app, "users").Count("status = 'active'")
//
//	// Count with parameters
//	count, err := dsl.Collection(app, "users").Count("status = {:status}", dbx.Params{"status": "active"})
func (c *CollectionQueryBuilder) Count(filter string, params ...dbx.Params) (int64, error) {
	return c.app.CountRecords(c.collection, dbx.NewExp(filter, params...))
}

// Collection creates a new CollectionQueryBuilder for the specified collection.
//
// The app parameter should be a valid PocketBase app instance, and collection
// should be the name or ID of the collection to operate on.
//
// Example:
//
//	collection := dsl.Collection(app, "users")
//	records, err := collection.List(dsl.Query("status = 'active'"))
func Collection(app core.App, collection string) *CollectionQueryBuilder {
	return &CollectionQueryBuilder{
		app:        app,
		collection: collection,
	}
}
