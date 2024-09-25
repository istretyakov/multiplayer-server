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
	Players   []WorldStatePlayer `json:"players"`
	Weather   Weather            `json:"weather"`
	Timestamp time.Time          `json:"timestamp"`
}

type WorldStatePlayer struct {
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
	for _, client := range worldState.Clients {
		closeClients := getCloseClients(*client)

		closePlayers := make([]WorldStatePlayer, 0)

		for _, closeClient := range closeClients {
			closePlayers = append(closePlayers, WorldStatePlayer{
				Id: closeClient.Id,
				Position: Vector3{
					X: closeClient.Player.Position.X,
					Y: closeClient.Player.Position.Y,
					Z: closeClient.Player.Position.Z,
				},
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

func getCloseClients(client Client) []Client {
	var closeClients []Client
	for _, otherClient := range worldState.Clients {
		if distance(client.Player.Position, otherClient.Player.Position) < 300 {
			closeClients = append(closeClients, *otherClient)
		}
	}
	return closeClients
}
