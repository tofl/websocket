package websocket

import (
	"bufio"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

func upgrade(c *net.Conn, req *http.Request, path string, allowedOrigin string, allowedProtocols []string, allowedExtensions []string) (*Connection, error) {
	handshakeFail := "the websocket handshake has failed: "
	newConn := Connection{}

	// Check that the request is valid
	if req.Method != http.MethodGet || req.URL.String() != path {
		return nil, errors.New(handshakeFail + "wrong url")
	}

	if !strings.EqualFold(req.Header.Get("upgrade"), "websocket") {
		return nil, errors.New(handshakeFail + "missing or wrong 'upgrade' header")
	}

	if !strings.EqualFold(req.Header.Get("Connection"), "upgrade") {
		return nil, errors.New(handshakeFail + "missing or wrong 'connection' header")
	}

	// |Sec-WebSocket-Key|: base64-encoded value that, when decoded, is 16 bytes in length.
	secWebsocketKey := req.Header.Get("Sec-WebSocket-Key")
	secWebsocketKeyDecoded, err := base64.StdEncoding.DecodeString(secWebsocketKey)
	if len(secWebsocketKeyDecoded) != 16 || err != nil {
		return nil, errors.New(handshakeFail + "missing or wrong 'sec-websocket-key' header")
	}

	if req.Header.Get("Sec-WebSocket-Version") != "13" {
		return nil, errors.New(handshakeFail + "missing or wrong 'sec-websocket-version' header")
	}

	// Optional: |Origin|
	origin := req.Header.Get("Origin")
	newConn.Origin = origin
	if allowedOrigin != "*" && origin != allowedOrigin {
		return nil, errors.New(handshakeFail + "origin not allowed")
	}

	// Optional: |Sec-WebSocket-Protocol| and |Sec-WebSocket-Extensions|
	protocols := strings.Split(req.Header.Get("Sec-WebSocket-Protocol"), ", ")
	extensions := strings.Split(req.Header.Get("Sec-WebSocket-Extensions"), ", ")

	newConn.isOpen = true

	newConn.Id = uuid.NewString()

	// Shape and send the response
	wsKeyDec := sha1.Sum([]byte(secWebsocketKey + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"))
	wsKey := base64.StdEncoding.EncodeToString(wsKeyDec[:])

	response := fmt.Sprintf("HTTP/1.1 101 Switching Protocols\r\n"+
		"upgrade: websocket\r\n"+
		"Connection: upgrade\r\n"+
		"Sec-WebSocket-Accept: %s\r\n", wsKey)

	for _, v := range protocols {
		if slices.Contains(allowedProtocols, v) {
			newConn.Protocol = v
			break
		}
	}
	if newConn.Protocol != "" {
		response += fmt.Sprintf("Sec-WebSocket-Protocol: %s\r\n", newConn.Protocol)
	}

	for _, v := range extensions {
		if slices.Contains(allowedExtensions, v) {
			newConn.Extensions = append(newConn.Extensions, v)
		}
	}
	if len(newConn.Extensions) > 0 {
		response += fmt.Sprintf("Sec-WebSocket-Extensions:", strings.Join(newConn.Extensions, ", "))
	}

	response += "\r\n"

	// Return the connection to the caller
	(*c).Write([]byte(response))

	newConn.Conn = c
	return &newConn, nil
}

func handleWebsocket(c *net.Conn, r func(c *Connection, cp *ConnectionPool, frame Frame), cp *ConnectionPool, path string) {
	// 1. Read http request
	buf := bufio.NewReader(*c)
	req, err := http.ReadRequest(buf)
	if err != nil {
		fmt.Println("Error reading the http request:", err)
		return
	}

	// 2. Check handshake
	connection, err := upgrade(c, req, path, "*", nil, nil)
	if err != nil {
		r := http.Response{Status: strconv.Itoa(http.StatusBadRequest)}
		r.Write(*c)
		return
	}

	// 3. Manage connection (add it to a pool)
	cp.Add(connection)

	// 4. Read frame
	for (*connection).isOpen {
		(*connection).OnRead(r, cp)
	}
}

func NewServer(address, path string, r func(c *Connection, cp *ConnectionPool, frame Frame)) {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("error listening to connections")
		return
	}

	defer listener.Close()

	connectionPool := ConnectionPool{Connections: make(map[string]*Connection)}

	var conn net.Conn
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("error accepting connection")
			return
		}

		go handleWebsocket(&conn, r, &connectionPool, path)
	}

	defer conn.Close()
}
