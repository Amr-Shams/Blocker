package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/Amr-Shams/Blocker/cmd"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"github.com/spf13/cobra"
)

type wsMessage struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func executeCommand(baseCmd *cobra.Command, command string) ([]byte, error) {
	var outputBuf bytes.Buffer
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	_, w, _ := os.Pipe()
	os.Stdout = w
	os.Stderr = w
	baseCmd.SetOut(&outputBuf)
	baseCmd.SetErr(&outputBuf)
	baseCmd.SetArgs(strings.Split(command, " "))

	err := baseCmd.Execute()

	w.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	if err != nil {
		return nil, fmt.Errorf("failed to execute command: %v", err)
	}

	return outputBuf.Bytes(), nil
}

func handleWSConnection(conn net.Conn, n serverNode) {
	baseCmd := cmd.NewBaseCommand()

	for {
		msg, _, err := wsutil.ReadClientData(conn)
		if err != nil {
			if err != io.EOF {
				log.Printf("WebSocket read error: %v", err)
			}
			return
		}

		node, ok := n.(Node)
		if !ok || node == nil {
			log.Printf("Invalid node type or nil node")
			return
		}

		response, err := handleWSMessage(msg, baseCmd)
		if err != nil {
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

func handleWSMessage(msg []byte, baseCmd *cobra.Command) ([]byte, error) {
	var msgType wsMessage
	var response []byte
	var err error

	if err := json.Unmarshal(msg, &msgType); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %v", err)
	}

	command := msgType.Payload
	response, err = executeCommand(baseCmd, command)
	return response, err
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
			handleWSConnection(conn, n)
		}),
	}).Serve(l)
}
