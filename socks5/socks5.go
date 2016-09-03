package main

import (
    "net"
    "fmt"
     ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
)

func main() {
    var UH ss.S5UH
    UH.Init()
    var LH ss.S5LH
	ss.Log.Init("/tmp/ss.debug.log", ss.INFO)
    LH.Init()
    TCPAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:1081")
    if nil != err {
        fmt.Println("Error get TCP Address")
    }
    LH.Listen(TCPAddr)
	ss.Log.Finish()
}
