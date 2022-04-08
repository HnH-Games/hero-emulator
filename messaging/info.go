package messaging

import (
	"hero-emulator/utils"
)

var (
	INFO_MESSAGE   = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x71, 0x08, 0x00, 0x55, 0xAA}
	SYSTEM_MESSAGE = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x01, 0x03, 0x55, 0xAA}
)

func InfoMessage(message string) []byte {

	resp := INFO_MESSAGE
	resp.SetLength(int16(len(message)) + 3) // packet length
	resp[6] = byte(len(message))            // message length
	resp.Insert([]byte(message), 7)         // message

	return resp
}

func SystemMessage(code uint64) []byte {

	resp := SYSTEM_MESSAGE
	resp.Insert(utils.IntToBytes(code, 2, true), 6) // err code
	return resp
}
