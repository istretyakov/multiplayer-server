package models

import "net"

type Client struct {
	Connection *net.Conn
	Id         int
	Player     Player
}

type Player struct {
	Position Vector3
	Velocity Vector3
}
