package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gochat/core"
	"log"
	"net"
	"sync"
	"time"
)

type SInterface interface {
	Init()
	Serve()
	Run()
	Send()
	Receive()
	CheckOnline()
}

type Server struct {
	Network     string
	Address     string
	OnlineUsers map[int32]*net.Conn
	AllUsers    map[int32]core.User
	rwMutex     sync.RWMutex
}

// Init client initialization
func (s *Server) Init(network string, address string, users []core.User) {
	s.Network = network
	s.Address = address
	s.OnlineUsers = make(map[int32]*net.Conn)
	s.AllUsers = make(map[int32]core.User)
	for _, u := range users {
		s.AllUsers[u.Id] = u
	}
}

// Send package the write process
// Returns the error during the process
func (s *Server) Send(conn *net.Conn, sendData core.Data) error {
	writer := bufio.NewWriter(*conn)
	jsonBytes, err := json.Marshal(sendData)
	if err != nil {
		log.Println("json marshal error:", err)
		fmt.Println("json marshal error:", err)
		return err
	}
	_, err = writer.WriteString(string(jsonBytes))
	if err != nil {
		log.Println("write json data error:", err)
		fmt.Println("write json data error:", err)
		return err
	}
	err = writer.Flush()
	if err != nil {
		log.Println("writer flush error:", err)
		fmt.Println("writer flush error:", err)
		return err
	}
	log.Println("Send data package to:", (*conn).RemoteAddr())
	return nil
}

// Receive package the read process
// Returns pointer of Data
func (s *Server) Receive(conn *net.Conn, buf [4096]byte) (*core.Data, error) {
	reader := bufio.NewReader(*conn)
	n, err := reader.Read(buf[:])
	if err != nil {
		log.Println("reader.Read error:", err)
		fmt.Println("reader.Read error:", err)
		return nil, err
	}
	recvData := string(buf[:n])
	var data core.Data
	err = json.Unmarshal([]byte(recvData), &data)
	if err != nil {
		log.Println("json unmarshal error:", err)
		fmt.Println("json unmarshal error:", err)
		return nil, err
	}
	log.Println("Receive data package from:", (*conn).RemoteAddr())
	return &data, nil
}

// Serve server $ main method $
// Handle packages from clients
func (s *Server) Serve(conn *net.Conn) {
	defer func() {
		err := (*conn).Close()
		if err != nil {
			log.Println("connection close failed")
			fmt.Println("connection close failed")
			return
		}
	}()
	var buf [4096]byte
	for {
		data, err := s.Receive(conn, buf)
		if err != nil {
			log.Println("Client status error")
			fmt.Println("Client status error")
			return
		}
		if data.DataType == core.DataTypeOperation {
			// Operation Data Package
			fromId := data.FromId
			if data.Operation == core.OperationLogin {
				// Login
				s.rwMutex.Lock()
				s.OnlineUsers[fromId] = conn
				s.rwMutex.Unlock()
				log.Println("User:", s.AllUsers[fromId].Nickname, "login")
				fmt.Println("User:", s.AllUsers[fromId].Nickname, "login")
			} else if data.Operation == core.OperationLogout {
				// Logout
				s.rwMutex.Lock()
				if _, ok := s.OnlineUsers[fromId]; ok {
					delete(s.OnlineUsers, fromId)
					log.Println("User:", s.AllUsers[fromId].Nickname, "logout")
					fmt.Println("User:", s.AllUsers[fromId].Nickname, "logout")
				}
				s.rwMutex.Unlock()
			}
		} else if data.DataType == core.DataTypeMessage {
			// Message Data Package
			fromId := data.FromId
			toId := data.ToId
			if toId == -1 {
				// toId -1 means broadcast data package
				// Broadcast to all online clients
				s.rwMutex.RLock()
				for k := range s.OnlineUsers {
					tmpConn := s.OnlineUsers[k]
					err = s.Send(tmpConn, *data)
					if err != nil {
						log.Println("Send broadcast message data to:", k, "error")
						fmt.Println("Send broadcast message data to:", k, "error")
						break
					}
				}
				s.rwMutex.RUnlock()
			} else {
				// Client->Server->Client
				s.rwMutex.RLock()
				if tmpConn, ok := s.OnlineUsers[toId]; ok && toId != fromId {
					err = s.Send(tmpConn, *data)
					if err != nil {
						log.Println("Send client message data to:", toId, "error")
						fmt.Println("Send client message data to:", toId, "error")
					}
					log.Println("Send client message data to:", toId)
					fmt.Println("Send client message data to:", toId)
				}
				s.rwMutex.RUnlock()
			}
		}
	}
}

// CheckOnline check whether user is online every N seconds
// Delete user which isn't online
func (s *Server) CheckOnline() {
	tick := time.Tick(5 * time.Second)
	for {
		select {
		case <-tick:
			// s.OnlineUsers may have concurrency bugs!!
			for k, conn := range s.OnlineUsers {
				// Heartbeat Package
				sendData := core.Data{
					FromId:    0,
					ToId:      k,
					DataType:  core.DataTypeOperation,
					Operation: core.OperationKeepAlive,
				}
				err := s.Send(conn, sendData)
				if err != nil {
					s.rwMutex.Lock()
					delete(s.OnlineUsers, k) // delete user
					s.rwMutex.Unlock()
					log.Println("User", s.AllUsers[k].Nickname, "offline")
					fmt.Println("User", s.AllUsers[k].Nickname, "offline")
				}
			}
		default:
			time.Sleep(1 * time.Second)
		}
	}
}

// Run start run the server
func (s *Server) Run() {
	listener, err := net.Listen(s.Network, s.Address)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	fmt.Println("Gochat server start")

	go s.CheckOnline()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("listener accept error", err)
			fmt.Println("listener accept error", err)
			continue
		}
		log.Println("Accept connection from address:", conn.RemoteAddr())
		fmt.Println("Accept connection from address:", conn.RemoteAddr())
		go s.Serve(&conn)
	}
}
