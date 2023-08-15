package data

import "time"

type Room struct {
	ID              string             `json:"id"`
	Topic           string             `json:"topic"`
	MaxParticipants int                `json:"max_participants"`
	Participants    []*RoomParticipant `json:"participants"`
	RoomMetadata
}

type RoomParticipant struct {
	BasicUserResp
	SID    string `json:"sid,omitempty"`
	Status string `json:"status,omitempty"`
}

type RoomMetadata struct {
	Language           string           `json:"language"`
	Owner              *BasicUserResp   `json:"owner"`
	CoOwners           []*BasicUserResp `json:"co_owners"`
	WelcomeMessage     string           `json:"welcome_message"`
	KickedParticipants []*Kick          `json:"kicked_participants"`
}

type Kick struct {
	Kicked   int       `json:"kicked"`
	KickedBy int       `json:"kicked_by"`
	KickedAt time.Time `json:"kicked_at"`
	Timeout  int       `json:"timeout"`
	Reason   string    `json:"reason"`
}
