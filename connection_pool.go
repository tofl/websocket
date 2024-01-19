package websocket

import "sync"

type ConnectionPool struct {
	mutex       sync.Mutex
	Connections map[string]*Connection
}

func (p *ConnectionPool) Broadcast(f Frame) {
	for _, c := range p.Connections {
		c.Write(f)
	}
}

func (p *ConnectionPool) SendTo(id string, f Frame) {
	p.Connections[id].Write(f)
}

func (p *ConnectionPool) Add(c *Connection) {
	p.mutex.Lock()
	p.Connections[c.Id] = c
	p.mutex.Unlock()
}

func (p *ConnectionPool) Remove(c *Connection) {
	p.mutex.Lock()
	delete(p.Connections, c.Id)
	p.mutex.Unlock()
}
