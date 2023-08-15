package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/kickbu2towski/brb-api/internal/data"
	"github.com/livekit/protocol/auth"
	"github.com/livekit/protocol/livekit"
	"github.com/livekit/protocol/webhook"
)

const ROOM_EMPTY_TIMEOUT = 300

func (app *application) GetRoomsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	res, err := app.lkRoomSvc.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	rooms, err := app.GetRoomsFromLKRooms(ctx, res.Rooms)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"rooms": rooms}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) GetRoomHandler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	lkRoom, err := app.IsRoomExists(ctx, r)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	rooms, err := app.GetRoomsFromLKRooms(ctx, []*livekit.Room{lkRoom})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"room": rooms[0]}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) UpdateRoomHandler(w http.ResponseWriter, r *http.Request) {
	var input struct {
		CoOwner        *data.BasicUserResp `json:"coOwner"`
		WelcomeMessage *string             `json:"welcomeMessage"`
		Kick           *data.Kick          `json:"kick"`
	}
	err := app.readJSON(r, &input)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	ctx := context.Background()
	lkRoom, err := app.IsRoomExists(ctx, r)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	rooms, err := app.GetRoomsFromLKRooms(ctx, []*livekit.Room{lkRoom})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	room := rooms[0]
	u := app.getUserContext(r)
	if ok := app.IsRoomOwnerOrCoOwner(room, u.ID); !ok {
		app.forbiddenResponse(w, r)
		return
	}

	if input.CoOwner != nil {
		coIdx := -1
		for i, co := range room.CoOwners {
			if co.ID == input.CoOwner.ID {
				coIdx = i
				break
			}
		}
		if coIdx == -1 {
			room.CoOwners = append(room.CoOwners, input.CoOwner)
		} else {
			room.CoOwners = append(room.CoOwners[:coIdx], room.CoOwners[coIdx+1:]...)
		}
	}

	if input.WelcomeMessage != nil {
		room.WelcomeMessage = *input.WelcomeMessage
	}

	if input.Kick != nil {
		var isKicked bool
		for _, k := range room.KickedParticipants {
			if k.Kicked == u.ID {
				isKicked = true
				break
			}
		}
		if !isKicked {
			room.KickedParticipants = append(room.KickedParticipants, input.Kick)
			app.lkRoomSvc.RemoveParticipant(ctx, &livekit.RoomParticipantIdentity{
				Room:     lkRoom.Name,
				Identity: fmt.Sprintf("%d", input.Kick.Kicked),
			})
		}
	}

	metadata := data.RoomMetadata{
		CoOwners:           room.CoOwners,
		Language:           room.Language,
		Owner:              room.Owner,
		WelcomeMessage:     room.WelcomeMessage,
		KickedParticipants: room.KickedParticipants,
	}

	js, err := json.Marshal(metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.lkRoomSvc.UpdateRoomMetadata(ctx, &livekit.UpdateRoomMetadataRequest{
		Room:     lkRoom.Name,
		Metadata: string(js),
	})

	err = app.writeJSON(w, http.StatusOK, envelope{"room": room}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) CreateRoomHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Topic           string `json:"topic"`
		MaxParticipants int    `json:"max_participants"`
		Language        string `json:"language"`
	}

	err := app.readJSON(r, &body)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	var owner data.BasicUserResp
	u := app.getUserContext(r)
	owner.ID = u.ID
	owner.Username = u.Username
	owner.Avatar = u.Avatar

	metadata := data.RoomMetadata{
		Language:           body.Language,
		Owner:              &owner,
		CoOwners:           make([]*data.BasicUserResp, 0),
		KickedParticipants: make([]*data.Kick, 0),
	}

	js, err := json.Marshal(metadata)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	_, err = app.lkRoomSvc.CreateRoom(context.Background(), &livekit.CreateRoomRequest{
		Name:            body.Topic,
		MaxParticipants: uint32(body.MaxParticipants),
		Metadata:        string(js),
		EmptyTimeout:    ROOM_EMPTY_TIMEOUT,
	})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"message": "room created successfully"}, nil)
}

func (app *application) LiveKitWebhookHandler(w http.ResponseWriter, r *http.Request) {
	authProvider := auth.NewSimpleKeyProvider(
		app.config.livekit.key, app.config.livekit.secret,
	)

	event, err := webhook.ReceiveWebhookEvent(r, authProvider)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	fmt.Printf("event: %+v\n", event)

	bm := &BroadcastMessage{
		toEveryone: true,
		Data: map[string]any{
			"name": "PublishEvent",
		},
	}

	payload := make(map[string]any)
	switch event.Event {
	case "room_started":
		err = json.Unmarshal([]byte(event.Room.Metadata), &payload)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		payload["id"] = event.Room.Sid
		payload["topic"] = event.Room.Name
		payload["max_participants"] = event.Room.MaxParticipants
		payload["participants"] = make([]data.BasicUserResp, 0)

		bm.Data["payload"] = payload
		bm.Data["type"] = "RoomStarted"
		app.hub.broadcast <- bm
	case "room_finished":
		payload["id"] = event.Room.Sid
		bm.Data["type"] = "RoomFinished"
		bm.Data["payload"] = payload
		app.hub.broadcast <- bm
	case "participant_joined":
		var u data.RoomParticipant
		err = json.Unmarshal([]byte(event.Participant.Metadata), &u)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		u.SID = event.Participant.Sid
		bm.Data["type"] = "ParticipantJoined"
		bm.Data["payload"] = map[string]any{
			"roomID":      event.Room.Sid,
			"participant": u,
		}
		app.hub.broadcast <- bm
	case "participant_left":
		bm.Data["type"] = "ParticipantLeft"
		bm.Data["payload"] = map[string]any{
			"roomID":        event.Room.Sid,
			"participantID": event.Participant.Identity,
		}
		app.hub.broadcast <- bm
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"message": "webhook acknowledged"}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}

func (app *application) CreateRoomTokenHandler(w http.ResponseWriter, r *http.Request) {
	u := app.getUserContext(r)
	lkRoom, err := app.IsRoomExists(context.Background(), r)
	if err != nil {
		app.badRequestResponse(w, r, err.Error())
		return
	}

	var m data.RoomMetadata
	err = json.Unmarshal([]byte(lkRoom.Metadata), &m)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	kickIdx := -1
	for i, k := range m.KickedParticipants {
		if u.ID == k.Kicked {
			kickIdx = i
			break
		}
	}

	if kickIdx != -1 {
		kick := m.KickedParticipants[kickIdx]
		timeoutExceeded := true

		if kick.Timeout == -1 {
			timeoutExceeded = false
		}

		currentTime := time.Now().UTC()
		timeoutDuration := time.Duration(kick.Timeout) * time.Second
		timeoutExceeded = currentTime.Sub(kick.KickedAt) > timeoutDuration

		if !timeoutExceeded {
			app.writeJSON(w, http.StatusForbidden, envelope{"error": kick}, nil)
			return
		}
		m.KickedParticipants = append(m.KickedParticipants[:kickIdx], m.KickedParticipants[kickIdx+1:]...)

		js, err := json.Marshal(m)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		app.lkRoomSvc.UpdateRoomMetadata(context.Background(), &livekit.UpdateRoomMetadataRequest{
			Room:     lkRoom.Name,
			Metadata: string(js),
		})
	}

	at := auth.NewAccessToken(app.config.livekit.key, app.config.livekit.secret)
	publishOwnMetadata := true
	at.AddGrant(&auth.VideoGrant{
		Room:                 lkRoom.Name,
		RoomJoin:             true,
		CanUpdateOwnMetadata: &publishOwnMetadata,
		RoomCreate:           false,
		RoomList:             false,
		RoomAdmin:            false,
	})

	at.SetIdentity(fmt.Sprintf("%d", u.ID))
	at.SetValidFor(time.Minute)

	var owner data.BasicUserResp
	owner.ID = u.ID
	owner.Username = u.Username
	owner.Avatar = u.Avatar

	b, err := json.Marshal(owner)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	at.SetMetadata(string(b))

	token, err := at.ToJWT()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	err = app.writeJSON(w, http.StatusOK, envelope{"token": token}, nil)
	if err != nil {
		app.serverErrorResponse(w, r, err)
	}
}
