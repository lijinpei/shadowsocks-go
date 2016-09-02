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

// Socks5LowerHalf
type S5LH struct{
	UH *SocksHalf
	Deadtime time.Time
	IsRunning bool
	PacketMaxLength int
}

func (s S5LH) Init {
	s.PacketMaxLength = 
}
func (s S5LH) LowerrHalf() *SocksHalf {
	return nil
}

func (s S5LH) Connect(net.TCPAddr) error {
	return error.New("Call connect on Socks5 Lower Half")
}

func (s S5LH) BindRequest(net.TCPAddr) error {
	return error.New("Call BindRequest on Socks5 Lower Half")
}

func (s S5LH) Listen(addr net.TCPAddr) error {
	listener := net.ListenTCP(addr)
	listener.SetDeadline(s.Deadtime)
	for s.IsRunning {
		conn, err := listener.AcceptTCP()
		go s.Deal(&SocksConn{TCPConn:conn, Pair:nil, Chan:nil})
	}
	return nil
}

func (s S5LH) Relay(conn *SocksConn) error {
	conn.SetDeadline(s.Deadtime)
	for s.IsRunning {
		b = make([]byte, s.PacketMaxLength)
		n, err := conn.Read(b)
		if nil != err {
			return err
		}
		(*con.Chan) <- &b[:n]
	}
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
				rep = 
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
				rep = 
			}
			err = UHF.BindAccept(bind, conn)
			if nil != err {
				rep = 
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

type S5UH struct {
	LH *ScoksHalf
	Deadtime time.Time
	IsRunning bool
	MaxPacketLength int
}

func (S5UH) UpperHald() *SocksHalf {
	return nil
}

func (S5UH) Deal(net.TCPAddr) error {
	return err.New("Shouldn't call Deal on Socks5 Upper Half")
{

func (S5UH) Listen(net.TCPAddr) error {
	return err.New("Shouldn't call Listen on Socks5 Upper Half")
}

func (S5UH) BindListen(*SocksConn, addr net.TCPAddr) (*net.TCPListener error) {
	listener, err := net.ListenTCP("tcp", laddr *TCPAddr)
	return listener, err
}

func (s S5UH) BindAccept(*SocksConn, listener *net.TCPListener) (*net.TCPConn error) {
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
	for b :=<- conn.Chan {
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
