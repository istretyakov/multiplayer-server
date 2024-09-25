package main

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

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
}

type Weather struct {
	Condition   string  `json:"condition"`
	Temperature float64 `json:"temperature"`
}

var worldState WorldState

var worldStateMutex sync.Mutex
var clientsMutex sync.Mutex

func sendWorldStateToClients() {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	for _, client := range worldState.Clients {
		closeClients := getCloseClients(client)

		closePlayers := make([]SyncPlayer, 0)

		for _, closeClient := range closeClients {
			closePlayers = append(closePlayers, SyncPlayer{
				Id:       closeClient.Id,
				Position: closeClient.Player.Position,
			})
		}

		worldStateForCurrentClient := SyncWorldState{
			Players:   closePlayers,
			Weather:   worldState.Weather,
			Timestamp: worldState.Timestamp,
		}

		data, err := json.Marshal(worldStateForCurrentClient)
		if err != nil {
			fmt.Println("Error marshalling to JSON:", err)
			break
		}
		msg := Message{
			Type:    "world_state",
			Payload: data,
		}
		sendMessage(*client, msg)
	}
}

func getCloseClients(client *Client) []*Client {
	var closeClients []*Client
	for _, otherClient := range worldState.Clients {
		if client != otherClient && distance(client.Player.Position, otherClient.Player.Position) < 300 {
			closeClients = append(closeClients, otherClient)
		}
	}
	return closeClients
}
