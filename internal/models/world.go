package models

import "time"

type WorldState struct {
	Clients   map[int]*Client
	Weather   Weather
	Timestamp time.Time
}

type SyncWorldState struct {
	Players   []SyncPlayer `json:"players"`
	Weather   Weather      `json:"weather"`
	Timestamp time.Time    `json:"timestamp"`
}

type SyncPlayer struct {
	Id       int     `json:"id"`
	Position Vector3 `json:"position"`
	Velocity Vector3 `json:"velocity"`
}

type Weather struct {
	Condition   string  `json:"condition"`
	Temperature float64 `json:"temperature"`
}
