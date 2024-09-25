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
type SyncWorldState struct {
	Players   []WorldStatePlayer `json:"players"`
	Weather   Weather            `json:"weather"`
	Timestamp time.Time          `json:"timestamp"`
}

type WorldState struct {
	Clients   []Client
	Weather   Weather
	Timestamp time.Time
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

var worldState SyncWorldState
var clients = make(map[int]*Client)
var clientsMutex sync.Mutex
var worldStateMutex sync.Mutex

func main() {
	// Инициализация начального состояния мира.
	worldState = SyncWorldState{
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

func sendWorldStateToClients() {
	for _, client := range clients {
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

func toJson(v interface{}) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Println("Error marshalling to JSON:", err)
	}
	return data
}
