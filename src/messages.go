package main

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Type    string          `json:"type"`    // Тип сообщения: "position", "chat", "exit"
	Payload json.RawMessage `json:"payload"` // Данные сообщения зависят от типа
}

func sendMessage(client Client, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		fmt.Println("Error marshalling message:", err)
		return
	}

	data = append(data, byte(0))

	if _, err := (*client.Connection).Write(data); err != nil {
		fmt.Println("Error sending data to client:", err)
		(*client.Connection).Close()
		delete(worldState.Clients, client.Id)
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

	data = append(data, byte(0))

	for id, client := range worldState.Clients {
		if _, err := (*client.Connection).Write(data); err != nil {
			fmt.Println("Error sending data to client:", err)
			(*client.Connection).Close()
			delete(worldState.Clients, id)
		}
	}
}
