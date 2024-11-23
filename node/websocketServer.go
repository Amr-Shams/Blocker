package node

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var (
	availableCommands = []string{"list", "log"}
)

type wsMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func wsServer(l net.Listener, n serverNode) error {
	return (&http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := w.Header()
			header.Add("Sec-WebSocket-Accept", "true")
			header.Add("Connection", "Upgrade")
			header.Add("Upgrade", "websocket")
			conn, _, _, err := ws.UpgradeHTTP(r, w)
			if err != nil {
				log.Printf("Failed to upgrade connection: %v", err)
				return
			}
			go handleWSConnection(conn, n)
		}),
	}).Serve(l)
}
func handleWSConnection(conn net.Conn, n serverNode) {
	for {
		msg, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			if err != io.EOF {
				log.Printf("WebSocket read error: %v", err)
			}
			return
		}
		var node = n.(Node)
		if node == nil {
			log.Printf("Node is nil")
			return
		}
		response, err := handleWSMessage(msg, node)
		if err != nil {
			log.Printf("Error processing message: %v", err)
			errMsg, _ := json.Marshal(map[string]string{"error": err.Error()})
			if err := wsutil.WriteServerText(conn, errMsg); err != nil {
				log.Printf("Failed to write error message: %v", err)
				return
			}
			continue
		}

		if err := wsutil.WriteServerText(conn, response); err != nil {
			log.Printf("Failed to write response: %v", err)
			return
		}

	}
}
// ADD(20): CMD to the list of available commands
// to excute more commands like createing a blockchain etc 
// that means I have to refactore the code to work with the new commands 
func handleWSMessage(msg []byte, n Node) (response []byte, err error) {
	var msgType wsMessage
	if err := json.Unmarshal(msg, &msgType); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	}
	command := msgType.Payload
	switch command {
	case "list":
		response, err = json.Marshal(n.GetWallets())
	case "log":
		response, err = json.Marshal(stateName[n.GetStatus()])
	default:
		response, err = json.Marshal(availableCommands)
	}
	return response, err
}
