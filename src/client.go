package main

import (
	"encoding/json"
	"fmt"
	"net"
)

type Client struct {
	Connection *net.Conn
	Id         int
	Player     Player
}

type Player struct {
	Position Vector3
	Velocity Vector3
}

type ChatMessage struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type PlayerEvent struct {
	Id    int    `json:"id"`
	Event string `json:"event"` // Тип события: "joined", "left"
}

func handleConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		clientsMutex.Lock()

		client := Client{
			Connection: &conn,
			Id:         getSmallestAvailabeId(),
			Player: Player{
				Position: Vector3{X: 0.0, Y: 0.0, Z: 0.0},
				Velocity: Vector3{X: 0.0, Y: 0.0, Z: 0.0},
			},
		}

		worldState.Clients[client.Id] = &client
		clientsMutex.Unlock()

		broadcastPlayerEvent(client.Id, "joined")

		go handleClient(&client)
	}
}

func handleClient(client *Client) {
	defer func() {
		clientsMutex.Lock()
		delete(worldState.Clients, client.Id)
		clientsMutex.Unlock()
		(*client.Connection).Close()
	}()

	decoder := json.NewDecoder(*client.Connection)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			fmt.Println("Error reading from client:", err)
			break
		}

		switch msg.Type {
		case "position":
			var playerPos UpdatedPlayerState
			if err := json.Unmarshal(msg.Payload, &playerPos); err != nil {
				fmt.Println("Error unmarshalling position:", err)
				continue
			}
			fmt.Printf("Received position from player %d: (%f, %f, %f) (%f, %f, %f)\n", client.Id, playerPos.Position.X, playerPos.Position.Y, playerPos.Position.Z, playerPos.Velocity.X, playerPos.Velocity.Y, playerPos.Velocity.Z)
			updatePlayerPosition(client, playerPos)
		case "chat":
			var chatMsg ChatMessage
			if err := json.Unmarshal(msg.Payload, &chatMsg); err != nil {
				fmt.Println("Error unmarshalling chat message:", err)
				continue
			}
			fmt.Printf("Chat message from player %s: %s\n", chatMsg.ID, chatMsg.Message)
			// Здесь можно добавить код для рассылки чата всем клиентам
			broadcastChatMessage(chatMsg)
		case "exit":
			fmt.Printf("UpdatedPlayerState %s has exited\n", client.Id)
			// Уведомление о выходе игрока
			broadcastPlayerEvent(client.Id, "left")
			return
		default:
			fmt.Println("Received unknown message type:", msg.Type)
		}
	}
}

func broadcastChatMessage(chatMsg ChatMessage) {
	data, err := json.Marshal(chatMsg)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	msg := Message{
		Type:    "chat",
		Payload: data,
	}
	broadcastMessage(msg)
}

func broadcastPlayerEvent(playerId int, event string) {
	playerEvent := PlayerEvent{
		Id:    playerId,
		Event: event,
	}

	data, err := json.Marshal(playerEvent)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
		return
	}

	msg := Message{
		Type:    "player_event",
		Payload: data,
	}
	broadcastMessage(msg)
}

func updatePlayerPosition(client *Client, playerPos UpdatedPlayerState) {
	worldStateMutex.Lock()
	defer worldStateMutex.Unlock()

	client.Player.Position = playerPos.Position
	client.Player.Velocity = playerPos.Velocity
}

func getSmallestAvailabeId() int {
	smallestId := 0

	for id := 0; id < len(worldState.Clients)+1; id++ {
		if _, ok := worldState.Clients[id]; !ok {
			smallestId = id
			break
		}
	}

	return smallestId
}
