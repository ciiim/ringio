package server

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"net"
	"strconv"
	"time"
)

func GetIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Println(err)
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
	return ""
}

type TaskType int

const (
	TaskTypeUpload TaskType = iota
	TaskTypeDownload
)

func genTaskID(fileHash string, taskType TaskType) string {
	timeStr := strconv.Itoa(int(time.Now().UnixMilli())) + fileHash
	sum := sha1.Sum([]byte(timeStr))
	return hex.EncodeToString(sum[:])

}
