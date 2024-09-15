// Package egothic is a modified version of original gothic package for the Echo server.
// The original gothic package is a wrapper for the Goth library.
// This package is based on https://github.com/markbates/goth/blob/edc3e96387cb58c3f3d58e70db2f115815ccdf1e/gothic/gothic.go
package egothic

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

const (
	stateQueryParam  = "state"
	providerURLParam = "provider"
	postHTTPMethod   = "POST"
)

// SetStore sets the store for the gothic session.
func SetStore(store sessions.Store) {
	gothic.Store = store
}

// Store returns the store for the gothic session.
func Store() sessions.Store {
	return gothic.Store
}

/*
BeginAuthHandler is a convenience handler for starting the authentication process.
It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".

BeginAuthHandler will redirect the user to the appropriate authentication end-point
for the requested provider.
*/
func BeginAuthHandler(e echo.Context) error {
	url, err := GetAuthURL(e)
	if err != nil {
		return err
	}
	return e.Redirect(http.StatusTemporaryRedirect, url)
}

// SetState sets the state string associated with the given request.
// If no state string is associated with the request, one will be generated.
// This state is sent to the provider and can be retrieved during the
// callback.
var SetState = func(e echo.Context) string {
	return gothic.SetState(e.Request())
}

// GetState gets the state returned by the provider during the callback.
// This is used to prevent CSRF attacks, see
// http://tools.ietf.org/html/rfc6749#section-10.12
var GetState = func(e echo.Context) string {
	return gothic.GetState(e.Request())
}

/*
GetAuthURL starts the authentication process with the requested provided.
It will return a URL that should be used to send users to.

It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".

I would recommend using the BeginAuthHandler instead of doing all of these steps
yourself, but that's entirely up to you.
*/
func GetAuthURL(e echo.Context) (string, error) {

	// get the provider name
	providerName, err := GetProviderName(e)
	if err != nil {
		return "", err
	}

	// get the provider
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return "", err
	}

	// begin the authentication process
	sess, err := provider.BeginAuth(SetState(e))
	if err != nil {
		return "", err
	}

	// get the auth URL
	url, err := sess.GetAuthURL()
	if err != nil {
		return "", err
	}

	// store the session data
	err = StoreInSession(e, providerName, sess.Marshal())
	if err != nil {
		return "", err
	}
	return url, err
}

/*
CompleteUserAuth does what it says on the tin. It completes the authentication
process and fetches all of the basic information about the user from the provider.

It expects to be able to get the name of the provider from the query parameters
as either "provider" or ":provider".
*/
var CompleteUserAuth = func(e echo.Context) (goth.User, error) {

	// ensure that the user is logged out after the request
	defer func() {
		// TODO: log?
		_ = Logout(e)
	}()

	// get the provider name
	providerName, err := GetProviderName(e)
	if err != nil {
		return goth.User{}, err
	}

	// get the provider
	provider, err := goth.GetProvider(providerName)
	if err != nil {
		return goth.User{}, err
	}

	// get the session data
	value, err := GetFromSession(e, providerName)
	if err != nil {
		return goth.User{}, err
	}

	// unmarshal the session data
	sess, err := provider.UnmarshalSession(value)
	if err != nil {
		return goth.User{}, err
	}

	// validate the state token
	err = validateState(e, sess)
	if err != nil {
		return goth.User{}, err
	}

	// fetch the user
	user, err := provider.FetchUser(sess)
	if err == nil {
		// user can be found with existing session data
		return user, err
	}

	// get the query parameters from the request
	params := e.Request().URL.Query()

	// if the request is a POST, parse the form data
	if params.Encode() == "" && e.Request().Method == postHTTPMethod {
		err = e.Request().ParseForm()
		if err != nil {
			return goth.User{}, err
		}
		params = e.Request().Form
	}

	// get new token and retry fetch
	_, err = sess.Authorize(provider, params)
	if err != nil {
		return goth.User{}, err
	}

	// store the new session data
	if err = StoreInSession(e, providerName, sess.Marshal()); err != nil {
		return goth.User{}, err
	}

	// fetch the user
	gu, err := provider.FetchUser(sess)
	return gu, err
}

// validateState ensures that the state token param from the original
// AuthURL matches the one included in the current (callback) request.
func validateState(e echo.Context, sess goth.Session) error {

	// get the original auth URL
	rawAuthURL, err := sess.GetAuthURL()
	if err != nil {
		return err
	}

	// parse the original auth URL
	authURL, err := url.Parse(rawAuthURL)
	if err != nil {
		return err
	}

	// get the state token from the current request
	reqState := GetState(e)

	// get the state token from the original auth URL
	originalState := authURL.Query().Get(stateQueryParam)

	// ensure that the state tokens match
	if originalState != "" && (originalState != reqState) {
		return errors.New("state token mismatch")
	}
	return nil
}

// Logout invalidates a user session.
func Logout(e echo.Context) error {
	return gothic.Logout(e.Response(), e.Request())
}

// GetProviderName is a function used to get the name of a provider
// for a given request. By default, this provider is fetched from
// the URL query string. If you provide it in a different way,
// assign your own function to this variable that returns the provider
// name for your request.
var GetProviderName = getProviderName

func getProviderName(e echo.Context) (string, error) {
	if p := e.Param(providerURLParam); p != "" {
		return p, nil
	}
	return gothic.GetProviderName(e.Request())
}

// StoreInSession stores a specified key/value pair in the session.
func StoreInSession(e echo.Context, key string, value string) error {
	return gothic.StoreInSession(key, value, e.Request(), e.Response())
}

// GetFromSession retrieves a previously-stored value from the session.
// If no value has previously been stored at the specified key, it will return an error.
func GetFromSession(e echo.Context, key string) (string, error) {
	return gothic.GetFromSession(key, e.Request())
}
