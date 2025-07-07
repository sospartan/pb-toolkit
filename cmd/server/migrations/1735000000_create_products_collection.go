package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

func init() {
	m.Register(func(app core.App) error {
		// init a new base collection for products
		collection := core.NewCollection(core.CollectionTypeBase, "products")

		// add custom fields
		collection.Fields.Add(
			&core.TextField{
				Name:     "name",
				Required: true,
				Max:      100,
			},
			&core.NumberField{
				Name:     "price",
				Required: true,
				Min:      &[]float64{0}[0],
			},
			&core.TextField{
				Name:     "description",
				Required: false,
				Max:      500,
			},
			&core.AutodateField{
				Name:     "created",
				OnCreate: true,
			},
			&core.AutodateField{
				Name:     "updated",
				OnUpdate: true,
				OnCreate: true,
			},
		)

		// add index for better query performance
		collection.AddIndex("idx_products_name", false, "name", "")

		return app.Save(collection)
	}, func(app core.App) error {
		// optional revert operation
		collection, err := app.FindCollectionByNameOrId("products")
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
