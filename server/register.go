package server

import (
	"fmt"
	"log"
	"sync"
	"time"

	"hero-emulator/database"
	"hero-emulator/utils"
)

var (
	MapRegister    = make([]map[int16]map[uint16]interface{}, database.SERVER_COUNT+1)
	mrMutex        sync.RWMutex
	PlayerRegister = make(map[uint16]interface{}, database.SERVER_COUNT+1)
	prMutex        sync.RWMutex
	Init           = make(chan bool, 1)
)

func init() {

	for j := 0; j <= database.SERVER_COUNT; j++ {
		MapRegister[j] = make(map[int16]map[uint16]interface{})
	}

	for i := int16(1); i <= 255; i++ {
		for j := 0; j <= database.SERVER_COUNT; j++ {
			MapRegister[j][i] = make(map[uint16]interface{})
		}
	}

	database.RemoveFromRegister = func(c *database.Character) {
		pseudo := c.PseudoID
		c.PseudoID = 0

		time.AfterFunc(time.Second*10, func() {
			prMutex.Lock()
			defer prMutex.Unlock()
			delete(PlayerRegister, pseudo)
		})
	}

	database.RemovePetFromRegister = func(c *database.Character) {
		user, err := database.FindUserByID(c.UserID)
		if err != nil || user == nil {
			log.Println("RemoveFromRegister failed:", err)
			return
		}

		slots, err := c.InventorySlots()
		if err != nil {
			return
		}

		pet := slots[0x0A].Pet
		if pet == nil || pet.PseudoID == 0 {
			return
		}

		mrMutex.Lock()
		defer mrMutex.Unlock()
		delete(MapRegister[user.ConnectedServer][c.Map], uint16(pet.PseudoID))
		pet.PseudoID = 0
	}

	database.GetFromRegister = func(server int, mapID int16, ID uint16) interface{} {
		mrMutex.RLock()
		defer mrMutex.RUnlock()
		return MapRegister[server][mapID][ID]
	}

	database.GenerateID = GenerateID
	database.FindCharacterByPseudoID = FindCharacter
	database.GeneratePetID = GenerateIDForPet

	Init <- true
}

func GenerateID(character *database.Character) error {

	prMutex.Lock()
	defer prMutex.Unlock()
	for i := uint16(1); i <= 2000; i++ {
		if _, ok := PlayerRegister[i]; !ok {
			log.Println(i)
			PlayerRegister[i] = character
			character.PseudoID = i
			return nil
		}
	}

	return fmt.Errorf("all pseudo ids taken")
}

func GenerateIDForAI(AI *database.AI) {
	mrMutex.Lock()
	defer mrMutex.Unlock()
	for {
		i := uint16(utils.RandInt(40000, 50000))
		if _, ok := MapRegister[AI.Server][AI.Map][i]; !ok {
			AI.PseudoID = i
			MapRegister[AI.Server][AI.Map][i] = AI
			return
		}
	}
}

func GenerateIDForPet(owner *database.Character, pet *database.PetSlot) {
	mrMutex.Lock()
	defer mrMutex.Unlock()

	server := owner.Socket.User.ConnectedServer
	for i := uint16(2500); i <= 3500; i++ {
		if _, ok := MapRegister[server][owner.Map][i]; !ok {
			pet.PseudoID = int(i)
			MapRegister[server][owner.Map][i] = pet
			return
		}
	}
}

func GenerateIDForNPC(NPCPos *database.NpcPosition) {
	mrMutex.Lock()
	defer mrMutex.Unlock()
	c := 1
	//for c := 1; c <= database.SERVER_COUNT; c++ {
	for {
		i := uint16(utils.RandInt(20000, 30000))
		if _, ok := MapRegister[c][NPCPos.MapID][i]; !ok {
			//NPCPos.PseudoID = uint16(NPCPos.NPCID)
			NPCPos.PseudoID = i
			MapRegister[c][NPCPos.MapID][NPCPos.PseudoID] = NPCPos
			break
		}
	}
	//}
}

func FindCharacter(server int, ID uint16) *database.Character {
	prMutex.RLock()
	defer prMutex.RUnlock()
	if c, ok := PlayerRegister[ID].(*database.Character); ok {
		return c
	}
	return nil
}
