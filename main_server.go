package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/yyzcoder/crointra/crointra_data"
	"github.com/yyzcoder/crointra/util"
	"github.com/yyzcoder/yyznet/channel"
	"github.com/yyzcoder/yyznet/protocol"
	"github.com/yyzcoder/yyznet/server"
	"os"
	"sync"
)

var connectionsLock sync.Mutex
var connections map[int]*server.Connect
var channelClient channel.Client
var remoteServerPort int

func init() {

	channelAddr, _, channelPort, _, remotePort, err := util.GetConf()
	if err != nil {
		fmt.Println("start fail:", err)
		os.Exit(1)
	}
	remoteServerPort = remotePort
	//启动channelClient
	connections = make(map[int]*server.Connect, 0)
	channelClient = channel.Client{
		ChannelAddr: channelAddr,
		ChannelPort: channelPort,
	}
	channelClient.On("RETURN", func(data string) {
		//fmt.Println("RETURN")
		yyzData := new(crointra_data.CrointraData)
		if err := json.Unmarshal([]byte(data), yyzData); err != nil {
			fmt.Println("jsonunmarshal错误", err)
			return
		}
		//util.LogWrite(yyzData.Data, "")
		connptr, ok := connections[yyzData.ConnectId]
		if !ok {
			//fmt.Println("尝试返回信息时，连接池未找到这个连接id=", yyzData.ConnectId)
			return
		}
		d, err := base64.StdEncoding.DecodeString(yyzData.Data)
		if err != nil {
			//fmt.Println("base64 err:", err)
		}
		connptr.Write(d)
	})
	channelClient.On("SERVERCLOSE", func(data string) {
		//fmt.Println("SERVERCLOSE")
		yyzData := new(crointra_data.CrointraData)
		if err := json.Unmarshal([]byte(data), yyzData); err != nil {
			fmt.Println("jsonunmarshal错误", err)
			return
		}
		//fmt.Printf("收到server关闭返回，连接id=%d %s\n",yyzData.ConnectId,yyzData.Data)
		connptr, ok := connections[yyzData.ConnectId]
		if !ok {
			//fmt.Println("尝试关闭连接时，连接池未找到这个连接id=", yyzData.ConnectId)
			return
		}
		connectionsLock.Lock()
		delete(connections, yyzData.ConnectId)
		connectionsLock.Unlock()
		connptr.Close()
	})
	go channelClient.Run()
}

func main() {
	//启动本地tcp服务端
	tcpServer := server.Tcp{
		ListenAddr: "0.0.0.0",
		ListenPort: remoteServerPort,
		Protocol:   protocol.Tcp{},
	}

	tcpServer.OnStart = func() {
		fmt.Printf("local server start successful 0.0.0.0:%d\n", remoteServerPort)
	}

	tcpServer.OnConnect = func(conn *server.Connect) {
		//fmt.Println("有浏览器连接")
		yyzData := crointra_data.CrointraData{
			ConnectId: conn.Id,
		}
		jsonBytes, err := json.Marshal(yyzData)
		if err != nil {
			fmt.Println(err)
			return
		}
		connectionsLock.Lock()
		connections[conn.Id] = conn
		connectionsLock.Unlock()
		channelClient.Publish("CONNECT", string(jsonBytes))
	}
	tcpServer.OnMessage = func(conn *server.Connect, data []byte) {
		//fmt.Printf("有浏览器消息%d",len(data))
		yyzData := crointra_data.CrointraData{
			ConnectId: conn.Id,
			Data:      base64.StdEncoding.EncodeToString(data),
		}
		jsonBytes, err := json.Marshal(yyzData)
		if err != nil {
			fmt.Println(err)
			return
		}
		channelClient.Publish("MESSAGE", string(jsonBytes))
	}
	tcpServer.OnClose = func(conn *server.Connect) {
		yyzData := crointra_data.CrointraData{
			ConnectId: conn.Id,
		}
		jsonBytes, err := json.Marshal(yyzData)
		if err != nil {
			fmt.Println(err)
			return
		}
		connectionsLock.Lock()
		delete(connections, conn.Id)
		connectionsLock.Unlock()
		channelClient.Publish("USERCLOSE", string(jsonBytes))
	}
	tcpServer.Run()
	fmt.Println("程序结束")
}
