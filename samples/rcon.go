package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	rcon "github.com/galexrt/go-rcon"
	log "github.com/sirupsen/logrus"
)

var (
	debug          bool
	connectTimeout string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.StringVar(&connectTimeout, "timeout", "1s", "Connection timeout")
}

func main() {
	flag.Parse()
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	addr := os.Getenv("ADDR")
	pass := os.Getenv("RCON_PASSWORD")
	if addr == "" || pass == "" {
		fmt.Println("Please set ADDR & RCON_PASSWORD.")
		return
	}
	connectTimeoutParsed, err := time.ParseDuration(connectTimeout)
	if err != nil {
		log.Fatal(err)
	}
	for {
		rcon, err := rcon.Connect(addr, &rcon.ConnectOptions{
			RCONPassword: pass,
			Timeout:      connectTimeoutParsed,
		})
		if err != nil {
			fmt.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}
		defer rcon.Close()
		for {
			resp, err := rcon.Send("status")
			if err != nil {
				fmt.Println(err)
				break
			}
			fmt.Println(resp)
			time.Sleep(5 * time.Second)
		}
	}
}
