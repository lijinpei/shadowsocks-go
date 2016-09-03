package shadowsocks

import (
	"net"
	"time"
)

const (
	SOCKS5_CONNECT      byte = 1
	SOCKS_BIND          byte = 2
	SOCKS_UDP_ASSOCIATE byte = 3
)

type Error string

func (err Error) Error() string {
	return string(err)
}

const (
	ERR_SOCKS5_NO_AVAIL_AUTH Error = Error("no available socks5 authentication method")
)

// Suppose only one upper half and lower half now

// Socks5LowerHalf
type S5LH struct{
	SocksHalf
	Deadtime time.Duration
	IsRunning bool
	PacketMaxLength uint
	AuthRep [2]byte
}

func (s S5LH) Init() {
	s.PacketMaxLength =256
	s.IsRunning = true
	s.Deadtime, _ = time.ParseDuration("1s")
	s.AuthRep = [2]byte{0x05, 0x00}
}

func (s S5LH) Authenticate(conn * ConnPair) error {
	b := make([]byte, 2 + 255)
	_, err := conn.Down.Read(b) // the struct/(pointer to struct) rule cann't be applied recursively ???
	err = ERR_SOCKS5_NO_AVAIL_AUTH
	for _, v := range b {
		if 0x00 == v {
			err = nil
			break
		}
	}
	_, err = conn.Down.Write(s.AuthRep[:])
	return err
}

func (s S5LH) GetRequestType(conn *ConnPair) (cmd byte, addr net.IP, port uint16, err error) {
	b1 := make([]byte, 4)
	_, err = conn.Down.Read(b1)
	cmd = b1[1]
	atype := b1[3]
	switch atype {
		case 0x01:
			addr = make([]byte, 4)
			_, err = conn.Down.Read(addr)
		case 0x03:
			b2 := make([]byte, 1)
			_, err = conn.Down.Read(b2)
			l := b2[0]
			b3 := make([]byte, l)
			_, err = conn.Down.Read(b2)
			host := string(b3[:])
			var ips []net.IP
			ips, err = net.LookupIP(host)
			addr = ips[0]
		case 0x04:
			addr = make([]byte, 16)
			_, err = conn.Down.Read(addr)
		default:
			return
	}
	b4 := make([]byte, 2)
	_, err = conn.Down.Read(b4)
	port = (uint16(b4[0]) << 8) | uint16(b4[1])
	return
}

func (s S5LH) Deal(conn *ConnPair) (err error) {
	err = s.Authenticate(conn)
	if nil != err {
		return
	}
	cmd, addr, port, err := s.GetRequestType(conn)
	if nil != err {
		return
	}
	switch cmd {
		case SOCKS5_CONNECT:
			bndAddr, bndPort, err := s.Connect(&addr, port, conn)
			if nil != err {
				return err
			}
			addrLen := len(bndAddr)
			repPacket := make([]byte, addrLen + 6)
			copy(repPacket[0:4], []byte{0x05, 0x00, 0x00, byte(addrLen)})
			copy(repPacket[4:4+addrLen], bndAddr)
			copy(repPacket[4+addrLen:], []byte{byte(bndPort >> 8), byte(bndPort & 0xff)})
			_, err = conn.Down.Write(repPacket)
			go s.SocksHalf.Relay(conn)
			err = s.Relay(conn)
		case SOCKS_BIND:
			var bindAddr *net.TCPAddr
			bindListener, _ := s.BindListen(bindAddr, conn)
			s.BindAccept(bindListener, conn)
			// reply
			go s.SocksHalf.Relay(conn)
			s.Relay(conn)
		case SOCKS_UDP_ASSOCIATE:
	}
	return
}

func (s S5LH) Listen(addr *net.TCPAddr) error {
	listener, _ := net.ListenTCP("tcp", addr)
	listener.SetDeadline(time.Now().Add(s.Deadtime))
	for s.IsRunning {
		conn, _ := listener.AcceptTCP()
		go s.Deal(&ConnPair{Down:conn, Up:nil, Chan:make(chan *Packet, 20)})
	}
	return nil
}

/*
func (s S5LH) Connect(addr *net.TCPAddr, conn *ConnPair) (*net.IP, uint16, error) {
	return s.UH.Connect(*net.TCPAddr, *ConnPair)
}

func (s S5LH) BindListen(addr *net.TCPAddr, conn *ConnPair) (*net.TCPListener, error) {
	return s.UH.BindListen(addr, conn)
}

func (s S5LH) BindAccept(listener *net.TCPListener, conn *ConnPair) (*net.TCPConn, error) {
	return s.UH.BindAccept(listener, conn)
}
*/

func (s S5LH) Relay(conn *ConnPair) error {
	conn.Down.SetDeadline(time.Now().Add(s.Deadtime))
	for {
		b := make([]byte, s.PacketMaxLength)
		_, err := conn.Down.Read(b)
		if nil != err {
			return err
		}
		p := Packet(b)
		conn.Chan <- &p
	}
}

type S5UH struct {
	// LH SocksHalf
	SocksHalf
	Deadtime time.Duration
	IsRunning bool
	MaxPacketLength int
}
/*
func (s S5UH) Deal(addr *net.TCPAddr) error {
	return s.LH.Deal(addr)
}

func (s S5UH) Listen(addr *net.TCPAddr) error {
	return s.LH.Listen(addr)
}
*/
func (s S5LH) SetDeadline(t time.Duration) error {
	s.Deadtime = t
	return nil
}

func (s S5UH) Init() {
	return
}

func (s S5UH) Connect(addr *net.IP, port uint16, conn *ConnPair) (bndAddr net.IP, bndPort uint16, err error) {
	dialer := net.Dialer{Timeout:s.Deadtime, DualStack:false}
	address := addr.String() + ":" + string(port)
	network := "tcp"
	var newConn net.Conn
	newConn, err = dialer.Dial(network, address)
	if nil != err {
		return
	}
	localAddr := newConn.LocalAddr().String()
	var tcpAddr *net.TCPAddr
	tcpAddr, err = net.ResolveTCPAddr(network, localAddr)
	bndAddr = tcpAddr.IP
	bndPort = uint16(tcpAddr.Port)
	conn.Up = newConn
	return
}

func (S5UH) BindListen(lAddr *net.TCPAddr, conn *ConnPair) (*net.TCPListener, error) {
	listener, err := net.ListenTCP("tcp", lAddr)
	return listener, err
}

func (s S5UH) BindAccept(listener *net.TCPListener, conn *ConnPair) error {
	err := listener.SetDeadline(time.Now().Add(s.Deadtime))
	newConn, err := listener.Accept()
	conn.Up = newConn
	return err
}

func (s S5UH) Relay(conn *ConnPair) error {
	for {
		b := <- conn.Chan
		if nil == b {
			break
		}
		_, err := conn.Up.Write(*b)
		if nil != err {
			return err
		}
	}
	return nil
}

func (s S5UH) SetDeadline(t time.Duration) error {
	s.Deadtime = t
	return nil
}
