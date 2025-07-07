package migrations

import (
	"github.com/pocketbase/pocketbase/core"
	m "github.com/pocketbase/pocketbase/migrations"
)

const (
	CollectionNameWechatAuth = "wechat_auth"
	FieldWeOpenid = "we_openid"
	FieldWeUnionid = "we_unionid"
	FieldWeAuthinfo = "we_authinfo"
	FieldWeTokenExpired = "we_token_expired"
	FieldWeAccessToken = "we_access_token"
	FieldLastAuthCode = "last_auth_code"
)

func init() {
	m.Register(func(app core.App) error {
		collection := core.NewCollection(core.CollectionTypeAuth, CollectionNameWechatAuth)

		collection.Fields.Add(
			&core.TextField{
				Name:     FieldWeOpenid,
				Required: true,
				Max:      100,
			},
			&core.TextField{
				Name:     FieldWeUnionid,
				Required: false, // may be null
				Max:      100,
			},
			&core.JSONField{
				Name:     FieldWeAuthinfo,
				Required: true,
			},
			&core.DateField{
				Name:     FieldWeTokenExpired,
				Required: true,
			},
			&core.JSONField{
				Name:     FieldWeAccessToken,
				Required: true,
			},
			&core.TextField{
				Name:     FieldLastAuthCode,
				Required: true,
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

		collection.PasswordAuth = core.PasswordAuthConfig{
			Enabled: false,
		}

		// add index for better query performance
		collection.AddIndex("idx_openid", true, FieldWeOpenid, "")
		collection.AddIndex("idx_last_auth_code", false, FieldLastAuthCode, "")

		return app.Save(collection)
	}, func(app core.App) error {
		// add down queries...
		collection, err := app.FindCollectionByNameOrId(CollectionNameWechatAuth)
		if err != nil {
			return err
		}

		return app.Delete(collection)
	})
}
