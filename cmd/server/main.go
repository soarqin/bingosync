package main

import (
	"bingosync/internal/overlay"
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

	// Overlay HTML page for OBS Browser Source
	http.HandleFunc("/overlay", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Cache-Control", "no-cache")
		w.Write(overlay.HTML)
	})

	// SSE stream endpoint - pushes state_update events to OBS overlay
	http.HandleFunc("/stream", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "missing token", http.StatusBadRequest)
			return
		}

		roomID, ok := handler.ResolveStreamToken(token)
		if !ok {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "streaming not supported", http.StatusInternalServerError)
			return
		}

		// Subscribe to room state updates
		ch, unsubscribe := handler.SubscribeSSE(roomID)
		defer unsubscribe()

		// Send current state immediately or a not_in_room event
		if initial := handler.GetRoomState(roomID); initial != nil {
			fmt.Fprintf(w, "event: state_update\ndata: %s\n\n", initial)
		} else {
			fmt.Fprintf(w, "event: not_in_room\ndata: {}\n\n")
		}
		flusher.Flush()

		// Stream state updates until client disconnects
		clientGone := r.Context().Done()
		for {
			select {
			case <-clientGone:
				return
			case payload, ok := <-ch:
				if !ok {
					return
				}
				fmt.Fprintf(w, "event: state_update\ndata: %s\n\n", payload)
				flusher.Flush()
			}
		}
	})

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("Starting BingoSync WebSocket server on %s", addr)
	log.Printf("Data directory: %s", *dataDir)
	log.Printf("Empty room TTL: %v", *roomTTL)

	server := &http.Server{Addr: addr}

	// Graceful shutdown: store.Close() is handled by defer above
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}
