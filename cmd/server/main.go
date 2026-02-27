package main

import (
	"bingosync/internal/storage"
	"bingosync/internal/websocket"
	"bingosync/pkg/protocol"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/lxzan/gws"
)

func main() {
	port := flag.Int("port", 8765, "WebSocket server port")
	dataDir := flag.String("data", "./data", "Data directory for persistence")
	roomTTL := flag.Duration("room-ttl", 30*time.Minute, "Empty room TTL before deletion (0 to disable)")
	flag.Parse()

	// Initialize storage
	store, err := storage.New(*dataDir)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize handler with storage and TTL
	handler := websocket.NewHandler(store, *roomTTL)

	upgrader := gws.NewUpgrader(handler, &gws.ServerOption{
		ParallelEnabled: true,
		Recovery:        gws.Recovery,
		PermessageDeflate: gws.PermessageDeflate{
			Enabled: true,
		},
	})

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		// Check protocol version from query parameter
		vStr := r.URL.Query().Get("v")
		clientVersion := 0
		if vStr != "" {
			clientVersion, _ = strconv.Atoi(vStr)
		}

		if clientVersion < protocol.ProtocolVersion {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUpgradeRequired) // 426
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":          "client_outdated",
				"message":        "Client version is outdated, please update",
				"server_version": protocol.ProtocolVersion,
			})
			return
		}
		if clientVersion > protocol.ProtocolVersion {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest) // 400
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":          "server_outdated",
				"message":        "Protocol version mismatch, server needs update",
				"server_version": protocol.ProtocolVersion,
			})
			return
		}

		// Version matches, proceed with upgrade
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
	log.Printf("Data directory: %s", *dataDir)
	log.Printf("Empty room TTL: %v", *roomTTL)

	server := &http.Server{Addr: addr}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		store.Close()
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
