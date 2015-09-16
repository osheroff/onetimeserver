package onetimeserver

import (
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type Server interface {
	Boot([]string) (map[string]interface{}, error)
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

func pidExists(pid int) bool {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

func WatchServer(ppid int, server Server) {
	channel := make(chan os.Signal, 1)

	signal.Notify(channel, os.Interrupt, os.Kill)
	go func() {
		for true {
			if !pidExists(ppid) || !pidExists(server.Pid()) {
				channel <- os.Kill
			}
			time.Sleep(1 * time.Second)
		}
	}()
	<-channel

	server.Kill()
}
