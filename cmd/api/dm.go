package main

import (
	"context"
	"net/http"
)

// createDMHandler creates the dm if not exists and returs the dm ID
func (app *application) createDMHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	var input struct {
		Participants []int `json:"participants"`
	}

	err := app.readJSON(r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	if len(input.Participants) != 2 {
		app.badRequestResponse(w, r, "incorrect participants length")
		return
	}

	isFriends, err := app.models.Users.IsFriends(ctx, input.Participants)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !isFriends {
		app.badRequestResponse(w, r, "participants should be friends")
		return
	}

	dmID, err := app.models.DMs.GetDMForParticipants(ctx, input.Participants)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"dm_id": dmID}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
