package database

import (
	"fmt"
	"time"

	"hero-emulator/messaging"

	"hero-emulator/nats"
	"hero-emulator/utils"
)

var (
	ANNOUNCEMENT    = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x71, 0x06, 0x00, 0x55, 0xAA}
	START_WAR       = utils.Packet{0xaa, 0x55, 0x23, 0x00, 0x65, 0x01, 0x00, 0x00, 0x17, 0x00, 0x00, 0x00, 0x10, 0x27, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0d, 0x00, 0x00, 0x00, 0x10, 0x27, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xb0, 0x04, 0x00, 0x00, 0x55, 0xaa}
	OrderCharacters = make(map[int]*Character)
	ShaoCharacters  = make(map[int]*Character)
	LobbyCharacters = make(map[int]*Character)

	WarRequirePlayers = 10
	OrderPoints       = 10000
	ShaoPoints        = 10000
	CanJoinWar        = false
	WarStarted        = false
	WarStonesIDs      = []uint16{}
	WarStones         = make(map[int]*WarStone)
	ActiveWars        = make(map[int]*ActiveWar)
)

type WarStone struct {
	PseudoID      uint16 `db:"-" json:"-"`
	NpcID         int    `db:"-" json:"-"`
	ConqueredID   int    `db:"-" json:"-"`
	ConquereValue int    `db:"-" json:"-"`
	NearbyZuhang  int    `db:"-" json:"-"`
	NearbyShao    int    `db:"-" json:"-"`
	NearbyZuhangV []int  `db:"-" json:"-"`
	NearbyShaoV   []int  `db:"-" json:"-"`
}
type ActiveWar struct {
	WarID         uint16       `db:"-" json:"-"`
	ZuhangPlayers []*Character `db:"-" json:"-"`
	ShaoPlayers   []*Character `db:"-" json:"-"`
}

func makeAnnouncement(msg string) {
	length := int16(len(msg) + 3)

	resp := ANNOUNCEMENT
	resp.SetLength(length)
	resp[6] = byte(len(msg))
	resp.Insert([]byte(msg), 7)

	p := nats.CastPacket{CastNear: false, Data: resp}
	p.Cast()
}
func JoinToWarLobby(char *Character) {
	LobbyCharacters[char.ID] = char
	warReady := false
	zuhangPlayers := 0
	shaoPlayers := 0
	if len(LobbyCharacters) >= WarRequirePlayers {
		newWar := &ActiveWar{WarID: 1}
		ActiveWars[int(1)] = newWar
		for _, char := range LobbyCharacters {
			if char.Faction == 1 {
				if zuhangPlayers < WarRequirePlayers/2 {
					zuhangPlayers++
					ActiveWars[int(1)].ZuhangPlayers = append(ActiveWars[int(1)].ZuhangPlayers, char)
				}
			} else {
				if shaoPlayers < WarRequirePlayers/2 {
					shaoPlayers++
					ActiveWars[int(1)].ShaoPlayers = append(ActiveWars[int(1)].ShaoPlayers, char)
				}
			}
			if zuhangPlayers >= WarRequirePlayers/2 && shaoPlayers >= WarRequirePlayers/2 {
				warReady = true
				continue
			}
		}
	}
	if warReady {
		for _, char := range ActiveWars[int(1)].ShaoPlayers {
			char.Socket.Write(messaging.InfoMessage(fmt.Sprintf("Your war is ready. /accept war ")))
		}
		for _, char := range ActiveWars[int(1)].ZuhangPlayers {
			char.Socket.Write(messaging.InfoMessage(fmt.Sprintf("Your war is ready. /accept war ")))
		}
		CheckALlPlayersReady()
	}
}
func secondsToMinutes(inSeconds int) (int, int) {
	minutes := inSeconds / 60
	seconds := inSeconds % 60
	return minutes, seconds
}

func StartWarTimer(prepareWarStart int) {
	min, sec := secondsToMinutes(prepareWarStart)
	msg := fmt.Sprintf("%d minutes %d second after the Great War will start.", min, sec)
	msg2 := fmt.Sprintf("Please participate war by ")
	makeAnnouncement(msg)
	makeAnnouncement(msg2)
	if prepareWarStart > 0 {
		time.AfterFunc(time.Second*10, func() {
			StartWarTimer(prepareWarStart - 10)
		})
	} else {
		StartWar()
	}
}

func ResetWar() {
	time.AfterFunc(time.Second*10, func() {
		for _, char := range OrderCharacters {
			if char.IsOnline {

			} else {
				delete(OrderCharacters, char.ID)
			}
			char.WarKillCount = 0
			char.WarContribution = 0
			char.IsinWar = false
			gomap, _ := char.ChangeMap(1, nil)
			char.Socket.Write(gomap)
			delete(OrderCharacters, char.ID)
		}
		for _, char := range ShaoCharacters {
			if char.IsOnline {

			} else {
				delete(OrderCharacters, char.ID)
			}
			char.WarKillCount = 0
			char.WarContribution = 0
			char.IsinWar = false
			gomap, _ := char.ChangeMap(1, nil)
			char.Socket.Write(gomap)
			delete(ShaoCharacters, char.ID)
		}
		for _, stones := range WarStones {
			stones.ConquereValue = 0
			stones.ConqueredID = 0
		}
	})
}

func StartWar() {
	resp := START_WAR
	byteOrders := utils.IntToBytes(uint64(len(OrderCharacters)), 4, false)
	byteShaos := utils.IntToBytes(uint64(len(ShaoCharacters)), 4, false)
	resp.Overwrite(byteOrders, 8)
	resp.Overwrite(byteShaos, 22)
	for _, char := range OrderCharacters {
		char.Socket.Write(resp)
	}
	for _, char := range ShaoCharacters {
		char.Socket.Write(resp)
	}

	CanJoinWar = false
	WarStarted = true
	StartInWarTimer()
}

func (self *WarStone) RemoveZuhang(id int) {
	for i, other := range self.NearbyZuhangV {
		if other == id {
			self.NearbyZuhangV = append(self.NearbyZuhangV[:i], self.NearbyZuhangV[i+1:]...)
			break
		}
	}
}

func (self *WarStone) RemoveShao(id int) {
	for i, other := range self.NearbyShaoV {
		if other == id {
			self.NearbyShaoV = append(self.NearbyShaoV[:i], self.NearbyShaoV[i+1:]...)
			break
		}
	}
}
