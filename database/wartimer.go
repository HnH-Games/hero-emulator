package database

import (
	"fmt"
	"os"
	"time"

	"hero-emulator/messaging"

	"hero-emulator/utils"
)

var (
	TIMER_MENU     = utils.Packet{0xAA, 0x55, 0x08, 0x00, 0x65, 0x03, 0x00, 0x55, 0xAA}
	WAR_SCOREPANEL = utils.Packet{0xAA, 0x55, 0x30, 0x00, 0x65, 0x06, 0x55, 0xAA}
)

func CheckALlPlayersReady() {
	timein := time.Now().Add(time.Minute * 1)
	deadtime := timein.Format(time.RFC3339)

	v, err := time.Parse(time.RFC3339, deadtime)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for range time.Tick(1 * time.Second) {
		readyShaoPlayers := 0
		readyZuhangPlayers := 0
		timeRemaining := getTimeRemaining(v)
		if timeRemaining.t <= 0 {
			for _, char := range ActiveWars[int(1)].ShaoPlayers {
				if !char.IsAcceptedWar {
					ActiveWars[int(1)].RemoveLobbyShao(char)
					delete(LobbyCharacters, char.ID)
					char.Socket.Write(messaging.InfoMessage(fmt.Sprintf("You missed the ready check. You kicked from the lobby")))
				}
				char.IsAcceptedWar = false
			}
			for _, char := range ActiveWars[int(1)].ZuhangPlayers {
				if !char.IsAcceptedWar {
					ActiveWars[int(1)].RemoveLobbyZuhang(char)
					delete(LobbyCharacters, char.ID)
					char.Socket.Write(messaging.InfoMessage(fmt.Sprintf("You missed the ready check. You kicked from the lobby")))
				}
				char.IsAcceptedWar = false
			}
			break
		}
		for _, char := range ActiveWars[int(1)].ShaoPlayers {
			if char.IsAcceptedWar {
				readyShaoPlayers++
			}
		}
		for _, char := range ActiveWars[int(1)].ZuhangPlayers {
			if char.IsAcceptedWar {
				readyZuhangPlayers++
			}
		}
		if readyZuhangPlayers >= WarRequirePlayers/2 && readyShaoPlayers >= WarRequirePlayers/2 {
			for _, char := range ActiveWars[int(1)].ZuhangPlayers {
				x := 75.0
				y := 45.0
				char.IsinWar = true
				OrderCharacters[char.ID] = char
				data, _ := char.ChangeMap(230, ConvertPointToLocation(fmt.Sprintf("%.1f,%.1f", x, y)))
				char.Socket.Write(data)
			}
			for _, char := range ActiveWars[int(1)].ShaoPlayers {
				x := 81.0
				y := 475.0
				char.IsinWar = true
				ShaoCharacters[char.ID] = char
				data, _ := char.ChangeMap(230, ConvertPointToLocation(fmt.Sprintf("%.1f,%.1f", x, y)))
				char.Socket.Write(data)
			}
			StartWarTimer(int(60))
			break
		}
	}
}
func (self *ActiveWar) RemoveLobbyShao(char *Character) {
	for i, other := range self.ShaoPlayers {
		if other == char {
			self.ShaoPlayers = append(self.ShaoPlayers[:i], self.ShaoPlayers[i+1:]...)
			break
		}
	}
}
func (self *ActiveWar) RemoveLobbyZuhang(char *Character) {
	for i, other := range self.ZuhangPlayers {
		if other == char {
			self.ZuhangPlayers = append(self.ZuhangPlayers[:i], self.ZuhangPlayers[i+1:]...)
			break
		}
	}
}
func StartInWarTimer() {
	timein := time.Now().Add(time.Minute * 20)
	deadtime := timein.Format(time.RFC3339)

	v, err := time.Parse(time.RFC3339, deadtime)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for range time.Tick(1 * time.Second) {
		timeRemaining := getTimeRemaining(v)
		if timeRemaining.t <= 0 || OrderPoints <= 0 || ShaoPoints <= 0 {
			WarScorePanel()
			ResetWar()
			break
		}
		timeCounters := CalculateWarCountdown(timeRemaining)
		data := utils.IntToBytes(uint64(timeCounters), 4, true)
		shaoStones := 0
		ZuhangStones := 0
		index := 7
		resp := TIMER_MENU
		byteOrders := utils.IntToBytes(uint64(len(OrderCharacters)), 4, true)
		ordersPoint := utils.IntToBytes(uint64(OrderPoints), 4, true)
		byteShaos := utils.IntToBytes(uint64(len(ShaoCharacters)), 4, true)
		shaoPoint := utils.IntToBytes(uint64(ShaoPoints), 4, true)
		for _, stones := range WarStones {
			if stones.ConqueredID == 1 {
				ShaoPoints -= 2
				ZuhangStones++
			} else if stones.ConqueredID == 2 {
				OrderPoints -= 2
				shaoStones++
			}
		}
		resp.Insert(byteOrders, index)
		index += 4
		resp.Insert(ordersPoint, index)
		index += 4
		resp.Insert([]byte{0x00, 0x00, 0x00, 0x00}, index)
		index += 4
		if ZuhangStones > 0 {
			resp.Insert(utils.IntToBytes(uint64(ZuhangStones), 1, false), index)
			index++
			for _, stones := range WarStones {
				if stones.ConqueredID == 1 {
					resp.Insert(utils.IntToBytes(uint64(stones.NpcID), 4, true), index)
					index += 4
				}
			}
			resp.Insert([]byte{0x00}, index)
			index++
		} else {
			resp.Insert([]byte{0x00, 0x00}, index) //IDE JÖN MAJD HOGY KINEK HÁNY KÖVE VAN
			index += 2
		}
		resp.Insert(byteShaos, index)
		index += 4
		resp.Insert(shaoPoint, index)
		index += 4
		resp.Insert([]byte{0x00, 0x00, 0x00, 0x00}, index)
		index += 4
		if shaoStones >= 1 {
			resp.Insert(utils.IntToBytes(uint64(shaoStones), 1, false), index)
			index++
			for _, stones := range WarStones {
				if stones.ConqueredID == 2 {
					resp.Insert(utils.IntToBytes(uint64(stones.NpcID), 4, true), index)
					index += 4
				}
			}
		} else {
			resp.Insert([]byte{0x00}, index-2)
			index++
		}
		resp.Insert(data, index)
		index += 4
		/*resp.Insert(data2, index)
		index++*/
		length := index - 4
		resp.SetLength(int16(length))
		for _, char := range OrderCharacters {
			if char.IsOnline {
				char.Socket.Write(resp)
			} else {
				delete(OrderCharacters, char.ID)
			}
		}
		for _, char := range ShaoCharacters {
			if char.IsOnline {
				char.Socket.Write(resp)
			} else {
				delete(OrderCharacters, char.ID)
			}
		}
		for _, stones := range WarStones {
			if len(stones.NearbyZuhangV) > len(stones.NearbyShaoV) {
				if stones.ConquereValue > 0 {
					stones.ConquereValue--
				}
				if stones.ConquereValue >= 0 && stones.ConquereValue <= 30 {
					stones.ConqueredID = 1
				} else if stones.ConquereValue > 170 {
					stones.ConqueredID = 0
				}
			} else if len(stones.NearbyShaoV) > len(stones.NearbyZuhangV) {
				if stones.ConquereValue < 200 {
					stones.ConquereValue++
				}
				if stones.ConquereValue >= 170 && stones.ConquereValue <= 200 {
					stones.ConqueredID = 2
				} else if stones.ConquereValue < 30 {
					stones.ConqueredID = 0
				}
			}
		}
	}
}

type countdown struct {
	t int
	d int
	h int
	m int
	s int
}

func WarScorePanel() {
	zuhang_nyert := true
	resp := WAR_SCOREPANEL
	index := 6
	if OrderPoints > ShaoPoints {
		resp.Insert([]byte{0x00, 0x28, 0x00}, index)
	} else {
		resp.Insert([]byte{0x01, 0x28, 0x00}, index)
	}
	if OrderPoints > ShaoPoints {
		zuhang_nyert = true
	}
	index += 3
	for _, char := range OrderCharacters {
		resp.Insert(utils.IntToBytes(uint64(len(char.Name)), 1, false), index)
		index++
		resp.Insert([]byte(char.Name), index)
		index += len(char.Name)
		resp.Insert(utils.IntToBytes(uint64(char.Faction), 1, false), index)
		index++
		data := utils.IntToBytes(uint64(char.WarContribution), 2, true)
		resp.Insert(data, index)
		index += 2
		resp.Insert([]byte{0x00, 0x00}, index)
		index += 2
		data2 := utils.IntToBytes(uint64(char.WarKillCount), 2, true)
		resp.Insert(data2, index)
		index += 2
		resp.Insert([]byte{0x00, 0x00}, index)
		index += 2
		if zuhang_nyert {
			item := &InventorySlot{ItemID: 99009118, Quantity: uint(1)}
			r, _, err := char.AddItem(item, -1, false)
			if err == nil {
				char.Socket.Write(*r)
			}
		} else {
			item := &InventorySlot{ItemID: 99009117, Quantity: uint(1)}
			r, _, err := char.AddItem(item, -1, false)
			if err == nil {
				char.Socket.Write(*r)
			}
		}
	}
	for _, char := range ShaoCharacters {
		resp.Insert(utils.IntToBytes(uint64(len(char.Name)), 1, false), index)
		index++
		resp.Insert([]byte(char.Name), index)
		index += len(char.Name)
		resp.Insert(utils.IntToBytes(uint64(char.Faction), 1, false), index)
		index++
		data := utils.IntToBytes(uint64(char.WarContribution), 2, true)
		resp.Insert(data, index)
		index += 2
		resp.Insert([]byte{0x00, 0x00}, index)
		index += 2
		data2 := utils.IntToBytes(uint64(char.WarKillCount), 2, true)
		resp.Insert(data2, index)
		index += 2
		resp.Insert([]byte{0x00, 0x00}, index)
		index += 2
		if !zuhang_nyert {
			item := &InventorySlot{ItemID: 99009117, Quantity: uint(1)}
			r, _, err := char.AddItem(item, -1, false)
			if err == nil {
				char.Socket.Write(*r)
			}
		} else {
			item := &InventorySlot{ItemID: 99009118, Quantity: uint(1)}
			r, _, err := char.AddItem(item, -1, false)
			if err == nil {
				char.Socket.Write(*r)
			}
		}
	}
	length := index - 4
	resp.SetLength(int16(length))
	for _, char := range ShaoCharacters {
		if char.IsOnline {
			char.Socket.Write(resp)
		} else {
			delete(OrderCharacters, char.ID)
		}
	}
	for _, char := range OrderCharacters {
		if char.IsOnline {
			char.Socket.Write(resp)
		} else {
			delete(OrderCharacters, char.ID)
		}
	}
}
func CalculateResult(number int) []int {
	remaining := number
	divCount := []int{0, 0, 0, 0}
	divNumbers := []int{1, 16, 256, 4096, 65536, 1048576}
	for i := len(divNumbers) - 1; i >= 0; i-- {
		if remaining < divNumbers[i] || remaining == 0 {
			continue
		}
		test := remaining / divNumbers[i]
		if test > 15 {
			test = 15
		}
		divCount[i] = test
		test2 := test * divNumbers[i]
		remaining -= test2
	}
	return divCount
}

func CalculateWarCountdown(time countdown) int {
	//remaining := time.t
	return time.t
}

func divmod(numerator, denominator int64) (quotient, remainder int64) {
	quotient = numerator / denominator // integer division, decimals are truncated
	remainder = numerator % denominator
	return
}

func getTimeRemaining(t time.Time) countdown {
	currentTime := time.Now()
	difference := t.Sub(currentTime)

	total := int(difference.Seconds())
	days := int(total / (60 * 60 * 24))
	hours := int(total / (60 * 60) % 24)
	minutes := int(total/60) % 60
	seconds := int(total % 60)
	return countdown{
		t: total,
		d: days,
		h: hours,
		m: minutes,
		s: seconds,
	}
}
