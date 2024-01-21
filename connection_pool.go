package websocket

import "sync"

type ConnectionPool struct {
	mutex       sync.Mutex
	Connections map[string]*Connection
}

// Broadcast sends the frame f to all clients
func (p *ConnectionPool) Broadcast(f Frame) {
	for _, c := range p.Connections {
		c.Write(f)
	}
}

// SendTo sends a frame to a specific client
func (p *ConnectionPool) SendTo(id string, f Frame) {
	p.Connections[id].Write(f)
}

func (p *ConnectionPool) add(c *Connection) {
	p.mutex.Lock()
	p.Connections[c.Id] = c
	p.mutex.Unlock()
}

// Remove removes a connection from the connection pool
func (p *ConnectionPool) Remove(c *Connection) {
	p.mutex.Lock()
	delete(p.Connections, c.Id)
	p.mutex.Unlock()
}
