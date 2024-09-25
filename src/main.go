package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"sync"
	"time"
)

type Vector3 struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type Client struct {
	Connection *net.Conn
	Id         int
	Player     Player
}

type Player struct {
	Position Vector3
}

// Структура для описания позиции игрока
type UpdatedPlayerState struct {
	Position Vector3 `json:"position"`
}

// Структура для "погоды"
type Weather struct {
	Condition   string  `json:"condition"`
	Temperature float64 `json:"temperature"`
}

type WorldStatePlayer struct {
	Id       int     `json:"id"`
	Position Vector3 `json:"position"`
}

// Структура для "мира" с игроками и погодой
type WorldState struct {
	Players   []WorldStatePlayer `json:"players"`
	Weather   Weather            `json:"weather"`
	Timestamp time.Time          `json:"timestamp"`
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
	Id    int    `json:"id"`
	Event string `json:"event"` // Тип события: "joined", "left"
}

var worldState WorldState
var clients = make(map[int]*Client)
var clientsMutex sync.Mutex
var worldStateMutex sync.Mutex

func main() {
	// Инициализация начального состояния мира.
	worldState = WorldState{
		Players:   []WorldStatePlayer{},
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
		sendWorldStateToClients()
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

		client := Client{
			Connection: &conn,
			Id:         getSmallestAvailabeId(),
			Player:     Player{Position: Vector3{X: 0.0, Y: 0.0, Z: 0.0}},
		}

		clients[client.Id] = &client
		clientsMutex.Unlock()

		broadcastPlayerEvent(client.Id, "joined")

		// Обработка сообщений от клиента в отдельной горутине
		go handleClient(&client)
	}
}

func handleClient(client *Client) {
	defer func() {
		clientsMutex.Lock()
		delete(clients, client.Id)
		clientsMutex.Unlock()
		(*client.Connection).Close()
	}()

	// Чтение сообщений от клиента
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
			fmt.Printf("Received position from player %s: (%f, %f, %f)\n", client.Id, playerPos.Position.X, playerPos.Position.Y, playerPos.Position.Z)
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

func updatePlayerPosition(client *Client, playerPos UpdatedPlayerState) {
	worldStateMutex.Lock()
	defer worldStateMutex.Unlock()

	client.Player.Position = playerPos.Position
}

func broadcastChatMessage(chatMsg ChatMessage) {
	msg := Message{
		Type:    "chat",
		Payload: toJson(chatMsg),
	}
	broadcastMessage(msg)
}

func broadcastPlayerEvent(playerID int, event string) {
	playerEvent := PlayerEvent{
		Id:    playerID,
		Event: event,
	}
	msg := Message{
		Type:    "player_event",
		Payload: toJson(playerEvent),
	}
	broadcastMessage(msg)
}

func sendWorldStateToClients() {
	for _, client := range clients {
		closeClients := getCloseClients(*client)

		closePlayers := make([]WorldStatePlayer, 0)

		for _, closeClient := range closeClients {
			closePlayers = append(closePlayers, WorldStatePlayer{Id: closeClient.Id, Position: Vector3{X: closeClient.Player.Position.X, Y: closeClient.Player.Position.Y, Z: closeClient.Player.Position.Z}})
		}

		worldStateForCurrentClient := WorldState{
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
	for _, otherClient := range clients {
		if distance(client.Player.Position, otherClient.Player.Position) < 300 {
			closeClients = append(closeClients, *otherClient)
		}
	}
	return closeClients
}

func distance(a, b Vector3) float64 {
	return math.Sqrt(math.Pow(a.X-b.X, 2) + math.Pow(a.Y-b.Y, 2) + math.Pow(a.Z-b.Z, 2))
}

func sendMessage(client Client, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return
	}

	if _, err := (*client.Connection).Write(data); err != nil {
		fmt.Println("Error sending data to client:", err)
		(*client.Connection).Close()
		delete(clients, client.Id)
	}
}

func broadcastMessage(msg Message) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()

	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return
	}

	for id, client := range clients {
		if _, err := (*client.Connection).Write(data); err != nil {
			fmt.Println("Error sending data to client:", err)
			(*client.Connection).Close()
			delete(clients, id)
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

func getSmallestAvailabeId() int {
	smallestId := 0

	for id := 0; id < len(clients)+1; id++ {
		if _, ok := clients[id]; !ok {
			smallestId = id
			break
		}
	}

	return smallestId
}
