package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/kickbu2towski/brb-api/internal/data"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var oauthConfig = oauth2.Config{
	Endpoint: google.Endpoint,
	Scopes: []string{
		"https://www.googleapis.com/auth/userinfo.profile",
		"https://www.googleapis.com/auth/userinfo.email",
	},
}

const userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

func (app *application) getLoggedInUserHandler(w http.ResponseWriter, r *http.Request) {
	u := app.getUserContext(r)
	app.writeJSON(w, http.StatusOK, envelope{"user": u}, nil)
}

func (app *application) getRedirectURLHandler(w http.ResponseWriter, r *http.Request) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	oauthState := base64.StdEncoding.EncodeToString(b)
	cookie := &http.Cookie{
		Name:     "oauthState",
		Value:    oauthState,
		Secure:   true,
		HttpOnly: true,
		Expires:  time.Now().Add(5 * time.Minute),
		Path:     "/",
	}
	http.SetCookie(w, cookie)

	oauthConfig.ClientID = app.config.google.clientID
	oauthConfig.ClientSecret = app.config.google.clientSecret
	oauthConfig.RedirectURL = app.config.google.redirectURL
	redirectURL := oauthConfig.AuthCodeURL(oauthState)

	err = app.writeJSON(w, http.StatusOK, envelope{"redirectURL": redirectURL}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) callbackHandler(w http.ResponseWriter, r *http.Request) {
	state, code := r.FormValue("state"), r.FormValue("code")
	cookie, err := r.Cookie("oauthState")

	if err != nil {
		switch {
		case errors.Is(err, http.ErrNoCookie):
			app.badRequestResponse(w, r, "missing required cookie: oauthState")
		default:
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if cookie.Value != state {
		app.badRequestResponse(w, r, "invalid cookie found: oauthState")
		return
	}

	token, err := oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	res, err := http.Get(userInfoURL + token.AccessToken)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		app.serverErrorResponse(w, r, err)
		return
	}

	var userInfo struct {
		ID            string `json:"id"`
		Name          string `json:"name"`
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Picture       string `json:"picture"`
	}

	err = json.NewDecoder(res.Body).Decode(&userInfo)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	user := &data.User{
		ID:            userInfo.ID,
		Username:      userInfo.Name,
		Email:         userInfo.Email,
		EmailVerified: userInfo.VerifiedEmail,
		Avatar:        userInfo.Picture,
	}

	err = app.models.Users.AddUser(context.Background(), user)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	t, err := data.NewToken(user.ID, 24*time.Hour, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.models.Tokens.Insert(context.Background(), t)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	sessionCookie := &http.Cookie{
		Name:     "sessionID",
		Value:    t.PlainText,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		Expires:  t.ExpiryTime,
		SameSite: http.SameSiteNoneMode,
	}
	http.SetCookie(w, sessionCookie)

	http.Redirect(w, r, app.config.webURL, http.StatusTemporaryRedirect)
}

func (app *application) logoutHandler(w http.ResponseWriter, r *http.Request) {
	user := app.getUserContext(r)

	err := app.models.Tokens.DeleteForUser(context.Background(), user.ID, data.ScopeAuthentication)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	sessionCookie := &http.Cookie{
		Name:     "sessionID",
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
		Expires:  time.Unix(0, 0),
	}
	http.SetCookie(w, sessionCookie)

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "logged out successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
