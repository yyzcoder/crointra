package main

import (
	"fmt"
	"github.com/yyzcoder/crointra/util"
	"github.com/yyzcoder/yyznet/channel"
)

func main() {
	_, _, channelPort, _, _, err := util.GetConf()
	server := channel.Server{
		ListenAddr: "0.0.0.0",
		ListenPort: channelPort,
	}
	err = server.Run()
	if err != nil {
		fmt.Println("start error:", err)
	}
}
