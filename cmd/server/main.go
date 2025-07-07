package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/sospartan/pb-toolkit/cmd/server/migrations"
	_ "github.com/sospartan/pb-toolkit/cmd/server/migrations" // import migrations
	"github.com/sospartan/pb-toolkit/pkg/rpc"
	"github.com/sospartan/pb-toolkit/pkg/wechat"

	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/plugins/migratecmd"
)

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

	appID := os.Getenv("WECHAT_APP_ID")
	appSecret := os.Getenv("WECHAT_APP_SECRET")
	if appID == "" || appSecret == "" {
		log.Fatal("WECHAT_APP_ID and WECHAT_APP_SECRET must be set")
	}
	wechatHandler := NewWechatAuthHandler(app, appID, appSecret)

	// Add RPC routes to PocketBase
	app.OnServe().BindFunc(func(se *core.ServeEvent) error {

		// Path-based RPC endpoint
		g := se.Router.Group("/rpc")

		// g.Bind(apis.RequireAuth("users"))
		rpcServer.Bind(g)

		// redirect to wechat auth url
		se.Router.GET("/redirect-wechat-auth", func(e *core.RequestEvent) error {
			// Construct the WeChat OAuth2 URL
			redirectURI := "https://localhost:8099/w/callback" // adjust this to your domain
			authURL := wechat.BuildAuthUrl(wechatHandler.GetAuthConfig().AppID, redirectURI, "snsapi_userinfo", "STATE")

			return e.Redirect(http.StatusTemporaryRedirect, authURL)
		})
		w := se.Router.Group("/w")

		//callback with code
		w.GET("/callback", wechat.HandleAuthResponseWithCode(wechatHandler))

		// after auth, get user info as regular record auth
		authed := w.Group("/authed")
		authed.Bind(apis.RequireAuth(migrations.CollectionNameWechatAuth))
		authed.GET("/profile", func(e *core.RequestEvent) error {
			record := e.Auth
			user := record.Get(migrations.FieldWeAuthinfo).(map[string]any)
			return e.JSON(http.StatusOK, user)
		})

		return se.Next()
	})

	if err := app.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println("Server Stopped")
}
