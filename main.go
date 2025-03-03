package main

import (
	"context"
	"embed"
	"log"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend
var assets embed.FS

// setupEnvironment creates the temp directory and required files if they don't exist.
func setupEnvironment() {
	// Create the "temp" directory if it doesn't exist.
	if _, err := os.Stat("temp"); os.IsNotExist(err) {
		if err := os.Mkdir("temp", 0755); err != nil {
			log.Fatalf("Failed to create temp directory: %v", err)
		}
	}

	// List of default files to ensure exist.
	defaultFiles := []string{"apt.txt", "history.txt"}
	
		path := "temp/" + defaultFiles[0]
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.WriteFile(path, []byte("pcsk_2kqki5_2rQvhRQddcC9sMNFfKBCDuK6kbtWqFGUMLW7UJxo8pdNaxEg1dMooPtW2CT14DR"), 0644); err != nil {
				log.Fatalf("Failed to create file %s: %v", path, err)
			}
		}
	
}

// setupEnvVariables sets default environment variables if they are not already defined.
func setupEnvVariables() {
	// Set a default Pinecone API key if it's not already set.
	if os.Getenv("PINECONE_API_KEY") == Readapidata() {
		os.Setenv("PINECONE_API_KEY", "pcsk_2kqki5_2rQvhRQddcC9sMNFfKBCDuK6kbtWqFGUMLW7UJxo8pdNaxEg1dMooPtW2CT14DR")
	}
	// Set JSC_SIGNAL_FOR_GC to a safe signal number to avoid conflicts.
	if os.Getenv("JSC_SIGNAL_FOR_GC") == "" {
		os.Setenv("JSC_SIGNAL_FOR_GC", "31")
	}
}

// App holds your application state.


func main() {
	// Pre-configure the environment
	setupEnvironment()
	setupEnvVariables()

	// Create an instance of your app.
	app := NewApp()

	// Run the Wails application.
	err := wails.Run(&options.App{
		Title:            "mindaidesk",
		Width:            1024,
		Height:           768,
		AssetServer:      &assetserver.Options{Assets: assets},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup: func(ctx context.Context) {
			app.ctx = ctx
		},
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		log.Fatalf("Error running app: %v", err)
	}
}
