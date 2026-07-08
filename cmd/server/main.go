package main

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"text/template"
	"time"

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

func getRunningPID() (int, bool) {
	home, _ := os.UserHomeDir()
	pidFilePath := filepath.Join(home, ".mac-remote-server.pid")
	data, err := os.ReadFile(pidFilePath)
	if err != nil {
		return 0, false
	}
	var pid int
	if _, err := fmt.Sscanf(string(data), "%d", &pid); err != nil {
		return 0, false
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return 0, false
	}
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		os.Remove(pidFilePath)
		return 0, false
	}
	return pid, true
}

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: mac-remote-server <command> [options]\n\n")
		fmt.Printf("Commands:\n")
		fmt.Printf("  start     Start the remote mouse WebSocket server (-d to run in background)\n")
		fmt.Printf("  stop      Stop the remote mouse WebSocket server running in the background\n\n")
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
		daemonFlag := startCmd.Bool("d", false, "Run server in the background (daemon mode)")

		if err := startCmd.Parse(os.Args[2:]); err != nil {
			log.Fatal("Failed to parse flags:", err)
		}

		if *daemonFlag {
			if pid, running := getRunningPID(); running {
				log.Fatalf("Server is already running in the background with PID %d.\n", pid)
			}

			binaryPath, err := os.Executable()
			if err != nil {
				log.Fatal("Failed to get executable path:", err)
			}

			args := []string{"start"}
			if *portFlag != "8025" {
				args = append(args, "-port", *portFlag)
			}
			if *hostFlag != "0.0.0.0" {
				args = append(args, "-host", *hostFlag)
			}
			if *devFlag {
				args = append(args, "-dev")
			}
			if *debugFlag {
				args = append(args, "-debug")
			}

			cmd := exec.Command(binaryPath, args...)
			cmd.SysProcAttr = &syscall.SysProcAttr{
				Setsid: true,
			}

			home, _ := os.UserHomeDir()
			logFilePath := filepath.Join(home, ".mac-remote-server.log")
			logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
			if err != nil {
				log.Fatal("Failed to open log file:", err)
			}
			defer logFile.Close()

			cmd.Stdout = logFile
			cmd.Stderr = logFile

			if err := cmd.Start(); err != nil {
				log.Fatal("Failed to start background process:", err)
			}

			pidFilePath := filepath.Join(home, ".mac-remote-server.pid")
			if err := os.WriteFile(pidFilePath, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644); err != nil {
				log.Println("Warning: Failed to write PID file:", err)
			}

			// Wait a bit to ensure it doesn't crash immediately (e.g. port already bound)
			time.Sleep(500 * time.Millisecond)
			if err := cmd.Process.Signal(syscall.Signal(0)); err != nil {
				log.Fatal("Server failed to start. Check the logs at: ", logFilePath)
			}

			fmt.Printf("✦ Mac Remote Server started in the background (PID: %d) ✦\n", cmd.Process.Pid)
			fmt.Printf("Logs are written to: %s\n", logFilePath)
			fmt.Printf("To stop the server, run: mac-remote-server stop\n")
			os.Exit(0)
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
		binaryPath, err := os.Executable()
		if err != nil {
			log.Println("Warning: failed to get absolute path of executable:", err)
			binaryPath, _ = filepath.Abs(os.Args[0])
		}

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

	case "stop":
		pid, running := getRunningPID()
		if !running {
			log.Fatal("Server is not running in the background.")
		}

		process, err := os.FindProcess(pid)
		if err != nil {
			log.Fatalf("Failed to find process %d: %v\n", pid, err)
		}

		err = process.Signal(syscall.SIGTERM)
		if err != nil {
			log.Fatalf("Failed to stop process %d: %v\n", pid, err)
		}

		home, _ := os.UserHomeDir()
		pidFilePath := filepath.Join(home, ".mac-remote-server.pid")
		os.Remove(pidFilePath)

		fmt.Printf("Stopped Mac Remote Server (PID %d)\n", pid)
		os.Exit(0)

	default:
		fmt.Printf("Unknown command: %s\n\n", subcommand)
		flag.Usage()
		os.Exit(1)
	}
}
