package util

import (
	"fmt"
	"gopkg.in/ini.v1"
	"os"
	"time"
)

func LogWrite(str string, filename string) {
	if filename == "" {
		filename = "runtime_log.txt"
	}
	f, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		fmt.Println("log err:", err)
		return
	}
	now := time.Now()
	log := fmt.Sprintf("[%s] %s\n", now.Format("2006/01/02 15:04"), str)
	f.Write([]byte(log))
	f.Close()
}

func GetConf() (channelAddr, localAddr string, channelPort, localPort, remotePort int, err error) {
	cfg, err := ini.Load("my.ini")
	if err != nil {
		return
	}
	channelAddr = cfg.Section("").Key("CHANNEL_ADDR").String()
	if len(channelAddr) < 1 {
		err = fmt.Errorf("missing CHANNEL_ADDR config")
		return
	}
	channelPort, err = cfg.Section("").Key("CHANNEL_PORT").Int()
	if err != nil {
		return
	}
	localPort, err = cfg.Section("").Key("LOCAL_PORT").Int()
	if err != nil {
		return
	}
	remotePort, err = cfg.Section("").Key("REMOTE_PORT").Int()
	if err != nil {
		return
	}
	localAddr = cfg.Section("").Key("LOCAL_ADDR").String()
	if len(localAddr) < 1 {
		err = fmt.Errorf("missing LOCAL_ADDR config")
	}
	return
}
