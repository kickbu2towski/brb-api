package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/justinas/alice"
)

func (app *application) routes() http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	corsMw := alice.New(app.logRequest, app.enableCORS)
	authMw := alice.New(app.isAuthenticated)

	// authentication
	router.HandlerFunc(http.MethodGet, "/v1/auth/redirectURL", app.getRedirectURLHandler)
	router.HandlerFunc(http.MethodGet, "/v1/auth/callback", app.callbackHandler)
	router.Handler(http.MethodDelete, "/v1/auth/logout", authMw.Then(http.HandlerFunc(app.logoutHandler)))

	// users
	router.Handler(http.MethodGet, "/v1/users", authMw.Then(http.HandlerFunc(app.getUsersHandler)))
	router.Handler(http.MethodGet, "/v1/users/:userID", authMw.Then(http.HandlerFunc(app.getUserHandler)))
	router.Handler(http.MethodPost, "/v1/users/:userID/follow", authMw.Then(http.HandlerFunc(app.followUserHandler)))
	router.Handler(http.MethodDelete, "/v1/users/:userID/unfollow", authMw.Then(http.HandlerFunc(app.unfollowUserHandler)))

	// messages
	router.Handler(http.MethodGet, "/v1/messages", authMw.Then(http.HandlerFunc(app.getMessagesHandler)))

	// dms
	router.Handler(http.MethodPost, "/v1/dms", authMw.Then(http.HandlerFunc(app.createDMHandler)))

	// logged in user routes
	router.Handler(http.MethodGet, "/v1/me", authMw.Then(http.HandlerFunc(app.getLoggedInUserHandler)))
	router.Handler(http.MethodGet, "/v1/me/following", authMw.Then(http.HandlerFunc(app.getUsersForRelationHandler)))
	router.Handler(http.MethodGet, "/v1/me/friends", authMw.Then(http.HandlerFunc(app.getUsersForRelationHandler)))
	router.Handler(http.MethodGet, "/v1/me/followers", authMw.Then(http.HandlerFunc(app.getUsersForRelationHandler)))
	router.Handler(http.MethodGet, "/v1/me/dms", authMw.Then(http.HandlerFunc(app.getUserDMList)))

	// websocket
	router.Handler(http.MethodGet, "/ws", authMw.Then(http.HandlerFunc(app.wsHandler)))

	return corsMw.Then(router)
}
