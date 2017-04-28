package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/galexrt/go-steam"
)

var (
	debug          bool
	connectTimeout string
)

func init() {
	flag.BoolVar(&debug, "debug", false, "debug")
	flag.StringVar(&connectTimeout, "timeout", "15s", "Connection timeout")
}

func main() {
	flag.Parse()
	if debug {
		log.SetLevel(log.DebugLevel)
	}
	addr := os.Getenv("ADDR")
	if addr == "" {
		fmt.Println("Please set ADDR.")
		return
	}
	for {
		server, err := steam.Connect(addr, &steam.ConnectOptions{
			Timeout: connectTimeout,
		})
		if err != nil {
			panic(err)
		}
		defer server.Close()
		ping, err := server.Ping()
		if err != nil {
			fmt.Printf("steam: could not ping %v: %v\n", addr, err)
			return
		}
		fmt.Printf("steam: ping to %v: %v\n", addr, ping)
		info, err := server.Info()
		if err != nil {
			fmt.Printf("steam: could not get server info from %v: %v\n", addr, err)
			return
		}
		fmt.Printf("steam: info of %v: %v\n", addr, info)
		playersInfo, err := server.PlayersInfo()
		if err != nil {
			fmt.Printf("steam: could not get players info from %v: %v\n", addr, err)
			return
		}
		if len(playersInfo.Players) > 0 {
			fmt.Printf("steam: player infos for %v:\n", addr)
			for _, player := range playersInfo.Players {
				fmt.Printf("steam: %v %v\n", player.Name, player.Score)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
