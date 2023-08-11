package main

import (
	"context"
	"net/http"
)

func (app *application) createDMHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Participants []int `json:"participants"`
	}

	err := app.readJSON(r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	dmID, err := app.models.DMs.GetDMForParticipants(context.Background(), input.Participants)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"dm_id": dmID}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
