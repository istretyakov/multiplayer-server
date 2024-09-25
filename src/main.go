package main

import (
	"fmt"
	"net"
	"time"
)

type UpdatedPlayerState struct {
	Position Vector3 `json:"position"`
	Velocity Vector3 `json:"velocity"`
}

func main() {
	worldState = WorldState{
		Clients:   make(map[int]*Client),
		Weather:   Weather{Condition: "Sunny", Temperature: 25.0},
		Timestamp: time.Now(),
	}

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server started on port 8080")

	go handleConnections(listener)

	ticker := time.NewTicker(time.Second / 10)
	defer ticker.Stop()

	for range ticker.C {
		worldState.Timestamp = time.Now()

		sendWorldStateToClients()
	}
}
