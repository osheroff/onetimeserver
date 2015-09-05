package main

import (
	"flag"
	"fmt"
	"github.com/osheroff/onetimeserver"
	"math/rand"
	"os"
	"runtime"
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

	fmt.Printf("%s\n", runtime.GOOS)
	s.Boot(config.extraArgs)
	fmt.Printf("Yeah, I got a server!  It's all %s\n", s)
}
