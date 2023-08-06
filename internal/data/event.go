package data

type Event struct {
	Type        string         `json:"type"`
	BroadcastTo []string       `json:"broadcastTo"`
	Payload     map[string]any `json:"payload"`
}
