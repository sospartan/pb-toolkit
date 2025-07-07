// Package wechat provides WeChat OAuth authentication, user management, and template message functionality
// for PocketBase applications.
package wechat

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// WeChat API endpoints
const (
	// OAuth access token endpoint
	access_token_url = "https://api.weixin.qq.com/sns/oauth2/access_token"
	// OAuth refresh token endpoint
	refresh_token_url = "https://api.weixin.qq.com/sns/oauth2/refresh_token"
	// User info endpoint
	user_info_url = "https://api.weixin.qq.com/sns/userinfo"
	// API access token endpoint (for template messages)
	api_token_url = "https://api.weixin.qq.com/cgi-bin/token"
	// Template message sending endpoint
	send_template_message_url = "https://api.weixin.qq.com/cgi-bin/message/template/send"
)

// WechatAuth represents a WeChat application configuration with AppID and Secret
type WechatAuth struct {
	AppID  string // WeChat application ID
	Secret string // WeChat application secret
}

// VerifySignature verifies the WeChat signature for webhook validation
// Parameters:
//   - signature: The signature from WeChat
//   - timestamp: The timestamp from WeChat
//   - nonce: The nonce from WeChat
//
// Returns true if the signature is valid, false otherwise
func (w *WechatAuth) VerifySignature(signature, timestamp, nonce string) bool {
	tmpArr := []string{w.Secret, timestamp, nonce}
	sort.Strings(tmpArr)
	h := sha1.New()
	h.Write([]byte(strings.Join(tmpArr, "")))
	result := fmt.Sprintf("%x", h.Sum(nil))
	return result == signature
}

// AccessTokenResponse represents the response from WeChat OAuth access token request
type AccessTokenResponse struct {
	AccessToken    string `json:"access_token"`    // OAuth access token
	ExpiresIn      int    `json:"expires_in"`      // Token expiration time in seconds
	RefreshToken   string `json:"refresh_token"`   // Refresh token for getting new access token
	OpenID         string `json:"openid"`          // User's unique identifier
	Scope          string `json:"scope"`           // OAuth scope
	IsSnapshotUser int    `json:"is_snapshotuser"` // Whether user is a snapshot user
	UnionID        string `json:"unionid"`         // Union ID for cross-platform identification
}

// RefreshTokenResponse represents the response from WeChat refresh token request
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`  // New OAuth access token
	ExpiresIn    int    `json:"expires_in"`    // Token expiration time in seconds
	RefreshToken string `json:"refresh_token"` // New refresh token
	OpenID       string `json:"openid"`        // User's unique identifier
	Scope        string `json:"scope"`         // OAuth scope
}

// UserInfoResponse represents the response from WeChat user info request
type UserInfoResponse struct {
	OpenID     string   `json:"openid"`     // User's unique identifier
	Nickname   string   `json:"nickname"`   // User's nickname
	Sex        int      `json:"sex"`        // User's gender (1=male, 2=female, 0=unknown)
	Province   string   `json:"province"`   // User's province
	City       string   `json:"city"`       // User's city
	Country    string   `json:"country"`    // User's country
	HeadImgURL string   `json:"headimgurl"` // User's avatar URL
	Privilege  []string `json:"privilege"`  // User's privileges
	UnionID    string   `json:"unionid"`    // Union ID for cross-platform identification
}

// FetchAccessToken exchanges the authorization code for an access token
// Parameters:
//   - code: The authorization code from WeChat OAuth callback
//
// Returns the access token response or an error if the request fails
func (w *WechatAuth) FetchAccessToken(code string) (*AccessTokenResponse, error) {
	url := fmt.Sprintf("%s?appid=%s&secret=%s&code=%s&grant_type=authorization_code", access_token_url, w.AppID, w.Secret, code)

	// Send GET request to WeChat API
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request wechat api failed: %s", resp.Status)
	}

	// Parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var r AccessTokenResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}

	if r.AccessToken == "" {
		fmt.Println("Response body:", string(body))
		return nil, fmt.Errorf("request wechat api failed: token is empty")
	}

	return &r, nil
}

// RefreshAccessToken refreshes an expired access token using the refresh token
// Parameters:
//   - refreshToken: The refresh token from previous OAuth flow
//
// Returns the new access token response or an error if the request fails
func (w *WechatAuth) RefreshAccessToken(refreshToken string) (*RefreshTokenResponse, error) {
	url := fmt.Sprintf("%s?appid=%s&grant_type=refresh_token&refresh_token=%s", refresh_token_url, w.AppID, refreshToken)

	// Send GET request to WeChat API
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request wechat api failed: %s", resp.Status)
	}

	// Parse response body
	var r RefreshTokenResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	if r.AccessToken == "" {
		return nil, fmt.Errorf("request wechat api failed: token is empty")
	}

	return &r, nil
}

// FetchUserInfo retrieves user information using access token and openid
// Parameters:
//   - token: The OAuth access token
//   - openid: The user's OpenID
//
// Returns the user info response or an error if the request fails
func (w *WechatAuth) FetchUserInfo(token, openid string) (*UserInfoResponse, error) {
	url := fmt.Sprintf("%s?access_token=%s&openid=%s&lang=zh_CN", user_info_url, token, openid)

	// Send GET request to WeChat API
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request wechat api failed: %s", resp.Status)
	}

	// Parse response body
	var r UserInfoResponse
	err = json.NewDecoder(resp.Body).Decode(&r)
	if err != nil {
		return nil, err
	}

	if r.OpenID == "" {
		return nil, fmt.Errorf("request wechat api failed: openid is empty")
	}

	return &r, nil
}

// ApiTokenResponse represents the response from WeChat API token request
// This token is used for sending template messages and other API operations
type ApiTokenResponse struct {
	AccessToken string `json:"access_token"` // API access token
	ExpiresIn   int    `json:"expires_in"`   // Token expiration time in seconds
}

// Global variables for API token caching
var (
	apiTokenCache     *ApiTokenResponse // Cached API token
	apiTokenCacheTime time.Time         // Time when token was cached
)

// getApiToken returns a cached API token or fetches a new one if needed
// The token is cached for efficiency and automatically refreshed when expired
// Returns the API token or an error if the request fails
func (w *WechatAuth) getApiToken() (*ApiTokenResponse, error) {

	// Check if cache exists and is still valid (with 60 seconds buffer)
	if apiTokenCache != nil {
		if time.Since(apiTokenCacheTime).Seconds() < float64(apiTokenCache.ExpiresIn-60) {
			return apiTokenCache, nil
		}
	}

	// Fetch new token from WeChat API
	token, err := w.FetchApiToken()
	if err != nil {
		return nil, err
	}

	// Update cache with new token
	apiTokenCache = token
	apiTokenCacheTime = time.Now()

	return token, nil
}

// FetchApiToken fetches a new API token from WeChat using client credentials
// This token is used for sending template messages and other API operations
// Returns the API token response or an error if the request fails
func (w *WechatAuth) FetchApiToken() (*ApiTokenResponse, error) {
	url := fmt.Sprintf("%s?grant_type=client_credential&appid=%s&secret=%s", api_token_url, w.AppID, w.Secret)

	// Send GET request to WeChat API
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request wechat api failed: %s", resp.Status)
	}

	// Parse response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body failed: %v", err)
	}

	var r ApiTokenResponse
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, fmt.Errorf("decode response body failed: %v, body: %s", err, string(body))
	}

	if r.AccessToken == "" {
		return nil, fmt.Errorf("request wechat api failed: token is empty, body: %s", string(body))
	}

	return &r, nil
}

// TemplateMessageResponse represents the response from WeChat template message API
type TemplateMessageResponse struct {
	Errcode int    `json:"errcode"` // Error code (0 means success)
	Errmsg  string `json:"errmsg"`  // Error message
	MsgID   int    `json:"msgid"`   // Message ID if successful
}

// TemplateMessageRequest represents the request to send a template message
type TemplateMessageRequest struct {
	ToUser     string                         `json:"touser"`      // Recipient's OpenID
	TemplateID string                         `json:"template_id"` // Template message ID
	URL        string                         `json:"url"`         // URL to open when message is clicked
	Data       map[string]TemplateMessageData `json:"data"`        // Template data
}

// TemplateMessageData represents a single data field in a template message
type TemplateMessageData struct {
	Value string `json:"value"` // The value to display in the template
}

// SendTemplateMessage sends a template message to a WeChat user
// Parameters:
//   - openid: The recipient's OpenID
//   - templateId: The template message ID
//   - url: The URL to open when the message is clicked
//   - data: The template data to fill in the message
//
// Returns an error if the message sending fails
func (w *WechatAuth) SendTemplateMessage(openid, templateId, url string, data map[string]TemplateMessageData) error {
	// Get API token (cached or fresh)
	apiToken, err := w.getApiToken()
	if err != nil {
		return err
	}
	reqURL := fmt.Sprintf("%s?access_token=%s", send_template_message_url, apiToken.AccessToken)

	// Prepare the message request
	msg := TemplateMessageRequest{
		ToUser:     openid,
		TemplateID: templateId,
		URL:        url,
		Data:       data,
	}

	// Marshal the request to JSON
	jsonData, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	// Send POST request to WeChat API
	resp, err := http.Post(reqURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request wechat api failed: %s", resp.Status)
	}

	// Parse the response
	var result TemplateMessageResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return err
	}

	// Check for API errors
	if result.Errcode != 0 {
		return fmt.Errorf("send template message failed: %s", result.Errmsg)
	}

	return nil
}

// BuildAuthUrl builds the WeChat OAuth authorization URL
// Parameters:
//   - appID: The WeChat application ID
//   - redirectURI: The callback URL after authorization
//   - scope: The OAuth scope (e.g., "snsapi_base" or "snsapi_userinfo")
//   - state: A state parameter for security
//
// Returns the complete authorization URL
func BuildAuthUrl(appID, redirectURI, scope, state string) string {
	return fmt.Sprintf(
		"https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s#wechat_redirect",
		appID,
		url.QueryEscape(redirectURI),
		scope,
		state,
	)
}
