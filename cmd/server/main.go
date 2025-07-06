package main

import (
	"log"
	"os"
	"strings"

	_ "github.com/sospartan/pb-toolkit/migrations" // import migrations
	"github.com/sospartan/pb-toolkit/pkg/dsl"
	"github.com/sospartan/pb-toolkit/pkg/rpc"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
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

func main() {
	app := pocketbase.New()

	// Register migration commands
	// loosely check if it was executed using "go run"
	isGoRun := strings.HasPrefix(os.Args[0], os.TempDir())
	migratecmd.MustRegister(app, app.RootCmd, migratecmd.Config{
		// enable auto creation of migration files when making collection changes in the Dashboard
		// (the isGoRun check is to enable it only during development)
		Automigrate: isGoRun,
	})

	// Create RPC server
	rpcServer := rpc.NewServer()

	// Register products service
	productsService := &ProductsService{app: app}
	if err := rpcServer.RegisterService("products", productsService); err != nil {
		log.Fatal("Failed to register products service:", err)
	}

	// Add RPC routes to PocketBase
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {

		// Path-based RPC endpoint
		g := se.Router.Group("/rpc")
		// Remove auth requirement for testing
		// g.Bind(apis.RequireAuth("users"))
		rpcServer.Bind(g)

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println("Server Stopped")
}
