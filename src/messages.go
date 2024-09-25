package main

import (
	"encoding/json"
	"fmt"
)

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

	data = append(data, byte(0))

	for id, client := range clients {
		if _, err := (*client.Connection).Write(data); err != nil {
			fmt.Println("Error sending data to client:", err)
			(*client.Connection).Close()
			delete(clients, id)
		}
	}
}
