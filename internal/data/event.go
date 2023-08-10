package data

type Event struct {
	UserID      string         `json:"user_id"`
	Type        string         `json:"type"`
	BroadcastTo []string       `json:"broadcastTo"`
	Payload     map[string]any `json:"payload"`
}
