package shadowsocks

import (
	"net"
)

type SocksServer interface {
	Listen
	Authenticate
	Request
	Read
	Write
	Bind
}

type SocksClient interface {
	Listen
	Authenticate
	Request
	Read
	Write
	Bind
}

type Socks5Server struct {
}

type Sockes5Client struct {
}

type ShadowsocksServer struct {
}

type ShadowsocksClient struct {
}

type RawServer struct {
}

type RawClient struct {
}

type Relay interface {
	Listen
	Connect
	UpwardRead
	UpwardWrite
	DownwardRead
	DownwardWrite
}

type SSLocal struct {
}

type SSServer struct {
}

func (*Socks5) Read(b []byte) (n int, err error) {
}

func (*Socks5) Write(b []byte) (n int. err error) {
}

func (*Socks5) Close() error {
}

func (*Socks5) LocalAddr() net.Addr {
}

func (*Socks5) RemoteAddr() net.Addr {
}

func (*Socks5) SetDeadline(t time.Time) error {
}

func (*Socks5) SetReadDeadline(t time.Time) error {
}

func (*Socks5) SetWriteDeadline(t time.Time) error {
}

func (*Socks5) Dial(network, address string) (net.Conn, error) {
}

func (*Socks5) DialTimeout(network, address string, timeout time.Duration) (Conn, error) {
}

