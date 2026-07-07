package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"text/template"
	"bytes"

	"github.com/getlantern/systray"
	"mac-remote-server/internal/infrastructure/macos"
	"mac-remote-server/internal/infrastructure/network"
	"mac-remote-server/internal/logging"
)

//go:embed web/*
var webAssets embed.FS

//go:embed trayicon.png
var trayIconBytes []byte

const plistLabel = "My Mac Remote"

var plistTemplate = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>{{.Label}}</string>
    <key>ProgramArguments</key>
    <array>
        <string>{{.BinaryPath}}</string>
        <string>start</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <false/>
</dict>
</plist>
`

func plistPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", plistLabel+".plist")
}

func isLoginItemInstalled() bool {
	_, err := os.Stat(plistPath())
	return err == nil
}

func installLoginItem(binaryPath string) error {
	type plistData struct {
		Label      string
		BinaryPath string
	}
	tmpl, err := template.New("plist").Parse(plistTemplate)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, plistData{Label: plistLabel, BinaryPath: binaryPath}); err != nil {
		return err
	}
	return os.WriteFile(plistPath(), buf.Bytes(), 0644)
}

func uninstallLoginItem() error {
	return os.Remove(plistPath())
}

func main() {
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
		portFlag := startCmd.String("port", "8025", "Port to listen on")
		hostFlag := startCmd.String("host", "0.0.0.0", "Host address to bind to")
		devFlag := startCmd.Bool("dev", false, "Serve assets directly from disk instead of embed")
		debugFlag := startCmd.Bool("debug", false, "Enable verbose debug logging")

		if err := startCmd.Parse(os.Args[2:]); err != nil {
			log.Fatal("Failed to parse flags:", err)
		}

		logging.SetDebug(*debugFlag)

		controller := macos.NewMacCursorController()

		var webSub fs.FS
		if *devFlag {
			fmt.Println("🛠️  Development Mode: Serving assets from local disk './cmd/server/web'")
			webSub = os.DirFS("./cmd/server/web")
		} else {
			var err error
			webSub, err = fs.Sub(webAssets, "web")
			if err != nil {
				log.Fatal("Failed to load embedded web directory:", err)
			}
		}

		srv := network.NewServer(*hostFlag, *portFlag, controller, webSub)

		go func() {
			if err := srv.Start(); err != nil {
				log.Fatal("Server encountered an error:", err)
			}
		}()

		// Resolve binary absolute path for the plist
		binaryPath, _ := filepath.Abs(os.Args[0])

		systray.Run(func() {
			systray.SetIcon(trayIconBytes)
			systray.SetTooltip("My Mac Remote")

			mURL := systray.AddMenuItem("Open: http://localhost:"+*portFlag, "Server address")
			mURL.Disable()
			systray.AddSeparator()

			// Open at Login toggle
			var mLogin *systray.MenuItem
			if isLoginItemInstalled() {
				mLogin = systray.AddMenuItem("✓ Open at Login", "Remove from login items")
			} else {
				mLogin = systray.AddMenuItem("Open at Login", "Add to login items")
			}

			systray.AddSeparator()
			mQuit := systray.AddMenuItem("Stop Server", "Stop the server and quit")

			go func() {
				for {
					select {
					case <-mLogin.ClickedCh:
						if isLoginItemInstalled() {
							if err := uninstallLoginItem(); err != nil {
								log.Println("Failed to remove login item:", err)
							} else {
								mLogin.SetTitle("Open at Login")
							}
						} else {
							if err := installLoginItem(binaryPath); err != nil {
								log.Println("Failed to install login item:", err)
							} else {
								mLogin.SetTitle("✓ Open at Login")
							}
						}
					case <-mQuit.ClickedCh:
						systray.Quit()
						os.Exit(0)
					}
				}
			}()
		}, func() {})

	default:
		fmt.Printf("Unknown command: %s\n\n", subcommand)
		flag.Usage()
		os.Exit(1)
	}
}
