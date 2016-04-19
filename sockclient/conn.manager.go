package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
)

type PROXY_STATU int

const (
	PROXY_STATU_REQIN PROXY_STATU = 1
	PROXY_STATU_CHECK PROXY_STATU = 2
	PROXY_STATU_CONN  PROXY_STATU = 3
	PROXY_STATU_BIND  PROXY_STATU = 4
	PROXY_STATU_TRANS PROXY_STATU = 5
)

type HTTP_STYLE int

const (
	HTTP_STYLE_HTTP    HTTP_STYLE = 1
	HTTP_STYLE_CONNECT HTTP_STYLE = 2
)

type ConnMngr struct {
	frontcon    net.Conn
	aftercon    net.Conn
	curstatu    PROXY_STATU
	stop        bool
	lastsize    int
	lastmsgsize int
	lastmsg     []byte
}

func NewConnMngr(con net.Conn) ConnMngr {
	var cmt ConnMngr
	cmt.frontcon = con
	cmt.curstatu = PROXY_STATU_REQIN
	return cmt
}

func (cm *ConnMngr) DealReqin(buf []byte, size int) {
	rbuf := make([]byte, 100)
	rbuf[0] = 0x05
	rsize := 0
	if size >= 3 {
		if buf[0] == 0x05 {
			if buf[1] == 0x01 {
				if buf[2] == 0x00 {
					rbuf[1] = 0x00
					rsize = 2
					cm.curstatu = PROXY_STATU_CONN
					fmt.Println("1.匿名访问")
				} else if buf[2] == 0x02 {
					cm.stop = true
					cm.frontcon.Close()
					fmt.Println("验证访问拒绝")
				}

			} else {
				fmt.Println("2.匿名访问")
				rbuf[1] = 0x00
				rsize = 2
				cm.curstatu = PROXY_STATU_CONN
			}
		} else {
			fmt.Println("SOCK版本错误")
			cm.frontcon.Close()
			cm.stop = true
		}
	} else {
		cm.stop = true
		cm.frontcon.Close()
		fmt.Println("消息大小错误")
	}
	if rsize > 0 {
		//fmt.Println(rbuf[:rsize])
		tmp := rbuf[:rsize]
		_, err := cm.frontcon.Write(tmp)
		if err != nil {
			fmt.Println(err)
			cm.stop = true
		}
	}

}

func (cm *ConnMngr) AfterConn(addr string, port int) (err error) {
	return err
}

func (cm *ConnMngr) ConAESServer(buf []byte, size int) (net.Conn, error) {

	enbuf, err := aesEncrypt(buf[:size])
	if err != nil {
		//fmt.Println(err)
		return nil, err
	}
	conn, err := net.DialTimeout("tcp", AESADDR, time.Duration(3*time.Second))
	if err != nil {
		return nil, err
	}
	sendbuf := make([]byte, 200)
	sendbuf[0] = byte(int8(len(enbuf)))
	sendbuf = append(sendbuf[:1], enbuf...)
	fmt.Println("ConAESServer sendbuf size:=", len(sendbuf))
	_, err = conn.Write(sendbuf)
	if err != nil {
		conn.Close()
		return nil, err
	}
	readbuf := make([]byte, 100)
	_, err = conn.Read(readbuf)
	if err != nil {
		conn.Close()
		return nil, err
	}
	return conn, err
}

func (cm *ConnMngr) DealConn(buf []byte, size int) {
	rbuf := make([]byte, 100)
	rbuf[0] = 0x05
	rsize := 0
	if size >= 3 {
		if buf[0] == 0x05 {
			if buf[1] == 0x01 && buf[2] == 0x00 {
				if buf[3] == 0x01 {
					addr_len := buf[4]
					fmt.Println(addr_len)
					addr_bufer := make([]byte, addr_len)
					addr_bufer = buf[5 : addr_len+5]
					addr_str := string(addr_bufer)
					fmt.Println(addr_str)
					var port uint16
					br := bytes.NewReader(buf[addr_len+5:])
					err := binary.Read(br, binary.BigEndian, &port)
					if err != nil {
						cm.stop = true
						cm.frontcon.Close()
					}
					fmt.Println(port)

				} else if buf[3] == 0x03 {
					addr_len := buf[4]
					fmt.Println(addr_len)
					addr_bufer := make([]byte, addr_len)
					addr_bufer = buf[5 : addr_len+5]
					addr_str := string(addr_bufer)
					//fmt.Println(addr_str)
					var port uint16
					//fmt.Println(buf[addr_len+5 : addr_len+7])
					br := bytes.NewReader(buf[addr_len+5 : addr_len+7])
					binary.Read(br, binary.BigEndian, &port)

					port_str := strconv.Itoa(int(port))
					fmt.Println("conn ", addr_str+":"+port_str)

					conn, err := cm.ConAESServer(buf, size)

					if err != nil {
						cm.stop = true
						cm.frontcon.Close()

					} else {
						fmt.Println("conn success!")
						cm.aftercon = conn
						go cm.DealAftercon()
						cm.curstatu = PROXY_STATU_TRANS
						rbuf[1] = 0x00
						rbuf[2] = 0x00
						rbuf[3] = 0x01
						rbuf[4] = 0x00
						rbuf[5] = 0x00
						rbuf[6] = 0x00
						rbuf[7] = 0x00
						rbuf[8] = 0x00
						rbuf[9] = 0x00
						rsize = 10
						fmt.Println(rbuf[:rsize])
					}

				} else {
					cm.stop = true
					cm.frontcon.Close()
				}

			} else {
				cm.stop = true
				cm.frontcon.Close()
			}

		} else {
			cm.stop = true
			cm.frontcon.Close()
		}
	}
	if rsize > 0 {
		fmt.Println(rbuf[:rsize])
		tmp := rbuf[:rsize]
		_, err := cm.frontcon.Write(tmp)
		if err != nil {
			fmt.Println(err)
			cm.stop = true
		}
	}
}

func (cm *ConnMngr) DealCheck(buf []byte, size int) {

}

func (cm *ConnMngr) DealBind(buf []byte, size int) {

}

func (cm *ConnMngr) DealTrans(buf []byte, size int) {
	//	fmt.Println(string(buf[:size]))
	ttbuf := make([]byte, size)
	copy(ttbuf, buf[:size])
	enbuf, err := aesEncrypt(ttbuf)
	if err != nil {
		fmt.Println(err)
		cm.stop = true
	}

	senbuf := make([]byte, 150)
	senbuf[0] = byte(int8(len(enbuf)))

	senbuf = append(senbuf[:1], enbuf...)

	_, err = cm.aftercon.Write(senbuf)

	if err != nil {
		fmt.Println(err)
		cm.stop = true
	}
}

func (cm *ConnMngr) DealAftercon() {
	readbuf := make([]byte, 150)
	for {
		if cm.stop {
			cm.frontcon.Close()
			cm.aftercon.Close()
			break
		}
		size, err := cm.aftercon.Read(readbuf)
		if err != nil {
			fmt.Println("dealaftercon", err)
			cm.stop = true
			cm.frontcon.Close()
			break
		}
		//fmt.Println(readbuf[:size])
		deal_size := 0
		for deal_size < size {

			if cm.lastsize == 0 {
				msglen := int(readbuf[deal_size])

				deal_size = deal_size + 1
				if int(msglen) <= size-deal_size {
					okbuf := readbuf[deal_size:(deal_size + msglen)]
					debuf, err := aesDecrypt(okbuf)
					if err != nil {
						cm.stop = true
						break
					}
					_, err = cm.frontcon.Write(debuf)
					if err != nil {
						fmt.Println(err)
						cm.stop = true
						break
					}
					deal_size = deal_size + msglen
				} else {
					cm.lastmsg = append(cm.lastmsg[:0], readbuf[deal_size:]...)
					cm.lastmsgsize = msglen
					cm.lastsize = msglen - (size - deal_size)
					deal_size = size
				}
			} else if cm.lastsize > 0 {
				if cm.lastsize <= size-deal_size {
					okbuf := append(cm.lastmsg[:cm.lastmsgsize-cm.lastsize], readbuf[deal_size:deal_size+cm.lastsize]...)
					debuf, err := aesDecrypt(okbuf)
					if err != nil {
						fmt.Println("last size!=0", err)
						cm.stop = true
						break
					}
					_, err = cm.frontcon.Write(debuf)
					if err != nil {
						fmt.Println(err)
						cm.stop = true
						break
					}
					deal_size = deal_size + cm.lastsize
					cm.lastsize = 0
				} else {
					//copy(cm.lastmsg)
					cm.lastmsg = append(cm.lastmsg[:cm.lastmsgsize-cm.lastsize], readbuf[deal_size:]...)
					cm.lastsize = cm.lastsize - (size - deal_size)
					deal_size = size
				}
			}
		}

	}

}

func (cm *ConnMngr) DealHTTPConn(readbuf []byte, size int) {

	readstr := string(readbuf[:size])
	//fmt.Println(readstr)
	cmap := ConvertStr2Map(readstr)
	buf := make([]byte, 100)
	buf[0] = 0x05
	buf[1] = 0x01
	buf[2] = 0x00
	buf[3] = 0x03
	hoststr := cmap["Host"]
	if hoststr == "" {
		fmt.Println("can't get host or header > 1024")
		cm.stop = true
		cm.frontcon.Close()
		return
	}
	//check proxy style

	pos := strings.IndexRune(readstr, ' ')
	var style HTTP_STYLE
	if pos > 4 && pos < 8 {
		style = HTTP_STYLE_CONNECT
	}
	if pos < 5 {
		style = HTTP_STYLE_HTTP
	}

	hoststr = strings.Trim(hoststr, " ")
	hostarr := strings.Split(hoststr, ":")

	buf[4] = byte(int8(len(hostarr[0])))
	buf = append(buf[:5], hostarr[0][0:]...)
	var port int
	if len(hostarr) > 1 {
		port, _ = strconv.Atoi(hostarr[1])
	} else {
		port = 80
	}

	b_buf := new(bytes.Buffer)
	binary.Write(b_buf, binary.BigEndian, uint16(port))
	fmt.Println(b_buf.Bytes())
	buf = append(buf[:(5+len(hostarr[0]))], b_buf.Bytes()[0:]...)

	fmt.Println("DealHTTPConn ", hostarr[0], port)

	conn, err := cm.ConAESServer(buf, 7+len(hostarr[0]))

	if err != nil {
		fmt.Println("conn failed!", err)
		cm.stop = true
		return
	} else {
		fmt.Println("conn success!")
		cm.aftercon = conn
		go cm.DealAftercon()
		cm.curstatu = PROXY_STATU_TRANS
	}
	if style == HTTP_STYLE_HTTP {

		readstr = strings.Replace(readstr, "http://"+hoststr, "", 1)
		readstr = strings.Replace(readstr, "Proxy-Connection", "Connection", 1)
		fmt.Println(readstr)

		default_send_size := 100
		size = len(readstr)
		lastsize := size
		for ; lastsize > default_send_size; lastsize = lastsize - default_send_size {
			start_size := size - lastsize
			end_size := default_send_size + start_size
			fmt.Println(start_size, end_size)
			tt := readstr[start_size:end_size]
			cm.DealTrans([]byte(tt), default_send_size)
		}
		if lastsize > 0 {
			start_size := size - lastsize
			fmt.Println("caonima last size", lastsize)
			cm.DealTrans([]byte(readstr[start_size:]), lastsize)
		}
	} else {
		fmt.Println(readstr)
		retstr := []byte("HTTP/1.0 200 Connection established\r\n\r\n")
		fmt.Println(string(retstr))
		cm.frontcon.Write(retstr)
	}

}

func (cm *ConnMngr) DealHttpConn(buf []byte, size int) {
	default_send_size := 100
	lastsize := size
	for ; lastsize > default_send_size; lastsize = lastsize - default_send_size {
		start_size := size - lastsize
		end_size := default_send_size + start_size
		//fmt.Println(start_size, end_size)
		tt := buf[start_size:end_size]
		cm.DealTrans(tt, default_send_size)
	}
	if lastsize > 0 {
		start_size := size - lastsize
		cm.DealTrans(buf[start_size:], lastsize)
	}
}

func (cm *ConnMngr) DealHttp() {
	fmt.Println(cm.frontcon.RemoteAddr())
	readbuf := make([]byte, 2048)
	defer cm.frontcon.Close()
	for {
		if cm.stop {
			break
		}
		size, err := cm.frontcon.Read(readbuf)
		//fmt.Println(size, readbuf[:size], cm.curstatu)
		if err != nil {
			fmt.Println("deal", err)
			cm.stop = true
			cm.frontcon.Close()
			if cm.aftercon != nil {
				cm.aftercon.Close()
			}
			break
		}

		switch cm.curstatu {
		case PROXY_STATU_REQIN:
			cm.DealHTTPConn(readbuf, size)
		case PROXY_STATU_TRANS:
			cm.DealHttpConn(readbuf, size)
		}
	}
}

func (cm *ConnMngr) DealSock() {
	fmt.Println(cm.frontcon.RemoteAddr())
	readbuf := make([]byte, 100)
	defer cm.frontcon.Close()
	for {
		if cm.stop {
			break
		}
		size, err := cm.frontcon.Read(readbuf)
		//fmt.Println(size, readbuf[:size], cm.curstatu)
		if err != nil {
			fmt.Println("deal", err)
			cm.stop = true
			cm.frontcon.Close()
			cm.aftercon.Close()
			break
		}

		switch cm.curstatu {
		case PROXY_STATU_REQIN:
			cm.DealReqin(readbuf, size)
		case PROXY_STATU_CHECK:
			cm.DealCheck(readbuf, size)
		case PROXY_STATU_CONN:
			cm.DealConn(readbuf, size)
		case PROXY_STATU_TRANS:
			cm.DealTrans(readbuf, size)
		case PROXY_STATU_BIND:
			cm.DealBind(readbuf, size)
		}
	}
}
