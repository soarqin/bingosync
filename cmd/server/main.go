package main

import (
	"bingosync/internal/websocket"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/lxzan/gws"
)

func main() {
	port := flag.Int("port", 8765, "WebSocket server port")
	flag.Parse()

	handler := websocket.NewHandler()

	upgrader := gws.NewUpgrader(handler, &gws.ServerOption{
		ParallelEnabled: true,
		Recovery:        gws.Recovery,
		PermessageDeflate: gws.PermessageDeflate{
			Enabled: true,
		},
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		socket, err := upgrader.Upgrade(w, r)
		if err != nil {
			log.Printf("Upgrade error: %v", err)
			return
		}
		go socket.ReadLoop()
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting BingoSync WebSocket server on %s", addr)
	
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}

const (
	PingInterval = 30 * time.Second
	PingWait     = 60 * time.Second
)
