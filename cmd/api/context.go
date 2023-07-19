package main

import (
	"context"
	"net/http"

	"github.com/kickbu2towski/brb-api/internal/data"
)

type contextKey string

const userContextKey = contextKey("user")

func (app *application) setUserContext(r *http.Request, u *data.User) *http.Request {
	ctx := context.WithValue(r.Context(), userContextKey, u)
	return r.WithContext(ctx)
}

func (app *application) getUserContext(r *http.Request) *data.User {
	u, ok := r.Context().Value(userContextKey).(*data.User)
	if !ok {
		panic("missing required user context")
	}
	return u
}
