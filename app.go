package main

import (
	"context"
	"os"
)

// App struct
type App struct {
	ctx       context.Context
	serverURL string
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		serverURL: getEnvOrDefault("BINGO_SERVER_URL", "ws://localhost:8765/ws"),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetServerURL returns the WebSocket server URL
func (a *App) GetServerURL() string {
	return a.serverURL
}

// SetServerURL sets the WebSocket server URL
func (a *App) SetServerURL(url string) {
	a.serverURL = url
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
