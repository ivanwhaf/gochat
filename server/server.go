package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gochat/core"
	"log"
	"net"
)

type SInterface interface {
	Init()
	Serve()
	Run()
	Send()
	Receive()
}

type Server struct {
	Network     string
	Address     string
	OnlineUsers map[int32]*net.Conn
	AllUsers    map[int32]core.User
}

func (s *Server) Init(network string, address string, users []core.User) {
	s.Network = network
	s.Address = address
	s.OnlineUsers = make(map[int32]*net.Conn)
	s.AllUsers = make(map[int32]core.User)
	for _, u := range users {
		s.AllUsers[u.Id] = u
	}
}

func (s *Server) Send(conn *net.Conn, sendData core.Data) {
	writer := bufio.NewWriter(*conn)
	jsonBytes, err := json.Marshal(sendData)
	if err != nil {
		fmt.Println("json marshal error : ", err)
		log.Println(err)
	}
	_, err = writer.WriteString(string(jsonBytes))
	if err != nil {
		fmt.Println("send data error : ", err)
		log.Println(err)
	}
	fmt.Println("send data to", (*conn).RemoteAddr())
	log.Println("send data to:", (*conn).RemoteAddr())
}

func (s *Server) Receive(conn *net.Conn, buf [4096]byte) (*core.Data, error) {
	reader := bufio.NewReader(*conn)

	n, err := reader.Read(buf[:])
	if err != nil {
		fmt.Println("reader.Read  error : ", err)
		return nil, err
	}
	recvData := string(buf[:n])
	var data core.Data
	err = json.Unmarshal([]byte(recvData), &data)
	if err != nil {
		fmt.Println("json unmarshal  error : ", err)
		return nil, err
	}

	return &data, nil
}

func (s *Server) Serve(conn *net.Conn) {
	defer func() {
		err := (*conn).Close()
		if err != nil {
			log.Println("connection close failed")
			return
		}
	}()
	var buf [4096]byte
	for {
		data, err := s.Receive(conn, buf)
		if err != nil {
			log.Println(err)
			return
		}
		// parse json data
		if data.DataType == core.DataTypeOperation {
			fromId := data.FromId
			if data.Operation == core.OperationLogin {
				// Login
				if _, ok := s.OnlineUsers[fromId]; !ok {
					s.OnlineUsers[fromId] = conn
					fmt.Println("User:", s.AllUsers[fromId].Nickname, "login")
					log.Println("User:", s.AllUsers[fromId].Nickname, "login")
				}
			} else if data.Operation == core.OperationLogout {
				// Logout
				if _, ok := s.OnlineUsers[fromId]; ok {
					delete(s.OnlineUsers, fromId)
					fmt.Println("User:", s.AllUsers[fromId].Nickname, "logout")
					log.Println("User:", s.AllUsers[fromId].Nickname, "logout")
				}
			}
		} else if data.DataType == core.DataTypeMessage {
			fromId := data.FromId
			toId := data.ToId
			if toId == -1 {
				//broadcast
				for k := range s.OnlineUsers {
					conn := s.OnlineUsers[k]
					sendData := core.Data{
						FromId:   fromId,
						ToId:     k,
						DataType: core.DataTypeMessage,
						Message:  data.Message,
					}
					s.Send(conn, sendData)
				}
			} else {
				// p2p
				if conn, ok := s.OnlineUsers[toId]; ok {
					sendData := core.Data{
						FromId:   fromId,
						ToId:     toId,
						DataType: core.DataTypeMessage,
						Message:  data.Message,
					}
					s.Send(conn, sendData)
				}
			}
		}
	}
}

func (s *Server) Run() {
	listener, err := net.Listen(s.Network, s.Address)
	if err != nil {
		log.Println(err)
		panic(err)
	}
	fmt.Println("gochat server start")
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println(err)
			log.Println(err)
			continue
		}
		fmt.Println("accept connection from:", conn.RemoteAddr())
		log.Println("accept connection from:", conn.RemoteAddr())
		go s.Serve(&conn)
	}
}
