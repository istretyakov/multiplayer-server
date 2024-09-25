package main

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"time"
)

// Структура для описания позиции игрока
type Player struct {
	ID string  `json:"id"`
	X  float64 `json:"x"`
	Y  float64 `json:"y"`
	Z  float64 `json:"z"`
}

// Структура для "погоды"
type Weather struct {
	Condition   string  `json:"condition"`
	Temperature float64 `json:"temperature"`
}

// Структура для "мира" с игроками и погодой
type WorldState struct {
	Players   []Player  `json:"players"`
	Weather   Weather   `json:"weather"`
	Timestamp time.Time `json:"timestamp"`
}

// Структуры для различных типов сообщений
type Message struct {
	Type    string          `json:"type"`    // Тип сообщения: "position", "chat", "exit"
	Payload json.RawMessage `json:"payload"` // Данные сообщения зависят от типа
}

type ChatMessage struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

type PlayerEvent struct {
	ID    string `json:"id"`
	Event string `json:"event"` // Тип события: "joined", "left"
}

var worldState WorldState
var clients = make(map[*net.Conn]bool)
var clientsMutex sync.Mutex
var worldStateMutex sync.Mutex

func main() {
	// Инициализация начального состояния мира.
	worldState = WorldState{
		Players:   []Player{},
		Weather:   Weather{Condition: "Sunny", Temperature: 25.0},
		Timestamp: time.Now(),
	}

	// Старт сервера
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server started on port 8080")

	// Запуск обработки клиентских соединений в отдельной горутине
	go handleConnections(listener)

	// Запуск рассылки обновлений состояния мира 10 раз в секунду
	ticker := time.NewTicker(time.Second / 10)
	defer ticker.Stop()

	for range ticker.C {
		// Обновление времени состояния мира
		worldState.Timestamp = time.Now()

		// Отправка обновления всем клиентам
		sendWorldState()
	}
}

func handleConnections(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		clientsMutex.Lock()
		clients[&conn] = true
		clientsMutex.Unlock()

		// Обработка сообщений от клиента в отдельной горутине
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer func() {
		clientsMutex.Lock()
		delete(clients, &conn)
		clientsMutex.Unlock()
		conn.Close()
	}()

	// Чтение сообщений от клиента
	decoder := json.NewDecoder(conn)
	for {
		var msg Message
		if err := decoder.Decode(&msg); err != nil {
			fmt.Println("Error reading from client:", err)
			break
		}

		switch msg.Type {
		case "position":
			var playerPos Player
			if err := json.Unmarshal(msg.Payload, &playerPos); err != nil {
				fmt.Println("Error unmarshalling position:", err)
				continue
			}
			fmt.Printf("Received position from player %s: (%f, %f, %f)\n", playerPos.ID, playerPos.X, playerPos.Y, playerPos.Z)
			updatePlayerPosition(playerPos)
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
			var player Player
			if err := json.Unmarshal(msg.Payload, &player); err != nil {
				fmt.Println("Error unmarshalling exit message:", err)
				continue
			}
			fmt.Printf("Player %s has exited\n", player.ID)
			// Уведомление о выходе игрока
			broadcastPlayerEvent(player.ID, "left")
			removePlayer(player)
			return
		default:
			fmt.Println("Received unknown message type:", msg.Type)
		}
	}
}

func updatePlayerPosition(playerPos Player) {
	worldStateMutex.Lock()
	defer worldStateMutex.Unlock()

	for i, player := range worldState.Players {
		if player.ID == playerPos.ID {
			worldState.Players[i] = playerPos
			return
		}
	}
	worldState.Players = append(worldState.Players, playerPos)
	// Уведомление о новом игроке
	broadcastPlayerEvent(playerPos.ID, "joined")
}

func removePlayer(player Player) {
	worldStateMutex.Lock()
	defer worldStateMutex.Unlock()

	for i, p := range worldState.Players {
		if p.ID == player.ID {
			worldState.Players = append(worldState.Players[:i], worldState.Players[i+1:]...)
			break
		}
	}
}

func broadcastChatMessage(chatMsg ChatMessage) {
	msg := Message{
		Type:    "chat",
		Payload: toJson(chatMsg),
	}
	broadcastMessage(msg)
}

func broadcastPlayerEvent(playerID, event string) {
	playerEvent := PlayerEvent{
		ID:    playerID,
		Event: event,
	}
	msg := Message{
		Type:    "player_event",
		Payload: toJson(playerEvent),
	}
	broadcastMessage(msg)
}

func sendWorldState() {
	stateData := toJson(worldState)
	msg := Message{
		Type:    "world_state",
		Payload: stateData,
	}
	broadcastMessage(msg)
}

func broadcastMessage(msg Message) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return
	}

	for conn := range clients {
		if _, err := (*conn).Write(data); err != nil {
			fmt.Println("Error sending data to client:", err)
			(*conn).Close()
			delete(clients, conn)
		}
	}
}

func toJson(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
	}
	return data
}
