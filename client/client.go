package client

import (
	"bufio"
	"encoding/json"
	"fmt"
	"gochat/config"
	"gochat/core"
	"log"
	"net"
	"time"
)

type CInterface interface {
	Init()
	Run()
	Login()
	Connect()
	Send()
	Receive()
	HandleRead()
	HandleWrite()
}

type Client struct {
	Network        string
	Address        string
	User           core.User
	Friends        map[int32]core.User
	serverConn     *net.Conn
	isReconnecting bool
}

func (c *Client) Init(network string, address string) {
	c.Network = network
	c.Address = address
	c.Friends = make(map[int32]core.User)
	for _, u := range config.GetConfig().Users {
		c.Friends[u.Id] = u
	}
	c.User = config.GetConfig().Users[0]
	c.User = config.GetConfig().Users[1]
}

func (c *Client) Send(conn *net.Conn, sendData core.Data) error {
	writer := bufio.NewWriter(*conn)

	jsonBytes, err := json.Marshal(sendData)
	if err != nil {
		log.Println("json marshal error:", err)
		fmt.Println("json marshal error:", err)
		return err
	}
	_, err = writer.WriteString(string(jsonBytes))
	if err != nil {
		log.Println("send data error:", err)
		fmt.Println("send data error:", err)
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

func (c *Client) Receive(conn *net.Conn, buf [4096]byte) (*core.Data, error) {
	reader := bufio.NewReader(*conn)
	n, err := reader.Read(buf[:])
	if err != nil {
		fmt.Println("reader.Read error:", err)
		return nil, err
	}
	recvData := string(buf[:n])
	var data core.Data
	err = json.Unmarshal([]byte(recvData), &data)
	if err != nil {
		fmt.Println("json unmarshal error:", err)
		return nil, err
	}
	log.Println("Receive data package from:", (*conn).RemoteAddr())
	return &data, nil
}

func (c *Client) HandleRead() {
	var buf [4096]byte
	for {
		var data *core.Data
		var err error
		for i := 0; i <= 5; i++ {
			data, err = c.Receive(c.serverConn, buf)
			if err == nil {
				c.isReconnecting = false
				break
			} else {
				log.Println("Receive from server error, try to receive again")
				fmt.Println("Receive from server error, try to receive again")
			}
			c.isReconnecting = true
			time.Sleep(2 * time.Second)
			if i == 5 {
				log.Println("Lose connection from server, try to reconnect")
				fmt.Println("Lose connection from server, try to reconnect")
				for {
					tmpConn, err := c.Connect()
					if err != nil {
						continue
					}
					c.serverConn = &tmpConn
					log.Println("Reconnect to server successfully")
					fmt.Println("Reconnect to server successfully")
					isSuccess := c.Login()
					if !isSuccess {
						continue
					}
					break
				}
			}
		}

		if data != nil && data.DataType == core.DataTypeMessage && data.ToId == c.User.Id {
			fmt.Println("Receive message from user:", c.Friends[data.FromId].Nickname, "Message:", data.Message)
			log.Println("Receive message from user:", c.Friends[data.FromId].Nickname, "Message:", data.Message)
		}
		if data != nil && data.DataType == core.DataTypeOperation && data.Operation == core.OperationKeepAlive {
		}
	}
}

func (c *Client) HandleWrite() {
	var toId int32
	var message string
	for {
		fmt.Println("Please input $id and $message (split by space) :")
		_, err := fmt.Scanln(&toId, &message)
		if err != nil {
			fmt.Println("scan line error")
			continue
		}
		sendData := core.Data{
			FromId:   c.User.Id,
			ToId:     toId,
			DataType: core.DataTypeMessage,
			Message:  message,
		}
		err = c.Send(c.serverConn, sendData)
		if err != nil {
			continue
		}
		fmt.Println("Send message to user:", c.Friends[toId].Nickname, "Message:", message)
		log.Println("Send message to user:", c.Friends[toId].Nickname, "Message:", message)
	}
}

func (c *Client) Login() bool {
	sendData := core.Data{
		FromId:    c.User.Id,
		ToId:      0, // To server
		DataType:  core.DataTypeOperation,
		Operation: core.OperationLogin,
	}
	err := c.Send(c.serverConn, sendData)
	if err != nil {
		fmt.Println("Login failed")
		return false
	}
	fmt.Println("Login successfully")
	return true
}

func (c *Client) Connect() (net.Conn, error) {
	conn, err := net.Dial(c.Network, c.Address)
	if err != nil {
		fmt.Println("Connect to server error")
		log.Println(err)
		return nil, err
	}
	fmt.Println("Connected to server:", conn.RemoteAddr())
	log.Println("Connected to server:", conn.RemoteAddr())
	return conn, nil
}

func (c *Client) Run() {
	fmt.Println("Client start")
	conn, err := c.Connect()
	if err != nil {
		panic(err)
	}
	c.serverConn = &conn

	defer func() {
		err := conn.Close()
		if err != nil {
			return
		}
	}()

	isSuccess := c.Login()
	if isSuccess == true {
		go c.HandleRead()
		go c.HandleWrite()
		select {}
	}
}
