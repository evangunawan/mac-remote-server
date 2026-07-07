package network

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"mac-remote-server/internal/domain/input"
	"mac-remote-server/internal/logging"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow local Wi-Fi devices to connect
		return true
	},
}

// wsMessage matches the JSON structure sent from the phone web client.
type wsMessage struct {
	Action    string  `json:"action"`
	Dx        float64 `json:"dx,omitempty"`
	Dy        float64 `json:"dy,omitempty"`
	Button    string  `json:"button,omitempty"`
	Direction string  `json:"direction,omitempty"`
	Text      string  `json:"text,omitempty"`
	Name      string  `json:"name,omitempty"`
	Active    bool    `json:"active,omitempty"`
}

type Server struct {
	host       string
	port       string
	controller input.CursorController
	webAssets  fs.FS
	clients    map[*websocket.Conn]bool
	clientsMu  sync.Mutex
}

// NewServer initializes a new Server infrastructure layer.
func NewServer(host, port string, controller input.CursorController, webAssets fs.FS) *Server {
	return &Server{
		host:       host,
		port:       port,
		controller: controller,
		webAssets:  webAssets,
		clients:    make(map[*websocket.Conn]bool),
	}
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	s.clientsMu.Lock()
	s.clients[conn] = true
	s.clientsMu.Unlock()

	defer func() {
		s.clientsMu.Lock()
		delete(s.clients, conn)
		s.clientsMu.Unlock()
	}()

	log.Printf("Client connected from: %s\n", conn.RemoteAddr().String())

	for {
		_, payload, err := conn.ReadMessage()
		if err != nil {
			log.Println("Client disconnected:", err)
			break
		}

		var msg wsMessage
		if err := json.Unmarshal(payload, &msg); err != nil {
			log.Println("Invalid WebSocket payload:", err)
			continue
		}

		if msg.Action != "move" && msg.Action != "drag" {
			logging.Debugf("WS Action received: %s (Payload: %+v)\n", msg.Action, msg)
		}

		switch msg.Action {
		case "move":
			s.controller.Move(msg.Dx, msg.Dy)
		case "drag":
			s.controller.Drag(msg.Dx, msg.Dy)
		case "mousedown":
			s.controller.MouseDown()
		case "mouseup":
			s.controller.MouseUp()
		case "click":
			if msg.Button == "right" {
				s.controller.RightClick()
			} else {
				s.controller.LeftClick()
			}
		case "scroll":
			// dx and dy are received from the trackpad client and mapped to Scroll
			s.controller.Scroll(int(msg.Dx), int(msg.Dy))
		case "zoom":
			s.controller.Zoom(msg.Direction)
		case "playpause":
			s.controller.PlayPause()
		case "next":
			s.controller.NextTrack()
		case "previous":
			s.controller.PreviousTrack()
		case "volup":
			s.controller.VolumeUp()
		case "voldown":
			s.controller.VolumeDown()
		case "mute":
			s.controller.Mute()
		case "type":
			s.controller.TypeString(msg.Text)
		case "key":
			s.controller.PressKey(msg.Name)
		}
	}
}

// broadcast sends a WebSocket payload to all connected clients.
func (s *Server) broadcast(msg wsMessage) {
	s.clientsMu.Lock()
	defer s.clientsMu.Unlock()

	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}

	for conn := range s.clients {
		err := conn.WriteMessage(websocket.TextMessage, payload)
		if err != nil {
			log.Println("Broadcast failed, closing connection:", err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}

// Start launches the HTTP server, serves static files, and starts WebSocket listening.
func (s *Server) Start() error {
	trusted := s.controller.IsTrusted()
	mux := http.NewServeMux()

	if s.webAssets != nil {
		mux.Handle("/", http.FileServer(http.FS(s.webAssets)))
	}

	mux.HandleFunc("/ws", s.handleWebSocket)

	addr := net.JoinHostPort(s.host, s.port)
	ips := s.getLocalIPs()

	fmt.Printf("\n\x1b[1;36m====================================================\x1b[0m\n")
	fmt.Printf("\x1b[1;32m✦ Mac Remote Server is running! ✦\x1b[0m\n")
	if trusted {
		fmt.Printf("🔒 macOS Accessibility: \x1b[1;32mTRUSTED\x1b[0m (Focus detection active)\n")
	} else {
		fmt.Printf("⚠️  macOS Accessibility: \x1b[1;31mNOT TRUSTED\x1b[0m (Focus detection will be disabled)\n")
	}
	fmt.Printf("Connect your phone/tablet to the same Wi-Fi network\n")
	fmt.Printf("and visit the following address in your browser:\n\n")
	
	if len(ips) == 0 {
		fmt.Printf("  👉 \x1b[1;33mhttp://localhost:%s\x1b[0m (No Wi-Fi IP detected)\n", s.port)
	} else {
		for _, ip := range ips {
			fmt.Printf("  👉 \x1b[1;33mhttp://%s:%s\x1b[0m\n", ip, s.port)
		}
	}
	fmt.Printf("\n\x1b[1;30m(Press Ctrl+C to stop the server)\x1b[0m\n")
	fmt.Printf("\x1b[1;36m====================================================\x1b[0m\n\n")

	return http.ListenAndServe(addr, mux)
}

func (s *Server) getLocalIPs() []string {
	var ips []string
	ifaces, err := net.Interfaces()
	if err != nil {
		return ips
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			ips = append(ips, ip.String())
		}
	}
	return ips
}
