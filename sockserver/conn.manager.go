package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"
)

//protocol  00+buffer

type PROXY_STATU int

const (
	PROXY_STATU_REQIN PROXY_STATU = 1
	PROXY_STATU_CHECK PROXY_STATU = 2
	PROXY_STATU_CONN  PROXY_STATU = 3
	PROXY_STATU_BIND  PROXY_STATU = 4
	PROXY_STATU_TRANS PROXY_STATU = 5
)

type ConnMngr struct {
	frontcon    net.Conn
	aftercon    net.Conn
	curstatu    PROXY_STATU
	lastsize    int
	lastmsgsize int
	lastmsg     []byte
	stop        bool
	ch          chan []byte
	chclose     bool
}

func NewConnMngr(con net.Conn) ConnMngr {
	var cmt ConnMngr
	cmt.frontcon = con
	cmt.curstatu = PROXY_STATU_CONN
	cmt.lastsize = 0
	cmt.lastmsgsize = 0
	cmt.lastmsg = make([]byte, 150)
	cmt.ch = make(chan []byte)
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
		fmt.Println(rbuf[:rsize])
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

					var port uint16

					br := bytes.NewReader(buf[addr_len+5 : addr_len+7])
					binary.Read(br, binary.BigEndian, &port)

					port_str := strconv.Itoa(int(port))
					fmt.Println("conn ", addr_str+":"+port_str)
					conn, err := net.DialTimeout("tcp", addr_str+":"+port_str, time.Duration(3*time.Second))
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
		tmp := rbuf[:rsize]
		enbuf, err := aesEncrypt(tmp)
		_, err = cm.frontcon.Write(enbuf)
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
	_, err := cm.aftercon.Write(buf[:size])
	if err != nil {
		fmt.Println(err)
		cm.stop = true
	}
}

func (cm *ConnMngr) DealAftercon() {
	readbuf := make([]byte, 100)
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
			cm.aftercon.Close()
			break
		}
		if size > 0 {
			enbuf, err := aesEncrypt(readbuf[:size])
			if err != nil {
				fmt.Println(err)
				cm.stop = true
				break
			}

			senbuf := make([]byte, 200)
			senbuf[0] = byte(int8(len(enbuf)))

			senbuf = append(senbuf[:1], enbuf...)
			_, err = cm.frontcon.Write(senbuf)

			if err != nil {
				fmt.Println(err)
				cm.stop = true
				break
			}
		}
	}

}

func (cm *ConnMngr) GetllegeMsg(readbuf []byte, size int) {

	deal_size := 0
	for deal_size < size {
		if cm.lastsize == 0 {
			msglen := int(readbuf[deal_size])
			fmt.Println("get msg_len", msglen)
			deal_size = deal_size + 1
			if int(msglen) <= size-deal_size {
				okbuf := readbuf[deal_size:(deal_size + msglen)]
				debuf, err := aesDecrypt(okbuf)
				if err != nil {
					cm.stop = true
					if !cm.chclose {
						close(cm.ch)
						cm.chclose = true
					}
					break
				}
				cm.ch <- debuf
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
					if !cm.chclose {
						close(cm.ch)
						cm.chclose = true
					}
					break
				}
				cm.ch <- debuf
				deal_size = deal_size + cm.lastsize
				cm.lastsize = 0
			} else {
				cm.lastmsg = append(cm.lastmsg[:cm.lastmsgsize-cm.lastsize], readbuf[deal_size:]...)
				cm.lastsize = cm.lastsize - (size - deal_size)
				deal_size = size
			}
		}
	}

}

func (cm *ConnMngr) ConnRead() {
	fmt.Println(cm.frontcon.RemoteAddr())
	readbuf := make([]byte, 150)

	for {
		size, err := cm.frontcon.Read(readbuf)

		if err != nil {
			fmt.Println("deal", err)
			cm.stop = true
			if !cm.chclose {
				close(cm.ch)
				cm.chclose = true
			}
			break
		}
		cm.GetllegeMsg(readbuf, size)
	}
}

func (cm *ConnMngr) Deal() {

	defer cm.frontcon.Close()

	go cm.ConnRead()

	for debuf := range cm.ch {
		if cm.stop {
			break
		}

		switch cm.curstatu {
		case PROXY_STATU_REQIN:
			cm.DealReqin(debuf, len(debuf))
		case PROXY_STATU_CHECK:
			cm.DealCheck(debuf, len(debuf))
		case PROXY_STATU_CONN:
			cm.DealConn(debuf, len(debuf))
		case PROXY_STATU_TRANS:
			cm.DealTrans(debuf, len(debuf))
		case PROXY_STATU_BIND:
			cm.DealBind(debuf, len(debuf))
		}
	}
}
