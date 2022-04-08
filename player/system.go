package player

import (
	"hero-emulator/database"
	"hero-emulator/utils"
)

type CharacterMenuHandler struct {
}

type ServerMenuHandler struct {
}

type QuitGameHandler struct {
}

var (
	CHARACTER_MENU = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x09, 0x09, 0x00, 0x55, 0xAA}
	SERVER_MENU    = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x09, 0x08, 0x00, 0x55, 0xAA}
	QUIT_GAME      = utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x21, 0x02, 0x00, 0x55, 0xAA}
)

func (h *CharacterMenuHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if c := s.Character; c != nil {
		s.User.ConnectingIP = s.ClientAddr
		s.User.ConnectingTo = s.User.ConnectedServer
		c.Logout()
	}

	return CHARACTER_MENU, nil
}

func (h *ServerMenuHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if c := s.Character; c != nil {
		c.Logout()
	}

	resp := SERVER_MENU
	return resp, nil
}

func (h *QuitGameHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if c := s.Character; c != nil {
		resp := QUIT_GAME
		resp.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 6)

		s.User.ConnectingIP = ""
		s.OnClose()
		database.DeleteUserFromCache(s.User.ID)
		return resp, nil
	}

	return nil, nil
}
