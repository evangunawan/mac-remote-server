package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"

	"mac-remote-server/internal/infrastructure/macos"
	"mac-remote-server/internal/infrastructure/network"
)

//go:embed web/*
var webAssets embed.FS

func main() {
	// Custom usage output
	flag.Usage = func() {
		fmt.Printf("Usage: mac-remote-server <command> [options]\n\n")
		fmt.Printf("Commands:\n")
		fmt.Printf("  start     Start the remote mouse WebSocket server\n\n")
		fmt.Printf("Run 'mac-remote-server <command> -h' to see options for that command.\n")
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	subcommand := os.Args[1]

	switch subcommand {
	case "start":
		startCmd := flag.NewFlagSet("start", flag.ExitOnError)
		portFlag := startCmd.String("port", "8080", "Port to listen on")
		hostFlag := startCmd.String("host", "0.0.0.0", "Host address to bind to")
		devFlag := startCmd.Bool("dev", false, "Serve assets directly from disk instead of embed")

		// Parse custom sub-flags for the 'start' command
		if err := startCmd.Parse(os.Args[2:]); err != nil {
			log.Fatal("Failed to parse flags:", err)
		}

		// Initialize Infrastructure dependencies
		controller := macos.NewMacCursorController()

		var webSub fs.FS
		if *devFlag {
			fmt.Println("🛠️  Development Mode: Serving assets from local disk './cmd/server/web'")
			webSub = os.DirFS("./cmd/server/web")
		} else {
			// Read static HTML/JS/CSS files from the embedded folder and strip the 'web' folder prefix
			var err error
			webSub, err = fs.Sub(webAssets, "web")
			if err != nil {
				log.Fatal("Failed to load embedded web directory:", err)
			}
		}

		// Initialize and launch the network server
		srv := network.NewServer(*hostFlag, *portFlag, controller, webSub)
		if err := srv.Start(); err != nil {
			log.Fatal("Server encountered an error:", err)
		}

	default:
		fmt.Printf("Unknown command: %s\n\n", subcommand)
		flag.Usage()
		os.Exit(1)
	}
}
