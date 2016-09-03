package shadowsocks

import (
	"io"
	"net"
	"time"
	"strconv"
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

// RelayHF:
// src -> c1
// src <- c2
// this function doesn't close connections
func RelayHF(src *net.Conn, c1 c2 *chan (*[]byte), maxLen int, res int) error {
	var errString string
	resStr := " " + strconv.Itoa(res)
	if Debug {
		Log.Info("Socks5 Relay Half started " + strconv.Itoa(res))
	}

	t1 := true
	t2 := true
	for t1 && t2 {
		if t1 {
		var b *[]byte
		b = <- c2
		if nil == b {
			if Debug {
				Log.Info("Socks5 Relay to dst finished " + resStr)
			}
			t1 = false
		}
		src.SetWriteDeadline(time.Now().Add(s.Deadtime))
		_, err = src.Write(*b)
		if nil != err {
			if Debug {
				Log.Info("Socks5 Relay write to dst error " + err.Error() + resStr)
			}
			return err
		}
		}
		if t2 {
		//----------------------------------------------------------------------
		b = make([]byte, maxLen)
		src.SetReadDeadline(time.Now().Add(s.Deadtime))
		_, err := src.Read(b)
		if err == io.EOF {
			c1 <- nil
			if Debug {
				Log.Info("Socks5 Relay read from src finished " + resStr)
			}
			t2 = false
		}
		if nil != err {
			if Debug {
				Log.Info("Socks5 Relay read from src error " + err.Error() + resStr)
			}
			return err
		}
		c1 <- b
		}
	}
}

func Relay(conn *SocksPair) error {
	var c chan bool
	var e1, e2 error
	go func () {
		e1 = RelayHF(conn.Down, conn.UpChan, conn.DownChan, 255, 1)
		c <- true
	}
	go func () {
		e2 = RelayHF(conn.Up, conn.DownChan, conn.UpChan, 255, 2)
		c <- true
	}
	<- c
	<- c
	conn.Down.Close()
	conn.Up.Close()
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
	Deadtime time.Duration
	IsRunning bool
	PacketMaxLength uint
	AuthRep [2]byte
}

func (s *S5LH) Init() {
	s.PacketMaxLength =256
	s.IsRunning = true
	s.Deadtime, _ = time.ParseDuration("1s")
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
	_, err = conn.Down.Write(s.AuthRep[:])
	if Debug {
		Log.Info("AUthenticate finished")
		var errString string
		if nil != err {
			errString = err.Error()
		}
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
			}
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
			err = Relay(conn)
		case SOCKS_BIND:
			if Debug {
				Log.Info("Deal bind")
			}
			var bindAddr *net.TCPAddr
			bindListener, _ := s.BindListen(bindAddr, conn)
			s.BindAccept(bindListener, conn)
			// reply
			go s.SocksHalf.Relay(conn)
			s.Relay(conn)
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
	}
	listener, _ := net.ListenTCP("tcp", addr)
	//listener.SetDeadline(time.Now().Add(s.Deadtime))
	for s.IsRunning {
		conn, _ := listener.AcceptTCP()
		if Debug {
			Log.Info("Accept connection")
		}
		go s.Deal(&ConnPair{Down:conn, Up:nil, UpChan:make(chan *Packet, 20), DownChan:make(chan *Packet, 20)})
	}
	return nil
}

type S5UH struct {
	// LH SocksHalf
	SocksHalf
	Deadtime time.Duration
	IsRunning bool
	MaxPacketLength int
}

func (s S5LH) SetDeadline(t time.Duration) error {
	s.Deadtime = t
	return nil
}

func (s S5UH) Init() {
	return
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
	}
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
