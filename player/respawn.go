package player

import (
	"hero-emulator/database"
	"hero-emulator/nats"
	"hero-emulator/utils"
)

type RespawnHandler struct {
}

var ()

func (h *RespawnHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	resp := utils.Packet{}
	respawnType := data[5]

	stat := s.Stats
	if stat == nil {
		return nil, nil
	}

	switch respawnType {
	case 1: // Respawn at Safe Zone

		save := database.SavePoints[byte(s.Character.Map)]
		point := database.ConvertPointToLocation(save.Point)
		teleportData := s.Character.Teleport(point)
		resp.Concat(teleportData)

		s.Character.IsActive = false
		stat.HP = stat.MaxHP
		stat.CHI = stat.MaxCHI
		s.Character.Respawning = false
		hpData := s.Character.GetHPandChi()
		resp.Concat(hpData)

		p := nats.CastPacket{CastNear: true, CharacterID: s.Character.ID, Type: nats.PLAYER_RESPAWN}
		p.Cast()
		break

	case 4: // Respawn at Location
		return nil, nil // FIX later (hp does not update on client)
		if s.Character.Gold > 10000 {
			s.Character.Gold -= 10000
			s.Character.IsActive = false
			stat.HP = stat.MaxHP / 10
			stat.CHI = stat.MaxCHI / 10
			s.Character.Respawning = false

			hpData := s.Character.GetHPandChi()
			resp.Concat(hpData)

			coordinate := database.ConvertPointToLocation(s.Character.Coordinate)
			teleportData := s.Character.Teleport(coordinate)
			resp.Concat(teleportData)

			h := GetGoldHandler{}
			goldData, _ := h.Handle(s)
			resp.Concat(goldData)
			resp.Print()
		}
		break
	}

	go s.Character.ActivityStatus(30)
	return resp, nil
}
