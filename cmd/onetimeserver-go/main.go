package main

import (
	"flag"
	"fmt"
	"github.com/osheroff/onetimeserver"
	"math/rand"
	"os"
	"time"
)

type config struct {
	ppid       int
	serverType string
	extraArgs  []string
}

func getconf() config {
	c := config{}
	flag.IntVar(&c.ppid, "ppid", os.Getppid(), "parent PID")
	flag.StringVar(&c.serverType, "type", "", "server type: one of mysql")
	flag.Parse()

	c.extraArgs = flag.Args()
	return c
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	config := getconf()

	fmt.Printf("ppid: %d, extraArgs: %s\n", config.ppid, config.extraArgs)
	var s onetimeserver.Server

	switch config.serverType {
	case "mysql":
		s = onetimeserver.NewMysql()
	default:
		fmt.Fprintf(os.Stderr, "Please provide 'type' command line option\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	s.Boot(config.extraArgs)
	fmt.Printf("port: %d\n", s.Port())
	fmt.Printf("booted: true\n")

	onetimeserver.WatchServer(config.ppid, s)
}
