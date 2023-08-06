package main

import (
	"context"
	"net/http"
	"strconv"
)

func (app *application) getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	dmID, err := strconv.Atoi(r.FormValue("dm_id"))
	if err != nil {
		app.badRequestResponse(w, r, "invalid query param: dm_id")
		return
	}

	messages, err := app.models.Messages.GetMessages(context.Background(), dmID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"messages": messages}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
