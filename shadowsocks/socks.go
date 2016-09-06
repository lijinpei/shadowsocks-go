package shadowsocks

import (
	"net"
)

// ConnPair represents a pair of connections to be relayed
type ConnPair struct {
	// Up: as upward connection in connect mode, and bind connection in bind mode
	Up *net.TCPConn
	UDPUp *net.UDPConn
	Down *net.TCPConn
	UDPDown *net.UDPConn
	//  UpChan, DownChan: UDP Realy sent four slices per packet
	// atype, addr, port, data
	// addr/port are in network endian order
	UpChan chan []byte
	DownChan chan []byte
	UDPRunning bool
	ClientUDPAddr *net.UDPAddr
	UDPAllowedIP *net.IP
	UDPAllowedPort *uint16
}

// SocksHalf represents half of relay server
// SocksHalf should be bound to nic
type Socks interface {
	// Methods specific to lower half
	// upper half should call corresponding lower half's methods
	Listen(*net.TCPAddr) error
	// Methods specifig to upper half
	// Lower half should call corresponding upper half's methods
	Connect(byte, net.IP, uint16, *ConnPair) (net.IP, uint16, error)
	// addr: address to listen for bind as expected by client
	BindListen(addr *net.TCPAddr, conn *ConnPair) (*net.TCPListener, error)
	BindAccept(*net.TCPListener, *ConnPair) (error)
	// Methods shared by both half
	Relay(*ConnPair) error
	UDPRelay(*ConnPair) error
}

// Make Sure those two inferface have no methods with the same name
type SocksLH interface {
	Listen(*net.TCPAddr) error
	ReadLH([]byte, *ConnPair) error
	WriteLH(*ConnPair) error
	UDPReadLH(*ConnPair, net.IP, uint16) error
	UDPWriteLH(*ConnPair) error
}

type SocksUH interface {
	Connect(*net.IP, uint16, *ConnPair) (net.IP, uint16, error)
	BindListen(addr *net.TCPAddr, conn *ConnPair) (*net.TCPListener, error)
	BindAccept(*net.TCPListener, *ConnPair) (error)
	ReadUH(*ConnPair)  error
	WriteUH(*ConnPair) error
	UDPReadUH(*ConnPair) error
	UDPWriteUH(*ConnPair) error
	UDPListen(ipv6 bool) (*net.UDPConn, error)
}
