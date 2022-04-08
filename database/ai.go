package database

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"sort"
	"sync"
	"time"

	"hero-emulator/nats"
	"hero-emulator/utils"

	"github.com/thoas/go-funk"
)

type Damage struct {
	DealerID int
	Damage   int
}

type AI struct {
	ID           int     `db:"id" json:"id"`
	PosID        int     `db:"pos_id" json:"pos_id"`
	Server       int     `db:"server" json:"server"`
	Faction      int     `db:"faction" json:"faction"`
	Map          int16   `db:"map" json:"map"`
	Coordinate   string  `db:"coordinate" json:"coordinate"`
	WalkingSpeed float64 `db:"walking_speed" json:"walking_speed"`
	RunningSpeed float64 `db:"running_speed" json:"running_speed"`

	DamageDealers  utils.SMap          `db:"-"`
	TargetLocation utils.Location      `db:"-"`
	PseudoID       uint16              `db:"-"`
	CHI            int                 `db:"-" json:"chi"`
	HP             int                 `db:"-" json:"hp"`
	IsDead         bool                `db:"-" json:"is_dead"`
	IsMoving       bool                `db:"-" json:"is_moving"`
	MovementToken  int64               `db:"-" json:"-"`
	OnSightPlayers map[int]interface{} `db:"-" json:"players"`
	PlayersMutex   sync.RWMutex        `db:"-"`
	TargetPlayerID int                 `db:"-" json:"target_player"`
	TargetPetID    int                 `db:"-" json:"target_pet"`
	Handler        func()              `db:"-" json:"-"`
	Once           bool                `db:"-"`
}

var (
	AIs             = make(map[int]*AI)
	DungeonsAiByMap []map[int16][]*AI
	AIsByMap        []map[int16][]*AI
	DungeonsByMap   []map[int16]int
	eventBosses     = []int{50009, 50010}

	MOB_MOVEMENT    = utils.Packet{0xAA, 0x55, 0x21, 0x00, 0x33, 0x00, 0xBC, 0xDB, 0x9F, 0x41, 0x52, 0x70, 0xA2, 0x41, 0x00, 0x55, 0xAA}
	MOB_ATTACK      = utils.Packet{0xAA, 0x55, 0x0C, 0x00, 0x41, 0x01, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x55, 0xAA}
	MOB_SKILL       = utils.Packet{0xAA, 0x55, 0x1B, 0x00, 0x42, 0x0A, 0x00, 0xDF, 0x28, 0xFA, 0xBE, 0x01, 0x01, 0x55, 0xAA}
	MOB_DEAL_DAMAGE = utils.Packet{0xAA, 0x55, 0x28, 0x00, 0x16, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x27, 0x00, 0x00, 0x55, 0xAA}

	ITEM_DROPPED = utils.Packet{0xAA, 0x55, 0x42, 0x00, 0x67, 0x02, 0x01, 0x01, 0x7A, 0xFB, 0x7B, 0xBF, 0x00, 0xA2, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x55, 0xAA}

	STONE_APPEARED = utils.Packet{0xAA, 0x55, 0x57, 0x00, 0x31, 0x01, 0x01, 0x00, 0x00, 0x00, 0x0c, 0x45, 0x6d, 0x70, 0x69, 0x72, 0x65, 0x20, 0x53, 0x74, 0x6f, 0x6e, 0x65, 0x01, 0x01,
		0x8E, 0xE5, 0x38, 0xC0, 0xD9, 0xB8, 0x05, 0xC0, 0x00, 0x00, 0x00, 0x40, 0xFF, 0xFF, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64, 0x00, 0x55, 0xAA}

	MOB_APPEARED = utils.Packet{0xAA, 0x55, 0x54, 0x00, 0x31, 0x01, 0xFF, 0xFF, 0xFF, 0xFF, 0x01, 0x01,
		0x8E, 0xE5, 0x38, 0xC0, 0xD9, 0xB8, 0x05, 0xC0, 0x00, 0x00, 0x00, 0x40, 0xFF, 0xFF, 0x00, 0xFC, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64, 0x00, 0x55, 0xAA}

	DROP_DISAPPEARED = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x67, 0x04, 0x55, 0xAA}

	dropOffsets = []*utils.Location{&utils.Location{0, 0}, &utils.Location{0, 1}, &utils.Location{1, 0}, &utils.Location{1, 1}, &utils.Location{-1, 0},
		&utils.Location{-1, 1}, &utils.Location{-1, -1}, &utils.Location{0, -1}, &utils.Location{1, -1}, &utils.Location{-1, 2}, &utils.Location{0, 2},
		&utils.Location{2, 2}, &utils.Location{2, 1}, &utils.Location{2, 0}, &utils.Location{2, -1}, &utils.Location{2, -2}, &utils.Location{1, -2},
		&utils.Location{0, -2}, &utils.Location{-1, -2}, &utils.Location{-2, -2}, &utils.Location{-2, -1}, &utils.Location{-2, 0}, &utils.Location{-2, 1},
		&utils.Location{-2, 2}, &utils.Location{-2, 3}, &utils.Location{-1, 3}, &utils.Location{0, 3}, &utils.Location{1, 3}, &utils.Location{2, 3},
		&utils.Location{3, 3}, &utils.Location{3, 2}, &utils.Location{3, 1}, &utils.Location{3, 0}, &utils.Location{3, -1}, &utils.Location{3, -2},
		&utils.Location{3, -3}, &utils.Location{2, -3}, &utils.Location{1, -3}, &utils.Location{0, -3}, &utils.Location{-1, -3}, &utils.Location{-2, -3},
		&utils.Location{-3, -3}, &utils.Location{-3, -2}, &utils.Location{-3, -1}, &utils.Location{-3, 0}, &utils.Location{-3, 1}, &utils.Location{-3, 2}, &utils.Location{-3, 3}}
)

func FindAIByID(ID int) *AI {
	return AIs[ID]
}

func (ai *AI) SetCoordinate(coordinate *utils.Location) {
	ai.Coordinate = fmt.Sprintf("(%.1f,%.1f)", coordinate.X, coordinate.Y)
}

func (ai *AI) Create() error {
	return db.Insert(ai)
}

func GetAllAI() error {
	var arr []*AI
	query := `select * from hops.ai order by id`

	if _, err := db.Select(&arr, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("GetAllAI: %s", err.Error())
	}

	for _, a := range arr {
		AIs[a.ID] = a
	}

	return nil
}
func getAllDungeon(mob int) error {
	return nil
}

func (ai *AI) FindTargetCharacterID() (int, error) {

	var (
		distance = 15.0
	)

	if len(characters) == 0 {
		return 0, nil
	}

	npcPos := NPCPos[ai.PosID]
	minCoordinate := ConvertPointToLocation(npcPos.MinLocation)
	maxCoordinate := ConvertPointToLocation(npcPos.MaxLocation)
	aiCoordinate := ConvertPointToLocation(ai.Coordinate)

	characterMutex.RLock()
	allChars := funk.Values(characters)
	characterMutex.RUnlock()
	filtered := funk.Filter(allChars, func(c *Character) bool {

		if c.Socket == nil || !c.IsOnline {
			return false
		}

		user := c.Socket.User
		stat := c.Socket.Stats

		if user == nil || stat == nil {
			return false
		}

		characterCoordinate := ConvertPointToLocation(c.Coordinate)

		seed := utils.RandInt(0, 1000)

		return user.ConnectedServer == ai.Server && c.Map == ai.Map && stat.HP > 0 && !c.Invisible &&
			characterCoordinate.X >= minCoordinate.X && characterCoordinate.X <= maxCoordinate.X &&
			characterCoordinate.Y >= minCoordinate.Y && characterCoordinate.Y <= maxCoordinate.Y &&
			utils.CalculateDistance(characterCoordinate, aiCoordinate) <= distance && seed < 500
	})

	filtered = funk.Shuffle(filtered)
	characters := filtered.([]*Character)

	npc := npcPos.NPCID
	if len(characters) > 0 {

		for _, factionNPC := range ZhuangFactionMobs {
			if factionNPC == npc && characters[0].Faction == 1 {
				return 0, nil
			}
		}
		for _, factionNPC := range ShaoFactionMobs {
			if factionNPC == npc && characters[0].Faction == 2 {
				return 0, nil
			}
		}
		return characters[0].ID, nil
	}

	return 0, nil
}

func (ai *AI) FindTargetPetID(characterID int) (*InventorySlot, error) {

	enemy, err := FindCharacterByID(characterID)
	if err != nil || enemy == nil {
		return nil, err
	}

	slots, err := enemy.InventorySlots()
	if err != nil {
		return nil, err
	}

	pet := slots[0x0A].Pet
	if pet == nil || !pet.IsOnline {
		return nil, nil
	}

	return slots[0x0A], nil
}

func (ai *AI) Move(targetLocation utils.Location, runningMode byte) []byte {

	resp := MOB_MOVEMENT
	currentLocation := ConvertPointToLocation(ai.Coordinate)

	resp.Insert(utils.IntToBytes(uint64(ai.PseudoID), 2, true), 5) // mob pseudo id
	resp[7] = runningMode
	resp.Insert(utils.FloatToBytes(currentLocation.X, 4, true), 8)  // current coordinate-x
	resp.Insert(utils.FloatToBytes(currentLocation.Y, 4, true), 12) // current coordinate-y
	resp.Insert(utils.FloatToBytes(targetLocation.X, 4, true), 20)  // current coordinate-x
	resp.Insert(utils.FloatToBytes(targetLocation.Y, 4, true), 24)  // current coordinate-y

	speeds := []float64{0, ai.WalkingSpeed, ai.RunningSpeed}
	resp.Insert(utils.FloatToBytes(speeds[runningMode], 4, true), 32) // current coordinate-y

	return resp
}

func (ai *AI) Attack() []byte {

	resp := MOB_ATTACK
	character, err := FindCharacterByID(ai.TargetPlayerID)
	if err != nil || character == nil {
		return nil
	}

	pos := NPCPos[ai.PosID]
	if pos == nil {
		return nil
	}

	npc := NPCs[pos.NPCID]
	if npc == nil {
		return nil
	}
	for _, factionNPC := range ZhuangFactionMobs {
		if factionNPC == npc.ID && character.Faction == 1 {
			return nil
		}
	}
	for _, factionNPC := range ShaoFactionMobs {
		if factionNPC == npc.ID && character.Faction == 2 {
			return nil
		}
	}

	stat := character.Socket.Stats

	rawDamage := int(utils.RandInt(int64(npc.MinATK), int64(npc.MaxATK)))
	damage := int(math.Max(float64(rawDamage-stat.DEF), 3))

	reqAcc := float64(stat.Dodge) + float64(character.Level-int(npc.Level))*10
	//probability := reqAcc
	if utils.RandInt(0, 2000) < int64(reqAcc) {
		damage = 0
	}

	resp.Insert(utils.IntToBytes(uint64(ai.PseudoID), 2, true), 6) // mob pseudo id
	//resp.Insert([]byte{0}, 8)
	resp.Insert(utils.IntToBytes(uint64(character.PseudoID), 2, true), 8) // character pseudo id

	resp[11] = 2
	if damage > 0 {
		resp[12] = 1 // damage sound
	}

	resp.Concat(ai.DealDamage(damage))
	return resp
}

func (ai *AI) CastSkill() []byte {

	character, err := FindCharacterByID(ai.TargetPlayerID)
	if err != nil || character == nil {
		return nil
	}

	pos := NPCPos[ai.PosID]
	if pos == nil {
		return nil
	}

	npc := NPCs[pos.NPCID]
	if npc == nil {
		return nil
	}
	for _, factionNPC := range ZhuangFactionMobs {
		if factionNPC == npc.ID && character.Faction == 1 {
			return nil
		}
	}
	for _, factionNPC := range ShaoFactionMobs {
		if factionNPC == npc.ID && character.Faction == 2 {
			return nil
		}
	}

	stat := character.Socket.Stats

	rawDamage := int(utils.RandInt(int64(npc.MinArtsATK), int64(npc.MaxArtsATK)))
	damage := int(math.Max(float64(rawDamage-stat.ArtsDEF), 3))

	reqAcc := float64(stat.Dodge) + float64(character.Level-int(npc.Level))*10
	if utils.RandInt(0, 2000) < int64(reqAcc) {
		damage = 0
	}

	mC := ConvertPointToLocation(ai.Coordinate)

	resp := MOB_SKILL
	resp.Insert(utils.IntToBytes(uint64(ai.PseudoID), 2, true), 7)         // mob pseudo id
	resp.Insert(utils.IntToBytes(uint64(npc.SkillID), 4, true), 9)         // pet skill id
	resp.Insert(utils.FloatToBytes(mC.X, 4, true), 13)                     // pet-x
	resp.Insert(utils.FloatToBytes(mC.Y, 4, true), 17)                     // pet-x
	resp.Insert(utils.IntToBytes(uint64(character.PseudoID), 2, true), 25) // target pseudo id
	resp.Insert(utils.IntToBytes(uint64(character.PseudoID), 2, true), 28) // target pseudo id

	//time.AfterFunc(time.Second, func() {
	resp.Concat(ai.DealDamage(damage))
	//p := nats.CastPacket{CastNear: true, MobID: ai.ID, Data: r, Type: nats.MOB_ATTACK}
	//p.Cast()
	//})

	return resp
}

func (ai *AI) AttackPet() []byte {

	resp := MOB_ATTACK
	pet, ok := GetFromRegister(ai.Server, ai.Map, uint16(ai.TargetPetID)).(*PetSlot)
	if !ok || pet == nil || !pet.IsOnline || pet.HP <= 0 {
		return nil
	}

	pos := NPCPos[ai.PosID]
	if pos == nil {
		return nil
	}

	npc := NPCs[pos.NPCID]
	if npc == nil {
		return nil
	}
	for _, factionNPC := range ZhuangFactionMobs {
		if factionNPC == npc.ID {
			return nil
		}
	}
	for _, factionNPC := range ShaoFactionMobs {
		if factionNPC == npc.ID {
			return nil
		}
	}

	if pet.Target == 0 {
		pet.Target = int(ai.PseudoID)
	}

	rawDamage := int(utils.RandInt(int64(npc.MinATK), int64(npc.MaxATK)))
	damage := int(math.Max(float64(rawDamage-pet.DEF), 3))

	reqAcc := float64(int(pet.Level)-int(npc.Level)) * 10
	if utils.RandInt(0, 1000) < int64(reqAcc) {
		damage = 0
	}

	resp.Insert(utils.IntToBytes(uint64(ai.PseudoID), 2, true), 6) // mob pseudo id
	//resp.Insert([]byte{0}, 8)
	resp.Insert(utils.IntToBytes(uint64(pet.PseudoID), 2, true), 8) // character pseudo id

	resp[11] = 2
	if damage > 0 {
		resp[12] = 1 // damage sound
	}

	resp.Concat(ai.DealDamageToPet(damage))
	return resp
}

func (ai *AI) CastSkillToPet() []byte {

	pet, ok := GetFromRegister(ai.Server, ai.Map, uint16(ai.TargetPetID)).(*PetSlot)
	if !ok || pet == nil || !pet.IsOnline || pet.HP <= 0 {
		return nil
	}

	pos := NPCPos[ai.PosID]
	if pos == nil {
		return nil
	}

	npc := NPCs[pos.NPCID]
	if npc == nil {
		return nil
	}
	for _, factionNPC := range ZhuangFactionMobs {
		if factionNPC == npc.ID {
			return nil
		}
	}
	for _, factionNPC := range ShaoFactionMobs {
		if factionNPC == npc.ID {
			return nil
		}
	}

	rawDamage := int(utils.RandInt(int64(npc.MinArtsATK), int64(npc.MaxArtsATK)))
	damage := int(math.Max(float64(rawDamage-pet.ArtsDEF), 3))

	dodge := float64(pet.STR)
	reqAcc := dodge + float64(int(pet.Level)-int(npc.Level))*10
	if utils.RandInt(0, 1000) < int64(reqAcc) {
		damage = 0
	}

	mC := ConvertPointToLocation(ai.Coordinate)

	resp := MOB_SKILL
	resp.Insert(utils.IntToBytes(uint64(ai.PseudoID), 2, true), 7)   // mob pseudo id
	resp.Insert(utils.IntToBytes(uint64(npc.SkillID), 4, true), 9)   // pet skill id
	resp.Insert(utils.FloatToBytes(mC.X, 4, true), 13)               // pet-x
	resp.Insert(utils.FloatToBytes(mC.Y, 4, true), 17)               // pet-x
	resp.Insert(utils.IntToBytes(uint64(pet.PseudoID), 2, true), 25) // target pseudo id
	resp.Insert(utils.IntToBytes(uint64(pet.PseudoID), 2, true), 28) // target pseudo id

	//time.AfterFunc(time.Second, func() {
	resp.Concat(ai.DealDamageToPet(damage))
	//p := nats.CastPacket{CastNear: true, MobID: ai.ID, Data: r, Type: nats.MOB_ATTACK}
	//p.Cast()
	//})

	return resp
}

func (ai *AI) DealDamage(damage int) []byte {

	resp := MOB_DEAL_DAMAGE
	character, err := FindCharacterByID(ai.TargetPlayerID)
	if err != nil || character == nil {
		return nil
	}

	stat := character.Socket.Stats

	stat.HP = int(math.Max(float64(stat.HP-damage), 0)) // deal damage
	if stat.HP <= 0 {
		ai.TargetPlayerID = 0
	}

	resp.Insert(utils.IntToBytes(uint64(character.PseudoID), 2, true), 5) // character pseudo id
	resp.Insert(utils.IntToBytes(uint64(ai.PseudoID), 2, true), 7)        // mob pseudo id
	resp.Insert(utils.IntToBytes(uint64(stat.HP), 4, true), 9)            // character hp
	resp.Insert(utils.IntToBytes(uint64(stat.CHI), 4, true), 13)          // character chi

	resp.Concat(character.GetHPandChi())
	return resp
}

func (ai *AI) DealDamageToPet(damage int) []byte {

	resp := MOB_DEAL_DAMAGE
	pet, ok := GetFromRegister(ai.Server, ai.Map, uint16(ai.TargetPetID)).(*PetSlot)
	if !ok || pet == nil || !pet.IsOnline || pet.HP <= 0 {
		return nil
	}

	pet.HP = int(math.Max(float64(pet.HP-damage), 0)) // deal damage
	pet.RefreshStats = true
	if pet.HP <= 0 {
		ai.TargetPetID = 0
	}

	resp.Insert(utils.IntToBytes(uint64(pet.PseudoID), 2, true), 5) // pet pseudo id
	resp.Insert(utils.IntToBytes(uint64(ai.PseudoID), 2, true), 7)  // mob pseudo id
	resp.Insert(utils.IntToBytes(uint64(pet.HP), 2, true), 9)       // pet hp
	resp.Insert(utils.IntToBytes(uint64(pet.CHI), 2, true), 11)     // pet chi
	resp.SetLength(0x24)

	return resp
}

func (ai *AI) MovementHandler(token int64, start, end *utils.Location, speed float64) {

	diff := utils.CalculateDistance(start, end)

	if diff < 1 {
		ai.SetCoordinate(end)
		ai.MovementToken = 0
		ai.IsMoving = false
		return
	}

	ai.SetCoordinate(start)
	ai.TargetLocation = *end

	r := []byte{}
	if speed == ai.RunningSpeed {
		r = ai.Move(*end, 2)
	} else {
		r = ai.Move(*end, 1)
	}

	p := nats.CastPacket{CastNear: true, MobID: ai.ID, Data: r, Type: nats.MOB_MOVEMENT}
	p.Cast()

	if diff <= speed { // target is so close
		*start = *end
		time.AfterFunc(time.Duration(diff/speed)*time.Millisecond, func() {
			if token == ai.MovementToken {
				ai.MovementHandler(token, start, end, speed)
			}
		})
	} else { // target is away
		start.X += (end.X - start.X) * speed / diff
		start.Y += (end.Y - start.Y) * speed / diff
		time.AfterFunc(1000*time.Millisecond, func() {
			if token == ai.MovementToken {
				ai.MovementHandler(token, start, end, speed)
			}
		})
	}
}

func (ai *AI) FindClaimer() (*Character, error) {
	dealers := ai.DamageDealers.Values()
	sort.Slice(dealers, func(i, j int) bool {
		di := dealers[i].(*Damage)
		dj := dealers[j].(*Damage)
		return di.Damage > dj.Damage
	})

	if len(dealers) == 0 {
		return nil, nil
	}

	return FindCharacterByID(dealers[0].(*Damage).DealerID)
}

func (ai *AI) DropHandler(claimer *Character) {

	var (
		err error
	)

	npcPos := NPCPos[ai.PosID]
	if npcPos == nil {
		return
	}

	npc := NPCs[npcPos.NPCID]
	if npc == nil {
		return
	}

	isEventBoss := true
	bossMultiplier, dropCount, count, minCount := 0.0, 0, 0, 0
	baseLocation := ConvertPointToLocation(ai.Coordinate)
	if funk.Contains(bosses, npc.ID) {
		bossMultiplier = 2.0
		minCount = 12

	} else if funk.Contains(eventBosses, npc.ID) {
		bossMultiplier = 7.0
		minCount = 48
		isEventBoss = true

	} else if !npcPos.Attackable && claimer.PickaxeActivated() {
		bossMultiplier = 0.4
	}
	if ai.Map == 27 {
		bossMultiplier = 0.1
	}

BEGIN:
	id := npc.DropID

	drop, ok := Drops[id]
	if !ok || drop == nil {
		return
	}

	itemID := 0
	end := false
	for ok {
		index := 0
		seed := int(utils.RandInt(0, 1000))
		items := drop.GetItems()

		probabilities := drop.GetProbabilities()
		totalDropRate := (DROP_RATE * (claimer.DropMultiplier + claimer.AdditionalDropMultiplier)) + bossMultiplier
		if ai.Map == 27 {
			totalDropRate = (1 * (claimer.DropMultiplier + claimer.AdditionalDropMultiplier)) + bossMultiplier
		} else {
			totalDropRate = (DROP_RATE * (claimer.DropMultiplier + claimer.AdditionalDropMultiplier)) + bossMultiplier
		}
		dropFailRate := float64(1000 - probabilities[len(probabilities)-1])
		dropFailRate /= totalDropRate
		newDropFailRate := 1000 - dropFailRate
		probMultiplier := float64(probabilities[len(probabilities)-1]) / newDropFailRate

		if float64(probabilities[len(probabilities)-1])*totalDropRate < 900 {
			probMultiplier = 1
			probabilities = funk.Map(probabilities, func(prob int) int {
				return int(float64(prob) * totalDropRate)
			}).([]int)
		}

		/*
			for _, prob := range probabilities {
				if float64(seed)*probMultiplier > float64(prob) {
					index++
					continue
				}
				break
			}
		*/
		/*
			probabilities = funk.Map(probabilities, func(prob int) int {
				return int(float64(prob) / probMultiplier)
			}).([]int)
		*/

		seed = int(float64(seed) * probMultiplier)
		index = sort.SearchInts(probabilities, seed)
		if index >= len(items) {
			if count >= minCount {
				end = true
				break
			} else {
				drop = Drops[id]
				continue
			}
		}

		itemID = items[index]
		item, exist := Items[int64(itemID)]

		if exist {
			itemType := item.GetType()
			if itemType == QUEST_TYPE || itemType == INGREDIENTS_TYPE {
				drop = Drops[id]
				continue
			}
		}

		drop, ok = Drops[itemID]

		if !ok {
		}
	}

	if itemID > 0 && !end { // can drop an item
		count++
		if count >= 100 {
			return
		}
		go func() {
			resp := utils.Packet{}
			isRelic := false

			if _, ok := Relics[itemID]; ok { // relic drop
				resp.Concat(claimer.RelicDrop(int64(itemID)))
				isRelic = true
			}

			item := Items[int64(itemID)]
			if item != nil {
				seed := int(utils.RandInt(0, 1000))
				plus := byte(0)
				for i := 0; i < len(plusRates) && !isRelic; i++ {
					if seed > plusRates[i] {
						plus++
						continue
					}
					break
				}

				drop := NewSlot()
				drop.ItemID = item.ID
				drop.ItemType = 1
				drop.Quantity = 1
				drop.Plus = plus
				if item.Timer > 0 {
					drop.Quantity = uint(item.Timer)
					drop.ItemType = 1
					drop.Plus = 1
				}
				var upgradesArray []byte
				itemType := item.GetType()
				if itemType == WEAPON_TYPE {
					upgradesArray = WeaponUpgrades
				} else if itemType == ARMOR_TYPE {
					upgradesArray = ArmorUpgrades
				} else if itemType == ACC_TYPE {
					upgradesArray = AccUpgrades
				} else if itemType == PENDENT_TYPE || item.ID == 254 || item.ID == 255 {
					if plus == 0 {
						plus = 1
						drop.Plus = 1
					}
					upgradesArray = []byte{byte(item.ID)}

				} else if itemType == SOCKET_TYPE {
					drop.ItemID = 235
					drop.Plus = socketOrePlus[item.ID]
					plus = socketOrePlus[item.ID]
					upgradesArray = []byte{235}

				} else {
					plus = 0
					drop.Plus = 0
				}

				for i := byte(0); i < plus; i++ {
					index := utils.RandInt(0, int64(len(upgradesArray)))
					drop.SetUpgrade(int(i), upgradesArray[index])
				}

				if isRelic || !npcPos.Attackable {

					slot := int16(-1)
					if npcPos.Attackable {
						slot, err = claimer.FindFreeSlot()
						if slot == 0 || err != nil {
							return
						}
					}

					data, _, err := claimer.AddItem(drop, slot, true)
					if err != nil || data == nil {
						return
					}
					claimer.Socket.Write(*data)

				} else {

					offset := dropOffsets[dropCount%len(dropOffsets)]
					dropCount++

					dr := &Drop{Server: ai.Server, Map: ai.Map, Claimer: claimer, Item: drop,
						Location: utils.Location{X: baseLocation.X + offset.X, Y: baseLocation.Y + offset.Y}}

					if isEventBoss {
						dr.Claimer = nil
					}
					time.AfterFunc(FREEDROP_LIFETIME, func() { //ALL PLAYER CAN PICKUP THE ITEMS
						dr.Claimer = nil
					})

					dr.GenerateIDForDrop(ai.Server, ai.Map)

					dropID := uint16(dr.ID)
					time.AfterFunc(DROP_LIFETIME, func() { // remove drop after timeout
						ai.RemoveDrop(ai.Server, ai.Map, dropID)
					})

					/*r := ITEM_DROPPED
					r.Insert(utils.IntToBytes(uint64(dropID), 2, true), 6) // drop id

					r.Insert(utils.FloatToBytes(offset.X+baseLocation.X, 4, true), 10) // drop coordinate-x
					r.Insert(utils.FloatToBytes(offset.Y+baseLocation.Y, 4, true), 18) // drop coordinate-y

					r.Insert(utils.IntToBytes(uint64(itemID), 4, true), 22) // item id
					if drop.Plus > 0 {
						r[27] = 0xA2
						r.Insert(drop.GetUpgrades(), 32)                                  // item upgrades
						r.Insert([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 47) // item sockets
						r.Insert(utils.IntToBytes(uint64(claimer.PseudoID), 2, true), 66) // claimer id
						r.SetLength(0x42)
					} else {
						r[27] = 0xA1
						r.Insert(utils.IntToBytes(uint64(claimer.PseudoID), 2, true), 36) // claimer id
						r.SetLength(0x24)
					}
					resp.Concat(r)*/
				}

				p := nats.CastPacket{CastNear: true, MobID: ai.ID, Data: resp, Type: nats.ITEM_DROP}
				if isRelic {
					p = nats.CastPacket{CastNear: false, Data: resp, Type: nats.ITEM_DROP}
				} else {
					p = nats.CastPacket{CastNear: true, MobID: ai.ID, Data: resp, Type: nats.BOSS_DROP}
				}

				if err := p.Cast(); err != nil {
					return
				}
			}
		}()
	}

	if !npcPos.Attackable {
		end = true
	}

	if !end {
		goto BEGIN
	}

	return
}

func (ai *AI) RemoveDrop(server int, mapID int16, dropID uint16) {
	drMutex.RLock()
	_, ok := DropRegister[server][mapID][dropID]
	drMutex.RUnlock()

	if ok {
		drMutex.Lock()
		delete(DropRegister[server][mapID], dropID)
		drMutex.Unlock()
	}
}

func (ai *AI) AIHandler() {

	if len(ai.OnSightPlayers) > 0 && ai.HP > 0 {

		timer := fmt.Sprintf("%s", time.Now().String())
		npcPos := NPCPos[ai.PosID]
		npc := NPCs[npcPos.NPCID]

		ai.PlayersMutex.RLock()
		ids := funk.Keys(ai.OnSightPlayers).([]int)
		ai.PlayersMutex.RUnlock()

		for _, id := range ids {
			remove := false

			c, err := FindCharacterByID(id)
			if err != nil || c == nil || !c.IsOnline || c.Map != ai.Map {
				remove = true
			}

			if c != nil {
				user, err := FindUserByID(c.UserID)
				if err != nil || user == nil || user.ConnectedIP == "" || user.ConnectedServer == 0 || user.ConnectedServer != ai.Server {
					remove = true
				}
			}

			if remove {
				ai.PlayersMutex.Lock()
				delete(ai.OnSightPlayers, id)
				ai.PlayersMutex.Unlock()
			}
		}

		if ai.TargetPetID > 0 {
			pet, ok := GetFromRegister(ai.Server, ai.Map, uint16(ai.TargetPetID)).(*PetSlot)
			if !ok || pet == nil || !pet.IsOnline || pet.HP <= 0 {
				ai.TargetPetID = 0
			}
		}

		if ai.TargetPlayerID > 0 {
			c, err := FindCharacterByID(ai.TargetPlayerID)
			if err != nil || c == nil || !c.IsOnline || c.Socket == nil || c.Socket.Stats.HP <= 0 {
				ai.TargetPlayerID = 0
				//ai.HP = npc.MaxHp
			} else {
				slots, _ := c.InventorySlots()
				petSlot := slots[0x0A]
				pet := petSlot.Pet
				petInfo, ok := Pets[petSlot.ItemID]
				if pet != nil && ok && pet.IsOnline && !petInfo.Combat {
					ai.TargetPlayerID = 0
					ai.TargetPetID = petSlot.Pet.PseudoID
				}
			}
		}

		var err error
		if ai.TargetPetID == 0 && ai.TargetPlayerID == 0 { // gotta find a target

			ai.TargetPlayerID, err = ai.FindTargetCharacterID() // 50% chance to trigger
			if err != nil {
				log.Println("AIHandler FindTargetPlayer error:", err)
			}

			petSlot, err := ai.FindTargetPetID(ai.TargetPlayerID)
			if err != nil {
				log.Println("AIHandler FindTargetPet error:", err)
			}

			if petSlot != nil {
				pet := petSlot.Pet
				petInfo, ok := Pets[petSlot.ItemID]

				seed := utils.RandInt(0, 1000)
				if pet != nil && ok && (seed > 500 || !petInfo.Combat) {
					ai.TargetPlayerID = 0
					ai.TargetPetID = pet.PseudoID
				}
			}
		}

		if ai.TargetPlayerID > 0 || ai.TargetPetID > 0 {
			ai.IsMoving = false
		}

		if ai.IsMoving {
			goto OUT
		}

		if ai.TargetPlayerID == 0 && ai.TargetPetID == 0 { // Idle mode
			coordinate := ConvertPointToLocation(ai.Coordinate)
			minCoordinate := ConvertPointToLocation(npcPos.MinLocation)
			maxCoordinate := ConvertPointToLocation(npcPos.MaxLocation)

			if utils.RandInt(0, 1000) < 750 { // 75% chance to move
				ai.IsMoving = true

				targetX := utils.RandFloat(minCoordinate.X, maxCoordinate.X)
				targetY := utils.RandFloat(minCoordinate.Y, maxCoordinate.Y)
				target := utils.Location{X: targetX, Y: targetY}
				ai.TargetLocation = target

				//d := ai.Move(target, 1)
				//p := nats.CastPacket{CastNear: true, MobID: ai.ID, Data: d, Type: nats.MOB_MOVEMENT}
				//p.Cast()

				token := ai.MovementToken
				for token == ai.MovementToken {
					ai.MovementToken = utils.RandInt(1, math.MaxInt64)
				}

				go ai.MovementHandler(ai.MovementToken, coordinate, &target, ai.WalkingSpeed)

			}

		} else if ai.TargetPetID > 0 {
			pet, ok := GetFromRegister(ai.Server, ai.Map, uint16(ai.TargetPetID)).(*PetSlot)
			if !ok || pet == nil || !pet.IsOnline || pet.HP <= 0 {
				ai.TargetPetID = 0
				goto OUT
			}

			aiCoordinate := ConvertPointToLocation(ai.Coordinate)
			distance := utils.CalculateDistance(&pet.Coordinate, aiCoordinate)

			if ai.ShouldGoBack() || distance > 50 { // better to retreat
				ai.TargetPlayerID = 0
				ai.MovementToken = 0
				ai.IsMoving = false
				//ai.HP = npc.MaxHp
			} else if distance <= 3 && pet.IsOnline && pet.HP > 0 { // attack
				seed := utils.RandInt(1, 1000)
				_, ok := SkillInfos[npc.SkillID]

				r := utils.Packet{}
				if seed < 400 && ok {
					r.Concat(ai.CastSkillToPet())
				} else {
					r.Concat(ai.AttackPet())
				}

				p := nats.CastPacket{CastNear: true, MobID: ai.ID, Data: r, Type: nats.MOB_ATTACK}
				p.Cast()

			} else if distance > 3 && distance <= 50 { // chase
				ai.IsMoving = true
				target := GeneratePoint(&pet.Coordinate)
				ai.TargetLocation = target

				token := ai.MovementToken
				for token == ai.MovementToken {
					ai.MovementToken = utils.RandInt(1, math.MaxInt64)
				}

				go ai.MovementHandler(ai.MovementToken, aiCoordinate, &target, ai.RunningSpeed)

			}

		} else if ai.TargetPlayerID > 0 { // Target mode player
			character, err := FindCharacterByID(ai.TargetPlayerID)
			if err != nil || character == nil || (character != nil && (!character.IsOnline || character.Invisible)) {
				//ai.HP = npc.MaxHp
				ai.TargetPlayerID = 0
				goto OUT
			}

			stat := character.Socket.Stats

			characterCoordinate := ConvertPointToLocation(character.Coordinate)
			aiCoordinate := ConvertPointToLocation(ai.Coordinate)
			distance := utils.CalculateDistance(characterCoordinate, aiCoordinate)

			if ai.ShouldGoBack() || distance > 50 { // better to retreat
				ai.TargetPlayerID = 0
				ai.MovementToken = 0
				ai.IsMoving = false
				//ai.HP = npc.MaxHp
			} else if distance <= 5 && character.IsActive && stat.HP > 0 { // attack
				seed := utils.RandInt(1, 1000)
				_, ok := SkillInfos[npc.SkillID]

				r := utils.Packet{}
				if seed < 400 && ok {
					r.Concat(ai.CastSkill())
				} else {
					r.Concat(ai.Attack())
				}

				p := nats.CastPacket{CastNear: true, MobID: ai.ID, Data: r, Type: nats.MOB_ATTACK}
				p.Cast()

			} else if distance > 5 && distance <= 50 { // chase
				ai.IsMoving = true
				target := GeneratePoint(characterCoordinate)
				ai.TargetLocation = target

				token := ai.MovementToken
				for token == ai.MovementToken {
					ai.MovementToken = utils.RandInt(1, math.MaxInt64)
				}

				go ai.MovementHandler(ai.MovementToken, aiCoordinate, &target, ai.RunningSpeed)

			}
		}

		timer = fmt.Sprintf("%s -> %s", timer, time.Now().String())
		//log.Println(timer)
	}

OUT:
	delay := utils.RandFloat(1.0, 1.5) * 1000
	time.AfterFunc(time.Duration(delay)*time.Millisecond, func() {
		ai.AIHandler()
	})
}

func (ai *AI) ShouldGoBack() bool {

	npcPos := NPCPos[ai.PosID]
	aiMinCoordinate := ConvertPointToLocation(npcPos.MinLocation)
	aiMaxCoordinate := ConvertPointToLocation(npcPos.MaxLocation)
	aiCoordinate := ConvertPointToLocation(ai.Coordinate)

	if aiCoordinate.X >= aiMinCoordinate.X && aiCoordinate.X <= aiMaxCoordinate.X &&
		aiCoordinate.Y >= aiMinCoordinate.Y && aiCoordinate.Y <= aiMaxCoordinate.Y {
		return false
	}

	return true
}

func GeneratePoint(location *utils.Location) utils.Location {

	r := 2.0
	alfa := utils.RandFloat(0, 360)
	targetX := location.X + r*float64(math.Cos(alfa*math.Pi/180))
	targetY := location.Y + r*float64(math.Sin(alfa*math.Pi/180))

	return utils.Location{X: targetX, Y: targetY}
}
