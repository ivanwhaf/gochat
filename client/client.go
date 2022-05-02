package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gochat/core"
	"log"
	"net"
)

type CInterface interface {
	Init()
	Run()
	Login()
	Send()
	Receive()
	HandleRead()
	HandleWrite()
}

type Client struct {
	Network string
	Address string
	User    core.User
}

func (c *Client) Init(network string, address string) {
	c.Network = network
	c.Address = address
	c.User = core.User{
		Id:       1000,
		Nickname: "ivan",
		Sex:      core.SexMale,
		Password: "111111",
		Auth:     core.AuthAdmin,
	}

	c.User = core.User{
		Id:       1001,
		Nickname: "test admin",
		Sex:      core.SexMale,
		Password: "test",
		Auth:     core.AuthAdmin,
	}
}

func (c *Client) Send(conn *net.Conn, sendData core.Data) {
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
	err = writer.Flush()
	if err != nil {
		return
	}
	fmt.Println("send data to:", (*conn).RemoteAddr())
	log.Println("send data to:", (*conn).RemoteAddr())
}

func (c *Client) Receive(conn *net.Conn, buf [4096]byte) (*core.Data, error) {
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
	fmt.Println("receive data")
	return &data, nil
}

func (c *Client) HandleRead(conn *net.Conn) {
	var buf [4096]byte
	for {
		data, err := c.Receive(conn, buf)
		if err != nil {
			log.Println(err)
		}
		if data != nil && data.DataType == core.DataTypeMessage && data.ToId == c.User.Id {
			fmt.Println("receive data from:", data.FromId, data.Message)
			log.Println("receive data from:", data.FromId, data.Message)
		}
	}
}

func (c *Client) HandleWrite(conn *net.Conn) {
	var toId int32
	var message string
	for {
		_, err := fmt.Scanln(&toId, &message)
		if err != nil {
			return
		}
		sendData := core.Data{
			FromId:   c.User.Id,
			ToId:     toId,
			DataType: core.DataTypeMessage,
			Message:  message,
		}
		c.Send(conn, sendData)
	}
}

func (c *Client) Login(conn *net.Conn) {
	sendData := core.Data{
		FromId:    c.User.Id,
		ToId:      0, // To server
		DataType:  core.DataTypeOperation,
		Operation: core.OperationLogin,
	}
	c.Send(conn, sendData)
}

func (c *Client) Run() {
	fmt.Println("client start")
	conn, err := net.Dial(c.Network, c.Address)
	if err != nil {
		fmt.Println("connect to server error")
		log.Println(err)
		panic(err)
	}
	fmt.Println("connected to server:", conn.RemoteAddr())
	log.Println("connected to server:", conn.RemoteAddr())

	defer func() {
		err := conn.Close()
		if err != nil {
			return
		}
	}()

	c.Login(&conn)
	go c.HandleRead(&conn)
	go c.HandleWrite(&conn)
	for {
	}
}
