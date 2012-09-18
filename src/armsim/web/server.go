package web

import (
	"log"
	"net/http"
	"armsim"
	"encoding/json"
	"code.google.com/p/go.net/websocket"
)

// Generic Message
type Message struct {
	Type string
	Content string
}

func (m *Message) Send(ws *websocket.Conn) {
		websocket.JSON.Send(ws, m)
}

type Server struct {
	Computer *armsim.Computer
	FilePath string
	Trace chan bool
	Halt chan bool
	Finished chan bool
}

var globalServer Server

func (s *Server) Serve(ws *websocket.Conn) {
	for {
		var m Message

		err := websocket.JSON.Receive(ws, &m)
		if err != nil {
			log.Println(err)
			ws.Close()
			return
		}

		log.Println(m)

		switch m.Type {
		case "hello": // Acknowledge ping
			s.SayHi(ws)
		case "load": // Load an ELF file by pathname
			s.Load(m, ws)
		case "reset": // Reset the simulator
			s.Reset(ws)
		case "start": // Run the program
			go s.Start(ws)
		case "status":
			s.UpdateStatus(ws)
		case "step": // Step the program
			s.Computer.Step()
			s.UpdateStatus(ws)
		case "stop": // Stop the program while running
			s.Stop(ws)
		case "quit": // Quit connection
			ws.Close()
			break
		default:
			m = Message{"error", "no command"}
			m.Send(ws)
		}
	}
}

func (s *Server) SayHi(ws *websocket.Conn) {
	m := Message{"status", "ready"}
	m.Send(ws)
	s.UpdateStatus(ws)
}

func (s *Server) Load(m Message, ws *websocket.Conn) {
	path := m.Content
	log.Println(path)
	s.FilePath = path

	s.Computer.Reset()
	s.Computer.LoadELF(path)
	m = Message{"status", "loaded"}
	m.Send(ws)
	s.UpdateStatus(ws)
}

func (s *Server) Start(ws *websocket.Conn) {
	go s.Computer.Run(s.Trace, s.Halt, s.Finished)
	m := Message{"status", "running"}
	m.Send(ws)

	// Wait for completion
	<- s.Finished
	s.UpdateStatus(ws)
	m = Message{"status", "finished"}
	m.Send(ws)

}

func (s *Server) Stop(ws *websocket.Conn) {
	s.Halt <- true
	m := Message{"status", "stopped"}
	m.Send(ws)
	s.UpdateStatus(ws)
}

func (s *Server) Reset(ws *websocket.Conn) {
	s.Computer.Reset()
	s.Computer.LoadELF(s.FilePath)
	s.UpdateStatus(ws)
}

func (s *Server) Quit(ws *websocket.Conn) {
}

func (s *Server) UpdateStatus(ws *websocket.Conn) {
	out, _ := json.Marshal(s.Computer.Status())
	m := Message{"update", string(out)}
	m.Send(ws)
}

func wsHandler(ws *websocket.Conn) {
	globalServer.Serve(ws)
}

func (s *Server) Launch() {
	globalServer = *s
	http.Handle("/", http.FileServer(http.Dir("assets/")))
	http.Handle("/ws", websocket.Handler(wsHandler))

	if err := http.ListenAndServe(":4567", nil); err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}
