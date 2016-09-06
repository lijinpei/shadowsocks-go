package shadowsocks

import (
	"io"
	"net"
	"time"
	"strconv"
	"fmt"
//	"encoding/hex"
)

func dummyFmt () {
	fmt.Println("123")
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
	ERR_NO_AVAIL_AUTH Error = Error("no available socks5 authentication method")
)

func lookupDNS(host []byte) (atype byte, addr []byte, err error) {
		var ips []net.IP
		ips, err = net.LookupIP(string(addr))
		if nil != err {
			return
		}
		addr = ips[0].To4()
		if nil == addr {
			addr = ips[0]
			atype = 0x04
		} else {
			atype = 0x01
		}
		if DbgCtlFlow {
			Log.Debug(fmt.Sprintf("Deal DNS lookup results: atype %v addr %v", atype, addr))
		}
}

type Socks5 struct {
	S5LH
	S5UH
}

func (s Socks5) Relay(conn *ConnPair) error {
	if DbgCtlFlow {
		Log.Debug("Relay started")
	}
	var c chan bool
	go func () {
		s.S5UH.ReadUH(conn)
		c <- true
		if DbgCtlFlow {
			Log.Debug("S5UH ReadUH finished")
		}
	} ()
	go func () {
		s.S5UH.WriteUH(conn)
		c <- true
		if DbgCtlFlow {
			Log.Debug("S5UH WriteUH finished")
		}
	} ()
	go func () {
		s.S5LH.ReadLH(conn)
		c <- true
		if DbgCtlFlow {
			Log.Debug("S5LH ReadLH finished")
		}
	} ()
	go func () {
		s.S5LH.WriteLH(conn)
		c <- true
		if DbgCtlFlow {
			Log.Debug("S5LH WriteLH finished")
		}
	} ()
	<-c;<-c;<-c;<-c;
	conn.Up.Close()
	conn.Down.Close()
	if DbgCtlFlow {
		Log.Debug("Relay finished")
	}
	return nil
}

// Suppose only one upper half and lower half now
// Socks5LowerHalf
type S5LH struct {
	S Socks
	Deadtime time.Duration
	MaxPacketLength uint
	AuthRep [2]byte
}

func (s *S5LH) Init() {
	s.MaxPacketLength =256
	s.Deadtime, _ = time.ParseDuration("200s")
	s.AuthRep = [2]byte{0x05, 0x00}
}

func (s S5LH) Authenticate(conn * ConnPair) error {
	if DbgCtlFlow {
		Log.Debug("Authenticate started")
	}
	b := make([]byte, 2 + 255)
	_, err := conn.Down.Read(b[0:2])
	var l = int(b[1])
	conn.Down.Read(b[2:2 + l])
	err = ERR_NO_AVAIL_AUTH
	for _, v := range b[2:2 + l] {
		if 0x00 == v {
			err = nil
			break
		}
	}
	_, err = conn.Down.Write(s.AuthRep[0:2])
	if DbgCtlFlow {
		Log.Debug("AUthenticate finished")
		var errString string
		if nil != err {
			errString = err.Error()
		}
		Log.Debug(fmt.Sprintf("Auth Replay length %v packet", len(s.AuthRep), s.AuthRep[0:2]))
		Log.Debug("Authentication results " + errString)
	}
	return err
}

func (s S5LH) GetRequestType(conn *ConnPair) (cmd, atype byte, addr net.IP, port uint16, err error) {
	if DbgCtlFlow {
		Log.Debug("GetRequestType started")
	}
	b1 := make([]byte, 4)
	var n int
	n, err = conn.Down.Read(b1)
	cmd = b1[1]
	atype = b1[3]
	if DbgCtlFlow {
		Log.Debug("GetRequestType info")
		Log.Debug("b1 bytes read " + strconv.Itoa(n))
		Log.Debug("cmd " + strconv.Itoa(int(cmd)))
		Log.Debug("atype " + strconv.Itoa(int(atype)))
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
				Log.Debug("Domain name length " + strconv.Itoa(int(l)))
			}
			addr = make([]byte, l)
			n, err = conn.Down.Read(addr)
			if Debug {
				Log.Debug("Domain name length read " + strconv.Itoa(int(n)))
			}
		case 0x04:
			addr = make([]byte, 16)
			_, err = conn.Down.Read(addr)
		default:
			if DbgCtlFlow {
				Log.Debug("GetRequest finished")
				Log.Debug("GetRequest wrong atype")
				}
			err = Error("GetRequest wrong atype")
			return
	}
	b4 := make([]byte, 2)
	_, err = conn.Down.Read(b4)
	port = (uint16(b4[0]) << 8) | uint16(b4[1])
	if DbgCtlFlow {
		Log.Debug("Getrequest finished")
		var errStr string
		if nil != err {
			errStr = err.Error()
		}
		var addrStr string
		if atype != 0x03 {
			addrStr = addr.String()
		} else {
			addrStr = string(addr)
		}
		Log.Debug(fmt.Sprintf("GetRequest results cmd %v atype %v addr %v port %v err %v", cmd, atype, addrStr, port, errStr))
	}
	return
}

func connToChan(conn *net.TCPConn, c chan []byte, t time.Duration, mpl uint) (err error) {
	if Debug {
		Log.Debug(fmt.Sprintf("connToChan timeout %v", t))
	}
	for {
		var b []byte = make([]byte, mpl)
		conn.SetReadDeadline(time.Now().Add(t))
		var n int
		n, err = conn.Read(b)
		if io.EOF == err {
			close(c)
			return
		}
		if DbgLogPacket {
			Log.Debug(fmt.Sprintf("connToChan packet %v", b))
		}
		if nil != err {
			return err
		}
		c <- b[:n]
	}
}

func chanToConn(conn *net.TCPConn, c chan []byte, t time.Duration) (err error) {
	if Debug {
		Log.Debug(fmt.Sprintf("chanToConn timeout %v", t))
	}
	for {
		var b []byte
		b = <- c
		if nil == b {
			return nil
		}
		conn.SetWriteDeadline(time.Now().Add(t))
		_, err = conn.Write(b)
		if DbgLogPacket {
			Log.Debug(fmt.Sprintf("chanToConn packet %v", b))
		}
		if nil != err {
			return err
		}
	}
}

func (s S5LH) ReadLH(conn *ConnPair) (err error) {
	if DbgCtlFlow {
		Log.Debug("S5LH ReadLH started")
	}
	err = connToChan(conn.Down, conn.UpChan, s.Deadtime, s.MaxPacketLength)
	if Debug {
		Log.Debug(err.Error())
	}
	return err
}

func (s S5LH) WriteLH(conn *ConnPair) (err error) {
	if DbgCtlFlow {
		Log.Debug("S5LH WriteLH started")
	}
	err = chanToConn(conn.Down, conn.DownChan, s.Deadtime)
	if Debug {
		var errStr string
		if nil != err {
			errStr = err.Error()
		}
		Log.Debug(errStr)
	}
	return err
}

func (s S5LH) Deal(conn *ConnPair) (err error) {
	if DbgCtlFlow {
		Log.Debug("Deal started")
	}
	err = s.Authenticate(conn)
	if nil != err {
		return
	}
	cmd, atype, addr, port, err := s.GetRequestType(conn)
	if DbgCtlFlow {
		var addrStr string
		if atype != 0x03 {
			addrStr = addr.String()
		} else {
			addrStr = string(addr)
		}
		Log.Debug(fmt.Sprintf("GetRequestType results: cmd %v atype %v addr %v port %v", cmd, atype, addrStr, port))
	}
	if nil != err {
		if DbgCtlFlow {
			Log.Debug("Deal returned")
		}
		return
	}
    if (0x03 == atype) {
        atype, addr, err = lookupDNS(addr)
    }
	if (0x01 != atype) && (0x04 != atype) {
		if DbgCtlFlow {
			Log.Debug("Deal returned")
		}
		return Error("Socks5 Request Wrong atype")
	}
	switch cmd {
		case SOCKS5_CONNECT:
			if DbgCtlFlow {
				Log.Debug("Deal connect")
			}
			var bndAddr net.IP
			var bndPort uint16
			bndAddr, bndPort, err = s.S.Connect(atype, addr, port, conn)
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
				if DbgCtlFlow {
					Log.Debug("Deal connect error " + err.Error())
				}
				return err
			}
			repPacket := make([]byte, l + 6)
			copy(repPacket[0:4], []byte{0x05, 0x00, 0x00, atype})
			copy(repPacket[4:4+l], bndAddr[l1:l2])
			copy(repPacket[4+l:], []byte{byte(bndPort >> 8), byte(bndPort & 0xff)})
			if DbgLogPacket {
				Log.Debug(fmt.Sprintf("Replay packet length %v packet %v", len(repPacket), repPacket))
			}
			_, err = conn.Down.Write(repPacket)
			err = s.S.Relay(conn)
		case SOCKS_BIND:
			if DbgCtlFlow {
				Log.Debug("Deal bind")
			}
			var bindAddr *net.TCPAddr
			bindListener, _ := s.S.BindListen(bindAddr, conn)
			s.S.BindAccept(bindListener, conn)
			// reply
		case SOCKS_UDP_ASSOCIATE:
			if DbgCtlFlow || DbgUDP {
				Log.Debug("Deal udp associate")
			}
	}
	if DbgCtlFlow {
		Log.Debug("Deal finished")
	}
	return
}

func (s S5LH) Listen(addr *net.TCPAddr) error {
	if DbgCtlFlow {
		Log.Debug("Listen started")
	}
	listener, _ := net.ListenTCP("tcp", addr)
	//listener.SetDeadline(time.Now().Add(s.Deadtime))
	for {
		conn, _ := listener.AcceptTCP()
		if DbgCtlFlow {
			Log.Debug("Accept connection")
		}
		go s.Deal(&ConnPair{Down:conn, Up:nil, UpChan:make(chan []byte, 20), DownChan:make(chan []byte, 20)})
	}
	return nil
}

func (s S5LH) UDPReadLH(conn *ConnPair, srcAddr net.IP, srcPort uint16) (err error) {
    for conn.UDPRunning {
        var n int
        var udpAddr *net.UDPAddr
        var b = make([]byte, s.MaxPacketLength)
        conn.Down.SetReadDeadline(time.Now.Add(s.Deadtime))
        n, b, udpAddr = conn.Down.ReadFromUDP(b)
        if !udpAddr.IP.Equal(srcAddr) || (udpAddr.Port != int(SrcPort)) {
            continue
        }
        if n < 7 {
            continue
        }
        if 0x00 != b[2] {
            continue
        }
		connn.UpChan <- b[2:]
}

func (s S5LH) UDPWriteLH(conn *ConnPair) error {
	for conn.UDPRuning {
		var b []byte
		b = <- conn.DownChan
		go func () {
			conn.Down.SetWriteDeadline(time.Now().Add(s.Deadtime))
			conn.Down.WriteToUDP(b, conn.ClientUDPAddr)
		}
	}
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
	s.Deadtime, _ = time.ParseDuration("200s")
}

func (s S5UH) Connect(atype byte, addr net.IP, port uint16, conn *ConnPair) (bndAddr net.IP, bndPort uint16, err error) {
	if DbgCtlFlow {
		Log.Debug("Connect started")
	}
	// TODO: ipv4-mapped ipv6 address
	var addrStr string
	if 0x01 == atype {
		addrStr = net.IP(addr).String() + ":" + strconv.Itoa(int(port))
	} else {
		addrStr = "[" + net.IP(addr).String() + "]:" + strconv.Itoa(int(port))
	}
	if DbgCtlFlow {
		Log.Debug("Connect address " + addrStr)
	}
	network := "tcp"
	var newConn net.Conn
	newConn, err = net.DialTimeout(network, addrStr, s.Deadtime)
	if nil != err {
		if DbgCtlFlow {
			Log.Debug("Connect error " + err.Error())
		}
		return
	}
	localAddrStr := newConn.LocalAddr().String()
	var tcpAddr *net.TCPAddr
	tcpAddr, err = net.ResolveTCPAddr(network, localAddrStr)
	bndAddr = tcpAddr.IP
	bndPort = uint16(tcpAddr.Port)
	conn.Up = newConn.(*net.TCPConn)
	if DbgCtlFlow {
		Log.Debug("Connect finished")
		Log.Debug(fmt.Sprintf("Connect local results: addr %v port %v err %v", bndAddr, bndPort, err))
	}
	return
}

func (s S5UH) ReadUH(conn *ConnPair) (err error) {
	if DbgCtlFlow {
		Log.Debug("S5UH ReadUH started")
	}
	err = connToChan(conn.Up, conn.DownChan, s.Deadtime, s.MaxPacketLength)
	if Debug {
		Log.Debug(err.Error())
	}
	return err
}

func (s S5UH) WriteUH(conn *ConnPair) (err error) {
	if DbgCtlFlow {
		Log.Debug("S5UH WriteUH started")
	}
	err = chanToConn(conn.Up, conn.UpChan, s.Deadtime)
	if Debug {
		var errStr string
		if nil != err {
			errStr = err.Error()
		}
		Log.Debug(errStr)
	}
	return err
}

func (S5UH) BindListen(lAddr *net.TCPAddr, conn *ConnPair) (*net.TCPListener, error) {
	if DbgCtlFlow {
		Log.Debug("BindListen started")
	}
	listener, err := net.ListenTCP("tcp", lAddr)
	return listener, err
}

func (s S5UH) BindAccept(listener *net.TCPListener, conn *ConnPair) error {
	if DbgCtlFlow {
		Log.Debug("BindAccept started")
	}
	err := listener.SetDeadline(time.Now().Add(s.Deadtime))
	newConn, err := listener.Accept()
	conn.Up = newConn.(*net.TCPConn)
	return err
}

func (s S5UH) UDPReadUH(conn *ConnPair) (err error) {
	for conn.UDPRunning {
		var addr *net.UDPAddr
		var n int
		var b []byte = make([]byte, s.MaxPacketLength + 22)
		conn.UDPUp.SetReadDeadline(time.Now().Add(conn.Deadtime))
		n, addr, err = conn.UDPUp.ReadUDP(b[22:])
		if nil != err {
			return
		}
		addr4 := addr.IP.To4()
		b[20] = byte(addr.Port >> 8)
		b[21] = byte(addr.Port & 0xff)
		if nil != addr4 {
			copy(b[16:20], addr4)
			b[15] = 0x01
			b = b[15:22+n]
		} else {
			copy(b[4:20], addr)
			b[3] = 0x04
			b = b[3:22+n]
		}
		conn.DownChan <- b
	}
}

func (s S5UH) UDPWritingUH(conn *ConnPair) (err error) {
	for conn.UDPRuning {
		b := <-conn.UpChan
		atype := b[0]
		//TODO: what is ipv6 zone?
		addr := &net.UDPAddr{}
		switch atype: {
		case 0x01:
			addr.IP = b[1:5]
			addr.Port = (int(b[5]) << 8) | int(b[6])
			go func() {
				conn.UDPUp.SetWriteDeadline(time.Now().Add(s.Deadtime))
				conn.UDPUp.WriteToUDP(b[7:], addr)
			} ()
		case 0x04:
			addr.IP = b[1:17]
			addr.Port = (int(b[17]) << 8) | int(b[18])
			go func() {
				conn.UDPUp.SetWriteDeadline(time.Now().Add(s.Deadtime))
				conn.UDPUp.WriteToUDP(b[19:], addr)
			} ()
		case 0x03:
			var ip []byte
			var l byte
			l = b[1]
			addr.Port = (int(b[l+1]) << 8) | int(b[l + 2])
			_, ip, err = lookupDNS(b[1:1+l])
			addr.IP = net.IP(ip)
			go func() {
				conn.UDPUp.SetWriteDeadline(time.Now().Add(s.Deadtime))
				conn.UDPUp.WriteToUDP(b[l+3:], addr)
			} ()
		}
	}
	return
}

func (s S5UH) SetDeadline(t time.Duration) error {
	s.Deadtime = t
	return nil
}
