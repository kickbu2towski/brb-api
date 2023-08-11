package data

type Event struct {
	Name        string `json:"name"`
	UserID      int         `json:"user_id"`
	Type        string         `json:"type"`
	BroadcastTo []int       `json:"broadcastTo"`
	Payload     map[string]any `json:"payload"`
}
