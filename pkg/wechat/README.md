# WeChat Integration Package

A comprehensive WeChat integration package for PocketBase applications, providing OAuth authentication, user management, and template message functionality.

## Features

- **OAuth Authentication**: Complete WeChat OAuth 2.0 flow implementation
- **User Management**: Automatic user creation and management
- **Template Messages**: Send WeChat template messages
- **Token Management**: Automatic access token refresh and caching
- **Signature Verification**: WeChat signature verification for webhooks
- **PocketBase Integration**: Seamless integration with PocketBase authentication system

## Installation

```bash
go get github.com/sospartan/pb-toolkit/pkg/wechat
```

## Quick Start

### 1. Initialize WeChat Auth

```go
import "github.com/sospartan/pb-toolkit/pkg/wechat"

// Create WeChat auth instance
auth := &wechat.WechatAuth{
    AppID:  "your_app_id",
    Secret: "your_app_secret",
}
```

### 2. Implement AuthHandler Interface

```go
type MyAuthHandler struct {
    App *pocketbase.PocketBase
    Auth *wechat.WechatAuth
}

func (h *MyAuthHandler) FindAuthRecordByCode(code string) (*core.Record, error) {
    // Find existing user by WeChat code
    // Return NoAuthRecordError if not found
    return nil, wechat.NoAuthRecordError
}

func (h *MyAuthHandler) SaveSignIn(token *wechat.AccessTokenResponse, info *wechat.UserInfoResponse, code string) (*core.Record, error) {
    // Save or update user record with WeChat data
    // Return the created/updated record
    return record, nil
}

func (h *MyAuthHandler) GetAuthConfig() *wechat.WechatAuth {
    return h.Auth
}

func (h *MyAuthHandler) ModifyAuthRecord(record *core.Record) error {
    // Modify record before sending response
    // Clean sensitive fields if needed
    return nil
}
```

### 3. Set Up Routes

```go
app.OnServe().BindFunc(func(se *core.ServeEvent) error {
    // WeChat OAuth callback route
    se.Router.GET("/auth/wechat/callback", wechat.HandleAuthResponseWithCode(authHandler))
    
    // WeChat webhook verification (optional)
    se.Router.GET("/webhook/wechat", func(c echo.Context) error {
        signature := c.QueryParam("signature")
        timestamp := c.QueryParam("timestamp")
        nonce := c.QueryParam("nonce")
        
        if auth.VerifySignature(signature, timestamp, nonce) {
            return c.String(http.StatusOK, c.QueryParam("echostr"))
        }
        return c.String(http.StatusForbidden, "Invalid signature")
    })
    
    return se.Next()
})
```

## API Reference

### WechatAuth

#### OAuth Authentication

```go
// Fetch access token using authorization code
token, err := auth.FetchAccessToken(code)

// Refresh access token
newToken, err := auth.RefreshAccessToken(refreshToken)

// Fetch user information
userInfo, err := auth.FetchUserInfo(token.AccessToken, token.OpenID)
```

#### Template Messages

```go
// Send template message
data := map[string]wechat.TemplateMessageData{
    "first":    {Value: "Hello!"},
    "keyword1": {Value: "Order #12345"},
    "keyword2": {Value: "Processing"},
    "remark":   {Value: "Thank you for your order"},
}

err := auth.SendTemplateMessage(
    "user_openid",
    "template_id",
    "https://example.com",
    data,
)
```

#### Signature Verification

```go
// Verify WeChat signature
isValid := auth.VerifySignature(signature, timestamp, nonce)
```

### Response Types

#### AccessTokenResponse

```go
type AccessTokenResponse struct {
    AccessToken    string `json:"access_token"`
    ExpiresIn      int    `json:"expires_in"`
    RefreshToken   string `json:"refresh_token"`
    OpenID         string `json:"openid"`
    Scope          string `json:"scope"`
    IsSnapshotUser int    `json:"is_snapshotuser"`
    UnionID        string `json:"unionid"`
}
```

#### UserInfoResponse

```go
type UserInfoResponse struct {
    OpenID     string   `json:"openid"`
    Nickname   string   `json:"nickname"`
    Sex        int      `json:"sex"`
    Province   string   `json:"province"`
    City       string   `json:"city"`
    Country    string   `json:"country"`
    HeadImgURL string   `json:"headimgurl"`
    Privilege  []string `json:"privilege"`
    UnionID    string   `json:"unionid"`
}
```

## Usage Examples

### Complete Authentication Flow

```go
package main

import (
    "log"
    "github.com/sospartan/pb-toolkit/pkg/wechat"
    "github.com/pocketbase/pocketbase"
    "github.com/pocketbase/pocketbase/core"
)

type WeChatAuthHandler struct {
    App  *pocketbase.PocketBase
    Auth *wechat.WechatAuth
}

func (h *WeChatAuthHandler) FindAuthRecordByCode(code string) (*core.Record, error) {
    // Check if user already exists with this code
    records, err := h.App.Dao().FindRecordsByFilter("users", "wechat_code = ?", code)
    if err != nil {
        return nil, err
    }
    if len(records) == 0 {
        return nil, wechat.NoAuthRecordError
    }
    return records[0], nil
}

func (h *WeChatAuthHandler) SaveSignIn(token *wechat.AccessTokenResponse, info *wechat.UserInfoResponse, code string) (*core.Record, error) {
    // Create or update user record
    collection, err := h.App.Dao().FindCollectionByNameOrId("users")
    if err != nil {
        return nil, err
    }
    
    record := core.NewRecord(collection)
    record.Set("wechat_openid", info.OpenID)
    record.Set("wechat_nickname", info.Nickname)
    record.Set("wechat_avatar", info.HeadImgURL)
    record.Set("wechat_code", code)
    record.Set("wechat_unionid", info.UnionID)
    
    if err := h.App.Dao().SaveRecord(record); err != nil {
        return nil, err
    }
    
    return record, nil
}

func (h *WeChatAuthHandler) GetAuthConfig() *wechat.WechatAuth {
    return h.Auth
}

func (h *WeChatAuthHandler) ModifyAuthRecord(record *core.Record) error {
    // Remove sensitive fields before sending to client
    record.Unset("wechat_code")
    return nil
}

func main() {
    app := pocketbase.New()
    
    auth := &wechat.WechatAuth{
        AppID:  "your_wechat_app_id",
        Secret: "your_wechat_app_secret",
    }
    
    authHandler := &WeChatAuthHandler{
        App:  app,
        Auth: auth,
    }
    
    app.OnServe().BindFunc(func(se *core.ServeEvent) error {
        se.Router.GET("/auth/wechat/callback", wechat.HandleAuthResponseWithCode(authHandler))
        return se.Next()
    })
    
    if err := app.Start(); err != nil {
        log.Fatal(err)
    }
}
```

### Building Authorization URL

```go
// Build WeChat authorization URL
authURL := wechat.BuildAuthUrl(
    "your_app_id",
    "https://your-domain.com/auth/wechat/callback",
    "snsapi_userinfo", // or "snsapi_base"
    "state_parameter",
)
```

### Sending Template Messages

```go
// Example: Order notification
func sendOrderNotification(auth *wechat.WechatAuth, openid, orderID string) error {
    data := map[string]wechat.TemplateMessageData{
        "first":    {Value: "您的订单状态已更新"},
        "keyword1": {Value: orderID},
        "keyword2": {Value: "已发货"},
        "keyword3": {Value: time.Now().Format("2006-01-02 15:04:05")},
        "remark":   {Value: "感谢您的购买！"},
    }
    
    return auth.SendTemplateMessage(
        openid,
        "your_template_id",
        "https://your-domain.com/orders/" + orderID,
        data,
    )
}
```

## Configuration

### WeChat App Configuration

1. Register your application at [WeChat Open Platform](https://open.weixin.qq.com/)
2. Configure your domain in the app settings
3. Set up template messages if needed
4. Configure OAuth redirect URI

### Environment Variables

```bash
WECHAT_APP_ID=your_app_id
WECHAT_APP_SECRET=your_app_secret
```

## Error Handling

The package provides comprehensive error handling:

```go
// Handle authentication errors
if err != nil {
    switch err {
    case wechat.NoAuthRecordError:
        // Handle new user registration
    default:
        // Handle other errors
        log.Printf("Authentication error: %v", err)
    }
}
```

## Best Practices

1. **Token Caching**: The package automatically caches API tokens for efficiency
2. **Error Handling**: Always handle authentication errors gracefully
3. **User Data**: Store only necessary user information
4. **Security**: Verify signatures for webhook endpoints
5. **Rate Limiting**: Respect WeChat API rate limits
6. **Template Messages**: Use appropriate template message formats

## Dependencies

- `github.com/pocketbase/pocketbase` - PocketBase integration
- Standard Go libraries for HTTP, JSON, and crypto operations

## License

This package is part of the PocketBase Toolkit and is licensed under the MIT License. 