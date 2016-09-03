package main

import (
    "net"
    "fmt"
    "github.com/shadowsocks/shadowsocks-go/shadowsocks"
)

func main() {
    var UH shadowsocks.S5UH
    UH.Init()
    var LH shadowsocks.S5LH
    LH.Init()
    UH.SocksHalf = LH
    LH.SocksHalf = UH
    TCPAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:1081")
    if nil != err {
        fmt.Println("Error get TCP Address")
    }
    fmt.Print(LH.IsRunning)
    LH.Listen(TCPAddr)
}
