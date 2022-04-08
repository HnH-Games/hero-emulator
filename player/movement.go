package player

import (
	"math"
	"strconv"
	"strings"
	"time"

	"hero-emulator/database"
	"hero-emulator/nats"
	"hero-emulator/utils"
)

type MovementHandler struct {
}

var (
	CHARACTER_MOVEMENT = utils.Packet{0xAA, 0x55, 0x22, 0x00, 0x22, 0x01, 0x00, 0x00, 0x00, 0x00, 0xC8, 0xB0, 0xFE, 0xBE, 0x00, 0x00, 0x55, 0xAA}
)

func (h *MovementHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	c := s.Character
	if c == nil {
		return nil, nil
	}

	if !c.IsActive {
		c.IsActive = true
		go c.ActivityStatus(0)
	}

	if len(data) < 26 {
		return nil, nil
	}
	if c.Map == 255 && database.IsFactionWarEntranceActive() {
		parts := strings.Split(c.Coordinate, ",")
		y := strings.Trim(parts[1], ")")
		py := strings.Split(y, ".")
		y = py[0]
		posY, _ := strconv.ParseInt(y, 10, 64)
		if c.Faction == 1 && posY < 440 {

			coordinate := &utils.Location{X: 325, Y: 465}
			gomap, _ := c.ChangeMap(255, coordinate)
			c.Socket.Write(gomap)

		}
		if c.Faction == 2 && posY > 80 {

			coordinate := &utils.Location{X: 179, Y: 45}
			gomap, _ := c.ChangeMap(255, coordinate)
			c.Socket.Write(gomap)

		}
	}

	movType := utils.BytesToInt(data[4:6], false)
	speed := float64(0.0)

	if movType == 8705 { // movement
		speed = 5.6
	} else if movType == 8706 || movType == 9732 { // running or flying
		speed = c.RunningSpeed + c.AdditionalRunningSpeed
	}

	resp := CHARACTER_MOVEMENT
	resp.Insert(utils.IntToBytes(uint64(s.Character.PseudoID), 2, true), 5) // character pseudo id

	resp[4] = data[4]
	resp[7] = data[5]            // running mode
	resp.Insert(data[6:14], 8)   // current coordinate-x & coordinate-y
	resp.Insert(data[18:26], 20) // target coordinate-x & coordinate-y

	resp.Insert(utils.FloatToBytes(speed, 4, true), 32) // speed

	p := &nats.CastPacket{CastNear: true, CharacterID: s.Character.ID, Data: resp, Type: nats.PLAYER_MOVEMENT}
	err := p.Cast()
	if err != nil {
		return nil, err
	}

	coordinate := &utils.Location{X: utils.BytesToFloat(data[6:10], true), Y: utils.BytesToFloat(data[10:14], true)}
	c.SetCoordinate(coordinate)
	token := utils.RandInt(0, math.MaxInt64)
	c.MovementToken = token

	target := &utils.Location{X: utils.BytesToFloat(data[18:22], true), Y: utils.BytesToFloat(data[22:26], true)}
	if c.IsinWar && !database.WarStarted {
		if coordinate.X >= 155 && c.Faction == 1 && target.X > 155 || target.Y > 65 && c.Faction == 1 {
			target.X = 155
			target.Y = coordinate.Y
			if target.Y > 65 {
				target.Y = 65
			}
			c.SetCoordinate(target)
			mapID, _ := s.Character.ChangeMap(c.Map, target)
			s.Conn.Write(mapID)
			return nil, nil
		} else if coordinate.X >= 147 && c.Faction == 2 && target.X > 147 || target.Y < 457 && c.Faction == 2 {
			target.X = 147
			target.Y = coordinate.Y
			if target.Y < 457 {
				target.Y = 457
			}
			c.SetCoordinate(target)
			mapID, _ := s.Character.ChangeMap(c.Map, target)
			s.Conn.Write(mapID)
			return nil, nil
		}
	}
	distance := utils.CalculateDistance(coordinate, target)
	delay := distance * 1000 / speed // delay (ms)
	time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		if c.MovementToken == token {
			c.SetCoordinate(target)
		}
	})

	if speed > 5.6 {
		s.Stats.CHI -= int(speed) / 2
		if s.Stats.CHI < 0 {
			s.Stats.CHI = 0
		}
		resp.Concat(c.GetHPandChi())
	}

	return resp, nil
}
