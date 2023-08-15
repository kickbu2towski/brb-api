package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/kickbu2towski/brb-api/internal/data"
	"github.com/livekit/protocol/livekit"
)

func Includes(input []int, key int) bool {
	var exists bool
	for _, v := range input {
		if v == key {
			exists = true
			break
		}
	}
	return exists
}

func (app *application) IsRoomExists(ctx context.Context, r *http.Request) (*livekit.Room, error) {
	params := httprouter.ParamsFromContext(r.Context())
	roomID := params.ByName("roomID")
	if roomID == "" {
		return nil, fmt.Errorf("invalid param roomID")
	}

	res, err := app.lkRoomSvc.ListRooms(ctx, &livekit.ListRoomsRequest{})
	if err != nil {
		return nil, err
	}

	var lkRoom *livekit.Room
	for _, room := range res.Rooms {
		if room.Sid == roomID {
			lkRoom = room
			break
		}
	}

	if lkRoom == nil {
		return nil, fmt.Errorf("the requested room doesn't exists")
	}

	return lkRoom, nil
}

func (app *application) GetRoomsFromLKRooms(ctx context.Context, lkRooms []*livekit.Room) ([]*data.Room, error) {
	rooms := make([]*data.Room, 0)
	for _, lkRoom := range lkRooms {
		var room data.Room
		var metadata data.RoomMetadata
		err := json.Unmarshal([]byte(lkRoom.Metadata), &metadata)
		if err != nil {
			return nil, err
		}

		// metadata
		room.Language = metadata.Language
		room.Owner = metadata.Owner
		room.CoOwners = metadata.CoOwners
		room.WelcomeMessage = metadata.WelcomeMessage
		room.KickedParticipants = metadata.KickedParticipants

		room.ID = lkRoom.Sid
		room.Topic = lkRoom.Name
		participants, err := app.GetParticipantsForRoom(ctx, lkRoom.Name)
		if err != nil {
			return nil, err
		}
		room.Participants = participants
		rooms = append(rooms, &room)
	}
	return rooms, nil
}

func (app *application) GetParticipantsForRoom(ctx context.Context, name string) ([]*data.RoomParticipant, error) {
	participants := make([]*data.RoomParticipant, 0)
	pRes, err := app.lkRoomSvc.ListParticipants(ctx, &livekit.ListParticipantsRequest{
		Room: name,
	})
	if err != nil {
		return nil, err
	}
	for _, p := range pRes.Participants {
		var u data.RoomParticipant
		err := json.Unmarshal([]byte(p.Metadata), &u)
		if err != nil {
			return nil, err
		}
		u.SID = p.Sid
		participants = append(participants, &u)
	}
	return participants, nil
}

func (app *application) IsRoomOwnerOrCoOwner(room *data.Room, pID int) bool {
	if room.Owner.ID == pID {
		return true
	}
	for _, co := range room.CoOwners {
		if co.ID == pID {
			return true
		}
	}
	return false
}
