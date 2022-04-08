package auth

import (
	"hero-emulator/database"
	"hero-emulator/utils"
)

type ListServersHandler struct {
	header string
}

var (
	SERVER_LIST = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x00, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
)

func (lsh *ListServersHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	lsh.header = "Dragon"
	return lsh.listServers(s)
}

func (lsh *ListServersHandler) listServers(s *database.Socket) ([]byte, error) {

	go disconnectMyCharacters(s.User)
	resp := SERVER_LIST
	resp.Insert([]byte{byte(len(lsh.header))}, 12) // header length
	resp.Insert([]byte(lsh.header), 13)

	servers, err := database.GetServers()
	if err != nil {
		return nil, err
	}

	length := len(lsh.header) + 11

	index := len(lsh.header) + 13
	resp.Insert([]byte{byte(len(servers))}, index) // server count

	index += 2
	for _, s := range servers {
		resp.Insert([]byte{byte(s.ID) - 1, 0x00}, index) //server index
		index += 2

		resp.Insert([]byte{byte(len(s.Name))}, index) // server name length
		index += 1

		resp.Insert([]byte(s.Name), index) // server name
		index += len(s.Name)

		resp.Insert(utils.IntToBytes(uint64(s.ConnectedUsers), 2, true), index) // server connections
		index += 2

		resp.Insert(utils.IntToBytes(uint64(s.MaxUsers), 2, true), index) // maximum connections
		index += 2

		resp.Insert([]byte{0x12, 0x00, 0x00, 0x00, 0x01, 0x00}, index)
		index += 6

		length += len(s.Name) + 13
	}

	resp.SetLength(int16(length))
	return resp, nil
}

func disconnectMyCharacters(user *database.User) {
	if user == nil {
		return
	}

	characters, err := database.FindCharactersByUserID(user.ID)
	if err != nil {
		return
	}

	for _, c := range characters {
		c.IsOnline = false
		c.Update()
	}
}
