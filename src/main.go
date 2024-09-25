package main

import (
	"encoding/json"
	"fmt"
	"net"
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

type WorldStatePlayer struct {
	Id       int     `json:"id"`
	Position Vector3 `json:"position"`
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

func main() {
	// Инициализация начального состояния мира.
	worldState = WorldState{
		Clients:   make(map[int]*Client),
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
