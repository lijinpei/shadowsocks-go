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

const (
	ERR_SOCKS5_NO_AVAIL_AUTH Error = Error("no available socks5 authentication method")
)

type Error string

func (me Error) Error() string {
	return string(me)
}

// Suppose only one upper half and lower half now

// Socks5LowerHalf
type S5LH struct{
	UH *SocksHalf
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
	_, err := (*(conn.Down)).Read(b) // the struct/(pointer to struct) rule cann't be applied recursively ???
	if nil != err {
		return err
	}
	err = ERR_SOCKS5_NO_AVAIL_AUTH
	for _, v := range b {
		if 0x00 == v {
			err = nil
			break
		}
	}
	if nil != err {
		return err
	}
	_, err = (*(conn.Down)).Write(s.AuthRep[:])
	return err
}

func (s S5LH) UpperHalf() *SocksHalf {
	return s.UH
}

func (s S5LH) GetRequestType(conn *ConnPair) (cmd byte, addr net.IP, port uint16, err error) {
	b1 := make([]byte, 4)
	n, err := (*(conn.Down)).Read(b1)
	if (4 != n) || (nil != err) {
		// TODO
		return
	}
	cmd = b1[1]
	atype := b1[3]
	switch atype {
		case 0x01:
			addr = make([]byte, 4)
			n, err = (*(conn.Down)).Read(addr)
			if (4 != n) || (nil != err) {
				// TODO
				return
			}
		case 0x03:
			b2 := make([]byte, 1)
			n, err = (*(conn.Down)).Read(b2)
			if (1 != n) || (nil != err) {
				// TODO
				return
			}
			l := b2[0]
			b3 := make([]byte, l)
			n, err = (*(conn.Down)).Read(b2)
			if (int(l) != n) || (nil != err) {
				// TODO
				return
			}
			host := string(b3[:])
			var ips []net.IP
			ips, err = net.LookupIP(host)
			if nil != err {
				// TODO
				return
			}
			addr = ips[0]
		case 0x04:
			addr = make([]byte, 16)
			n, err = (*(conn.Down)).Read(addr)
			if (16 != n) || (nil != err) {
				// TODO
				return
			}
		default:
			// TODO
			return
	}
	b4 := make([]byte, 2)
	n, err = (*(conn.Down)).Read(b4)
	if (2 != n) || (nil != err) {
		// TODO
		return
	}
	port = (uint16(b4[0]) << 8) | uint16(b4[1])
	return
}

func (s *S5LH) Deal(conn *ConnPair) error {
	var err error
	err = s.Authenticate(conn)
	if nil != err {
		return err
	}
	cmd, addr, port, err := s.GetRequestType(conn)
	if nil != err {
		return err
	}
	var rep byte
	UH := s.UpperHalf()
	switch cmd {
		case SOCKS5_CONNECT:
			bndAddr, bndPort, err = (*UH).Connect(&addr, port, conn)
			if nil != err {
				rep = 1
			}
			err = conn.Reply(rep)
			if nil != err {
				conn.Close()
				return repErr
			}
			go UHF.Relay(conn.Pair)
			err = s.Relay(conn)
		case SOCKS_BIND:
			bind, err = UHF.BindLisen(conn, addr, port)
			if nil != err {
				rep = 1
			}
			err = UHF.BindAccept(bind, conn)
			if nil != err {
				rep = 1
			}
			err = conn.Reply()
			if nil != err {
			}
			bindConn := s.Bind()
			go UHF.Relay(conn.Pair)
			err = s.Relay(conn)
		case SOCKS5_UDP:
/* TO BE IMPLEMENTED
			repErr := s.Reply(conn, rep)
			if nil != repErr {
				// close connection
				return repErr
			}
			err = conn.RelayUDP(udpConn)
*/
	}
	conn.Close()
	return err
}

func (s S5LH) Listen(addr *net.TCPAddr) error {
	listener := net.ListenTCP("tcp", addr)
	listener.SetDeadline(s.Deadtime)
	for s.IsRunning {
		conn, err := listener.AcceptTCP()
		go s.Deal(&ConnPair{Down:conn, Up:nil, Chan:make(Chan *Packet, 10)})
	}
	return nil
}

func (s S5LH) LowerrHalf() *SocksHalf {
	return nil
}

func (s S5LH) Connect(*net.TCPAddr, *ConnPair) error {
	return Error("Call connect on Socks5 Lower Half")
}

func (s S5LH) BindListen(*net.TCPAddr, *ConnPair) (*net.TCPListener, error) {
	return Error("Call BindListen on Socks5 Lower Half")
}

func (s S5LH) BindAccept(*net.TCPListener, *ConnPair) (*net.TCPConn, error) {
	return Error("Call BindAccept on Socks5 Lower Half")
}

func (s S5LH) Relay(conn *ConnPair) error {
	conn.Down.SetDeadline(s.Deadtime)
	for {
		b = make([]byte, s.PacketMaxLength)
		n, err := conn.Down.Read(b)
		if nil != err {
			return err
		}
		con.Chan <- &b
	}
}

type S5UH struct {
	LH *SocksHalf
	Deadtime time.Duration
	IsRunning bool
	MaxPacketLength int
}

func (S5UH) UpperHalf() *SocksHalf {
	return nil
}

func (S5UH) Deal(*net.TCPAddr) error {
	return err.New("Shouldn't call Deal on Socks5 Upper Half")
}

func (S5UH) Listen(*net.TCPAddr) error {
	return err.New("Shouldn't call Listen on Socks5 Upper Half")
}

func (S5UH) LowerHalf() *SocksHalf {
	return S5UH.LH
}

func (s S5UH) Connect(addr *net.IP, port uint16, conn *ConnPair) (bndAddr *net.IP, bndPort uint16, err error) {
	dialer := Dialer{Timeout:s.Deadtime, DualStack:false}
	address := addr.String() + ":" + string(port)
	network := "tcp"
	var newConn net.Conn
	newConn, err = dialer.Dial(network, address)
	if nil != err {
		return
	}
	localAddr := newConn.LocalAddr().String()
	var tcpAddr *TCPAddr
	tcpAddr, err = net.ResolveTCPAddr(network, localAddr)
	bndAddr = &tcpAddr.IP
	bndPort = &tcpAddr.Port
	conn.Up = newConn
	return
}

func (S5UH) BindListen(addr *net.TCPAddr, conn *ConnPair) (*net.TCPListener, error) {
	listener, err := net.ListenTCP("tcp", laddr *TCPAddr)
	return listener, err
}

func (s S5UH) BindAccept(conn *ConnPair, listener *net.TCPListener) (*net.TCPConn, error) {
	err := listener.SetDeadline(s.Deadtime)
	newConn, err := listener.Accept()
	conn.Up = newConn
	return newConn, err
}

func (s S5UH) Relay(conn *ConnPair) error {
	for {
		b := <- conn.Chann
		if nil == b {
			break
		}
		err := conn.Up.Write(*b)
		if nil != err {
			return err
		}
	}
	return nil
}

func (s S5UH) SetDeadline(t time.Duration) error {
	s.Deadline = t
	return nil
}
