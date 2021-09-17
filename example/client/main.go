package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/EnsurityTechnologies/config"
	"github.com/EnsurityTechnologies/logger"
)

func main() {
	var userName string
	var password string
	flag.StringVar(&userName, "u", "", "User Name")
	flag.StringVar(&password, "p", "", "Password")

	logOptions := &logger.LoggerOptions{
		Name:  "ClientExample",
		Color: logger.AutoColor,
	}

	log := logger.New(logOptions)

	log.Info("Starting Client...")

	if len(os.Args) < 2 {
		log.Error("Invalid Command")
		return
	}
	cfg, err := config.LoadConfig("config.json")

	if err != nil {
		fmt.Printf("%s\n", err.Error())
		log.Error("Invalid Config file")
		return
	}

	cmd := os.Args[1]

	os.Args = os.Args[1:]
	c, err := NewClient(cfg, log)
	if err != nil {
		log.Error(err.Error())
		return
	}
	flag.Parse()
	switch cmd {
	case "Login":
		err = c.Login(userName, password)
		if err != nil {
			log.Error(err.Error())
			return
		}
		log.Info("Login successfully completed")
	case "LoginSession":
		msg, err := c.LoginSession()
		if err != nil {
			log.Error(err.Error())
		} else {
			log.Info(msg)
		}
	default:
		log.Error("Invalid command")
	}
}
