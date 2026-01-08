package main

import (
	"fmt"
	"multiplayer-server/internal/config"
	"multiplayer-server/internal/models"
	"multiplayer-server/internal/server"
	"net"
	"time"
)

func main() {
	if err := config.Load(); err != nil {
		fmt.Println("Error loading config:", err)
		return
	}

	cfg := config.Get()

	worldState := models.WorldState{
		Clients:   make(map[int]*models.Client),
		Weather:   models.Weather{Condition: "Sunny", Temperature: 25.0},
		Timestamp: time.Now(),
	}

	server.Init(&worldState)

	address := ":" + cfg.ServerPort
	listener, err := net.Listen("tcp", address)
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer listener.Close()

	fmt.Printf("Server started on port %s\n", cfg.ServerPort)

	go server.HandleConnections(listener)

	ticker := time.NewTicker(time.Second / 10)
	defer ticker.Stop()

	for range ticker.C {
		server.UpdateTimestamp(time.Now())
		server.SendWorldStateToClients()
	}
}
