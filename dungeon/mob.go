package dungeon

import (
	"fmt"
	"math/rand"

	"hero-emulator/database"
	"hero-emulator/messaging"
	"hero-emulator/server"
	"hero-emulator/utils"
)

func FindEmptyServer() int {
	for i := 1; i < database.SERVER_COUNT; i++ {
		allMobs := database.DungeonsByMap[i][229]
		if allMobs == 0 {
			return i
		}
	}
	return 0
}

func StartDungeon(char *database.Socket) {
	server := FindEmptyServer()
	if server != 0 {
		char.Character.IsDungeon = true
		char.Character.DungeonLevel = 1
		char.Character.GeneratedNumber = 0
		char.Character.CanTip = 1
		char.Character.Socket.User.ConnectedServer = server
		data, _ := char.Character.ChangeMap(229, nil)
		char.Conn.Write(data)
		char.Conn.Write(messaging.InfoMessage(fmt.Sprintf("Welcome to Pecetek's Dungeon. You have 30 minutes, Survive & Slay the Monsters.")))
		Createmob(server)
		StartTimer(char)
	}
}

func DeleteMobs(server int) {
	allMobs := database.AIsByMap[server][229]
	//database.AIsByMap[server][229] = nil
	for _, mobs := range allMobs {
		//delete(database.AIs, mobs.ID)
		//database.AIs[mobs.ID] = nil
		mobs.IsDead = true
		mobs.Handler = nil
		delete(database.DungeonsAiByMap[server], mobs.Map)
		DungeonCount := database.DungeonsByMap[server][229] - 1
		database.DungeonsByMap[server][229] -= DungeonCount
	}
}

func CreateMobsToNcash(aiMapID int16) {
	NPCsSpawnPoint := []string{"313,343", "419,377", "393,143", "233,155", "129,69", "67,181"}
	//NPCsTest := []int{40761, 40532, 40597, 40729, 40821, 40541, 40522}
	NPCsTest := []int{1000016}
	for _, action := range NPCsSpawnPoint {
		for i := 0; i < int(30); i++ {
			npcPos := &database.NpcPosition{ID: len(database.NPCPos), NPCID: int(NPCsTest[0]), MapID: aiMapID, Rotation: 0, Attackable: true, IsNPC: false, RespawnTime: 30, Count: 30, MinLocation: "120,120", MaxLocation: "150,150"}
			database.NPCPos = append(database.NPCPos, npcPos)
			npc, _ := database.NPCs[NPCsTest[0]]
			newai := &database.AI{ID: len(database.AIs), HP: npc.MaxHp, Map: aiMapID, PosID: npcPos.ID, RunningSpeed: 10, Server: 100, WalkingSpeed: 5, Once: false}
			newai.OnSightPlayers = make(map[int]interface{})
			coordinate := database.ConvertPointToLocation(action)
			randomLocX := randFloats(coordinate.X, coordinate.X+30)
			randomLocY := randFloats(coordinate.Y, coordinate.Y+30)
			loc := utils.Location{X: randomLocX, Y: randomLocY}
			npcPos.MinLocation = fmt.Sprintf("%.1f,%.1f", randomLocX, randomLocY)
			maxX := randomLocX + 50
			maxY := randomLocY + 50
			npcPos.MaxLocation = fmt.Sprintf("%.1f,%.1f", maxX, maxY)
			newai.Coordinate = loc.String()
			fmt.Println(newai.Coordinate)
			newai.Handler = newai.AIHandler
			database.AIsByMap[newai.Server][newai.Map] = append(database.AIsByMap[newai.Server][newai.Map], newai)
			database.AIs[newai.ID] = newai
			database.DungeonsAiByMap[newai.Server][newai.Map] = append(database.AIsByMap[newai.Server][newai.Map], newai)
			DungeonCount := database.DungeonsByMap[newai.Server][newai.Map] + 1
			fmt.Println("Mobs Count: ", DungeonCount)
			database.DungeonsByMap[newai.Server][newai.Map] = DungeonCount
			server.GenerateIDForAI(newai)
			//ai.Init()
			if newai.WalkingSpeed > 0 {
				go newai.Handler()
			}
		}
	}
}

func Createmob(serverID int) {
	NPCsSpawnPoint := []string{"97,339", "343,89", "211,235", "283,319", "393,383", "347,123", "413,365"}
	//NPCsTest := []int{40761, 40532, 40597, 40729, 40821, 40541, 40522}
	NPCsTest := []int{493001, 493002, 23741, 41758}
	for _, action := range NPCsTest {
		for i := 0; i < int(20); i++ {
			randomInt := rand.Intn(len(NPCsSpawnPoint))
			npcPos := &database.NpcPosition{ID: len(database.NPCPos), NPCID: int(action), MapID: 229, Rotation: 0, Attackable: true, IsNPC: false, RespawnTime: 30, Count: 30, MinLocation: "120,120", MaxLocation: "150,150"}
			database.NPCPos = append(database.NPCPos, npcPos)
			npc, _ := database.NPCs[action]
			newai := &database.AI{ID: len(database.AIs), HP: npc.MaxHp, Map: 229, PosID: npcPos.ID, RunningSpeed: 10, Server: serverID, WalkingSpeed: 5, Once: true}
			newai.OnSightPlayers = make(map[int]interface{})
			coordinate := database.ConvertPointToLocation(NPCsSpawnPoint[randomInt])
			randomLocX := randFloats(coordinate.X, coordinate.X+30)
			randomLocY := randFloats(coordinate.Y, coordinate.Y+30)
			loc := utils.Location{X: randomLocX, Y: randomLocY}
			npcPos.MinLocation = fmt.Sprintf("%.1f,%.1f", randomLocX, randomLocY)
			maxX := randomLocX + 50
			maxY := randomLocY + 50
			npcPos.MaxLocation = fmt.Sprintf("%.1f,%.1f", maxX, maxY)
			newai.Coordinate = loc.String()
			fmt.Println(newai.Coordinate)
			newai.Handler = newai.AIHandler
			database.AIsByMap[newai.Server][newai.Map] = append(database.AIsByMap[newai.Server][newai.Map], newai)
			database.DungeonsAiByMap[newai.Server][newai.Map] = append(database.AIsByMap[newai.Server][newai.Map], newai)
			database.AIs[newai.ID] = newai
			DungeonCount := database.DungeonsByMap[newai.Server][newai.Map] + 1
			database.DungeonsByMap[newai.Server][newai.Map] = DungeonCount
			fmt.Println("Mobs Count: ", DungeonCount)
			server.GenerateIDForAI(newai)
			//ai.Init()
			if newai.WalkingSpeed > 0 {
				go newai.Handler()
			}
		}
	}
}

func randFloats(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func BossSpawn(mobInt int, serverID int) {
	NPCsSpawnPoint := "211,235"
	npcPos := &database.NpcPosition{ID: len(database.NPCPos), NPCID: int(mobInt), MapID: 229, Rotation: 0, Attackable: true, IsNPC: false, RespawnTime: 30, Count: 30, MinLocation: "120,120", MaxLocation: "150,150"}
	database.NPCPos = append(database.NPCPos, npcPos)
	npc, _ := database.NPCs[mobInt]
	newai := &database.AI{ID: len(database.AIs), HP: npc.MaxHp, Map: 229, PosID: npcPos.ID, RunningSpeed: 10, Server: serverID, WalkingSpeed: 5, Once: true}
	newai.OnSightPlayers = make(map[int]interface{})
	coordinate := database.ConvertPointToLocation(NPCsSpawnPoint)
	randomLocX := randFloats(coordinate.X, coordinate.X+30)
	randomLocY := randFloats(coordinate.Y, coordinate.Y+30)
	loc := utils.Location{X: randomLocX, Y: randomLocY}
	npcPos.MinLocation = fmt.Sprintf("%.1f,%.1f", randomLocX, randomLocY)
	maxX := randomLocX + 50
	maxY := randomLocY + 50
	npcPos.MaxLocation = fmt.Sprintf("%.1f,%.1f", maxX, maxY)
	newai.Coordinate = loc.String()
	fmt.Println(newai.Coordinate)
	newai.Handler = newai.AIHandler

	database.AIsByMap[newai.Server][newai.Map] = append(database.AIsByMap[newai.Server][newai.Map], newai)
	database.DungeonsAiByMap[newai.Server][newai.Map] = append(database.AIsByMap[newai.Server][newai.Map], newai)
	database.AIs[newai.ID] = newai
	DungeonCount := database.DungeonsByMap[newai.Server][newai.Map] + 1
	database.DungeonsByMap[newai.Server][newai.Map] = DungeonCount
	server.GenerateIDForAI(newai)
	//ai.Init()
	if newai.WalkingSpeed > 0 {
		go newai.Handler()
	}
}

func FindTheNumber(s *database.Socket) {
	s.Character.CanTip = 2
	s.Conn.Write(messaging.InfoMessage(fmt.Sprintf("Congratulations! Now guess the Boss's Favourite Number [1 - 10] Type:/number [no]")))
	if s.Character.GeneratedNumber == 0 {
		min := 1
		max := 10
		s.Character.GeneratedNumber = rand.Intn(max-min) + min
		fmt.Println("Gondolt szam: ", s.Character.GeneratedNumber)
	}

}

func MobsCreate(mobsID []int, serverID int) {
	NPCsSpawnPoint := []string{"97,339", "343,89", "211,235", "283,319", "393,383", "347,123", "413,365"}
	for _, action := range mobsID {
		randomInt := rand.Intn(len(NPCsSpawnPoint))
		for i := 0; i < int(20); i++ {
			npcPos := &database.NpcPosition{ID: len(database.NPCPos), NPCID: int(action), MapID: 229, Rotation: 0, Attackable: true, IsNPC: false, RespawnTime: 30, Count: 30, MinLocation: "120,120", MaxLocation: "150,150"}
			database.NPCPos = append(database.NPCPos, npcPos)
			npc, _ := database.NPCs[action]
			newai := &database.AI{ID: len(database.AIs), HP: npc.MaxHp, Map: 229, PosID: npcPos.ID, RunningSpeed: 10, Server: serverID, WalkingSpeed: 5, Once: true}
			newai.OnSightPlayers = make(map[int]interface{})
			coordinate := database.ConvertPointToLocation(NPCsSpawnPoint[randomInt])
			randomLocX := randFloats(coordinate.X, coordinate.X+30)
			randomLocY := randFloats(coordinate.Y, coordinate.Y+30)
			loc := utils.Location{X: randomLocX, Y: randomLocY}
			npcPos.MinLocation = fmt.Sprintf("%.1f,%.1f", randomLocX, randomLocY)
			maxX := randomLocX + 50
			maxY := randomLocY + 50
			npcPos.MaxLocation = fmt.Sprintf("%.1f,%.1f", maxX, maxY)
			newai.Coordinate = loc.String()
			fmt.Println(newai.Coordinate)
			newai.Handler = newai.AIHandler

			database.AIsByMap[newai.Server][newai.Map] = append(database.AIsByMap[newai.Server][newai.Map], newai)
			database.AIs[newai.ID] = newai
			DungeonCount := database.DungeonsByMap[newai.Server][newai.Map] + 1
			database.DungeonsByMap[newai.Server][newai.Map] = DungeonCount
			server.GenerateIDForAI(newai)
			if newai.WalkingSpeed > 0 {
				go newai.Handler()
			}
		}
	}
}

func ExploreDungeons() {
	characters, _ := database.FindOnlineCharacters()
	havePlayer := false
	for _, c := range characters {
		if c.Map == 120 && database.DungeonsByMap[100][120] == 0 {
			havePlayer = true
			CreateMobsToNcash(120)
			fmt.Println("Mobok spawnolva")
		} else if havePlayer == false && database.DungeonsByMap[100][120] != 0 && c.Map != 120 {
			database.DungeonsAiByMap[100][120] = nil
			database.AIsByMap[100][120] = nil

			database.DungeonsByMap[100][120] = 0
		}
	}
}
