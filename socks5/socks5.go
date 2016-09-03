package main

import (
    "net"
    "fmt"
    "github.com/shadowsocks/shadowsocks-go/shadowsocks"
)

type Socks5Relay struct {
    UH shadowsocks.SocksHalf
    LH shadowsocks.SocksHalf
}

func main() {
    var relay Socks5Relay
    var UH shadowsocks.S5UH
    UH.Init()
    var LH shadowsocks.S5LH
    LH.Init()
    UH.SocksHalf = relay.LH
    LH.SocksHalf = relay.UH
    relay.UH = UH
    relay.LH = LH

    TCPAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:1081")
    if nil != err {
        fmt.Println("Error get TCP Address")
    }
    relay.LH.Listen(TCPAddr)
}
