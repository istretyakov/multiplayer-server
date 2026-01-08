package server

import (
	"encoding/json"
	"fmt"
	"multiplayer-server/internal/broadcast"
	"multiplayer-server/internal/models"
	"net"
	"sync"
)

var (
	worldState      *models.WorldState
	worldStateMutex sync.Mutex
	clientsMutex    sync.Mutex
)

// Init инициализирует пакет server с состоянием мира
func Init(ws *models.WorldState) {
	worldState = ws
	broadcast.Init(ws)
}

func HandleConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		clientsMutex.Lock()

		client := models.Client{
			Connection: &conn,
			Id:         getSmallestAvailableId(),
			Player: models.Player{
				Position: models.Vector3{X: 0.0, Y: 0.0, Z: 0.0},
				Velocity: models.Vector3{X: 0.0, Y: 0.0, Z: 0.0},
			},
		}

		worldState.Clients[client.Id] = &client
		clientsMutex.Unlock()

		broadcast.BroadcastPlayerEvent(client.Id, "joined")

		go handleClient(&client)
	}
}

func handleClient(client *models.Client) {
	defer func() {
		clientsMutex.Lock()
		delete(worldState.Clients, client.Id)
		clientsMutex.Unlock()
		(*client.Connection).Close()
	}()

	decoder := json.NewDecoder(*client.Connection)
	for {
		var msg models.Message
		if err := decoder.Decode(&msg); err != nil {
			fmt.Println("Error reading from client:", err)
			break
		}

		switch msg.Type {
		case "position":
			var playerPos models.UpdatedPlayerState
			if err := json.Unmarshal(msg.Payload, &playerPos); err != nil {
				fmt.Println("Error unmarshalling position:", err)
				continue
			}
			fmt.Printf("Received position from player %d: (%f, %f, %f) (%f, %f, %f)\n", client.Id, playerPos.Position.X, playerPos.Position.Y, playerPos.Position.Z, playerPos.Velocity.X, playerPos.Velocity.Y, playerPos.Velocity.Z)
			updatePlayerPosition(client, playerPos)
		case "chat":
			var chatMsg models.ChatMessage
			if err := json.Unmarshal(msg.Payload, &chatMsg); err != nil {
				fmt.Println("Error unmarshalling chat message:", err)
				continue
			}
			fmt.Printf("Chat message from player %s: %s\n", chatMsg.ID, chatMsg.Message)
			broadcast.BroadcastChatMessage(chatMsg)
		case "exit":
			fmt.Printf("Player %d has exited\n", client.Id)
			broadcast.BroadcastPlayerEvent(client.Id, "left")
			return
		default:
			fmt.Println("Received unknown message type:", msg.Type)
		}
	}
}

func updatePlayerPosition(client *models.Client, playerPos models.UpdatedPlayerState) {
	worldStateMutex.Lock()
	defer worldStateMutex.Unlock()

	client.Player.Position = playerPos.Position
	client.Player.Velocity = playerPos.Velocity
}

func getSmallestAvailableId() int {
	smallestId := 0

	for id := 0; id < len(worldState.Clients)+1; id++ {
		if _, ok := worldState.Clients[id]; !ok {
			smallestId = id
			break
		}
	}

	return smallestId
}
