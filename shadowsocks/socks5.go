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

// Suppose only one upper half and lower half now

// Socks5LowerHalf
type S5LH struct{
	UH *SocksHalf
	Deadtime time.Duration
	IsRunning bool
	PacketMaxLength uint
}

func (s S5LH) Init() {
	s.PacketMaxLength =2566
	s.IsRunning = true
	s.Deadtime, _ = time.ParseDuration("1s")
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

func (s S5LH) Listen(addr net.TCPAddr) error {
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

func (s S5LH) Connect(net.TCPAddr) error {
	return error.New("Call connect on Socks5 Lower Half")
}

func (s S5LH) BindListen(net.TCPAddr, *ConnPair) (*net.TCPListener, error) {
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

func (S5UH) UpperHald() *SocksHalf {
	return nil
}

func (S5UH) Deal(net.TCPAddr) error {
	return err.New("Shouldn't call Deal on Socks5 Upper Half")
}

func (S5UH) Listen(net.TCPAddr) error {
	return err.New("Shouldn't call Listen on Socks5 Upper Half")
}

func (S5UH) BindListen(conn *SocksConn, addr net.TCPAddr) (*net.TCPListener, error) {
	listener, err := net.ListenTCP("tcp", laddr *TCPAddr)
	return listener, err
}

func (s S5UH) BindAccept(conn *SocksConn, listener *net.TCPListener) (*net.TCPConn, error) {
	err := listener.SetDeadline(s.Deadtime)
	newConn, err := listener.Accept()
	return newConn, err
}

func (S5UH) LowerHalf() *SocksHalf {
	return S5UH.S5LH
}

func (s S5UH) Connect(addr *net.TCPAddr, conn *SocksConn) error {
	// TODO: add timeout constraints
	newConn, err := net.DialTCP("tcp", nil, addr)
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

func (s S5UH) SetDeadline(t time.Time) error {
	s.Deadline = t
	return nil
}
