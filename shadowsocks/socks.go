package shadowsocks

import (
	"net"
	"time"
)

// Packet represents a TCP/UDP packet
type Packet []byte

// ConnPair represents a pair of connections to be relayed
type ConnPair struct {
	Up *net.Conn
	Down *net.Conn
	Chan *(chan Packet)
}

// SocksHalf represents half of relay server
// SocksHalf should be bound to nic
type SocksHalf interface {
	// Methods specific to lower half
	// Upper half should implement dummy functions
	UpperHalf() *SocksHalf
	Deal(*ConnPair) error
	Listen(*net.TCPAddr) error
	// Methods specifig to upper half
	// Lower half should implement dummy functions
	LowerHalf() *SocksHalf
	Connect(*net.IP, uint16, *ConnPair) (*net.IP, uint16, error)
	BindListen(*net.TCPAddr, *ConnPair) (*net.TCPListener, error)
	BindAccept(*net.TCPListener, *ConnPair) (*net.TCPConn, error)
	// Methods shared by both half
	Relay(*ConnPair) error
//	UDPRelay(, *UDPConn) error
	SetDeadline(time.Duration) error
}
