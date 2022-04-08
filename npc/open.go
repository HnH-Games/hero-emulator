package npc

import (
	"encoding/binary"
	"fmt"
	"log"
	"strings"

	"hero-emulator/database"
	"hero-emulator/messaging"
	"hero-emulator/nats"
	"hero-emulator/utils"

	"github.com/thoas/go-funk"
	"github.com/tidwall/gjson"
)

type OpenHandler struct {
}

type PressButtonHandler struct {
}

var (
	shops = map[int]int{20002: 7, 20003: 2, 20004: 4, 20005: 1, 20009: 8, 20010: 10, 20011: 10, 20013: 25,
		20024: 6, 20025: 6, 20026: 11, 20033: 21, 20034: 22, 20035: 23, 20036: 24, 20044: 21, 20047: 21, 20082: 21,
		20413: 25, 20379: 25, 20253: 25, 20251: 25, 20414: 25, 23725: 25, 20337: 25, 20323: 25, 203161: 25, 20290: 25, 20236: 25,
		20083: 21, 20084: 21, 20085: 23, 20086: 22, 20087: 21, 20094: 103, 20095: 100, 20105: 21,
		20146: 21, 20151: 6, 20173: 327, 20211: 25, 20202: 105, 20361: 329, 20364: 306, 20415: 21, 20206: 21, 25004: 362, 25005: 346, 25001: 363, 20205: 540, 20207: 551, 20204: 553, 20203: 554, 20292: 555, 23742: 556, 60123: 557, 20316: 558, 20329: 559, 20016: 560, 20018: 561, 20088: 562, 20325: 563, 20310: 564, 20305: 565, 21096: 566, 20160: 700, 20133: 569, 20301: 571, 45109: 572, 45110: 573, 44490: 574, 44489: 575, 20137: 576, 20343: 577, 41171: 578, 20117: 579, 20239: 580, 20293: 581, 50008: 582, 23720: 583}

	COMPOSITION_MENU    = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x0F, 0x01, 0x55, 0xAA}
	OPEN_SHOP           = utils.Packet{0xAA, 0x55, 0x07, 0x00, 0x57, 0x03, 0x01, 0x55, 0xAA}
	NPC_MENU            = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x57, 0x02, 0x55, 0xAA}
	STRENGTHEN_MENU     = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x08, 0x01, 0x55, 0xAA}
	JOB_PROMOTED        = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x09, 0x00, 0x55, 0xAA}
	NOT_ENOUGH_LEVEL    = utils.Packet{0xAA, 0x55, 0x0B, 0x00, 0x57, 0x02, 0x38, 0x42, 0x0F, 0x00, 0x00, 0x55, 0xAA}
	INVALID_CLASS       = utils.Packet{0xAA, 0x55, 0x0B, 0x00, 0x57, 0x02, 0x49, 0x2F, 0x00, 0x00, 0x00, 0x55, 0xAA}
	INVALID_REQUIREMENT = utils.Packet{0xAA, 0x55, 0x0B, 0x00, 0x57, 0x02, 0x4A, 0x2A, 0x00, 0x00, 0x00, 0x55, 0xAA}
	GUILD_MENU          = utils.Packet{0xAA, 0x55, 0x02, 0x00, 0x57, 0x0D, 0x55, 0xAA}
	DISMANTLE_MENU      = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x16, 0x01, 0x55, 0xAA}
	EXTRACTION_MENU     = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x17, 0x01, 0x55, 0xAA}
	ADV_FUSION_MENU     = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x32, 0x01, 0x55, 0xAA}
	TACTICAL_SPACE      = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x50, 0x01, 0x01, 0x55, 0xAA}
	CREATE_SOCKET_MENU  = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x39, 0x01, 0x55, 0xAA}
	UPGRADE_SOCKET_MENU = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x3A, 0x01, 0x55, 0xAA}
	CONSIGNMENT_MENU    = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x42, 0x01, 0x55, 0xAA}
	CO_PRODUCTION_MENU  = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x3B, 0x01, 0x55, 0xAA}
	SYNTHESIS_MENU      = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x45, 0x01, 0x55, 0xAA}
	HIGH_SYNTHETIC_MENU = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x46, 0x01, 0x55, 0xAA}
	APPEARANCE_MENU     = utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x41, 0x01, 0x55, 0xAA}

	FACTION_WAR = utils.Packet{0xAA, 0x55, 0x23, 0x00, 0x65, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
)

func makeAnnouncement(msg string) {
	length := int16(len(msg) + 3)

	resp := database.ANNOUNCEMENT
	resp.SetLength(length)
	resp[6] = byte(len(msg))
	resp.Insert([]byte(msg), 7)

	p := nats.CastPacket{CastNear: false, Data: resp}
	p.Cast()
}
func (h *OpenHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	c := s.Character
	if c == nil {
		return nil, nil
	}

	u := s.User
	if u == nil {
		return nil, nil
	}

	id := uint16(utils.BytesToInt(data[6:10], true))
	pos, ok := database.GetFromRegister(1, c.Map, id).(*database.NpcPosition)
	if !ok {
		return nil, nil
	}

	npc := database.NPCs[pos.NPCID]

	if npc.ID == 20147 { // Ice Palace Mistress Lord
		coordinate := &utils.Location{X: 163, Y: 350}
		return c.Teleport(coordinate), nil

	} else if npc.ID == 20055 { // Mysterious Tombstone
		coordinate := &utils.Location{X: 365, Y: 477}
		return c.Teleport(coordinate), nil

	} else if npc.ID == 41381 { // Old War
		coordinate := &utils.Location{X: 125, Y: 142}
		return c.ChangeMap(253, nil)
		return c.Teleport(coordinate), nil

	} else if npc.ID == 20056 { // Mysterious Tombstone (R)
		coordinate := &utils.Location{X: 70, Y: 450}
		return c.Teleport(coordinate), nil

	} else if npc.ID == 22351 { // Golden Castle Teleport Tombstone
		return c.ChangeMap(251, nil)
	} else if npc.ID == 22357 { // 2nd FL Entrance
		return c.ChangeMap(237, nil)

	} else if npc.ID == 22358 { // 3rd FL Entrance
		return c.ChangeMap(239, nil)
	}
	if npc.ID == 23728 { //  Golden Castle Teleport Tombstone
		return c.ChangeMap(121, nil)
	}
	if npc.ID == 22355 { //  Boss Map Teleport
		return c.ChangeMap(241, nil)
	}
	if npc.ID == 20134 { //  Gold Mine
		return c.ChangeMap(251, nil)
	}
	if npc.ID == 23801 { // Hongmu (Dream Valey Teleport Master )
		return c.ChangeMap(99, nil)
	}
	if npc.ID == 23722 { // Okamu (Master Teleport lv.225-250 )
		return c.ChangeMap(43, nil)
	}
	if npc.ID == 23719 { //Britney (Master Teleport 250-275)
		return c.ChangeMap(45, nil)
	}
	if npc.ID == 23714 { //Biryu (Master Teleport (275-300)
		return c.ChangeMap(44, nil)
	}
	if npc.ID == 21600 { //Premium Map
		return c.ChangeMap(120, nil)
	}
	if npc.ID == 20355 { //Premium Map
		return c.ChangeMap(112, nil)
	}
	if npc.ID == 23718 { //Premium Map2
		return c.ChangeMap(122, nil)
	} else if npc.ID == 22358 { // 3rd FL Entrance
		return c.ChangeMap(239, nil)
	}
	if npc.ID == 1000026 { //Biryu (Master Teleport (275-300)
		return c.ChangeMap(247, nil)
	}
	if npc.ID == 1000009 { //Premium Map2
		return c.ChangeMap(160, nil)
	}
	if npc.ID == 23751 { //Premium Map2
		return c.ChangeMap(10, nil)
	}
	if npc.ID == 16210013 { //Premium Map
		return c.ChangeMap(1, nil)
	}
	if npc.ID == 20245 { //Premium Map
		return c.ChangeMap(251, nil)
	}
	if npc.ID == 20296 { //Premium Map
		return c.ChangeMap(243, nil)
	}
	if npc.ID == 20289 { //Premium Map
		return c.ChangeMap(52, nil)
	}
	if npc.ID == 449065 { //Premium Map
		return c.ChangeMap(2, nil)
	}
	if npc.ID == 20290 { //Premium Map
		return c.ChangeMap(56, nil)
	}
	if npc.ID == 20292 { //Premium Map
		return c.ChangeMap(1, nil)
	}
	if npc.ID == 20315 { //Premium Map
		return c.ChangeMap(1, nil)
	}
	if npc.ID == 20265 { //Premium Map
		return c.ChangeMap(61, nil)
	}
	npcScript := database.NPCScripts[npc.ID]
	if npcScript == nil {
		return nil, nil
	}

	script := string(npcScript.Script)
	textID := gjson.Get(script, "text").Int()
	actions := []int{}

	for _, action := range gjson.Get(script, "actions").Array() {
		actions = append(actions, int(action.Int()))
	}

	resp := NPC_MENU
	resp.Insert(utils.IntToBytes(uint64(npc.ID), 4, true), 6)        // npc id
	resp.Insert(utils.IntToBytes(uint64(textID), 4, true), 10)       // text id
	resp.Insert(utils.IntToBytes(uint64(len(actions)), 1, true), 14) // action length

	index, length := 15, int16(11)
	for i, action := range actions {
		resp.Insert(utils.IntToBytes(uint64(action), 4, true), index) // action
		index += 4

		resp.Insert(utils.IntToBytes(uint64(npc.ID), 2, true), index) // npc id
		index += 2

		resp.Insert(utils.IntToBytes(uint64(i+1), 2, true), index) // action index
		index += 2

		length += 8
	}

	resp.SetLength(length)
	return resp, nil
}

func (h *PressButtonHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	c := s.Character
	if c == nil {
		return nil, nil
	}

	npcID := int(utils.BytesToInt(data[6:8], true))
	index := int(utils.BytesToInt(data[8:10], true))
	indexes := []int{index & 7, (index & 56) / 8, (index & 448) / 64, (index & 3584) / 512, (index & 28672) / 4096}
	indexes = funk.FilterInt(indexes, func(i int) bool {
		return i > 0
	})

	npcScript := database.NPCScripts[npcID]
	if npcScript == nil {
		return nil, nil
	}

	script := string(npcScript.Script)
	key := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(indexes)), "."), "[]")
	script = gjson.Get(script, key).String()
	if script != "" {
		textID := int(gjson.Get(script, fmt.Sprintf("text")).Int())
		actions := []int{}

		for _, action := range gjson.Get(script, "actions").Array() {
			actions = append(actions, int(action.Int()))
		}

		resp := GetNPCMenu(npcID, textID, index, actions)
		return resp, nil
	} else { // Action button

		key := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(indexes[:len(indexes)-1])), "."), "[]")
		script = string(npcScript.Script)
		if key != "" {
			script = gjson.Get(script, key).String()
		}

		actions := gjson.Get(script, "actions").Array()
		actIndex := indexes[len(indexes)-1] - 1
		actID := actions[actIndex].Int()
		log.Println(actID)

		resp := utils.Packet{}
		var err error
		book1, book2, job := 0, 0, 0
		switch actID {
		case 1: // Exchange
			shopNo := shops[npcID]
			resp = OPEN_SHOP
			resp.Insert(utils.IntToBytes(uint64(shopNo), 4, true), 7) // shop id

		case 2: // Compositon
			resp = COMPOSITION_MENU

		case 4: // Strengthen
			resp = STRENGTHEN_MENU

		case 6: // Deposit
			resp = c.BankItems()

		case 13: // Accept
			switch npcID {
			case 20006: // Hunter trainer
				book1, job = 16210003, 13
				resp, err = firstJobPromotion(c, book1, job, npcID)
				if err != nil {
					return nil, err
				}
			case 20020: // Warrior trainer
				book1, job = 16210001, 11
				resp, err = firstJobPromotion(c, book1, job, npcID)
				if err != nil {
					return nil, err
				}
			case 20021: // Physician trainer
				book1, job = 16210002, 12
				resp, err = firstJobPromotion(c, book1, job, npcID)
				if err != nil {
					return nil, err
				}
			case 20022: // Assassin trainer
				book1, job = 16210004, 14
				resp, err = firstJobPromotion(c, book1, job, npcID)
				if err != nil {
					return nil, err
				}
			case 20057: //HERO BATTLE MANAGER
				switch index {
				case 11: //THE GREAT WAR
					if database.CanJoinWar {
						if c.Faction == 1 {
							x := 75.0
							y := 45.0
							c.IsinWar = true
							database.OrderCharacters[c.ID] = c
							data, _ := c.ChangeMap(230, database.ConvertPointToLocation(fmt.Sprintf("%.1f,%.1f", x, y)))
							c.Socket.Write(data)
						} else {
							x := 81.0
							y := 475.0
							c.IsinWar = true
							database.ShaoCharacters[c.ID] = c
							data, _ := c.ChangeMap(230, database.ConvertPointToLocation(fmt.Sprintf("%.1f,%.1f", x, y)))
							c.Socket.Write(data)
						}
					}
				case 10: //FACTION WAR

				case 9: //FLAG KINGDOM
				}
			case 20415: // RDL tavern
				resp, _ = c.ChangeMap(254, nil)
			}
		case 64: // Create Guild
			if c.GuildID == -1 {
				resp = GUILD_MENU
			}
		case 3110: //New zone
			resp, _ = c.ChangeMap(41, nil)
		case 3106: //New zone
			resp, _ = c.ChangeMap(40, nil)
		case 4171: //Legendary fox Port
			resp, _ = c.ChangeMap(71, nil)

		case 3107: //New zone
			resp, _ = c.ChangeMap(44, nil)
		case 78: // Move to Dragon Castle
			resp, _ = c.ChangeMap(1, nil)
		case 706: // Move to Gm Map
			resp, _ = c.ChangeMap(46, nil)

		case 86: // Move to Spirit Spire
			resp, _ = c.ChangeMap(5, nil)
		case 103: // Move to Highlands
			resp, _ = c.ChangeMap(2, nil)

		case 104: // Move to Venom Swamp
			resp, _ = c.ChangeMap(3, nil)
		case 4190: //Move To Starter Zone
			resp, _ = c.ChangeMap(70, nil)
		case 106: // Move to Silent Valley
			resp, _ = c.ChangeMap(11, nil)
		case 148: // Become a Champion
			book1, book2, job = 16100039, 16100200, 21
			resp, err = secondJobPromotion(c, book1, book2, 11, job, npcID)
			if err != nil {
				return nil, err
			}
		case 149: // Become a Musa
			book1, book2, job = 16100040, 16100200, 22
			resp, err = secondJobPromotion(c, book1, book2, 11, job, npcID)
			if err != nil {
				return nil, err
			}
		case 151: // Become a Surgeon
			book1, book2, job = 16100041, 16100200, 23
			resp, err = secondJobPromotion(c, book1, book2, 12, job, npcID)
			if err != nil {
				return nil, err
			}
		case 152: // Become a Combat Medic
			book1, book2, job = 16100042, 16100200, 24
			resp, err = secondJobPromotion(c, book1, book2, 12, job, npcID)
			if err != nil {
				return nil, err
			}
		case 154: // Become a Slayer
			book1, book2, job = 16100043, 16100200, 27
			resp, err = secondJobPromotion(c, book1, book2, 14, job, npcID)
			if err != nil {
				return nil, err
			}
		case 155: // Become a Shinobi
			book1, book2, job = 16100044, 16100200, 28
			resp, err = secondJobPromotion(c, book1, book2, 14, job, npcID)
			if err != nil {
				return nil, err
			}
		case 157: // Become a Tracker
			book1, book2, job = 16100045, 16100200, 25
			resp, err = secondJobPromotion(c, book1, book2, 13, job, npcID)
			if err != nil {
				return nil, err
			}
		case 158: // Become a Ranger
			book1, book2, job = 16100046, 16100200, 26
			resp, err = secondJobPromotion(c, book1, book2, 13, job, npcID)
			if err != nil {
				return nil, err
			}

		case 194: // Dismantle
			resp = DISMANTLE_MENU

		case 195: // Extraction
			resp = EXTRACTION_MENU

		case 116:
			database.AddMemberToFactionWar(c)
			resp, _ = c.ChangeMap(255, nil)

		case 5214: // Exit Paid Zone
			if maps, ok := database.DKMaps[c.Map]; ok {
				resp, err = c.ChangeMap(maps[0], nil)
				if err != nil {
					return nil, err
				}
			}

		case 5215: // Enter Paid Zone
			f := func(item *database.InventorySlot) bool {
				return item.Activated
			}
			_, item, err := c.FindItemInInventory(f, 15700040, 15710087)
			if err != nil {
				return nil, err
			} else if item == nil { // You don't have ticket
				resp := GetNPCMenu(npcID, 999993, 0, nil)
				return resp, nil
			}

			if maps, ok := database.DKMaps[c.Map]; ok {
				resp, err = c.ChangeMap(maps[1], nil)
				if err != nil {
					return nil, err
				}
			}
		case 542:
			if database.CanJoinWar {
				if c.Faction == 1 {
					x := 75.0
					y := 45.0
					c.IsinWar = true
					database.OrderCharacters[c.ID] = c
					data, _ := c.ChangeMap(230, database.ConvertPointToLocation(fmt.Sprintf("%.1f,%.1f", x, y)))
					c.Socket.Write(data)
				} else {
					x := 81.0
					y := 475.0
					c.IsinWar = true
					database.ShaoCharacters[c.ID] = c
					data, _ := c.ChangeMap(230, database.ConvertPointToLocation(fmt.Sprintf("%.1f,%.1f", x, y)))
					c.Socket.Write(data)
				}
			}
		case 559: // Advanced Fusion
			resp = ADV_FUSION_MENU
		case 526: //Get Divine Skills

		case 631: // Tactical Space
			resp = TACTICAL_SPACE

		case 7312: // Flexible Castle Entry
			f := func(item *database.InventorySlot) bool {
				return item.Activated
			}
			_, item, err := c.FindItemInInventory(f, 15710087)
			if err != nil {
				return nil, err
			} else if item == nil { // You don't have ticket
				resp := GetNPCMenu(npcID, 999993, 0, nil)
				return resp, nil
			}

			if maps, ok := database.DKMaps[c.Map]; ok {
				resp, err = c.ChangeMap(maps[2], nil)
				if err != nil {
					return nil, err
				}
			}
		case 3315: // Premium Map
			slotID, item, err := c.FindItemInInventory(nil, 99002341)
			if err != nil {
				return nil, err
			} else if item == nil { // You don't have ticket
				resp := GetNPCMenu(npcID, 23740, 0, nil)
				return resp, nil
			} else {
				data := c.DecrementItem(slotID, 1)
				c.Socket.Write(*data)
				resp, _ = c.ChangeMap(244, nil)

			}
		case 3316: // Premium Map
			slotID, item, err := c.FindItemInInventory(nil, 990029421)
			if err != nil {
				return nil, err
			} else if item == nil { // You don't have ticket
				resp := GetNPCMenu(npcID, 23740, 0, nil)
				return resp, nil
			} else {
				data := c.DecrementItem(slotID, 1)
				c.Socket.Write(*data)
				resp, _ = c.ChangeMap(151, nil)

			}
		case 77: // Premium Map
			slotID, item, err := c.FindItemInInventory(nil, 990029422)
			if err != nil {
				return nil, err
			} else if item == nil { // You don't have ticket
				resp := GetNPCMenu(npcID, 23740, 0, nil)
				return resp, nil
			} else {
				data := c.DecrementItem(slotID, 1)
				c.Socket.Write(*data)
				resp, _ = c.ChangeMap(152, nil)

			}
		case 737: // Create Socket
			resp = CREATE_SOCKET_MENU

		case 738: // Upgrade Socket
			resp = UPGRADE_SOCKET_MENU

		case 739: // Co-production menu
			resp = CO_PRODUCTION_MENU
		case 906: //APPEARANCE CHANGE
			resp = APPEARANCE_MENU
		case 970: // Consignment
			resp = CONSIGNMENT_MENU
		case 3230: //High-grade synthetic
			resp = HIGH_SYNTHETIC_MENU
		case 3231: //High-grade synthetic
			resp = SYNTHESIS_MENU
		case 3306: // Aid 2hr
			_, item, err := c.FindItemInInventory(nil, 13000170)
			if item != nil || err != nil {
				return nil, nil
			}

			cost := 20000000

			if c.Gold < uint64(cost) {
				return nil, nil
			}

			itemData, _, err := c.AddItem(&database.InventorySlot{ItemID: 13000170, Quantity: 120}, -1, true)
			if err != nil {
				return nil, err
			}

			c.LootGold(-uint64(cost))
			resp.Concat(*itemData)
			resp.Concat(c.GetGold())
		case 3319: // Aid 4hr
			_, item, err := c.FindItemInInventory(nil, 13000074)
			if item != nil || err != nil {
				return nil, nil
			}

			cost := 40000000

			if c.Gold < uint64(cost) {
				return nil, nil
			}

			itemData, _, err := c.AddItem(&database.InventorySlot{ItemID: 13000074, Quantity: 240}, -1, true)
			if err != nil {
				return nil, err
			}

			c.LootGold(-uint64(cost))
			resp.Concat(*itemData)
			resp.Concat(c.GetGold())
		case 3320: // Aid 8hr
			_, item, err := c.FindItemInInventory(nil, 13000173)
			if item != nil || err != nil {
				return nil, nil
			}

			cost := 180000000

			if c.Gold < uint64(cost) {
				return nil, nil
			}

			itemData, _, err := c.AddItem(&database.InventorySlot{ItemID: 13000173, Quantity: 480}, -1, true)
			if err != nil {
				return nil, err
			}

			c.LootGold(-uint64(cost))
			resp.Concat(*itemData)
			resp.Concat(c.GetGold())
		case 3321: // Aid 16hr
			_, item, err := c.FindItemInInventory(nil, 23000141)
			if item != nil || err != nil {
				return nil, nil
			}

			cost := 11195000000

			if c.Gold < uint64(cost) {
				return nil, nil
			}

			itemData, _, err := c.AddItem(&database.InventorySlot{ItemID: 23000141, Quantity: 960}, -1, true)
			if err != nil {
				return nil, err
			}

			c.LootGold(-uint64(cost))
			resp.Concat(*itemData)
			resp.Concat(c.GetGold())

		case 3307: // Ncash map
			resp, _ = c.ChangeMap(231, nil)
		case 3318: //production2
			resp = CO_PRODUCTION_MENU

		case 3426: //DISCITEM
			resp := utils.Packet{0xaa, 0x55, 0x03, 0x00, 0x57, 0x49, 0x01, 0x55, 0xaa}
			return resp, nil
		case 3930: //Marbles
			resp := utils.Packet{0xaa, 0x55, 0x03, 0x00, 0x57, 0x46, 0x01, 0x55, 0xaa}
			return resp, nil
		case 3308: //GOLD TO NCASH
		/*cost := 10000000
		if c.Gold < uint64(cost) {
			return nil, nil
		}
		user, err := database.FindUserByID(c.UserID)
		if err != nil {
			return nil, err
		} else if user == nil {
			return nil, nil
		}
		c.LootGold(-uint64(cost))
		resp.Concat(c.GetGold())
		//user.NCash += uint64(1000)
		//user.Update()*/
		case 197101: // Move to Marketplace
			resp, _ = c.ChangeMap(254, nil)
		//case 753:
		case 742: //GEUK MAPPOK
			if c.Level <= 100 {
				resp, _ = c.ChangeMap(93, nil)
			}
		case 743: //GEUK MAPPOK
			if c.Level <= 100 {
				resp, _ = c.ChangeMap(94, nil)
			}
		case 744: //GEUK MAPPOK
			if c.Level <= 100 {
				resp, _ = c.ChangeMap(95, nil)
			}
		case 745: //GEUK MAPPOK
			if c.Level > 100 {
				resp, _ = c.ChangeMap(96, nil)
			}
		case 746: //GEUK MAPPOK
			if c.Level > 100 {
				resp, _ = c.ChangeMap(97, nil)
			}
		case 747: //GEUK MAPPOK
			if c.Level > 100 {
				resp, _ = c.ChangeMap(98, nil)
			}
		case 748: //GEUK MAPPOK
			if c.Level > 100 {
				resp, _ = c.ChangeMap(99, nil)
			}
		case 3233: // NON DIVINE Sawangcheon
			if c.Level <= 100 {
				resp, _ = c.ChangeMap(100, nil)
			}
		case 3235: //DIVINE Sawangcheon
			if c.Level > 100 {
				resp, _ = c.ChangeMap(101, nil)
			}
		case 3309: //Become God of War
			book1, book2, job = 100030020, 100030021, 41
			book3 := 100032001
			jobName := "God of War"
			resp, err = darknessJobPromotion(c, book1, book2, book3, job, npcID, jobName)
			if err != nil {
				return nil, err
			}
		case 3310: //Become God of Death
			book1, book2, job = 100030022, 100030023, 42
			book3 := 100032002
			jobName := "God of Death"
			resp, err = darknessJobPromotion(c, book1, book2, book3, job, npcID, jobName)
			if err != nil {
				return nil, err
			}
		case 3311: //Become God of Blade
			book1, book2, job = 100030024, 100030025, 43
			book3 := 100032003
			jobName := "God of Blade"
			resp, err = darknessJobPromotion(c, book1, book2, book3, job, npcID, jobName)
			if err != nil {
				return nil, err
			}
		case 33115: //Ncash Coin to Ncash
			canChange := true
			reqCoinCount := uint(10)
			slotID, _, _ := s.Character.FindItemInInventory(nil, 100080180)
			slots, err := s.Character.InventorySlots()
			if err != nil {
				return nil, err
			}
			item := slots[slotID]
			if item.Quantity < reqCoinCount {
				canChange = false
			}
			user, err := database.FindUserByID(c.UserID)
			if err != nil {
				return nil, err
			} else if user == nil {
				return nil, nil
			}
			if canChange {
				itemData := c.DecrementItem(slotID, reqCoinCount)
				resp.Concat(*itemData)
				user.NCash += uint64(10)
				user.Update()
				resp.Concat(messaging.InfoMessage(fmt.Sprintf("Successful change coin to Ncash. New ncash amount: %d", user.NCash))) //NOTICE TO PROMOTE
			}
		case 3317:
			if c.Exp >= 233332051410 && c.Level >= 100 && c.HonorRank < 10 {
				rank, err := database.FindRankByHonorID(c.HonorRank)
				if err != nil {
					return nil, err
				}
				if rank == nil {
					return nil, nil
				}
				st := c.Socket.Stats
				if st == nil {
					return nil, nil
				}
				rankid := rank.ID + 1
				c.HonorRank = int64(database.HONOR_RANKS[rankid])
				resp := database.CHANGE_RANK
				resp.Insert(utils.IntToBytes(uint64(s.Character.PseudoID), 2, true), 6)
				resp.Insert(utils.IntToBytes(uint64(s.Character.HonorRank), 4, true), 8)
				err = st.Reset()
				if err != nil {
					return nil, err
				}
				statData, _ := s.Character.GetStats()
				resp.Concat(statData)
				s.Write(resp)
				c.Level = 1
				c.Exp = 0
				if c.Level <= 100 {
					s.Character.Type = c.Type
				} else if c.Level > 100 && c.Level <= 200 {
					s.Character.Type = c.Type - 10
				} else {
					s.Character.Type = c.Type - 20
				}
				s.Character.Class = 0
				s.Character.Map = 1
				c.Socket.Skills.Delete()
				c.Socket.Skills.Create(c)
				s.Skills.Delete()
				s.Skills.Create(c)
				c.Socket.Skills.SkillPoints = 0
				st.StatPoints = 4
				st.NaturePoints = 0
				go c.Socket.Skills.Update()
				c.Update()
				s.User.Update()
				CharacterSelect := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x01, 0x05, 0x0A, 0x00, 0x55, 0xAA}
				CHARACTER_MENU := utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x09, 0x09, 0x00, 0x55, 0xAA}
				cresp := CHARACTER_MENU
				cresp.Concat(CharacterSelect)
				data, _ := c.ChangeMap(1, nil)
				cresp.Concat(data)
				s.Conn.Write(cresp)
			}
		case 3322: //Teleport To Summer Zone {
			if c.Faction == 1 {
				x := 567.0
				y := 747.0
				data, _ := c.ChangeMap(37, database.ConvertPointToLocation(fmt.Sprintf("%.1f,%.1f", x, y)))
				c.Socket.Write(data)
			} else if c.Faction == 2 {
				x := 567.0
				y := 747.0
				data, _ := c.ChangeMap(37, database.ConvertPointToLocation(fmt.Sprintf("%.1f,%.1f", x, y)))
				c.Socket.Write(data)
			}

		case 281: //ATARAXIA
			if c.Exp >= 233332051410 && c.Level == 100 {
				if c.Class != 0 {
					ATARAXIA := utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x57, 0x21, 0xE3, 0x55, 0xAA, 0xaa, 0x55, 0x0b, 0x00, 0x75, 0x00, 0x01, 0x00, 0x80, 0xa1, 0x43, 0x00, 0x00, 0x3d, 0x43, 0x55, 0xaa}
					resp := ATARAXIA
					resp[6] = byte(c.Type + 10) // character type
					KIIRAS := utils.Packet{0xaa, 0x55, 0x54, 0x00, 0x71, 0x14, 0x51, 0x55, 0xAA}
					kiirasuzenet := "At this moment I mark my name on list of Top master in Strong HERO."
					kiirasresp := KIIRAS
					index := 6
					kiirasresp[index] = byte(len(c.Name) + len(kiirasuzenet))
					index++
					kiirasresp.Insert([]byte("["+c.Name+"]"), index) // character name
					index += len(c.Name) + 2
					kiirasresp.Insert([]byte(kiirasuzenet), index) // character name
					kiirasresp.SetLength(int16(binary.Size(kiirasresp) - 6))

					resp.Concat(kiirasresp)
					c.Socket.Write(resp)
					c.Level = 100
					c.Type += 10
					c.Update()
					c.AddExp(1)
					c.Level = 101
					c.Socket.Skills.Delete()
					c.Socket.Skills.Create(c)
					s.Skills.Delete()
					s.Skills.Create(c)
					c.Update()
					s.User.Update()
					resp, _ = divineJobPromotion(c, npcID)
					statData, _ := c.GetStats()
					resp.Concat(statData)
					s.Conn.Write(resp)

				} else {

					resp.Concat(messaging.InfoMessage(fmt.Sprintf("You don't have class."))) //NOTICE TO NO SELECTED CLASS
				}
			}
		case 232: //Shin-Gang Region
			if c.Faction == 1 {
				resp, _ = c.ChangeMap(22, nil)
			} else if c.Faction == 2 {
				resp, _ = c.ChangeMap(23, nil)
			}
		}
		return resp, nil
	}
}

func GetNPCMenu(npcID, textID, index int, actions []int) []byte {
	resp := NPC_MENU
	resp.Insert(utils.IntToBytes(uint64(npcID), 4, true), 6)         // npc id
	resp.Insert(utils.IntToBytes(uint64(textID), 4, true), 10)       // text id
	resp.Insert(utils.IntToBytes(uint64(len(actions)), 1, true), 14) // action length

	counter, length := 15, int16(11)
	for i, action := range actions {
		resp.Insert(utils.IntToBytes(uint64(action), 4, true), counter) // action
		counter += 4

		resp.Insert(utils.IntToBytes(uint64(npcID), 2, true), counter) // npc id
		counter += 2

		actIndex := int(index) + (i+1)<<(len(actions)*3)
		resp.Insert(utils.IntToBytes(uint64(actIndex), 2, true), counter) // action index
		counter += 2

		length += 8
	}

	resp.SetLength(length)
	return resp
}

func firstJobPromotion(c *database.Character, book, job, npcID int) (utils.Packet, error) {
	resp := utils.Packet{}
	if c.Class == 0 && c.Level >= 10 {
		c.Class = job
		resp = JOB_PROMOTED
		resp[6] = byte(job)

		r, _, err := c.AddItem(&database.InventorySlot{ItemID: int64(book), Quantity: 1}, -1, false)
		if err != nil {
			return resp, err
		} else if r == nil {
			return nil, nil
		}

		resp.Concat(*r)

	} else if c.Level < 10 {
		resp = NOT_ENOUGH_LEVEL
		resp.Insert(utils.IntToBytes(uint64(npcID), 4, true), 6) // npc id
	} else {
		resp = INVALID_CLASS
		resp.Insert(utils.IntToBytes(uint64(npcID), 4, true), 6) // npc id
	}

	return resp, nil
}

func secondJobPromotion(c *database.Character, book1, book2, preJob, job, npcID int) (utils.Packet, error) {
	resp := utils.Packet{}
	if c.Class == preJob && c.Level >= 50 {
		c.Class = job
		resp = JOB_PROMOTED
		resp[6] = byte(job)

		r, _, err := c.AddItem(&database.InventorySlot{ItemID: int64(book1), Quantity: 1}, -1, false)
		if err != nil {
			return resp, err
		} else if r == nil {
			return nil, nil
		}

		resp.Concat(*r)

		r, _, err = c.AddItem(&database.InventorySlot{ItemID: int64(book2), Quantity: 1}, -1, false)
		if err != nil {
			return resp, err
		} else if r == nil {
			return nil, nil
		}

		resp.Concat(*r)

	} else if c.Level < 50 {
		resp := NOT_ENOUGH_LEVEL
		resp.Insert(utils.IntToBytes(uint64(npcID), 4, true), 6) // npc id
	} else {
		resp = INVALID_CLASS
		resp.Insert(utils.IntToBytes(uint64(npcID), 4, true), 6) // npc id
	}

	return resp, nil
}

func divineJobPromotion(c *database.Character, npcID int) (utils.Packet, error) {
	resp := utils.Packet{}
	if c.Class != 0 {
		jobName := ""
		book1 := 0
		book2 := 0
		book3 := 0
		if c.Class == 21 || c.Class == 22 { //WARLORD
			c.Class = 31
			book1 = 100031001
			book2 = 100030013
			book3 = 16100300
			jobName = "Warlord"
		} else if c.Class == 23 || c.Class == 24 { //Holy Hand
			c.Class = 32
			book1 = 100031003
			book2 = 100030015
			book3 = 16100300
			jobName = "HolyHand"
		} else if c.Class == 25 || c.Class == 26 { //BeastLord
			c.Class = 33
			book1 = 100031002
			book2 = 100030014
			book3 = 16100300
			jobName = "BeastLord"
		} else if c.Class == 27 || c.Class == 28 { //ShadowLord
			c.Class = 34
			book1 = 100031004
			book2 = 100030016
			book3 = 16100300
			jobName = "ShadowLord"
		}
		c.Update()
		resp = JOB_PROMOTED
		resp[6] = byte(c.Class)

		r, _, err := c.AddItem(&database.InventorySlot{ItemID: int64(book1), Quantity: 1}, -1, false)
		if err != nil {
			return resp, err
		} else if r == nil {
			return nil, nil
		}

		resp.Concat(*r)

		r, _, err = c.AddItem(&database.InventorySlot{ItemID: int64(book2), Quantity: 1}, -1, false)
		if err != nil {
			return resp, err
		} else if r == nil {
			return nil, nil
		}

		resp.Concat(*r)

		r, _, err = c.AddItem(&database.InventorySlot{ItemID: int64(book3), Quantity: 1}, -1, false)
		if err != nil {
			return resp, err
		} else if r == nil {
			return nil, nil
		}

		resp.Concat(*r)

		resp.Concat(messaging.InfoMessage(fmt.Sprintf("Promoted as a %s.", jobName))) //NOTICE TO PROMOTE
		//r = c.ChangeMap(1, nil)
	} else {
		resp = INVALID_CLASS
		resp.Insert(utils.IntToBytes(uint64(npcID), 4, true), 6) // npc id
	}
	return resp, nil
}

func darknessJobPromotion(c *database.Character, book1, book2, book3, job, npcID int, jobName string) (utils.Packet, error) {
	resp := utils.Packet{}
	if c.Class >= 30 && c.Class < 40 {
		c.Class = job
		c.Update()
		resp = JOB_PROMOTED
		resp[6] = byte(c.Class)

		r, _, err := c.AddItem(&database.InventorySlot{ItemID: int64(book1), Quantity: 1}, -1, false)
		if err != nil {
			return resp, err
		} else if r == nil {
			return nil, nil
		}

		resp.Concat(*r)

		r, _, err = c.AddItem(&database.InventorySlot{ItemID: int64(book2), Quantity: 1}, -1, false)
		if err != nil {
			return resp, err
		} else if r == nil {
			return nil, nil
		}

		resp.Concat(*r)

		r, _, err = c.AddItem(&database.InventorySlot{ItemID: int64(book3), Quantity: 1}, -1, false)
		if err != nil {
			return resp, err
		} else if r == nil {
			return nil, nil
		}

		resp.Concat(*r)
		c.Update()
		resp.Concat(messaging.InfoMessage(fmt.Sprintf("Promoted as a %s.", jobName))) //NOTICE TO PROMOTE
		//r = c.ChangeMap(1, nil)
	} else {
		resp = INVALID_CLASS
		resp.Insert(utils.IntToBytes(uint64(npcID), 4, true), 6) // npc id
	}
	return resp, nil
}
