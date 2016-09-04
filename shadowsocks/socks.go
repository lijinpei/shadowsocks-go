package shadowsocks

import (
	"net"
)

// ConnPair represents a pair of connections to be relayed
type ConnPair struct {
	Up net.Conn
	Down net.Conn
	UpChan chan *[]byte
	DownChan chan *[]byte
}

// SocksHalf represents half of relay server
// SocksHalf should be bound to nic
type Socks interface {
	// Methods specific to lower half
	// upper half should call corresponding lower half's methods
	Listen(*net.TCPAddr) error
	// Methods specifig to upper half
	// Lower half should call corresponding upper half's methods
	Connect(*net.IP, uint16, *ConnPair) (net.IP, uint16, error)
	// addr: address to listen for bind as expected by client
	BindListen(addr *net.TCPAddr, conn *ConnPair) (*net.TCPListener, error)
	BindAccept(*net.TCPListener, *ConnPair) (error)
	// Methods shared by both half
	Relay(*ConnPair) error
//	UDPRelay(, *UDPConn) error
}

// Make Sure those two inferface have no methods with the same name
type SocksLH interface {
	Listen(*net.TCPAddr) error
	ReadLH([]byte, *ConnPair) (int, error)
	WriteLH([]byte, conn *ConnPair) (int, error)
}

type SocksUH interface {
	Connect(*net.IP, uint16, *ConnPair) (net.IP, uint16, error)
	BindListen(addr *net.TCPAddr, conn *ConnPair) (*net.TCPListener, error)
	BindAccept(*net.TCPListener, *ConnPair) (error)
	ReadUH(conn *ConnPair) (int, error)
	WriteUH([]byte, conn *ConnPair) (int, error)
}
