pacakge main

import {
    "net"
    "fmt"
    "github/shadowsocks/shadowsocks-go/shadowsocks"
}

type Socks5Relay struct {
    UF *shadowsocks.SocksHalf
    DF *shadowsocks.SocksHalf

func main() {
    UF := new(shadowsocks.SocksHalf)
    UF.Init()
    DF := new(shadowsocks.SocksHalf)
    DF.Init()
    UF.DF = DF
    DF.UF = UF

    TCPAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:1081")
    if nil != err {
        fmt.Println("Error get TCP Address")
    }
    DF.Listen(TCPAddr)
}
