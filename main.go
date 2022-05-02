package main

import (
	"flag"
	"fmt"
	"gochat/client"
	"gochat/config"
	"gochat/server"
	"log"
	"os"
)

func setLogConfig(path string) {
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println("open log file failed, err:", err)
		return
	}
	log.SetOutput(logFile)
	log.SetFlags(log.Llongfile | log.Lmicroseconds | log.Ldate)
}

func main() {
	var mode string
	flag.StringVar(&mode, "mode", "server", "Running mode")
	flag.Parse()

	// read config file
	config.LoadConfig()
	cfg := config.GetConfig()

	if mode == "server" {
		// run as a server
		setLogConfig("./run_server.log")
		s := server.Server{}
		s.Init(cfg.Server.Network, cfg.Server.Address, cfg.Users)
		s.Run()
	} else if mode == "client" {
		// run as a client
		setLogConfig("./run_client.log")
		c := client.Client{}
		c.Init(cfg.Client.Network, cfg.Client.Address)
		c.Run()
	}
}
