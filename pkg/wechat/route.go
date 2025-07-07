// Package wechat provides WeChat OAuth authentication, user management, and template message functionality
// for PocketBase applications.
package wechat

import (
	"database/sql"
	"errors"
	"log"
	"net/http"

	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
)

// NoAuthRecordError is returned when no authentication record is found for a given code
var NoAuthRecordError = errors.New("no auth record found")

// AuthHandler defines the interface for handling WeChat authentication
// Implement this interface to customize authentication behavior
type AuthHandler interface {
	// FindAuthRecordByCode finds an existing authentication record by WeChat code
	// Return NoAuthRecordError if no record is found
	FindAuthRecordByCode(code string) (*core.Record, error)

	// Save saves or updates user authentication data
	// This method should create or update a user record with WeChat information
	Save(token *AccessTokenResponse, info *UserInfoResponse, code string) (*core.Record, error)

	// GetAuthConfig returns the WeChat authentication configuration
	GetAuthConfig() *WechatAuth

	// ModifyAuthRecord allows modification of the auth record before sending response
	// Use this to clean sensitive fields or add custom data
	ModifyAuthRecord(record *core.Record) error
}

// HandleAuthResponseWithCode creates a handler function for WeChat OAuth callback
// This function processes the authorization code from WeChat and returns a PocketBase auth response
// Parameters:
//   - store: An implementation of AuthHandler interface
//
// Returns a function that can be used as a PocketBase route handler
func HandleAuthResponseWithCode(store AuthHandler) func(e *core.RequestEvent) error {
	return func(e *core.RequestEvent) error {
		// Extract the authorization code from the request
		code := e.Request.URL.Query().Get("code")

		// Exchange the code for user information
		authRecord, err := exchangeWechatAuthInfo(store, code)
		if err != nil {
			log.Printf("fetch wechat user failed,%v \n", err)
			return e.JSON(http.StatusBadRequest, errors.New("fetch wechat user failed"))
		}

		// Allow modification of the auth record before response
		store.ModifyAuthRecord(authRecord)

		// Return PocketBase authentication response
		return apis.RecordAuthResponse(e, authRecord, "", nil)
	}
}

// exchangeWechatAuthInfo exchanges the authorization code for user information and saves it
// This function handles the complete OAuth flow including token exchange and user data retrieval
// Parameters:
//   - store: An implementation of AuthHandler interface
//   - code: The authorization code from WeChat
//
// Returns the user record or an error if the process fails
func exchangeWechatAuthInfo(store AuthHandler, code string) (*core.Record, error) {
	// Validate that code is provided
	if code == "" {
		log.Printf("code not found")
		return nil, errors.New("code required")
	}

	auth := store.GetAuthConfig()

	// Check if user already exists with this code (avoid re-signin with same code)
	record, err := store.FindAuthRecordByCode(code)
	if err != nil && err != NoAuthRecordError && err != sql.ErrNoRows {
		log.Printf("find user by code failed,%v \n", err)
		return nil, errors.New("find user by code failed")
	}

	// Return cached user if found (avoid re-signin with same code)
	if record != nil {
		return record, nil
	}

	// Exchange code for access token
	token, err := auth.FetchAccessToken(code)
	if err != nil {
		log.Printf("request wechat api failed,%v \n", err)
		return nil, errors.New("request wechat api failed")
	}

	// Fetch user information using the access token
	userInfo, err := auth.FetchUserInfo(token.AccessToken, token.OpenID)
	if err != nil {
		log.Printf("request wechat api failed,%v \n", err)
		return nil, errors.New("request wechat api failed")
	}

	// Save or update user authentication data
	record, err = store.Save(token, userInfo, code)
	if err != nil {
		log.Printf("upsert wechat user failed,%v \n", err)
		return nil, errors.New("upsert wechat user failed")
	}
	return record, nil
}
