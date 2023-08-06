package main

import (
	"context"
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/kickbu2towski/brb-api/internal/data"
)

func (app *application) getUsersHandler(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		app.badRequestResponse(w, r, "invalid query param: username")
		return
	}

	user := app.getUserContext(r)
	users, err := app.models.Users.GetUsers(context.Background(), user.ID, username)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"users": users}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) followUserHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	followingID := params.ByName("userID")
	if followingID == "" {
		app.badRequestResponse(w, r, "missing required param: userID")
		return
	}

	user := app.getUserContext(r)
	err := app.models.Users.FollowUser(context.Background(), followingID, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "followed user successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) unfollowUserHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	followingID := params.ByName("userID")
	if followingID == "" {
		app.badRequestResponse(w, r, "missing required param: userID")
		return
	}
	user := app.getUserContext(r)
	err := app.models.Users.UnfollowUser(context.Background(), followingID, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "unfollowed user successfully"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUsersForRelationHandler(w http.ResponseWriter, r *http.Request) {
	s := strings.Split(r.URL.Path, "/")
	var relation data.Relation
	switch s[len(s)-1] {
	case "friends":
		relation = data.RelationFriends
	case "following":
		relation = data.RelationFollowing
	case "followers":
		relation = data.RelationFollowers
	default:
		app.badRequestResponse(w, r, "invalid relation path")
		return
	}

	user := app.getUserContext(r)
	users, err := app.models.Users.GetUsersForRelation(context.Background(), relation, user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"users": users}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserHandler(w http.ResponseWriter, r *http.Request) {
	params := httprouter.ParamsFromContext(r.Context())
	userID := params.ByName("userID")

	if userID == "" {
		app.badRequestResponse(w, r, "invalid userID param")
		return
	}

	user, err := app.models.Users.GetUser(context.Background(), userID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"user": user}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) getUserDMList(w http.ResponseWriter, r *http.Request) {
	user := app.getUserContext(r)
	users, err := app.models.DMs.GetDMListForUser(context.Background(), user.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	err = app.writeJSON(w, http.StatusOK, envelope{"users": users}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
