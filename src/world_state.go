package main

import (
	"encoding/json"
	"fmt"
	"math"
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
			closePlayers = append(closePlayers, WorldStatePlayer{Id: closeClient.Id, Position: Vector3{X: closeClient.Player.Position.X, Y: closeClient.Player.Position.Y, Z: closeClient.Player.Position.Z}})
		}

		worldStateForCurrentClient := SyncWorldState{
			Players:   closePlayers,
			Weather:   worldState.Weather,
			Timestamp: worldState.Timestamp,
		}

		stateData := toJson(worldStateForCurrentClient)
		msg := Message{
			Type:    "world_state",
			Payload: stateData,
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

func distance(a, b Vector3) float64 {
	return math.Sqrt(math.Pow(a.X-b.X, 2) + math.Pow(a.Y-b.Y, 2) + math.Pow(a.Z-b.Z, 2))
}

func toJson(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
	}
	return data
}
