package database

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"hero-emulator/utils"

	"github.com/nats-io/nats.go"
)

var (
	Handler     func(*Socket, []byte, uint16) ([]byte, error)
	Sockets     = make(map[string]*Socket)
	socketMutex sync.RWMutex
)

type Socket struct {
	Conn       net.Conn
	ClientAddr string
	User       *User
	Character  *Character
	Stats      *Stat
	Skills     *Skills
	HoustonSub *nats.Subscription

	handlePing   func() error
	pingDuration time.Duration
}

func init() {
}

func (s *Socket) SetPingHandler(handler func() error) {
	if handler == nil {
		handler = func() error {
			buf := make([]byte, 1)
			r := bufio.NewReader(s.Conn)

			s.Conn.SetReadDeadline(time.Now().Add(time.Nanosecond))
			_, err := r.Read(buf)

			if neterr, ok := err.(net.Error); ok && neterr.Timeout() {
				fmt.Println("ping connection", neterr)
				time.AfterFunc(s.pingDuration, func() {
					s.handlePing()
				})
				return nil
			}

			log.Println(err)
			return err
		}
	}

	s.handlePing = handler
	s.handlePing()
}

func (s *Socket) SetPingDuration(duration time.Duration) {
	s.pingDuration = duration
}

func (s *Socket) Add(id string) {
	socketMutex.Lock()
	defer socketMutex.Unlock()
	Sockets[id] = s
}

func (s *Socket) Remove(id string) {
	socketMutex.Lock()
	defer socketMutex.Unlock()
	delete(Sockets, id)
}

func GetSocket(id string) *Socket {
	socketMutex.RLock()
	defer socketMutex.RUnlock()
	return Sockets[id]
}

func (s *Socket) Read() {

	//reader := bufio.NewReader(s.Conn)
	//writer := bufio.NewWriter(s.Conn)

	for {
		buf := make([]byte, 4096)
		n, err := s.Conn.Read(buf)
		if err != nil { // do not remove connecting ip here
			s.OnClose()
			break
		}

		go func() {
			resp, err := s.recognizePacket(buf[:n])
			if err != nil {
				log.Println("recognize packet error:", err)
			}

			if len(resp) > 0 {
				packets := bytes.SplitAfter(resp, []byte{0x55, 0xAA})
				for _, packet := range packets {
					if len(packet) == 0 {
						continue
					}

					_, err := s.Conn.Write(packet)
					//_, err := writer.Write(packet)
					//writer.Flush()
					if err != nil {
						log.Println("send response error:", err)
						break
					}
					time.Sleep(time.Duration(len(packet)/25) * time.Millisecond)
				}
			}
		}()

	}
}

func (s *Socket) OnClose() {
	s.Conn.Close()
	if u := s.User; u != nil {
		s.Remove(u.ID)
		if s.User.ConnectingIP == "" {
			u.Logout()
		}
	}
	if c := s.Character; c != nil {
		c.Logout()
	}
	if s.HoustonSub != nil {
		s.HoustonSub.Unsubscribe()
	}
}

func (s *Socket) recognizePacket(data []byte) ([]byte, error) {
	packets := bytes.SplitAfter(data, []byte{0x55, 0xAA})

	resp := utils.Packet{}
	for _, packet := range packets {

		if len(packet) < 6 {
			continue
		}

		if os.Getenv("PROXY_ENABLED") == "1" {
			header, body := []byte{}, []byte{}
			if bytes.Contains(packet, []byte{0xAA, 0x55}) {
				pParts := bytes.Split(packet, []byte{0xAA, 0x55})
				if len(pParts) == 1 {
					body = append([]byte{0xAA, 0x55}, pParts[0]...)

				} else {
					header = pParts[0]
					body = append([]byte{0xAA, 0x55}, pParts[1]...)
				}
			} else {
				header = packet
			}

			s.ParseHeader(header)

			if len(body) > 0 {
				sign := uint16(utils.BytesToInt(body[4:6], false))
				d, err := Handler(s, body, sign)
				if err != nil {
					return nil, err
				}

				resp.Concat(d)
			}

		} else {
			s.ClientAddr = s.Conn.RemoteAddr().String()

			sign := uint16(utils.BytesToInt(packet[4:6], false))
			d, err := Handler(s, packet, sign)
			if err != nil {
				return nil, err
			}

			resp.Concat(d)
		}
	}

	return resp, nil
}

func (s *Socket) Write(data []byte) error {
	if s != nil && s.Conn != nil {
		_, err := s.Conn.Write(data)
		return err
	}

	return nil
}

func (s *Socket) ParseHeader(header []byte) {

	if len(header) == 0 {
		return
	}

	sHeader := string(header)
	if !strings.HasPrefix(sHeader, "PROXY TCP4") {
		return
	}

	parts := strings.Split(sHeader, " ")
	clientIP := parts[2]
	clientPort := parts[4]

	s.ClientAddr = fmt.Sprintf("%s:%s", clientIP, clientPort)
}
