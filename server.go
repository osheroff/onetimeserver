package onetimeserver

import (
	"fmt"
	"math/rand"
	"net"
)

type Server interface {
	Boot([]string)
	Pid() int
	Port() int
	Kill()
	String() string
}

func GetPort(suggestedPort int) int {
	tryPort := suggestedPort

	for true {
		conn, err := net.Listen("tcp", fmt.Sprintf(":%d", tryPort))
		if err == nil {
			conn.Close()
			return tryPort
		} else {
			tryPort = rand.Intn(55000) + 9000
		}
	}
	return 0
}
