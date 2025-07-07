package main

import (
	"github.com/pocketbase/pocketbase/core"
	"github.com/sospartan/pb-toolkit/pkg/dsl"
)


type Product struct {
	ID          string `json:"id,omitempty"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
	Created     string `json:"created,omitempty"`
	Updated     string `json:"updated,omitempty"`
}

type ListRequest struct{}

type UpdateRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Description string `json:"description"`
}

type DeleteRequest struct {
	ID string `json:"id"`
}

type ProductsService struct {
	app core.App
}

func (s *ProductsService) Create(req Product) (Product, error) {
	data := map[string]any{
		"name":        req.Name,
		"price":       req.Price,
		"description": req.Description,
	}
	record, err := dsl.Collection(s.app, "products").Create(data)
	if err != nil {
		return Product{}, err
	}
	return Product{
		ID:          record.Id,
		Name:        record.GetString("name"),
		Price:       record.GetInt("price"),
		Description: record.GetString("description"),
		Created:     record.GetString("created"),
		Updated:     record.GetString("updated"),
	}, nil
}

func (s *ProductsService) GetProduct(id string) (Product, error) {
	record, err := dsl.Collection(s.app, "products").One(id)
	if err != nil {
		return Product{}, err
	}
	return Product{
		ID:          record.Id,
		Name:        record.GetString("name"),
		Price:       record.GetInt("price"),
		Description: record.GetString("description"),
		Created:     record.GetString("created"),
		Updated:     record.GetString("updated"),
	}, nil
}

func (s *ProductsService) List(req ListRequest) ([]Product, error) {
	query := dsl.Query("")
	records, err := dsl.Collection(s.app, "products").List(*query)
	if err != nil {
		return nil, err
	}
	products := make([]Product, len(records))
	for i, record := range records {
		products[i] = Product{
			ID:          record.Id,
			Name:        record.GetString("name"),
			Price:       record.GetInt("price"),
			Description: record.GetString("description"),
			Created:     record.GetString("created"),
			Updated:     record.GetString("updated"),
		}
	}
	return products, nil
}

func (s *ProductsService) Update(req UpdateRequest) (Product, error) {
	data := map[string]any{
		"name":        req.Name,
		"price":       req.Price,
		"description": req.Description,
	}
	record, err := dsl.Collection(s.app, "products").Update(req.ID, data)
	if err != nil {
		return Product{}, err
	}
	return Product{
		ID:          record.Id,
		Name:        record.GetString("name"),
		Price:       record.GetInt("price"),
		Description: record.GetString("description"),
		Created:     record.GetString("created"),
		Updated:     record.GetString("updated"),
	}, nil
}

func (s *ProductsService) Delete(req DeleteRequest) error {
	return dsl.Collection(s.app, "products").Delete(req.ID)
}

func (s *ProductsService) Clean() error {
	query := dsl.Query("")
	records, err := dsl.Collection(s.app, "products").List(*query)
	if err != nil {
		return err
	}

	for _, record := range records {
		if err := dsl.Collection(s.app, "products").Delete(record.Id); err != nil {
			return err
		}
	}

	return nil
}