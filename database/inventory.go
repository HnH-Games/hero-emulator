package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"hero-emulator/nats"
	"hero-emulator/utils"

	"github.com/thoas/go-funk"
	"gopkg.in/guregu/null.v3"
)

var (
	InventoryItems utils.SMap
)

type InventorySlot struct {
	ID            int             `db:"id"`
	UserID        null.String     `db:"user_id"`
	CharacterID   null.Int        `db:"character_id"`
	ItemID        int64           `db:"item_id"`
	SlotID        int16           `db:"slot_id"`
	Quantity      uint            `db:"quantity"`
	Plus          uint8           `db:"plus"`
	UpgradeArr    string          `db:"upgrades"`
	SocketCount   int8            `db:"socket_count"`
	SocketArr     string          `db:"sockets"`
	Activated     bool            `db:"activated"`
	InUse         bool            `db:"in_use"`
	PetInfo       json.RawMessage `db:"pet_info"`
	UpdatedAt     null.Time       `db:"updated_at"`
	Consignment   bool            `db:"consignment"`
	Appearance    int64           `db:"appearance"`
	ItemType      int16           `db:"item_type"`
	JudgementStat int64           `db:"judgement_stat"`
	Pet           *PetSlot        `db:"-" json:"-"`
	RFU           interface{}     `db:"-" json:"-"`
}

type PetSlot struct {
	Name     string `db:"name" json:"name"`
	Level    byte   `db:"level" json:"level"`
	Loyalty  byte   `db:"loyalty" json:"loyalty"`
	Fullness byte   `db:"fullness" json:"fullness"`
	HP       int    `db:"hp" json:"hp"`
	MaxHP    int    `db:"max_hp" json:"max_hp"`
	CHI      int    `db:"chi" json:"chi"`
	MaxCHI   int    `db:"max_chi" json:"max_chi"`
	Exp      uint64 `db:"exp" json:"exp"`
	STR      int    `db:"str" json:"str"`
	DEX      int    `db:"dex" json:"dex"`
	INT      int    `db:"int" json:"int"`
	MinATK   int    `db:"min_atk" json:"min_atk"`
	MaxATK   int    `db:"max_atk" json:"max_atk"`
	DEF      int    `db:"def" json:"def"`
	ArtsDEF  int    `db:"arts_def" json:"arts_def"`

	Casting        bool           `db:"-" json:"-"`
	Coordinate     utils.Location `db:"-" json:"-"`
	IsOnline       bool           `db:"-" json:"-"`
	IsMoving       bool           `db:"-" json:"-"`
	LastHit        int            `db:"-" json:"-"`
	MovementToken  int64          `db:"-" json:"-"`
	PseudoID       int            `db:"-" json:"-"`
	RefreshStats   bool           `db:"-" json:"-"`
	PetCombatMode  int16          `db:"-" json:"-"`
	Target         int            `db:"-" json:"-"`
	TargetLocation utils.Location `db:"-" json:"-"`
}

var (
	DropRegister = make([]map[int16]map[uint16]*Drop, SERVER_COUNT+1)
	drMutex      sync.RWMutex

	ITEM_SLOT = utils.Packet{0xAA, 0x55, 0x2E, 0x00, 0x57, 0x0A, 0x00, 0xA1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	ITEM_UPGRADED   = utils.Packet{0xAA, 0x55, 0x31, 0x00, 0x54, 0x02, 0xA1, 0x0F, 0x01, 0x00, 0xA2, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	SOCKET_OPENED   = utils.Packet{0xAA, 0x55, 0x30, 0x00, 0x54, 0x16, 0x0A, 0x00, 0x00, 0xA2, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	SOCKET_UPGRADED = utils.Packet{0xAA, 0x55, 0x30, 0x00, 0x54, 0x16, 0x0A, 0x00, 0x00, 0xA2, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	PET_STATS       = utils.Packet{0xAA, 0x55, 0x12, 0x00, 0x51, 0x08, 0x0A, 0x00, 0x55, 0xAA}
	SHOW_PET_BUTTON = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x57, 0x05, 0x0A, 0x00, 0x55, 0xAA}
	DISMISS_PET     = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x51, 0x02, 0x0A, 0x00, 0x55, 0xAA}
)

func init() {

	for j := 0; j <= SERVER_COUNT; j++ {
		DropRegister[j] = make(map[int16]map[uint16]*Drop)
	}

	for i := int16(1); i <= 255; i++ {
		for j := 0; j <= SERVER_COUNT; j++ {
			DropRegister[j][i] = make(map[uint16]*Drop)
		}
	}
}

func NewSlot() *InventorySlot {
	return &InventorySlot{UpgradeArr: "{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}", SocketArr: "{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}"}
}

func FindInventorySlotsByCharacterID(characterID int) ([]*InventorySlot, error) {

	var arr []*InventorySlot
	query := `select * from hops.items_characters where character_id=$1 and consignment=false and slot_id >= 0 order by slot_id asc`
	if _, err := db.Select(&arr, query, characterID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindInventorySlotsByCharacterID: %s", err.Error())
	}

	for _, s := range arr {
		if s.PetInfo != nil {
			json.Unmarshal(s.PetInfo, &s.Pet)
		}
		InventoryItems.Add(s.ID, s)
	}

	return arr, nil
}

func FindInventorySlotByID(id int) (*InventorySlot, error) {

	if s := InventoryItems.Get(id); s != nil {
		return s.(*InventorySlot), nil
	}

	s := NewSlot()
	query := `select * from hops.items_characters where id = $1`
	if err := db.SelectOne(s, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindInventorySlotByID: %s", err.Error())
	}

	InventoryItems.Add(s.ID, s)
	return s, nil
}

func FindBankSlotsByUserID(userID string) ([]*InventorySlot, error) {

	var arr []*InventorySlot
	query := `select * from hops.items_characters where user_id = $1 and character_id is NULL and slot_id >= 0 order by slot_id asc`
	if _, err := db.Select(&arr, query, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindBankSlotsByUserID: %s", err.Error())
	}

	for _, s := range arr {
		if s.PetInfo != nil {
			json.Unmarshal(s.PetInfo, &s.Pet)
		}
		InventoryItems.Add(s.ID, s)
	}

	return arr, nil
}

func (slot *InventorySlot) Insert() error {

	now := time.Now().UTC()
	slot.UpdatedAt = null.TimeFrom(now)
	if slot.UpgradeArr == "" {
		slot.UpgradeArr = "{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}"
	}
	if slot.SocketArr == "" {
		slot.SocketArr = "{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}"
	}

	if slot.Pet != nil {
		slot.PetInfo, _ = json.Marshal(slot.Pet)
	}

	if slot.PetInfo == nil {
		slot.PetInfo = json.RawMessage("{}")
	}

	err := db.Insert(slot)
	if err != nil {
		return err
	}

	InventoryItems.Add(slot.ID, slot)
	return nil
}

func (slot *InventorySlot) Update() error {

	if slot.ID == 0 {
		return nil
	}

	now := time.Now().UTC()
	slot.UpdatedAt = null.TimeFrom(now)

	if slot.UpgradeArr == "" {
		slot.UpgradeArr = "{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}"
	}
	if slot.SocketArr == "" {
		slot.SocketArr = "{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}"
	}

	if slot.Pet != nil {
		slot.PetInfo, _ = json.Marshal(slot.Pet)
	}

	if slot.PetInfo == nil {
		slot.PetInfo = json.RawMessage("{}")
	}

	_, err := db.Update(slot)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (slot *InventorySlot) Delete() error {
	InventoryItems.Delete(slot.ID)
	_, err := db.Delete(slot)
	if err != nil {
		log.Println(err)
	}

	return nil
}

func (slot *InventorySlot) GetUpgrades() []byte {
	upgs := strings.Split(strings.Trim(string(slot.UpgradeArr), "{}"), ",")
	return funk.Map(upgs, func(upg string) byte {
		u, _ := strconv.ParseUint(upg, 10, 8)
		return byte(u)
	}).([]byte)
}

func (slot *InventorySlot) SetUpgrade(i int, code byte) {
	upgs := slot.GetUpgrades()
	upgs[i] = code
	slot.SetUpgrades(upgs)
}

func (slot *InventorySlot) SetUpgrades(upgs []byte) {
	if len(upgs) == 15 {
		slot.UpgradeArr = fmt.Sprintf("{%s}", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(upgs)), ","), "[]"))
	}
}

func (slot *InventorySlot) GetSockets() []byte {
	upgs := strings.Split(strings.Trim(string(slot.SocketArr), "{}"), ",")
	return funk.Map(upgs, func(upg string) byte {
		u, _ := strconv.ParseUint(upg, 10, 8)
		return byte(u)
	}).([]byte)
}

func (slot *InventorySlot) SetSocket(i int, code byte) {
	socks := slot.GetSockets()
	socks[i] = code
	slot.SetSockets(socks)
}

func (slot *InventorySlot) SetSockets(socks []byte) {
	if len(socks) == 15 {
		slot.SocketArr = fmt.Sprintf("{%s}", strings.Trim(strings.Join(strings.Fields(fmt.Sprint(socks)), ","), "[]"))
	}
}

func (slot *InventorySlot) Upgrade(slotID int16, codes ...byte) []byte {

	if slot.ItemID == 0 {
		return nil
	} else if slot.Plus >= 15 {
		return nil
	}

	for _, code := range codes {
		slot.SetUpgrade(int(slot.Plus), code)
		slot.Plus++

		if slot.Plus == 15 {
			break
		}
	}

	resp := ITEM_UPGRADED
	resp.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), 9) // item id
	resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 17)     // slot id
	resp.Insert(slot.GetUpgrades(), 19)                            // item upgrades
	resp[34] = byte(slot.SocketCount)                              // socket count
	resp.Insert(slot.GetSockets(), 35)                             // item sockets
	c := 35 + 15
	if slot.ItemType != 0 {
		resp.Overwrite(utils.IntToBytes(uint64(slot.ItemType), 1, true), c-6)
		if slot.ItemType == 2 {
			resp.Overwrite(utils.IntToBytes(uint64(slot.JudgementStat), 4, true), c-5)
		}
	}
	err := slot.Update()
	if err != nil {
		return nil
	}

	return resp
}

func (slot *InventorySlot) CreateSocket(slotID int16, count int8) []byte {

	slot.SocketCount = count
	slot.SocketArr = "{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}"

	resp := SOCKET_OPENED
	resp.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), 8) // item id
	resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 16)     // slot id
	resp.Insert(slot.GetUpgrades(), 18)                            // item upgrades
	resp[33] = byte(slot.SocketCount)                              // socket count
	resp.Insert(slot.GetSockets(), 34)                             // sockets
	c := 34 + 15
	if slot.ItemType != 0 {
		resp.Overwrite(utils.IntToBytes(uint64(slot.ItemType), 1, true), c-6)
		if slot.ItemType == 2 {
			resp.Overwrite(utils.IntToBytes(uint64(slot.JudgementStat), 4, true), c-5)
		}
	}
	return resp
}

func (slot *InventorySlot) UpgradeSocket(slotID int16, codes []byte) []byte {

	for i := 0; i < len(codes); i++ {
		slot.SetSocket(i, codes[i])
	}

	resp := SOCKET_UPGRADED
	resp.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), 8) // item id
	resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 16)     // slot id
	resp.Insert(slot.GetUpgrades(), 18)                            // item upgrades
	resp[33] = byte(slot.SocketCount)                              // socket count
	resp.Insert(slot.GetSockets(), 34)                             // sockets
	c := 34 + 15
	if slot.ItemType != 0 {
		resp.Overwrite(utils.IntToBytes(uint64(slot.ItemType), 1, true), c-6)
		if slot.ItemType == 2 {
			resp.Overwrite(utils.IntToBytes(uint64(slot.JudgementStat), 4, true), c-5)
		}
	}
	err := slot.Update()
	if err != nil {
		return nil
	}

	return resp
}

func (slot *InventorySlot) GetData(slotID int16) []byte {
	resp, r2 := utils.Packet{}, utils.Packet{}
	if slot.ItemID > 0 {
		r := ITEM_SLOT
		item := Items[slot.ItemID]
		if item == nil {
			return nil
		}

		if item.GetType() == PET_TYPE { // pet

			pet := slot.Pet
			if pet == nil {
				return nil
			}

			if pet.Loyalty <= 0 {
				pet.Loyalty = 1
			}
			if pet.Fullness <= 0 {
				pet.Fullness = 1
			}

			r.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), 6) // item id
			r[11] = pet.Level
			r.Insert([]byte{pet.Loyalty, pet.Fullness}, 12)             // loyalty and fullness
			r.Insert(utils.IntToBytes(uint64(slotID), 2, true), 14)     // slot id
			r.Overwrite(utils.IntToBytes(uint64(pet.HP), 2, true), 16)  // pet hp
			r.Overwrite(utils.IntToBytes(uint64(pet.CHI), 2, true), 18) // pet chi
			r.Overwrite(utils.IntToBytes(uint64(pet.Exp), 8, true), 20) // pet exp

			//r2 = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x57, 0x05, 0x0A, 0x02, 0x55, 0xAA}

		} else { // normal item
			r.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), 6) // item id

			if slot.Activated { // using state
				if item.TimerType == 1 {
					r[10] = 3
				} else if item.TimerType == 3 {
					r[10] = 5
					r2 = GREEN_ITEM_COUNT
					r2.Insert(utils.IntToBytes(uint64(slotID), 2, true), 8)         // slot id
					r2.Insert(utils.IntToBytes(uint64(slot.Quantity), 4, true), 10) // item quantity
				}
			} else {
				r[10] = 0
			}

			if item.GetType() == FILLER_POTION_TYPE {
				r[11] = 0xA2
				r.Insert(utils.IntToBytes(1, 2, true), 12)              // item quantity
				r.Insert(utils.IntToBytes(uint64(slotID), 2, true), 14) // slot id

				qBytes := utils.IntToBytes(uint64(slot.Quantity), 8, true)
				for i := 0; i < 8; i++ {
					r[i+16] = qBytes[i]
				}

			} else {
				if slot.Plus > 0 || slot.SocketCount > 0 { // item text color
					r[11] = 0xA2
				}

				r.Insert(utils.IntToBytes(uint64(slot.Quantity), 2, true), 12) // item quantity
				r.Insert(utils.IntToBytes(uint64(slotID), 2, true), 14)        // slot id

				upgs := slot.GetUpgrades()
				socks := slot.GetSockets()

				for j := 0; j < 15; j++ {
					r[j+16] = upgs[j] // item upgrades
				}
				r[31] = byte(slot.SocketCount) // socket count
				for j := 0; j < 15; j++ {
					r[j+32] = socks[j] // socket features
				}
				r.Overwrite(utils.IntToBytes(uint64(slot.ItemType), 1, true), 41)
				r.Overwrite(utils.IntToBytes(uint64(slot.JudgementStat), 4, true), 42)
			}
		}
		if slot.Appearance != 0 && slot.Plus <= 0 {
			r.Overwrite(utils.IntToBytes(uint64(slot.Appearance), 4, true), 16)
		} else if slot.Appearance != 0 && slot.Plus > 0 {
			r.Overwrite(utils.IntToBytes(uint64(slot.Appearance), 4, true), 46)
		}
		resp.Concat(r)
		resp.Concat(r2)

	} else { // empty slot
		r := ITEM_SLOT
		r.Insert(utils.IntToBytes(0, 4, true), 6)               // item id
		r.Insert(utils.IntToBytes(0, 2, true), 12)              // item quantity
		r.Insert(utils.IntToBytes(uint64(slotID), 2, true), 14) // slot id
		resp.Concat(r)
	}

	return resp
}

func (slot *InventorySlot) GetPetStats(c *Character) []byte {

	pet := slot.Pet
	if pet == nil {
		return nil
	}

	petInfo := Pets[slot.ItemID]

	level := byte(petInfo.Level)
	levelDiff := int(pet.Level - level)
	STR := petInfo.BaseSTR + levelDiff*petInfo.AdditionalSTR
	DEX := petInfo.BaseDEX + levelDiff*petInfo.AdditionalDEX
	INT := petInfo.BaseINT + levelDiff*petInfo.AdditionalINT

	maxHP := petInfo.BaseHP + levelDiff*petInfo.AdditionalHP
	maxCHI := petInfo.BaseChi + levelDiff*petInfo.AdditionalChi

	minATK := STR
	maxATK := STR
	DEF := DEX
	artsDEF := DEX + 2*INT

	slots, err := c.InventorySlots()
	if err != nil {
		return nil
	}

	items := []*InventorySlot{slots[317], slots[318], slots[319]}
	for d, item := range items {
		if item.ItemID == 0 {
			continue
		}

		info := Items[item.ItemID]
		if info == nil || info.Slot != d+317 {
			continue
		}

		maxHP += info.MaxHp
		minATK += info.MinAtk
		maxATK += info.MaxAtk
		DEF += info.Def
		artsDEF += info.ArtsDef

		upgs := item.GetUpgrades()
		for i := byte(0); i < item.Plus; i++ {
			upg := Items[int64(upgs[i])]
			if upg == nil {
				continue
			}

			maxHP += upg.MaxHp
			minATK += upg.MinAtk
			maxATK += upg.MaxAtk
			DEF += upg.Def
			artsDEF += upg.ArtsDef
		}
	}

	pet.MaxHP = maxHP
	pet.MaxCHI = maxCHI
	pet.STR = STR
	pet.DEX = DEX
	pet.INT = INT
	pet.MinATK = minATK
	pet.MaxATK = maxATK
	pet.DEF = DEF
	pet.ArtsDEF = artsDEF

	resp := PET_STATS
	resp.Insert(utils.IntToBytes(uint64(pet.HP), 2, true), 8)       // pet hp
	resp.Insert(utils.IntToBytes(uint64(pet.MinATK), 2, true), 10)  // pet min atk
	resp.Insert(utils.IntToBytes(uint64(pet.MaxATK), 2, true), 12)  // pet max atk
	resp.Insert(utils.IntToBytes(uint64(pet.DEF), 4, true), 14)     // pet def
	resp.Insert(utils.IntToBytes(uint64(pet.ArtsDEF), 4, true), 18) // pet arts def

	return resp
}

func (pet *PetSlot) FindTargetMobID(owner *Character) (int, error) {

	var (
		distance = 15.0
	)

	ownerPos := ConvertPointToLocation(owner.Coordinate)
	ownerDist := utils.CalculateDistance(ownerPos, &pet.Coordinate)
	if ownerDist > 15 {
		return 0, nil
	}

	user := owner.Socket.User
	allMobs := AIsByMap[user.ConnectedServer][owner.Map]
	filtered := funk.Filter(allMobs, func(ai *AI) bool {

		pos := NPCPos[ai.PosID]
		if pos == nil {
			return false
		}

		aiCoordinate := ConvertPointToLocation(ai.Coordinate)
		seed := utils.RandInt(0, 1000)

		return user.ConnectedServer == ai.Server && owner.Map == ai.Map && pet.HP > 0 && pet.IsOnline && pos.Attackable &&
			utils.CalculateDistance(&pet.Coordinate, aiCoordinate) <= distance && seed < 750
	})

	filtered = funk.Shuffle(filtered)
	mobs := filtered.([]*AI)
	if len(mobs) > 0 {
		return int(mobs[0].PseudoID), nil
	}

	return 0, nil
}

func (pet *PetSlot) MovementHandler(token int64, start, end *utils.Location, speed float64) {

	diff := utils.CalculateDistance(start, end)

	if diff < 1 {
		pet.Coordinate = *end
		pet.MovementToken = 0
		pet.IsMoving = false
		return
	}

	pet.Coordinate = *start
	pet.TargetLocation = *end

	r := utils.Packet{}
	r = pet.Move(*end, 2)

	p := nats.CastPacket{CastNear: true, PetID: pet.PseudoID, Data: r, Type: nats.PET_MOVEMENT}
	p.Cast()

	if diff <= speed { // target is so close
		*start = *end
		time.AfterFunc(time.Duration(diff/speed)*time.Millisecond, func() {
			if token == pet.MovementToken {
				pet.MovementHandler(token, start, end, speed)
			}
		})
	} else { // target is away
		start.X += (end.X - start.X) * speed / diff
		start.Y += (end.Y - start.Y) * speed / diff
		time.AfterFunc(1000*time.Millisecond, func() {
			if token == pet.MovementToken {
				pet.MovementHandler(token, start, end, speed)
			}
		})
	}
}

func (pet *PetSlot) Move(targetLocation utils.Location, runningMode byte) []byte {

	resp := MOB_MOVEMENT
	currentLocation := pet.Coordinate

	resp.Insert(utils.IntToBytes(uint64(pet.PseudoID), 2, true), 5) // pet pseudo id
	resp[7] = runningMode
	resp.Insert(utils.FloatToBytes(currentLocation.X, 4, true), 8)  // current coordinate-x
	resp.Insert(utils.FloatToBytes(currentLocation.Y, 4, true), 12) // current coordinate-y
	resp.Insert(utils.FloatToBytes(targetLocation.X, 4, true), 20)  // current coordinate-x
	resp.Insert(utils.FloatToBytes(targetLocation.Y, 4, true), 24)  // current coordinate-y

	speeds := []float64{0.0, 5.0, 10.0}
	resp.Insert(utils.FloatToBytes(speeds[runningMode], 4, true), 32) // speed

	return resp
}

func (pet *PetSlot) PlayerAttack(owner *Character) []byte {
	resp := MOB_ATTACK
	mob := FindCharacterByPseudoID(owner.Socket.User.ConnectedServer, uint16(pet.Target))

	rawDamage := int(utils.RandInt(int64(pet.MinATK), int64(pet.MaxATK)))
	damage := int(math.Max(float64(rawDamage-mob.Socket.Stats.DEF), 3))

	reqAcc := float64(int(mob.Level)-int(pet.Level)) * 10
	if utils.RandInt(0, 1000) < int64(reqAcc) {
		damage = 0
	}

	resp.Insert(utils.IntToBytes(uint64(pet.PseudoID), 2, true), 6) // pet pseudo id
	//resp.Insert([]byte{0}, 8)
	resp.Insert(utils.IntToBytes(uint64(mob.PseudoID), 2, true), 8) // target pseudo id

	resp[11] = 2
	if damage > 0 {
		resp[12] = 1 // damage sound
		pet.AddExp(owner, 0)
		time.AfterFunc(time.Second/2, func() {
			owner.DealDamageToPlayer(mob, damage)
		})
	}

	return resp
}
func (pet *PetSlot) Attack(owner *Character) []byte {

	resp := MOB_ATTACK
	mob, ok := GetFromRegister(owner.Socket.User.ConnectedServer, owner.Map, uint16(pet.Target)).(*AI)
	if !ok || mob == nil || mob.HP <= 0 {
		return nil
	}

	pos := NPCPos[mob.PosID]
	if pos == nil {
		return nil
	}

	npc := NPCs[pos.NPCID]
	if npc == nil {
		return nil
	}

	rawDamage := int(utils.RandInt(int64(pet.MinATK), int64(pet.MaxATK)))
	damage := int(math.Max(float64(rawDamage-npc.DEF), 3))

	reqAcc := float64(int(npc.Level)-int(pet.Level)) * 10
	if utils.RandInt(0, 1000) < int64(reqAcc) {
		damage = 0
	}

	resp.Insert(utils.IntToBytes(uint64(pet.PseudoID), 2, true), 6) // pet pseudo id
	//resp.Insert([]byte{0}, 8)
	resp.Insert(utils.IntToBytes(uint64(mob.PseudoID), 2, true), 8) // target pseudo id

	resp[11] = 2
	if damage > 0 {
		resp[12] = 1 // damage sound
	}

	pet.AddExp(owner, 0)
	time.AfterFunc(time.Second/2, func() {
		owner.DealDamage(mob, damage)
	})

	return resp
}

func (pet *PetSlot) CastSkill(owner *Character) []byte {

	mob, ok := GetFromRegister(owner.Socket.User.ConnectedServer, owner.Map, uint16(pet.Target)).(*AI)
	if !ok || mob == nil || mob.HP <= 0 {
		return nil
	}

	pos := NPCPos[mob.PosID]
	if pos == nil {
		return nil
	}

	npc := NPCs[pos.NPCID]
	if npc == nil {
		return nil
	}

	slots, err := owner.InventorySlots()
	if err != nil {
		return nil
	}

	petSlot := slots[0x0A]
	petInfo := Pets[petSlot.ItemID]

	min := pet.MinATK + pet.MinATK*pet.INT/100
	max := pet.MaxATK + pet.MaxATK*pet.INT/100

	skillInfo := SkillInfos[petInfo.SkillID]
	castLocation := &pet.Coordinate

	if skillInfo.AreaCenter == 1 || skillInfo.AreaCenter == 2 {
		castLocation = ConvertPointToLocation(mob.Coordinate)
	}

	castRange := skillInfo.BaseRadius
	candidates := AIsByMap[mob.Server][mob.Map]
	candidates = funk.Filter(candidates, func(cand *AI) bool {
		nPos := NPCPos[cand.PosID]
		if nPos == nil {
			return false
		}

		aiCoordinate := ConvertPointToLocation(cand.Coordinate)
		return (cand.PseudoID == mob.PseudoID || (utils.CalculateDistance(aiCoordinate, castLocation) < castRange)) && cand.HP > 0 && nPos.Attackable
	}).([]*AI)

	resp := MOB_SKILL
	resp.Insert(utils.IntToBytes(uint64(pet.PseudoID), 2, true), 7)    // pet pseudo id
	resp.Insert(utils.IntToBytes(uint64(petInfo.SkillID), 4, true), 9) // pet skill id
	resp.Insert(utils.FloatToBytes(pet.Coordinate.X, 4, true), 13)     // pet-x
	resp.Insert(utils.FloatToBytes(pet.Coordinate.Y, 4, true), 17)     // pet-x
	resp.Insert(utils.IntToBytes(uint64(mob.PseudoID), 2, true), 25)   // target pseudo id
	resp.Insert(utils.IntToBytes(uint64(mob.PseudoID), 2, true), 28)   // target pseudo id

	pet.Casting = true
	pet.CHI -= skillInfo.BaseChi

	time.AfterFunc(time.Second*2, func() {
		pet.Casting = false
		for _, target := range candidates {

			pos := NPCPos[target.PosID]
			if pos == nil {
				continue
			}

			npc := NPCs[pos.NPCID]
			if npc == nil {
				continue
			}

			rawDamage := int(utils.RandInt(int64(min), int64(max)))
			damage := int(math.Max(float64(rawDamage-npc.ArtsDEF), 3))

			reqAcc := float64(int(npc.Level)-int(pet.Level)) * 10
			if utils.RandInt(0, 1000) < int64(reqAcc) {
				damage = 0
			}
			pet.AddExp(owner, 0)
			owner.DealDamage(target, damage)
		}
	})

	resp.Concat(petSlot.GetData(0x0A))
	return resp
}

func (pet *PetSlot) AddExp(owner *Character, amount uint64) {

	slots, err := owner.InventorySlots()
	if err != nil {
		return
	}

	petSlot := slots[0x0A]
	petInfo := Pets[petSlot.ItemID]
	petExpInfo := PetExps[int16(pet.Level)]
	if petExpInfo == nil {
		log.Println("Invalid pet level:", pet.Level)
		return
	}

	targetExps := []int{petExpInfo.ReqExpEvo1, petExpInfo.ReqExpEvo2, petExpInfo.ReqExpEvo3, petExpInfo.ReqExpHt, petExpInfo.ReqExpDivEvo1, petExpInfo.ReqExpDivEvo2, petExpInfo.ReqExpDivEvo3}
	if amount <= 0 {
		pet.Exp += 2
	} else {
		pet.Exp += amount
	}
	inform := true

	for pet.Exp >= uint64(targetExps[petInfo.Evolution-1]) {
		if pet.Level < 100 {
			pet.Level++
			petExpInfo = PetExps[int16(pet.Level)]
			targetExps = []int{petExpInfo.ReqExpEvo1, petExpInfo.ReqExpEvo2, petExpInfo.ReqExpEvo3, petExpInfo.ReqExpHt, petExpInfo.ReqExpDivEvo1, petExpInfo.ReqExpDivEvo2, petExpInfo.ReqExpDivEvo3}

		} else {
			targetExps = []int{petExpInfo.ReqExpEvo1, petExpInfo.ReqExpEvo2, petExpInfo.ReqExpEvo3, petExpInfo.ReqExpHt, petExpInfo.ReqExpDivEvo1, petExpInfo.ReqExpDivEvo2, petExpInfo.ReqExpDivEvo3}
			pet.Exp = uint64(targetExps[petInfo.Evolution-1])
			inform = false
			break
		}
	}

	petExpInfo = PetExps[int16(pet.Level)]
	targetExps = []int{petExpInfo.ReqExpEvo1, petExpInfo.ReqExpEvo2, petExpInfo.ReqExpEvo3, petExpInfo.ReqExpHt, petExpInfo.ReqExpDivEvo1, petExpInfo.ReqExpDivEvo2, petExpInfo.ReqExpDivEvo3}

	for int16(pet.Level) >= petInfo.TargetLevel { // evolution
		if petInfo.EvolvedID == 0 {
			break
		}

		petSlot.ItemID = petInfo.EvolvedID
		petInfo = Pets[petInfo.EvolvedID]
		pet.Exp = uint64(targetExps[petInfo.Evolution-1])
	}

	if inform {
		owner.Socket.Write(owner.GetPetStats())
	}
}
