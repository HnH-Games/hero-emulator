package ai

import (
	"fmt"
	"log"

	"hero-emulator/database"
	"hero-emulator/server"
	"hero-emulator/utils"
)

func createMobs() {
	posID := []int{5346, 5347, 5348, 5349, 5350}
	for _, action := range posID {
		npcPos := database.NPCPos[int(action)]
		npc, ok := database.NPCs[npcPos.NPCID]
		if !ok {
			fmt.Println("Error")
		}
		for i := 0; i < int(npcPos.Count); i++ {
			if npc.ID == 0 {
				continue
			}

			newai := &database.AI{ID: len(database.AIs), HP: npc.MaxHp, Map: npcPos.MapID, PosID: npcPos.ID, RunningSpeed: 10, Server: 98, WalkingSpeed: 5, Once: true}
			server.GenerateIDForAI(newai)
			newai.OnSightPlayers = make(map[int]interface{})

			minLoc := database.ConvertPointToLocation(npcPos.MinLocation)
			maxLoc := database.ConvertPointToLocation(npcPos.MaxLocation)
			loc := utils.Location{X: utils.RandFloat(minLoc.X, maxLoc.X), Y: utils.RandFloat(minLoc.Y, maxLoc.Y)}
			newai.Coordinate = loc.String()
			fmt.Println(newai.Coordinate)
			newai.Handler = newai.AIHandler
			database.AIsByMap[newai.Server][npcPos.MapID] = append(database.AIsByMap[newai.Server][npcPos.MapID], newai)
			database.AIs[newai.ID] = newai
			fmt.Println("New mob created", len(database.AIs))
			newai.Create()
			go newai.Handler()
		}
	}
	fmt.Println("Finished")
}
func Init() {
	database.AIsByMap = make([]map[int16][]*database.AI, database.SERVER_COUNT+1)
	for s := 0; s <= database.SERVER_COUNT; s++ {
		database.AIsByMap[s] = make(map[int16][]*database.AI)
	}
	database.DungeonsByMap = make([]map[int16]int, database.SERVER_COUNT+1)
	for s := 0; s <= database.SERVER_COUNT; s++ {
		database.DungeonsByMap[s] = make(map[int16]int)
	}
	database.DungeonsAiByMap = make([]map[int16][]*database.AI, database.SERVER_COUNT+1)
	for s := 0; s <= database.SERVER_COUNT; s++ {
		database.DungeonsAiByMap[s] = make(map[int16][]*database.AI)
	}
	func() {
		<-server.Init

		var err error

		database.NPCPos, err = database.GetAllNPCPos()
		if err != nil {
			log.Println(err)
			return
		}

		for _, pos := range database.NPCPos {
			if pos.IsNPC && !pos.Attackable {
				server.GenerateIDForNPC(pos)
			}
		}

		database.NPCs, err = database.GetAllNPCs()
		if err != nil {
			log.Println(err)
			return
		}
		//createMobs()
		err = database.GetAllAI()
		if err != nil {
			log.Println(err)
			return
		}

		for _, ai := range database.AIs {
			database.AIsByMap[ai.Server][ai.Map] = append(database.AIsByMap[ai.Server][ai.Map], ai)
		}

		for _, AI := range database.AIs {
			if AI.ID == 0 {
				continue
			}
			pos := database.NPCPos[AI.PosID]
			npc := database.NPCs[pos.NPCID]

			AI.TargetLocation = *database.ConvertPointToLocation(AI.Coordinate)
			AI.HP = npc.MaxHp
			AI.OnSightPlayers = make(map[int]interface{})
			AI.Handler = AI.AIHandler

			/*if npc.Level > 200 {
				continue
			}*/

			server.GenerateIDForAI(AI)
			if AI.ID == 55281 || AI.ID == 55283 || AI.ID == 55287 || AI.ID == 55289 || AI.ID == 55285 {
				log.Println(fmt.Sprintf("Pseudo: %d", AI.PseudoID))
				newStone := &database.WarStone{PseudoID: AI.PseudoID, NpcID: pos.NPCID, NearbyZuhang: 0, NearbyShao: 0, ConquereValue: 100}
				database.WarStonesIDs = append(database.WarStonesIDs, AI.PseudoID)
				database.WarStones[int(AI.PseudoID)] = newStone
			}
			if AI.WalkingSpeed > 0 {
				go AI.Handler()
			}
		}
	}()
}
