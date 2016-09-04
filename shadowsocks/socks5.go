package shadowsocks

import (
//	"io"
	"net"
	"time"
	"strconv"
	"errors"
	"fmt"
//	"encoding/hex"
)

func dummyFmt () {
	fmt.Println("123")
}

func dummyErr () error {
	return errors.New("This is a dummy error")
}

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

type Socks5 struct {
	S5LH
	S5UH
}

func (s Socks5) Relay(conn *ConnPair) error {
	if Debug {
		Log.Info("Relay started")
	}
	var mpl = s.S5LH.MaxPacketLength
	if mpl > s.S5UH.MaxPacketLength {
		mpl = s.S5UH.MaxPacketLength
	}
	var e1, e2 error
	var c chan bool
	// Upward
	go func () {
		if Debug {
			Log.Info("Upward started")
		}
		b := make([]byte, mpl)
		for {
			var n1 int
			n1, e1 = s.ReadLH(b, conn)
			if nil != e1{
				if Debug {
					Log.Info("Upward read " + e1.Error())
				}
				c <- true
				return
			}
			if Debug {
				Log.Info(fmt.Sprintf("Upward packet %v length %v", b, n1))
			}
			_, e1 = s.WriteUH(b[:n1], conn)
			if nil != e1 {
				if Debug {
					Log.Info("Upward write " + e1.Error())
				}
				c <- true
				return
			}
		}
	} ()
	// Downward
	func () {
		if Debug {
			Log.Info("Downward started")
		}
		b := make([]byte, mpl)
		var n1 int
		for {
			n1, e2 = s.ReadUH(b, conn)
			if nil != e2 {
				if Debug {
					Log.Info("Downward read " + e2.Error())
				}
				return
			}
			if Debug {
				Log.Info(fmt.Sprintf("Downward packet %v length %v", b, n1))
			}
			_, e2 = s.WriteLH(b[:n1], conn)
			if nil != e2 {
				if Debug {
					Log.Info("Downward write " + e2.Error())
				}
				return
			}
		}
	} ()
	<- c
	// Clear up
	conn.Down.Close()
	conn.Up.Close()
	if Debug {
		Log.Info("Relay finished")
	}
	if nil != e1 {
		return e1
	}
	if nil != e2 {
		return e2
	}
	return nil
}

// Suppose only one upper half and lower half now
// Socks5LowerHalf
type S5LH struct{
	S Socks
	Deadtime time.Duration
	IsRunning bool
	MaxPacketLength uint
	AuthRep [2]byte
}

func (s *S5LH) Init() {
	s.MaxPacketLength =256
	s.IsRunning = true
	s.Deadtime, _ = time.ParseDuration("20s")
	s.AuthRep = [2]byte{0x05, 0x00}
}

func (s S5LH) Authenticate(conn * ConnPair) error {
	if Debug {
		Log.Info("Authenticate started")
	}
	b := make([]byte, 2 + 255)
	_, err := conn.Down.Read(b[0:2])
	var l = int(b[1])
	conn.Down.Read(b[2:2 + l])
	err = ERR_SOCKS5_NO_AVAIL_AUTH
	for _, v := range b[2:2 + l] {
		if 0x00 == v {
			err = nil
			break
		}
	}
	_, err = conn.Down.Write(s.AuthRep[0:2])
	if Debug {
		Log.Info("AUthenticate finished")
		var errString string
		if nil != err {
			errString = err.Error()
		}
		Log.Info(fmt.Sprintf("Auth Replay length %v packet", len(s.AuthRep), s.AuthRep[0:2]))
		Log.Info("Authentication results " + errString)
	}
	return err
}

func (s S5LH) GetRequestType(conn *ConnPair) (cmd byte, addr net.IP, port uint16, err error) {
	if Debug {
		Log.Info("GetRequestType started")
	}
	b1 := make([]byte, 4)
	var n int
	n, err = conn.Down.Read(b1)
	cmd = b1[1]
	atype := b1[3]
	if Debug {
		Log.Info("GetRequestType info")
		Log.Info("b1 bytes read " + strconv.Itoa(n))
		Log.Info("cmd " + strconv.Itoa(int(cmd)))
		Log.Info("atype " + strconv.Itoa(int(atype)))
	}
	switch atype {
		case 0x01:
			addr = make([]byte, 4)
			_, err = conn.Down.Read(addr)
		case 0x03:
			b2 := make([]byte, 1)
			_, err = conn.Down.Read(b2)
			l := b2[0]
			if Debug {
				Log.Info("Domain name length " + strconv.Itoa(int(l)))
			}
			b3 := make([]byte, l)
			n, err = conn.Down.Read(b3)
			host := string(b3[:])
			if Debug {
				Log.Info("Domain name length read " + strconv.Itoa(n))
				Log.Info("Domain name " + host)
			}
			var ips []net.IP
			ips, err = net.LookupIP(host)
			addr = ips[0]
			if Debug {
				Log.Info("Ip address returned " + addr.String())
			}
		case 0x04:
			addr = make([]byte, 16)
			_, err = conn.Down.Read(addr)
		default:
			if Debug {
				Log.Info("GetRequest finished")
				Log.Info("GetRquest results ERROR default")
				}
			return
	}
	b4 := make([]byte, 2)
	_, err = conn.Down.Read(b4)
	port = (uint16(b4[0]) << 8) | uint16(b4[1])
	if Debug {
		Log.Info("Getrequest finished")
		var errString string
		if nil != err {
			errString = err.Error()
		}
		Log.Info("GetRequest results " + strconv.Itoa(int(cmd)) + " " + addr.String() + " " + strconv.Itoa(int(port)) + " " + errString)
	}
	return
}

func (s S5LH) ReadLH(b []byte, conn *ConnPair) (n int, err error) {
	conn.Down.SetReadDeadline(time.Now().Add(s.Deadtime))
	n, err = conn.Down.Read(b)
	return
}

func (s S5LH) WriteLH(b []byte, conn *ConnPair) (n int, err error) {
	conn.Down.SetWriteDeadline(time.Now().Add(s.Deadtime))
	n, err = conn.Down.Write(b)
	return
}

func (s S5LH) Deal(conn *ConnPair) (err error) {
	if Debug {
		Log.Info("Deal started")
	}
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
			if Debug {
				Log.Info("Deal connect")
				Log.Info("GetRequest results " + strconv.Itoa(int(cmd)) + " " + addr.String() + " " + strconv.Itoa(int(port)))
			}
			var bndAddr net.IP
			var bndPort uint16
			var atype byte
			bndAddr, bndPort, err = s.S.Connect(&addr, port, conn)
			var l2 = len(bndAddr)
			var l1 = l2
			var l = l2
			if nil != bndAddr.To4() {
				atype = 0x01
				l1 = l2 - 4
				l = 4
			} else {
				atype = 0x04
				l1 = l2 - 16
				l = 16
			}
			if nil != err {
				if Debug {
					Log.Info("Deal connect error " + err.Error())
				}
				return err
			}
			repPacket := make([]byte, l + 7)
			copy(repPacket[0:5], []byte{0x05, 0x00, 0x00, atype, byte(l)})
			copy(repPacket[5:5+l], bndAddr[l1:l2])
			copy(repPacket[5+l:], []byte{byte(bndPort >> 8), byte(bndPort & 0xff)})
			if Debug {
				Log.Info(fmt.Sprintf("Replay packet length %v packet %v", len(repPacket), repPacket))
			}
			_, err = conn.Down.Write(repPacket)
			err = s.S.Relay(conn)
		case SOCKS_BIND:
			if Debug {
				Log.Info("Deal bind")
			}
			var bindAddr *net.TCPAddr
			bindListener, _ := s.S.BindListen(bindAddr, conn)
			s.S.BindAccept(bindListener, conn)
			// reply
		case SOCKS_UDP_ASSOCIATE:
			if Debug {
				Log.Info("Deal udp associate")
			}
	}
	if Debug {
		Log.Info("Deal finished")
	}
	return
}

func (s S5LH) Listen(addr *net.TCPAddr) error {
	if Debug {
		Log.Info("Listen started")
		Log.Info(fmt.Sprintf("%v", s.IsRunning))
	}
	listener, _ := net.ListenTCP("tcp", addr)
	//listener.SetDeadline(time.Now().Add(s.Deadtime))
	for s.IsRunning {
		conn, _ := listener.AcceptTCP()
		if Debug {
			Log.Info("Accept connection")
		}
		go s.Deal(&ConnPair{Down:conn, Up:nil, UpChan:make(chan *[]byte, 20), DownChan:make(chan *[]byte, 20)})
	}
	return nil
}

func (s *S5LH) SetDeadline(t time.Duration) error {
	s.Deadtime = t
	return nil
}

type S5UH struct {
	S Socks
	Deadtime time.Duration
	IsRunning bool
	MaxPacketLength uint
}

func (s *S5UH) Init() {
	s.MaxPacketLength =256
	s.IsRunning = true
	s.Deadtime, _ = time.ParseDuration("20s")
}

func (s S5UH) Connect(addr *net.IP, port uint16, conn *ConnPair) (bndAddr net.IP, bndPort uint16, err error) {
	if Debug {
		Log.Info("Connect started")
	}
	dialer := net.Dialer{Timeout:s.Deadtime, DualStack:false}
	address := addr.String() + ":" + strconv.Itoa(int(port))
	if Debug {
		Log.Info("Connect address " + address)
	}
	network := "tcp"
	var newConn net.Conn
	newConn, err = dialer.Dial(network, address)
	if nil != err {
		if Debug {
			Log.Info("Connect error " + err.Error())
		}
		return
	}
	localAddr := newConn.LocalAddr().String()
	var tcpAddr *net.TCPAddr
	tcpAddr, err = net.ResolveTCPAddr(network, localAddr)
	bndAddr = tcpAddr.IP
	bndPort = uint16(tcpAddr.Port)
	conn.Up = newConn
	if Debug {
		Log.Info("Connect finished")
		Log.Info(fmt.Sprintf("Connect local results: addr %v port %v err %v", bndAddr, bndPort, err))
	}
	return
}

func (s S5UH) ReadUH(b []byte, conn *ConnPair) (n int, err error) {
	conn.Up.SetReadDeadline(time.Now().Add(s.Deadtime))
	n, err = conn.Up.Read(b)
	return
}

func (s S5UH) WriteUH(b []byte, conn *ConnPair) (n int, err error) {
	conn.Up.SetWriteDeadline(time.Now().Add(s.Deadtime))
	n, err = conn.Up.Write(b)
	return
}

func (S5UH) BindListen(lAddr *net.TCPAddr, conn *ConnPair) (*net.TCPListener, error) {
	if Debug {
		Log.Info("BindListen started")
	}
	listener, err := net.ListenTCP("tcp", lAddr)
	return listener, err
}

func (s S5UH) BindAccept(listener *net.TCPListener, conn *ConnPair) error {
	if Debug {
		Log.Info("BindAccept started")
	}
	err := listener.SetDeadline(time.Now().Add(s.Deadtime))
	newConn, err := listener.Accept()
	conn.Up = newConn
	return err
}

func (s S5UH) SetDeadline(t time.Duration) error {
	s.Deadtime = t
	return nil
}
