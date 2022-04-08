package dungeon

import (
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"time"

	"hero-emulator/database"
	"hero-emulator/messaging"
	"hero-emulator/utils"
)

var (
	TIMER_MENU = utils.Packet{0xAA, 0x55, 0x08, 0x00, 0x65, 0x03, 0x00, 0x00, 0x00, 0x55, 0xAA}
)

func StartTimer(char *database.Socket) {
	timein := time.Now().Add(time.Minute * 30)
	deadtime := timein.Format(time.RFC3339)

	v, err := time.Parse(time.RFC3339, deadtime)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	for range time.Tick(1 * time.Second) {
		//fmt.Println("CharName: ", char.Character.Name)
		timeRemaining := getTimeRemaining(v)
		if char.Character.DungeonLevel == 3 {
			break
		}
		if timeRemaining.t <= 0 || char.Character.Map != 229 || char.Character.IsOnline == false {
			resp := utils.Packet{}
			resp.Concat(messaging.InfoMessage(fmt.Sprintf("You have failed. Come again when you are stronger. Teleporting to safe zone.")))
			DeleteMobs(char.User.ConnectedServer)
			char.Character.Socket.User.ConnectedServer = 1
			data, _ := char.Character.ChangeMap(1, nil)
			resp.Concat(data)
			char.Conn.Write(resp)
			break
		}
		//index := 6
		//resp := TIMER_MENU
		timeCounters := CalculateCountdown(timeRemaining)
		var sb strings.Builder
		timer1 := fmt.Sprintf("%x", timeCounters[1])
		timer0 := fmt.Sprintf("%x", timeCounters[0])
		timer3 := fmt.Sprintf("%x", timeCounters[3])
		timer2 := fmt.Sprintf("%x", timeCounters[2])
		sb.WriteString("AA550800650300" + timer1 + timer0 + timer3 + timer2 + "00000055AA")
		data, err := hex.DecodeString(sb.String())
		if err != nil {
			panic(err)
		}
		char.Conn.Write(data)
		//fmt.Printf("Minutes: %d Seconds: %d\n", timeRemaining.m, timeRemaining.s)
	}
}

type countdown struct {
	t int
	d int
	h int
	m int
	s int
}

func CalculateCountdown(time countdown) []int {
	remaining := time.t
	divCount := []int{0, 0, 0, 0}
	divNumbers := []int{1, 16, 256, 4096}
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
