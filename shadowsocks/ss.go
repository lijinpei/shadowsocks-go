package shadowsocks

import (
	"io"
	"net"
	"time"
	"strconv"
	"fmt"
	"math/rand"
//	"encoding/hex"
)

func dummyFmt () {
	fmt.Println("123")
}

type SS5 struct {
	SS5LH
	SS5UH
}


func connToChan(conn *net.TCPConn, c chan []byte, t time.Duration, mpl uint, c *Cipher) (err error) {
	for {
		var b []byte = make([]byte, mpl)
		var n int
		conn.SetReadDeadline(time.Now().Add(t))
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
		var nb []byte = make([]byte, n)
		if nil != c {
			c.decrypt(nb, b[:n])
		} else {
			nb = b[:n]
		}
		c <- nb
	}
}

func chanToConn(conn *net.TCPConn, c chan []byte, t time.Duration, c *Cipher) (err error) {
	for {
		var b []byte
		b = <- c
		if nil == b {
			return nil
		}
		conn.SetWriteDeadline(time.Now().Add(t))
		var nb []byte
		if nil != c {
			nb = make(byte[], len(b))
			c.encrypt(nb, b)
		} else {
			nb = b
		}
		_, err = conn.Write(nb)
		if DbgLogPacket {
			Log.Debug(fmt.Sprintf("chanToConn packet %v", b))
		}
		if nil != err {
			return err
		}
	}
}

// Suppose only one upper half and lower half now
// Socks5LowerHalf
type SS5LH struct {
	S Socks
	Deadtime time.Duration
	MaxPacketLength uint
	AuthRep [2]byte
	Method string
	Password string
}

func (s *SS5LH) Init(cipher *Cipher) {
	s.MaxPacketLength =256
	s.Deadtime, _ = time.ParseDuration("200s")
	s.AuthRep = [2]byte{0x05, 0x00}
	s.Cipher = cipher
}

func (s SS5LH) Authenticate(conn * ConnPair) error {
	return nil
}

func (s SS5LH) ReadLH(conn *ConnPair) (err error) {
	if DbgCtlFlow {
		Log.Debug("S5LH ReadLH started")
	}
	err = connToChanCipher(conn.Down, conn.UpChan, s.Deadtime, s.MaxPacketLength, s.Cipher)
	if Debug && (nil != err) {
		Log.Debug("S5LU ReadLH error " + err.Error())
	}
	return err
}

func (s SS5LH) WriteLH(conn *ConnPair) (err error) {
	if DbgCtlFlow {
		Log.Debug("S5LH WriteLH started")
	}
	err = chanToConn(conn.Down, conn.DownChan, s.Deadtime, s.Cipher)
	if Debug && nil != err {
		Log.Debug("S5LH WriteLH error " + err.Error())
	}
	return err
}

func (s SS5LH) Deal(conn *ConnPair) (err error) {
	if DbgCtlFlow {
		Log.Debug("Deal started")
	}
	iv := make([]byte, s.c.info.ivLen)
	R.read(iv)
	conn.DownCipherWrite, _ = NewCipher(s.Method, s.Password, iv)
	_, err = conn.Down.Read(iv)
	conn.DownCipherRead, _ = NewCipher(s.Method, s.Password, iv)
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
			Log.Debug("Deal returned because GetRequestType Error")
		}
		return
	}
    if (0x03 == atype) {
		if DbgCtlFlow {
			Log.Debug("DNS lookup " + string(addr))
		}
        atype, addr, err = lookupDNS(addr)
		if DbgCtlFlow {
			Log.Debug(fmt.Sprintf("GetRequestType results: cmd %v atype %v addr %v port %v", cmd, atype, string(addr), port))
		}
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
			var l = len(bndAddr)
			if nil != bndAddr.To4() {
				atype = 0x01
				bndAddr = bndAddr[l-4:l]
				l = 4
			} else {
				atype = 0x04
				l = 16
			}
			if nil != err {
				if DbgCtlFlow {
					Log.Debug("Deal connect error " + err.Error())
				}
				return err
			}
			repPacket := make([]byte, l + 3)
			repPacket[0] = atype
			copy(repPacket[1:l+1], bndAddr)
			copy(repPacket[l+1:l+3], []byte{byte(bndPort >> 8), byte(bndPort & 0xff)})
			if DbgLogPacket {
				Log.Debug(fmt.Sprintf("Replay packet length %v packet %v", len(repPacket), repPacket))
			}
			nrepPacket := make([]byte, l + 3)
			conn.DownCipherWrite.encrypt(nrepPacketm repPacket)
			_, err1 := conn.Down.Write(conn.DownCipherWrite.iv)
			_, err2 := conn.Down.Write(nrepPacket)
			if nil != err1 {
				err = err1
			} else {
				err = err2
			}
			if nil != err {
				return
			}
			err = s.S.Relay(conn)
		case SOCKS_BIND:
			if DbgCtlFlow {
				Log.Debug("Deal bind")
			}
			// reply
		case SOCKS_UDP_ASSOCIATE:
			if DbgCtlFlow || DbgUDP {
				Log.Debug("Deal bind")
			}
	}
	if DbgCtlFlow {
		Log.Debug("Deal finished")
	}
	return
}

func (s SS5LH) Listen(addr *net.TCPAddr) error {
	if DbgCtlFlow {
		Log.Debug("SS5LH Listen started")
	}
	listener, _ := net.ListenTCP("tcp", addr)
	//listener.SetDeadline(time.Now().Add(s.Deadtime))
	for {
		conn, err := listener.AcceptTCP()
		if nil != err {
			Log.Debug("Accept connection error " + err.Error())
			continue
		}
		if DbgCtlFlow {
			Log.Debug("Accept connection")
		}
		go s.Deal(&ConnPair{Down:conn, Up:nil, UpChan:make(chan []byte, 20), DownChan:make(chan []byte, 20)})
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
	Methid string
	Password string
}

func (s *SS5UH) Init(c *Cipher) {
	s.MaxPacketLength =256
	s.IsRunning = true
	s.Deadtime, _ = time.ParseDuration("200s")
	s.c = c
}

func (s SS5UH) Connect(atype byte, addr []byte, port uint16, conn *ConnPair) (bndAddr net.IP, bndPort uint16, err error) {
	if DbgCtlFlow {
		Log.Debug("SS5UH Connect started")
	}
	var newConn net.Conn
	newConn, err = net.DialTimeout("tcp", s.Addr, s.Deadtime)
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
	conn.UpIV = make([]byte, s.c.info.ivLen)
	l := len(addr)
	b := make([]byte, 3+l)
	nb := make([]byte, 3+l)
	b[0] = atype
	b[l+1] = byte(port >> 8)
	b[l+2] = byte(port & 0xff)
	copy(b[1,l+1], addr)
	s.c.encrypt(nb, b)
	_, err = conn.Up.Write(nb)
	if DbgCtlFlow {
		Log.Debug("Connect finished")
		Log.Debug(fmt.Sprintf("Connect local results: addr %v port %v err %v", bndAddr, bndPort, err))
	}
	return
}

func (s SS5UH) ReadUH(conn *ConnPair) (err error) {
	if DbgCtlFlow {
		Log.Debug("S5UH ReadUH started")
	}
	err = connToChan(conn.Up, conn.DownChan, s.Deadtime, s.MaxPacketLength, conn.UpCipherRead)
	if Debug && (nil != err) {
		Log.Debug("S5UH ReadUH error " + err.Error())
	}
	return err
}

func (s SS5UH) WriteUH(conn *ConnPair) (err error) {
	if DbgCtlFlow {
		Log.Debug("S5UH WriteUH started")
	}
	err = chanToConn(conn.Up, conn.UpChan, s.Deadtime, conn.UpCipherWrite)
	if Debug && (nil != err) {
		Log.Debug("S5UH WriteUH error " + err.Error())
	}
	return err
}

func (s *SS5UH) SetDeadline(t time.Duration) error {
	s.Deadtime = t
	return nil
}
