package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/yyzcoder/crointra/crointra_data"
	"github.com/yyzcoder/crointra/util"
	"github.com/yyzcoder/yyznet/channel"
	"github.com/yyzcoder/yyznet/client"
	"os"
	"strings"
	"sync"
)

var localTcpClientsLock sync.Mutex
var localTcpClients map[int]*client.Tcp
var busId = ""
func init() {
	busId = os.Args[1]
	if strings.Trim(busId," ") == ""{
		fmt.Println("busId can't be empty")
		os.Exit(1)
	}
	localTcpClients = make(map[int]*client.Tcp, 0)
}

func main() {

	channelAddr, localAddr, channelPort, localPort, _, err := util.GetConf()
	if err != nil {
		fmt.Println("start fail:", err)
	}

	channelClient := &channel.Client{
		ChannelAddr: channelAddr,
		ChannelPort: channelPort,
	}
	channelClient.On(busId+"CONNECT", func(data string) {
		//fmt.Println("CONNECT")
		yyzData := new(crointra_data.CrointraData)
		if err := json.Unmarshal([]byte(data), yyzData); err != nil {
			fmt.Println("jsonunmarshal错误", err)
			return
		}
		localTcp := new(client.Tcp)
		localTcp.Addr = localAddr
		localTcp.Port = localPort
		localTcp.OnConnect = func() {
			fmt.Println("本地连接成功")
		}
		localTcp.OnMessage = func(data []byte) {
			fmt.Printf("数据回传%d\n", len(data))
			ryyzData := crointra_data.CrointraData{
				ConnectId: yyzData.ConnectId,
				Data:      base64.StdEncoding.EncodeToString(data),
			}
			jsonStr, _ := json.Marshal(ryyzData)
			channelClient.Publish(busId+"RETURN", string(jsonStr))
		}

		localTcp.OnClose = func() {
			fmt.Println("本地连接关闭")
			ryyzData := crointra_data.CrointraData{
				ConnectId: yyzData.ConnectId,
			}
			jsonStr, _ := json.Marshal(ryyzData)
			//本地连接关闭时，删除map，释放资源
			localTcpClientsLock.Lock()
			delete(localTcpClients, yyzData.ConnectId)
			localTcpClientsLock.Unlock()
			channelClient.Publish(busId+"SERVERCLOSE", string(jsonStr))
		}
		localTcpClientsLock.Lock()
		localTcpClients[yyzData.ConnectId] = localTcp
		localTcpClientsLock.Unlock()
		client.TcpWg.Add(1)
		go localTcp.Exec()
		client.TcpWg.Wait()
	})

	channelClient.On(busId+"MESSAGE", func(data string) {
		//fmt.Printf("收到消息%d\n", len(data))
		yyzData := new(crointra_data.CrointraData)
		if err := json.Unmarshal([]byte(data), yyzData); err != nil {
			fmt.Println("jsonunmarshal错误", err)
			return
		}
		localTcp, ok := localTcpClients[yyzData.ConnectId]
		if !ok {
			//fmt.Println("但本地连接尚未建立")
			return
		}
		d, err := base64.StdEncoding.DecodeString(yyzData.Data)
		if err != nil {
			fmt.Println("base64 err:", err)
		}
		localTcp.Send(d)
	})

	channelClient.On(busId+"USERCLOSE", func(data string) {
		//fmt.Printf("收到用户关闭%d\n", len(data))
		yyzData := new(crointra_data.CrointraData)
		if err := json.Unmarshal([]byte(data), yyzData); err != nil {
			fmt.Println("jsonunmarshal错误", err)
			return
		}
		localTcp, ok := localTcpClients[yyzData.ConnectId]
		if !ok {
			//fmt.Println("但本地连接尚未建立")
			return
		}
		localTcpClientsLock.Lock()
		delete(localTcpClients, yyzData.ConnectId)
		localTcpClientsLock.Unlock()
		localTcp.Close()
	})

	channelClient.Run()
	fmt.Println("程序结束")
}
