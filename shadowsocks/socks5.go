package shadowsocks

import (
	"net"
	"error"
	"time"
)

const (
	SOCKS5_CONNECT      byte = 1
	SOCKS_BIND          byte = 2
	SOCKS_UDP_ASSOCIATE byte = 3
)

const (
	ERR_SOCKS5_NO_AVAIL_AUTH error = error.New("No available socks5 authenticate method")
)
// Suppose only one upper half and lower half now

// Socks5LowerHalf
type S5LH struct{
	UH *SocksHalf
	Deadtime time.Duration
	IsRunning bool
	PacketMaxLength uint
	AuthRep []byte
}

func (s S5LH) Init() {
	s.PacketMaxLength =2566
	s.IsRunning = true
	s.Deadtime, _ = time.ParseDuration("1s")
	s.AuthRep = [...]byte{0x05, 0x00}
}

func (s S5LH) Authenticate(conn * SocksConn) error {
	b := make([]byte, 2 + 255)
	_, err := conn.Down.read(b)
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
	conn.Down.write(s.AuthRep)
}

func (s S5LH) UpperHalf() *SocksHalf {
	return s.UH
}

func (s *S5LH) Deal(conn *SocksConn) error {
	var err error
	err := s.Authenticate(conn)
	if nil != err {
		return err
	}
	cmd, addr, port, err := s.GetRequestType(conn)
	if nil != err {
		return err
	}
	var rep byte
	UHF := s.UpperHalf()
	switch cmd {
		case SOCKS5_CONNECT:
			conn.Pair, err = UHF.Connect(addr, port)
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
		go s.Deal(&SocksConn{Down:conn, Up:nil, Chan:nil})
	}
	return nil
}

func (s S5LH) LowerrHalf() *SocksHalf {
	return nil
}

func (s S5LH) Connect(*net.TCPAddr, *ConnPair) error {
	return error.New("Call connect on Socks5 Lower Half")
}

func (s S5LH) BindListen(*net.TCPAddr, *ConnPair) (*net.TCPListener, error) {
	return error.New("Call BindListen on Socks5 Lower Half")
}

func (s S5LH) BindAccept(*net.TCPListener, *ConnPair) (*net.TCPConn, error) {
	return error.New("Call BindAccept on Socks5 Lower Half")
}

func (s S5LH) Relay(conn *ConnPair) error {
	conn.Down.SetDeadline(s.Deadtime)
	for s.IsRunning {
		b = make([]byte, s.PacketMaxLength)
		n, err := conn.Down.Read(b)
		if nil != err {
			return err
		}
		con.Chan <- &b
	}
}

type S5UH struct {
	LH *ScoksHalf
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

func (s S5UH) Connect(addr *net.TCPAddr, conn *ConnPair) error {
	// TODO: add timeout constraints
	newConn, err := net.DialTCP("tcp", nil, addr)
	conn.Up = newConn
	return err
}

func (S5UH) BindListen(addr *net.TCPAddr, conn *ConnPair) (*net.TCPListener, error) {
	listener, err := net.ListenTCP("tcp", laddr *TCPAddr)
	return listener, err
}

func (s S5UH) BindAccept(conn *SocksConn, listener *net.TCPListener) (*net.TCPConn, error) {
	err := listener.SetDeadline(s.Deadtime)
	newConn, err := listener.Accept()
	return newConn, err
}

func (s S5UH) Relay(conn *SocksConn) error {
	for {
		b := <- conn.Chann
		if nil == b {
			break
		}
		err := conn.Write(b)
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
