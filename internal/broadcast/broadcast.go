package broadcast

import (
	"encoding/json"
	"fmt"
	"multiplayer-server/internal/models"
	"sync"
)

var (
	worldState      *models.WorldState
	worldStateMutex sync.Mutex
	clientsMutex    sync.Mutex
)

// Init инициализирует пакет broadcast с состоянием мира
func Init(ws *models.WorldState) {
	worldState = ws
}

func SendMessage(client *models.Client, msg models.Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return
	}

	data = append(data, byte(0))

	if _, err := (*client.Connection).Write(data); err != nil {
		fmt.Println("Error sending data to client:", err)
		(*client.Connection).Close()
		clientsMutex.Lock()
		delete(worldState.Clients, client.Id)
		clientsMutex.Unlock()
	}
}

func BroadcastMessage(msg models.Message) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return
	}

	data = append(data, byte(0))

	for id, client := range worldState.Clients {
		if _, err := (*client.Connection).Write(data); err != nil {
			fmt.Println("Error sending data to client:", err)
			(*client.Connection).Close()
			delete(worldState.Clients, id)
		}
	}
}

func BroadcastChatMessage(chatMsg models.ChatMessage) {
	data, err := json.Marshal(chatMsg)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	msg := models.Message{
		Type:    "chat",
		Payload: data,
	}
	BroadcastMessage(msg)
}

func BroadcastPlayerEvent(playerId int, event string) {
	playerEvent := models.PlayerEvent{
		Id:    playerId,
		Event: event,
	}

	data, err := json.Marshal(playerEvent)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	msg := models.Message{
		Type:    "player_event",
		Payload: data,
	}
	BroadcastMessage(msg)
}
