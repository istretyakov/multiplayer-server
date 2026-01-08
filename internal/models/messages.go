package models

import "encoding/json"

type Message struct {
	Type    string          `json:"type"`    // Тип сообщения: "position", "chat", "exit"
	Payload json.RawMessage `json:"payload"` // Данные сообщения зависят от типа
}

type ChatMessage struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type PlayerEvent struct {
	Id    int    `json:"id"`
	Event string `json:"event"` // Тип события: "joined", "left"
}

type UpdatedPlayerState struct {
	Position Vector3 `json:"position"`
	Velocity Vector3 `json:"velocity"`
}
