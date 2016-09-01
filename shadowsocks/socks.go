package shadowsocks

import (
	"net"
	"time"
)

type Packet []byte

type SocksConn struct {
	net.TCPConn
	Pair *net.TCPConn
	Chan *(chan packet)
}

type SocksHalf interface {
	// Methods specific to lower half
	// Upper half should implement dummy functions
	UpperHalf() *SockerHalf
	Listen(net.TCPAddr) error
	BindListen(*SocksConn, net.TCPAddr) (*net.TCPListener error)
	BindAccept(*SocksConn, *net.TCPListener) error
	// Methods specifig to upper half
	// Lower half should implement dummy functions
	LowerHalf() *SocksHalf
	Connect(net.TCPAddr) error
	BindRequest(net.TCPAddr) error
	// Methods shared by both half
	Deal(net.TCPAddr) error
	Relay(*SocksConn) error
	UDPRelay(*SocksConn, *UDPConn) error
	SetDeadline(time.Time) error
}

type SocksRelay interface {
	Run(net.TCPAddr) error
	Stop(net.TCPAddr) error
}

/*
type SocksUpperHalf interface {
	LowerHalf() (SocksLowerHalf, error)
	Connect(net.TCPAddr) error
	Deal(net.TCPAddr) error
	Bind(net.TCPAddr) error
	Relay(*SocksConn) error
	UDPRelay(*SocksConn, *UDPConn) error
	SetDeadline(time.Time) error
}

type SocksLowerHalf interface {
	UpperHalf() (SockerLowerHalf, error)
	Listen(net.TCPAddr) error
	Deal(net.TCPAddr) error
	BindRequest(net.TCPAddr) error
	Relay(*SocksConn) error
	UDPRelay(*SocksConn, *UDPConn) error
	SetDeadline(time.Time) error
}
*/
