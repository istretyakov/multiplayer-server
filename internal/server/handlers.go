package server

import (
	"encoding/json"
	"fmt"
	"multiplayer-server/internal/broadcast"
	"multiplayer-server/internal/models"
	"time"
)

// UpdateTimestamp обновляет timestamp в состоянии мира
func UpdateTimestamp(t time.Time) {
	worldStateMutex.Lock()
	defer worldStateMutex.Unlock()
	worldState.Timestamp = t
}

func SendWorldStateToClients() {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for _, client := range worldState.Clients {
		closeClients := getCloseClients(client)

		closePlayers := make([]models.SyncPlayer, 0)

		for _, closeClient := range closeClients {
			closePlayers = append(closePlayers, models.SyncPlayer{
				Id:       closeClient.Id,
				Position: closeClient.Player.Position,
				Velocity: closeClient.Player.Velocity,
			})
		}

		worldStateForCurrentClient := models.SyncWorldState{
			Players:   closePlayers,
			Weather:   worldState.Weather,
			Timestamp: worldState.Timestamp,
		}

		data, err := json.Marshal(worldStateForCurrentClient)
		if err != nil {
			fmt.Println("Error marshalling to JSON:", err)
			break
		}
		msg := models.Message{
			Type:    "world_state",
			Payload: data,
		}
		broadcast.SendMessage(client, msg)
	}
}

func getCloseClients(client *models.Client) []*models.Client {
	var closeClients []*models.Client
	for _, otherClient := range worldState.Clients {
		if client != otherClient && models.Distance(client.Player.Position, otherClient.Player.Position) < 300 {
			closeClients = append(closeClients, otherClient)
		}
	}
	return closeClients
}
