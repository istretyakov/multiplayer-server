package main

import (
	"fmt"
	"multiplayer-server/internal/models"
	"multiplayer-server/internal/server"
	"net"
	"time"
)

func main() {
	worldState := models.WorldState{
		Clients:   make(map[int]*models.Client),
		Weather:   models.Weather{Condition: "Sunny", Temperature: 25.0},
		Timestamp: time.Now(),
	}

	server.Init(&worldState)

	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server started on port 8080")

	go server.HandleConnections(listener)

	ticker := time.NewTicker(time.Second / 10)
	defer ticker.Stop()

	for range ticker.C {
		server.UpdateTimestamp(time.Now())
		server.SendWorldStateToClients()
	}
}
