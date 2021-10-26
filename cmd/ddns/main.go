package main

import (
	"flag"
	"log"
	"os"
	"os/signal"

	"github.com/samuelkadolph/ddns/internal/config"
	"github.com/samuelkadolph/ddns/internal/server"
)

var configFile string
var listens server.Listens

func main() {
	log.SetFlags(log.Ldate | log.Lmicroseconds | log.Lmsgprefix | log.Lshortfile | log.Ltime)

	flag.StringVar(&configFile, "config", "./ddns.yml", "Location of the config file to use")
	flag.Var(&listens, "listen", "ip:port to listen on")
	flag.Parse()

	if listens == nil {
		listens = []string{":4444"}
	}

	cfg, err := config.Read(configFile)
	if err != nil {
		log.Fatal(err)
	}

	done := make(chan bool, 1)
	intr := make(chan os.Signal, 1)
	signal.Notify(intr, os.Interrupt)

	server, err := server.NewServer(cfg, listens)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-intr
		log.Println("Shutting down...")
		if err := server.Shutdown(); err != nil {
			log.Println(err)
		}
		close(done)
	}()

	<-done
}
