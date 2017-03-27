package proto

import (
	"fmt"
	"log"
	"sync"

	"github.com/xtaci/kcp-go"
)

const (
	// DefaultReadBufferSize is the default read buffer size
	DefaultReadBufferSize = 4096 * 1024
	// DefaultWriteBufferSize is the default write buffer size
	DefaultWriteBufferSize = 4096 * 1024
	// DefaultDataShards is the default data shards
	DefaultDataShards = 10
	// DefaultParityShards is the default parity shards
	DefaultParityShards = 3
)

// unitizeSession sets default config to session
func unitizeSession(session *kcp.UDPSession) *kcp.UDPSession {
	session.SetReadBuffer(DefaultReadBufferSize)
	session.SetWriteBuffer(DefaultWriteBufferSize)
	session.SetNoDelay(1, 30, 2, 1)
	session.SetWindowSize(1024, 1024)
	session.SetACKNoDelay(true)
	return session
}

// ConnectionHandler is the handler of a single connection
type ConnectionHandler struct {
	Received func(typ uint16, obj interface{})
	Closed   func(reason string)
}

// Connection manage a x protocol connection
type Connection struct {
	// Session is a kcp session
	Session   *kcp.UDPSession
	Handler   *ConnectionHandler
	Converter Converter
	server    *Server
	suspend   *bool
	closed    bool
	lock      sync.Mutex
}

// CreateConnection Creates a connection to server
func CreateConnection(server string, converter Converter) (*Connection, error) {
	session, err := kcp.DialWithOptions(server, nil, DefaultDataShards, DefaultParityShards)
	if err != nil {
		return nil, err
	}
	return NewConnection(session, converter)
}

// NewConnection creates a connection from kcp session
func NewConnection(conn *kcp.UDPSession, converter Converter) (*Connection, error) {
	return &Connection{
		Session:   unitizeSession(conn),
		Converter: converter,
	}, nil
}

// Resume reads packets from connection
func (c *Connection) Resume(handler *ConnectionHandler) error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return fmt.Errorf("connection closed")
	}
	if c.suspend != nil && !*c.suspend {
		return fmt.Errorf("connection resumed")
	}
	c.Handler = handler
	suspend := false
	c.suspend = &suspend
	go func() {
		for !suspend {
			protocol := NewXProtocol()
			log.Println("===start===")
			_, err := protocol.ReadFrom(c.Session)
			log.Println("===start===", err)
			if err != nil {
				c.Close(err.Error())
				break
			}
			obj, err := c.Converter.ToObject(protocol)
			if err != nil {
				c.Close(err.Error())
				break
			}
			if c.Handler != nil && c.Handler.Received != nil {
				c.Handler.Received(protocol.Type, obj)
			}
		}
	}()
	return nil
}

// Suspend stops reading packets
func (c *Connection) Suspend() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return fmt.Errorf("connection closed")
	}
	if c.suspend == nil {
		return fmt.Errorf("connection suspended")
	}
	*c.suspend = true
	c.suspend = nil
	return nil
}

// Close closes current connection. A stoped connection can't send or recv any packet.
func (c *Connection) Close(reason string) {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.closed {
		return
	}
	c.closed = true
	if c.server != nil {
		c.server.Remove(c, reason)
	}
	c.Session.Close()
	if c.Handler != nil && c.Handler.Closed != nil {
		c.Handler.Closed(reason)
	}
	log.Printf("connection %s closed: %s", c.Session.RemoteAddr().String(), reason)
}

// Send sends a object data to remote
func (c *Connection) Send(obj interface{}) error {
	if c.closed {
		return fmt.Errorf("connection closed")
	}
	protocol, err := c.Converter.ToXProtocol(obj)
	if err != nil {
		return err
	}
	_, err = protocol.WriteTo(c.Session)
	return err
}

// ServerHandler is a handler of server
type ServerHandler struct {
	// Accepted handle a connection from server
	Accepted func(*Connection)
	// CLosed handle a closed connection
	Closed func(*Connection, string)
}

// Server manage a kcp server
type Server struct {
	Addr        string
	Listener    *kcp.Listener
	Connections map[string]*Connection
	Handler     *ServerHandler
	Converter   Converter
	lock        sync.Mutex
}

// NewServer creates a kcp server
func NewServer(addr string, converter Converter) (*Server, error) {
	listener, err := kcp.ListenWithOptions(addr, nil, DefaultDataShards, DefaultParityShards)
	if err != nil {
		return nil, err
	}
	listener.SetReadBuffer(DefaultReadBufferSize)
	listener.SetWriteBuffer(DefaultWriteBufferSize)
	return &Server{
		Addr:        addr,
		Listener:    listener,
		Connections: make(map[string]*Connection),
		Handler:     nil,
		Converter:   converter,
	}, nil
}

// Run runs the server
func (s *Server) Run(handler *ServerHandler) error {
	s.Handler = handler
	go func() {
		for true {
			session, err := s.Listener.AcceptKCP()
			if err != nil {
				log.Println("listen error", err)
			} else {
				log.Println("accept", session.RemoteAddr())
				s.handleConn(session)
			}
		}
	}()
	return nil
}

// Remove removes a connection from server
func (s *Server) Remove(conn *Connection, reason string) {
	key := conn.Session.RemoteAddr().String()
	if _, ok := s.Connections[key]; !ok {
		return
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	if _, ok := s.Connections[key]; ok {
		delete(s.Connections, key)
		conn.Close(reason)
		if s.Handler != nil && s.Handler.Closed != nil {
			s.Handler.Closed(conn, reason)
		}
	}
}

// handleConn handle a session
func (s *Server) handleConn(session *kcp.UDPSession) {
	conn, err := NewConnection(session, s.Converter)
	if err != nil {
		log.Println("connection error", err)
		return
	}
	s.Connections[conn.Session.RemoteAddr().String()] = conn
	if s.Handler != nil && s.Handler.Accepted != nil {
		s.Handler.Accepted(conn)
	}
}
