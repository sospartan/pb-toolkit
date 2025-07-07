package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/pocketbase/dbx"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"
	"github.com/pocketbase/pocketbase/tools/security"
	"github.com/sospartan/pb-toolkit/cmd/server/migrations"
	"github.com/sospartan/pb-toolkit/pkg/dsl"
	"github.com/sospartan/pb-toolkit/pkg/wechat"
)

func NewWechatAuthHandler(app core.App, appID, appSecret string) *WechatAuthHandler {
	return &WechatAuthHandler{app: app, appID: appID, appSecret: appSecret}
}

type WechatAuthHandler struct {
	app       core.App
	appID     string
	appSecret string
}

// FindAuthRecordByCode implements wechat.AuthHandler.
func (h *WechatAuthHandler) FindAuthRecordByCode(code string) (*core.Record, error) {
	collection := dsl.Collection(h.app, migrations.CollectionNameWechatAuth)
	query := dsl.Query(fmt.Sprintf("%s = {:code}", migrations.FieldLastAuthCode))
	record, err := collection.First(*query, dbx.Params{
		"code": code,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, wechat.NoAuthRecordError
		}
		return nil, err
	}
	return record, nil
}

// GetAuthConfig implements wechat.AuthHandler.
func (h *WechatAuthHandler) GetAuthConfig() *wechat.WechatAuth {
	return &wechat.WechatAuth{
		AppID:  h.appID,
		Secret: h.appSecret,
	}
}

// ModifyAuthRecord implements wechat.AuthHandler.
func (h *WechatAuthHandler) ModifyAuthRecord(record *core.Record) error {
	record.Hide(
		migrations.FieldLastAuthCode,
		migrations.FieldWeAccessToken,
		migrations.FieldWeTokenExpired,
	)
	return nil
}

// Save implements wechat.AuthHandler.
func (h *WechatAuthHandler) Save(token *wechat.AccessTokenResponse, info *wechat.UserInfoResponse, code string) (*core.Record, error) {
	collection := dsl.Collection(h.app, migrations.CollectionNameWechatAuth)
	record, err := collection.First(*dsl.Query(fmt.Sprintf("%s = {:openid}", migrations.FieldWeOpenid)), dbx.Params{
		"openid": info.OpenID,
	})
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if record == nil {
		return collection.Create(map[string]any{
			core.FieldNamePassword:         security.RandomString(10),
			core.FieldNameEmail:            info.OpenID + "@pb.com",
			migrations.FieldWeOpenid:       info.OpenID,
			migrations.FieldWeUnionid:      info.UnionID,
			migrations.FieldWeAuthinfo:     info,
			migrations.FieldWeAccessToken:  token,
			migrations.FieldWeTokenExpired: token.ExpiresIn,
			migrations.FieldLastAuthCode:   code,
		})
	}

	return collection.Update(record.Id, map[string]any{
		migrations.FieldWeAccessToken:  token,
		migrations.FieldWeTokenExpired: token.ExpiresIn,
		migrations.FieldLastAuthCode:   code,
		migrations.FieldWeAuthinfo:     info,
	})
}

func (h *WechatAuthHandler) SetupRoutes(g *router.RouterGroup[*core.RequestEvent]) {
	// Add wechat auth endpoint, with a code query param
	g.GET("/callback", wechat.HandleAuthResponseWithCode(h))

	authed := g.Group("/authed")
	authed.Bind(apis.RequireAuth(migrations.CollectionNameWechatAuth))

	authed.GET("/profile", func(e *core.RequestEvent) error {
		record := e.Auth
		user := record.Get(migrations.FieldWeAuthinfo).(map[string]any)
		return e.JSON(http.StatusOK, user)
	})

}
