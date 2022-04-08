package player

import (
	"hero-emulator/database"
	"hero-emulator/nats"
	"hero-emulator/utils"
)

type (
	AidHandler struct{}
)

func (h *AidHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if sale := database.FindSale(s.Character.PseudoID); sale != nil {
		return nil, nil
	}

	activated := data[5] == 1

	resp := utils.Packet{}
	if s.Character.HasAidBuff() && s.Character.AidTime < 60 {
		s.Character.AidTime = 60
		stData, _ := s.Character.GetStats()
		resp.Concat(stData)
	}

	s.Character.AidMode = activated

	resp.Concat(s.Character.AidStatus())

	p := &nats.CastPacket{CastNear: true, CharacterID: s.Character.ID, Data: s.Character.GetHPandChi()}
	p.Cast()

	return resp, nil
}
