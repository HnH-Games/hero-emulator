package utils

import "fmt"

type Packet []byte

func (p *Packet) SetLength(length int16) {
	n := length / 256
	(*p)[3] = byte(n)
	(*p)[2] = byte(length - n*256)
}

func (p *Packet) Insert(data []byte, i int) {
	_data := make([]byte, len(data))
	copy(_data, data)
	*p = append((*p)[:i], append(_data, (*p)[i:]...)...)
}

func (p *Packet) Overwrite(data []byte, i int) {
	_data := make([]byte, len(data))
	copy(_data, data)
	*p = append((*p)[:i], append(_data, (*p)[i+len(data):]...)...)
}

func (p *Packet) Concat(data []byte) {
	p.Insert(data, len(*p))
}

func (p *Packet) Print() {
	var print string
	for _, b := range *p {
		print += fmt.Sprintf("%02X ", b)
	}
	fmt.Printf("%s\n", print)
}
