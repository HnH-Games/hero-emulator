package player

import (
	"fmt"
	"log"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"hero-emulator/database"
	"hero-emulator/dungeon"
	"hero-emulator/messaging"
	"hero-emulator/nats"
	"hero-emulator/npc"
	"hero-emulator/server"
	"hero-emulator/utils"

	"github.com/robfig/cron"
	"gopkg.in/guregu/null.v3"

	"github.com/thoas/go-funk"
)

type ChatHandler struct {
	chatType  int64
	message   string
	receivers map[int]*database.Character
}

var (
	CHAT_MESSAGE  = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x00, 0x55, 0xAA}
	SHOUT_MESSAGE = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x71, 0x0E, 0x00, 0x00, 0x55, 0xAA}
	ANNOUNCEMENT  = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x71, 0x06, 0x00, 0x55, 0xAA}
	EVENT_NOTICE  = utils.Packet{0xAA, 0x55, 0x00, 0x75, 0x01, 0x00, 0x80, 0xa1, 0x00, 0x55, 0xAA}
)

func (h *ChatHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if s.Character == nil {
		return nil, nil
	}

	user, err := database.FindUserByID(s.Character.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	stat := s.Stats
	if stat == nil {
		return nil, nil
	}

	h.chatType = utils.BytesToInt(data[4:6], false)

	switch h.chatType {
	case 28929: // normal chat
		messageLen := utils.BytesToInt(data[6:8], true)
		h.message = string(data[8 : messageLen+8])

		return h.normalChat(s)
	case 28930: // private chat
		index := 6
		recNameLength := int(data[index])
		index++

		recName := string(data[index : index+recNameLength])
		index += recNameLength

		c, err := database.FindCharacterByName(recName)
		if err != nil {
			return nil, err
		} else if c == nil {
			return messaging.SystemMessage(messaging.WHISPER_FAILED), nil
		}

		h.receivers = map[int]*database.Character{c.ID: c}

		messageLen := int(utils.BytesToInt(data[index:index+2], true))
		index += 2

		h.message = string(data[index : index+messageLen])
		return h.chatWithReceivers(s, h.createChatMessage)

	case 28931: // party chat
		party := database.FindParty(s.Character)
		if party == nil {
			return nil, nil
		}

		messageLen := int(utils.BytesToInt(data[6:8], true))
		h.message = string(data[8 : messageLen+8])

		members := funk.Filter(party.GetMembers(), func(m *database.PartyMember) bool {
			return m.Accepted
		}).([]*database.PartyMember)
		members = append(members, &database.PartyMember{Character: party.Leader, Accepted: true})

		h.receivers = map[int]*database.Character{}
		for _, m := range members {
			if m.ID == s.Character.ID {
				continue
			}

			h.receivers[m.ID] = m.Character
		}

		return h.chatWithReceivers(s, h.createChatMessage)

	case 28932: // guild chat
		if s.Character.GuildID > 0 {
			guild, err := database.FindGuildByID(s.Character.GuildID)
			if err != nil {
				return nil, err
			}

			members, err := guild.GetMembers()
			if err != nil {
				return nil, err
			}

			messageLen := int(utils.BytesToInt(data[6:8], true))
			h.message = string(data[8 : messageLen+8])
			h.receivers = map[int]*database.Character{}

			for _, m := range members {
				c, err := database.FindCharacterByID(m.ID)
				if err != nil || c == nil || !c.IsOnline || c.ID == s.Character.ID {
					continue
				}

				h.receivers[m.ID] = c
			}

			return h.chatWithReceivers(s, h.createChatMessage)
		}

	case 28933, 28946: // roar chat
		if stat.CHI < 100 || time.Now().Sub(s.Character.LastRoar) < 10*time.Second {
			return nil, nil
		}

		s.Character.LastRoar = time.Now()
		characters, err := database.FindCharactersInServer(user.ConnectedServer)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		//delete(characters, s.Character.ID)
		h.receivers = characters

		stat.CHI -= 100

		index := 6
		messageLen := int(utils.BytesToInt(data[index:index+2], true))
		index += 2

		h.message = string(data[index : index+messageLen])

		resp := utils.Packet{}
		_, err = h.chatWithReceivers(s, h.createChatMessage)
		if err != nil {
			log.Println(err)
			return nil, err
		}

		//resp.Concat(chat)
		resp.Concat(s.Character.GetHPandChi())
		return resp, nil

	case 28935: // commands
		index := 6
		messageLen := int(data[index])
		index++

		h.message = string(data[index : index+messageLen])
		return h.cmdMessage(s, data)

	case 28943: // shout
		return h.Shout(s, data)

	case 28945: // faction chat
		characters, err := database.FindCharactersInServer(user.ConnectedServer)
		if err != nil {
			return nil, err
		}

		//delete(characters, s.Character.ID)
		for _, c := range characters {
			if c.Faction != s.Character.Faction {
				delete(characters, c.ID)
			}
		}

		h.receivers = characters
		index := 6
		messageLen := int(utils.BytesToInt(data[index:index+2], true))
		index += 2

		h.message = string(data[index : index+messageLen])
		resp := utils.Packet{}
		_, err = h.chatWithReceivers(s, h.createChatMessage)
		if err != nil {
			return nil, err
		}

		//resp.Concat(chat)
		return resp, nil

	}

	return nil, nil
}

func (h *ChatHandler) Shout(s *database.Socket, data []byte) ([]byte, error) {
	if time.Now().Sub(s.Character.LastRoar) < 10*time.Second {
		return nil, nil
	}

	characters, err := database.FindOnlineCharacters()
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//delete(characters, s.Character.ID)

	slot, _, err := s.Character.FindItemInInventory(nil, 15900001, 17500181, 17502689, 13000131)
	if err != nil {
		log.Println(err)
		return nil, err
	} else if slot == -1 {
		return nil, nil
	}

	resp := s.Character.DecrementItem(slot, 1)

	index := 6
	messageLen := int(data[index])
	index++

	h.chatType = 28942
	h.receivers = characters
	h.message = string(data[index : index+messageLen])

	_, err = h.chatWithReceivers(s, h.createShoutMessage)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	//resp.Concat(chat)
	return *resp, nil
}

func (h *ChatHandler) createChatMessage(s *database.Socket) *utils.Packet {

	resp := CHAT_MESSAGE

	index := 4
	resp.Insert(utils.IntToBytes(uint64(h.chatType), 2, false), index) // chat type
	index += 2

	if h.chatType != 28946 {
		resp.Insert(utils.IntToBytes(uint64(s.Character.PseudoID), 2, true), index) // sender character pseudo id
		index += 2
	}

	resp[index] = byte(len(s.Character.Name)) // character name length
	index++

	resp.Insert([]byte(s.Character.Name), index) // character name
	index += len(s.Character.Name)

	resp.Insert(utils.IntToBytes(uint64(len(h.message)), 2, true), index) // message length
	index += 2

	resp.Insert([]byte(h.message), index) // message
	index += len(h.message)

	length := index - 4
	resp.SetLength(int16(length)) // packet length

	return &resp
}

func (h *ChatHandler) createShoutMessage(s *database.Socket) *utils.Packet {

	resp := SHOUT_MESSAGE
	length := len(s.Character.Name) + len(h.message) + 6
	resp.SetLength(int16(length)) // packet length

	index := 4
	resp.Insert(utils.IntToBytes(uint64(h.chatType), 2, false), index) // chat type
	index += 2

	resp[index] = byte(len(s.Character.Name)) // character name length
	index++

	resp.Insert([]byte(s.Character.Name), index) // character name
	index += len(s.Character.Name)

	resp[index] = byte(len(h.message)) // message length
	index++

	resp.Insert([]byte(h.message), index) // message
	return &resp
}

func (h *ChatHandler) normalChat(s *database.Socket) ([]byte, error) {

	if _, ok := server.MutedPlayers.Get(s.User.ID); ok {
		msg := "Chatting with this account is prohibited. Please contact our customer support service for more information."
		return messaging.InfoMessage(msg), nil
	}

	resp := h.createChatMessage(s)
	p := &nats.CastPacket{CastNear: true, CharacterID: s.Character.ID, Data: *resp, Type: nats.CHAT_NORMAL}
	err := p.Cast()

	return nil, err
}

func (h *ChatHandler) chatWithReceivers(s *database.Socket, msgHandler func(*database.Socket) *utils.Packet) ([]byte, error) {

	if _, ok := server.MutedPlayers.Get(s.User.ID); ok {
		msg := "Chatting with this account is prohibited. Please contact our customer support service for more information."
		return messaging.InfoMessage(msg), nil
	}

	resp := msgHandler(s)

	for _, c := range h.receivers {
		if c == nil || !c.IsOnline {
			if h.chatType == 28930 { // PM
				return messaging.SystemMessage(messaging.WHISPER_FAILED), nil
			}
			continue
		}

		socket := database.GetSocket(c.UserID)
		if socket != nil {
			_, err := socket.Conn.Write(*resp)
			if err != nil {
				log.Println(err)
				return nil, err
			}
		}
	}

	return *resp, nil
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

func (h *ChatHandler) cmdMessage(s *database.Socket, data []byte) ([]byte, error) {

	var (
		err  error
		resp utils.Packet
	)

	if parts := strings.Split(h.message, " "); len(parts) > 0 {
		cmd := strings.ToLower(strings.TrimPrefix(parts[0], "/"))
		switch cmd {
		case "shout":
			return h.Shout(s, data)

		case "announce":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			msg := strings.Join(parts[1:], " ")
			makeAnnouncement(msg)
		case "event":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}
			return EVENT_NOTICE, nil
		case "deleteitemslot":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}
			slotID, err := strconv.ParseInt(parts[1], 10, 16)
			slotMax := int16(slotID)
			if err != nil {
				return nil, err
			}
			ch := s.Character
			if len(parts) >= 3 {
				chr, _ := database.FindCharacterByName(parts[2])
				ch = chr
			}
			r, err := ch.RemoveItem(slotMax)
			if err != nil {
				return nil, err
			}
			ch.Socket.Write(r)
		case "discitem":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			itemID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}

			quantity := int64(1)
			itemtype := int64(0)
			judgestat := int64(0)
			info := database.Items[itemID]
			if info.Timer > 0 {
				quantity = int64(info.Timer)
			}
			if len(parts) >= 3 {
				quantity, err = strconv.ParseInt(parts[2], 10, 64)
				if err != nil {
					return nil, err
				}
			}
			if len(parts) >= 4 {
				itemtype, err = strconv.ParseInt(parts[3], 10, 64)
				if err != nil {
					return nil, err
				}
			}
			if len(parts) >= 5 {
				judgestat, err = strconv.ParseInt(parts[4], 10, 64)
				if err != nil {
					return nil, err
				}
			}
			ch := s.Character
			if s.User.UserType >= server.HGM_USER {
				if len(parts) >= 6 {
					chID, err := strconv.ParseInt(parts[5], 10, 64)
					if err == nil {
						chr, err := database.FindCharacterByID(int(chID))
						if err == nil {
							ch = chr
						}
					}
				}
			}

			item := &database.InventorySlot{ItemID: itemID, Quantity: uint(quantity), ItemType: int16(itemtype), JudgementStat: int64(judgestat)}

			if info.GetType() == database.PET_TYPE {
				petInfo := database.Pets[itemID]
				expInfo := database.PetExps[petInfo.Level-1]
				targetExps := []int{expInfo.ReqExpEvo1, expInfo.ReqExpEvo2, expInfo.ReqExpEvo3, expInfo.ReqExpHt, expInfo.ReqExpDivEvo1, expInfo.ReqExpDivEvo2, expInfo.ReqExpDivEvo3}
				item.Pet = &database.PetSlot{
					Fullness: 100, Loyalty: 100,
					Exp:   uint64(targetExps[petInfo.Evolution-1]),
					HP:    petInfo.BaseHP,
					Level: byte(petInfo.Level),
					Name:  "",
					CHI:   petInfo.BaseChi}
			}

			r, _, err := ch.AddItem(item, -1, false)
			if err != nil {
				return nil, err
			}
			ch.Socket.Write(*r)
			return nil, nil
		case "item":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			itemID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}

			quantity := int64(1)
			if len(parts) >= 3 {
				quantity, err = strconv.ParseInt(parts[2], 10, 64)
				if err != nil {
					return nil, err
				}
			}

			ch := s.Character
			if len(parts) >= 4 {
				chID, err := strconv.ParseInt(parts[3], 10, 64)
				if err == nil {
					chr, err := database.FindCharacterByID(int(chID))
					if err == nil {
						ch = chr
					}
				}
			}
			info := database.Items[itemID]
			if info.Timer > 0 {
				quantity = int64(info.Timer)
			}
			item := &database.InventorySlot{ItemID: itemID, Quantity: uint(quantity)}
			if len(parts) >= 3 {
				quantity, err = strconv.ParseInt(parts[2], 10, 64)
				if err != nil {
					return nil, err
				}
			}

			if info.GetType() == database.PET_TYPE {
				petInfo := database.Pets[itemID]
				expInfo := database.PetExps[petInfo.Level-1]
				targetExps := []int{expInfo.ReqExpEvo1, expInfo.ReqExpEvo2, expInfo.ReqExpEvo3, expInfo.ReqExpHt, expInfo.ReqExpDivEvo1, expInfo.ReqExpDivEvo2, expInfo.ReqExpDivEvo3}
				item.Pet = &database.PetSlot{
					Fullness: 100, Loyalty: 100,
					Exp:   uint64(targetExps[petInfo.Evolution-1]),
					HP:    petInfo.BaseHP,
					Level: byte(petInfo.Level),
					Name:  petInfo.Name,
					CHI:   petInfo.BaseChi}
			}

			r, _, err := ch.AddItem(item, -1, false)
			if err != nil {
				return nil, err
			}

			ch.Socket.Write(*r)
			return nil, nil

		case "rank":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			rankID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}
			s.Character.HonorRank = rankID
			s.Character.Update()
			resp := database.CHANGE_RANK
			resp.Insert(utils.IntToBytes(uint64(s.Character.PseudoID), 2, true), 6)
			resp.Insert(utils.IntToBytes(uint64(s.Character.HonorRank), 4, true), 8)
			statData, _ := s.Character.GetStats()
			resp.Concat(statData)
			s.Write(resp)
		case "divine":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}
			data, levelUp := s.Character.AddExp(233332051410)
			if levelUp {
				statData, err := s.Character.GetStats()
				if err == nil && s.Character.Socket != nil {
					resp.Concat(statData)
				}
			}
			if s.Character.Socket != nil {
				resp.Concat(data)
			}
			s.Character.Class = 21
			s.Character.Update()
			gomap, _ := s.Character.ChangeMap(14, nil)
			resp.Concat(gomap)
			x := "261,420"
			coord := s.Character.Teleport(database.ConvertPointToLocation(x))
			resp.Concat(coord)
			return resp, nil
		case "class":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}

			id, _ := strconv.Atoi(parts[1])
			c, err := database.FindCharacterByID(int(id))
			if err != nil {
				return nil, err
			}

			t, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, err
			}
			c.Class = t
			c.Update()
			resp := utils.Packet{}
			resp = npc.JOB_PROMOTED
			resp[6] = byte(c.Class)
			return resp, nil
		case "gold":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			amount, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}

			s.Character.Gold += uint64(amount)
			h := &GetGoldHandler{}

			return h.Handle(s)

		case "upgrade":
			if s.User.UserType < server.GM_USER || len(parts) < 3 {
				return nil, nil
			}

			slots, err := s.Character.InventorySlots()
			if err != nil {
				return nil, err
			}

			slotID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}

			code, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return nil, err
			}

			count := int64(1)
			if len(parts) > 3 {
				count, err = strconv.ParseInt(parts[3], 10, 64)
				if err != nil {
					return nil, err
				}
			}

			codes := []byte{}
			for i := 0; i < int(count); i++ {
				codes = append(codes, byte(code))
			}

			item := slots[slotID]
			return item.Upgrade(int16(slotID), codes...), nil

		case "exp":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			amount, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}

			ch := s.Character
			if len(parts) > 2 {
				chID, err := strconv.ParseInt(parts[2], 10, 64)
				if err == nil {
					chr, err := database.FindCharacterByID(int(chID))
					if err == nil {
						ch = chr
					}
				}
			}

			data, levelUp := ch.AddExp(amount)
			if levelUp {
				statData, err := ch.GetStats()
				if err == nil && ch.Socket != nil {
					ch.Socket.Write(statData)
				}
			}

			if ch.Socket != nil {
				ch.Socket.Write(data)
			}

			return nil, nil
		case "petexp":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			amount, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}
			slots, err := s.Character.InventorySlots()
			if err != nil {
				log.Println(err)
				return nil, nil
			}

			petSlot := slots[0x0A]
			pet := petSlot.Pet
			if pet == nil || petSlot.ItemID == 0 || !pet.IsOnline {
				return nil, nil
			}
			pet.AddExp(s.Character, amount)

		case "map":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			mapID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}

			if len(parts) >= 3 {
				c, err := database.FindCharacterByName(parts[2])
				if err != nil {
					return nil, err
				}

				data, err := c.ChangeMap(int16(mapID), nil)
				if err != nil {
					return nil, err
				}

				database.GetSocket(c.UserID).Write(data)
				return nil, nil
			}

			return s.Character.ChangeMap(int16(mapID), nil)
		case "buff":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}
			infectionID, err := strconv.ParseUint(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}
			duration, err := strconv.ParseUint(parts[1], 10, 32)
			if err != nil {
				return nil, err
			}
			s.Character.NewBuffInfection(int64(infectionID), int(duration))
		case "cash":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}

			amount, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return nil, err
			}

			userID := parts[1]
			user, err := database.FindUserByName(userID)
			fmt.Println(userID)
			fmt.Println(user)
			fmt.Println(user.NCash)

			if err != nil {
				return nil, err
			} else if user == nil {
				return nil, nil
			}

			user.NCash += uint64(amount)
			user.Update()

			return messaging.InfoMessage(fmt.Sprintf("%d nCash loaded to %s (%s).", amount, user.Username, user.ID)), nil
		case "exprate":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}

			if len(parts) > 2 {
				if am, err := strconv.ParseFloat(parts[1], 64); err == nil {
					database.EXP_RATE = am

				}
				minute, err := strconv.ParseInt(parts[2], 10, 64)
				if err != nil {
					return nil, err

				}
				time.AfterFunc(time.Duration(minute)*time.Minute, func() {
					database.EXP_RATE = database.DEFAULT_EXP_RATE
				})
			}
			return messaging.InfoMessage(fmt.Sprintf("EXP Rate now: %f", database.EXP_RATE)), nil
		case "droprate":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}
			if len(parts) > 2 {
				if s, err := strconv.ParseFloat(parts[1], 64); err == nil {
					database.DROP_RATE = s
				}
				minute, err := strconv.ParseInt(parts[2], 10, 64)
				if err != nil {
					return nil, err
				}
				time.AfterFunc(time.Duration(minute)*time.Minute, func() {
					database.DROP_RATE = database.DEFAULT_DROP_RATE
				})
			}
			return messaging.InfoMessage(fmt.Sprintf("Drop Rate now: %f", database.DROP_RATE)), nil

		case "addguild":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}
			if len(parts) < 2 {
				return nil, nil
			}
			guildid, err := strconv.ParseInt(parts[1], 10, 32)
			if err != nil {
				return nil, err
			}
			ch := s.Character
			if len(parts) >= 3 {
				c, err := database.FindCharacterByName(parts[2])
				if err != nil {
					return nil, err
				}
				ch = c
			}
			removeid := -1
			if int(guildid) == removeid {
				guild, err := database.FindGuildByID(ch.GuildID)
				err = guild.RemoveMember(ch.ID)
				if err != nil {
					return nil, err
				}
				go guild.Update()
				ch.GuildID = int(guildid)
			} else {
				guild, err := database.FindGuildByID(int(guildid))
				if err != nil {
					return nil, err
				}
				guild.AddMember(&database.GuildMember{ID: ch.ID, Role: database.GROLE_MEMBER})
				go guild.Update()
				ch.GuildID = int(guildid)
			}
			spawnData, err := s.Character.SpawnCharacter()
			if err == nil {
				p := nats.CastPacket{CastNear: true, CharacterID: s.Character.ID, Type: nats.PLAYER_SPAWN, Data: spawnData}
				p.Cast()
				ch.Socket.Conn.Write(spawnData)
			}
			resp.Concat(messaging.InfoMessage(fmt.Sprintf("Player new guild id: %d", ch.GuildID)))
			return resp, nil
		case "findguild":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}
			guild, err := database.FindGuildByName(parts[1])
			if err != nil {
				return nil, err
			}
			return messaging.InfoMessage(fmt.Sprintf("Clan ID: %d", guild.ID)), nil
		case "dungeon":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}
			data, err := s.Character.ChangeMap(243, nil)
			if err != nil {
				return nil, err
			}
			resp.Concat(data)
			x := "377,246"
			coord := s.Character.Teleport(database.ConvertPointToLocation(x))
			resp.Concat(coord)
			return resp, nil
		case "refresh":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}
			if len(parts) < 2 {
				return nil, nil
			}
			command := parts[1]
			switch command {
			case "items":
				database.RefreshAllItems()
			case "scripts":
				database.RefreshScripts()
			case "htshop":
				database.RefreshHTItems()
			case "buffinf":
				database.RefreshBuffInfections()
			case "advancedfusions":
				database.RefreshAdvancedFusions()
			case "gamblings":
				database.RefreshGamblingItems()
			case "craftitems":
				database.RefreshCraftItem()
			case "productions":
				database.RefreshProductions()
			case "drops":
				database.RefreshAllDrops()
			case "shopitems":
				database.GetAllShopItems()
			case "npc":
				database.RefreshScripts()
				database.GetAllNPCs()
				database.GetAllNPCPos()
			case "exp":
				database.GetExps()
			case "users":
				database.RefreshUsers()
			case "npcpos":
				database.GetAllNPCPos()

			case "all":
				callBacks := []func() error{database.RefreshScripts, database.RefreshHTItems, database.RefreshBuffInfections, database.RefreshAdvancedFusions,
					database.RefreshGamblingItems, database.RefreshCraftItem, database.RefreshProductions, database.RefreshAllDrops, database.GetExps}
				for _, cb := range callBacks {
					if err := cb(); err != nil {
						fmt.Println("Error: ", err)
					}
				}
			}
		case "charinfo":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}
			c, err := database.FindCharacterByName(parts[1])
			if err != nil {
				return nil, err
			}

			resp.Concat(messaging.InfoMessage(fmt.Sprintf("%s player details:", c.Name)))
			resp.Concat(messaging.InfoMessage(fmt.Sprintf("CharID: %d | UserName: %s", c.Socket.Character.ID, c.Socket.User.Username)))
			resp.Concat(messaging.InfoMessage(fmt.Sprintf("Map: %d | Location: %s", c.Map, c.Coordinate)))
			resp.Concat(messaging.InfoMessage(fmt.Sprintf("Level: %d | Exp: %d", c.Level, c.Exp)))
			resp.Concat(messaging.InfoMessage(fmt.Sprint("Gold: ", c.Gold)))
			resp.Concat(messaging.InfoMessage(fmt.Sprint("Bank Gold: ", c.Socket.User.BankGold)))
			resp.Concat(messaging.InfoMessage(fmt.Sprint("Ncash: ", c.Socket.User.NCash)))
			resp.Concat(messaging.InfoMessage(fmt.Sprintf("AID: %d | AID-enabled:%t", c.AidTime, c.AidMode)))
			resp.Concat(messaging.InfoMessage(fmt.Sprint("SkillPoints: ", c.Socket.Skills.SkillPoints)))

		case "number":
			if len(parts) < 2 && s.Character.GeneratedNumber != 2 {
				return nil, nil
			}
			number, err := strconv.ParseInt(parts[1], 10, 32)
			if err != nil {
				return nil, err
			}
			if int(number) == s.Character.GeneratedNumber {
				s.Conn.Write(messaging.InfoMessage(fmt.Sprintf("You guessed right, Show the boss your power.")))
				s.Character.DungeonLevel++
			} else {
				s.Conn.Write(messaging.InfoMessage(fmt.Sprintf("You guessed poorly, survive & slay again!")))
				dungeon.MobsCreate([]int{40522}, s.User.ConnectedServer)
				s.Character.CanTip = 3
			}
		case "addmobs":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}
			npcId, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}
			count, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return nil, err
			}
			mapID := s.Character.Map
			cmdSpawnMobs(int(count), int(npcId), int(mapID), s.Character.Coordinate)
		case "resetallmobs":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			for _, npcPos := range database.NPCPos {
				npc, ok := database.NPCs[npcPos.NPCID]
				if !ok {
					fmt.Println("Error")
					continue
				}
				for k := 1; k <= 4; k++ {
					for i := 0; i < int(npcPos.Count); i++ {
						if npc.ID == 0 || npcPos.IsNPC || !ok || !npcPos.Attackable {
							continue
						}
						minCoordinate := database.ConvertPointToLocation(npcPos.MinLocation)
						maxCoordinate := database.ConvertPointToLocation(npcPos.MaxLocation)
						targetX := utils.RandFloat(minCoordinate.X, maxCoordinate.X)
						targetY := utils.RandFloat(minCoordinate.Y, maxCoordinate.Y)
						target := utils.Location{X: targetX, Y: targetY}
						newai := &database.AI{ID: len(database.AIs), HP: npc.MaxHp, Map: npcPos.MapID, PosID: npcPos.ID, RunningSpeed: float64(3), Server: k, WalkingSpeed: float64(3), Faction: npcPos.Faction}
						server.GenerateIDForAI(newai)
						newai.OnSightPlayers = make(map[int]interface{})
						newai.Coordinate = target.String()
						uploadAI := &database.AI{ID: len(database.AIs), PosID: npcPos.ID, Server: k, Faction: npcPos.Faction, Map: npcPos.MapID, Coordinate: newai.Coordinate, WalkingSpeed: float64(3), RunningSpeed: float64(3)}
						//fmt.Println(newai.Coordinate)
						aierr := uploadAI.Create()
						if aierr != nil {
							fmt.Println("Error: %s", aierr)
						}
						newai.Handler = newai.AIHandler
						database.AIsByMap[newai.Server][npcPos.MapID] = append(database.AIsByMap[newai.Server][npcPos.MapID], newai)
						database.AIs[newai.ID] = newai
						fmt.Println("New mob created", len(database.AIs))
						go newai.Handler()
					}
				}
			}
			fmt.Println("Finished")
			return nil, nil
		case "mob":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			posId, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}

			npcPos := database.NPCPos[int(posId)]
			npc, ok := database.NPCs[npcPos.NPCID]
			if !ok {
				return nil, nil
			}

			ai := &database.AI{ID: len(database.AIs), HP: npc.MaxHp, Map: npcPos.MapID, PosID: npcPos.ID, RunningSpeed: 10, Server: 1, WalkingSpeed: 5, Once: true}
			server.GenerateIDForAI(ai)
			ai.OnSightPlayers = make(map[int]interface{})

			minLoc := database.ConvertPointToLocation(npcPos.MinLocation)
			maxLoc := database.ConvertPointToLocation(npcPos.MaxLocation)
			loc := utils.Location{X: utils.RandFloat(minLoc.X, maxLoc.X), Y: utils.RandFloat(minLoc.Y, maxLoc.Y)}
			ai.Coordinate = loc.String()
			fmt.Println(ai.Coordinate)
			ai.Handler = ai.AIHandler
			go ai.Handler()

			makeAnnouncement(fmt.Sprintf("%s has been roaring.", npc.Name))

			database.AIsByMap[ai.Server][npcPos.MapID] = append(database.AIsByMap[ai.Server][npcPos.MapID], ai)
			database.AIs[ai.ID] = ai

		case "droplog":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}
			resp.Concat(messaging.InfoMessage(fmt.Sprintf("Today farmed relics: %d ea", len(database.RelicsLog))))
			for _, c := range database.RelicsLog {
				hour, min, sec := c.DropTime.Time.Hour(), c.DropTime.Time.Minute(), c.DropTime.Time.Second()
				resp.Concat(messaging.InfoMessage(fmt.Sprintf("Character ID: %d dropped item id: %d at %d:%d:%d ", c.CharID, c.ItemID, hour, min, sec)))
			}

		case "relic":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			itemID, err := strconv.ParseInt(parts[1], 10, 64)
			if err != nil {
				return nil, err
			}

			ch := s.Character
			if len(parts) >= 3 {
				chID, err := strconv.ParseInt(parts[2], 10, 64)
				if err == nil {
					chr, err := database.FindCharacterByID(int(chID))
					if err == nil {
						ch = chr
					}
				}
			}

			slot, err := ch.FindFreeSlot()
			if err != nil {
				return nil, nil
			}

			itemData, _, _ := ch.AddItem(&database.InventorySlot{ItemID: itemID, Quantity: 1}, slot, true)
			if itemData != nil {
				ch.Socket.Write(*itemData)

				relicDrop := ch.RelicDrop(int64(itemID))
				p := nats.CastPacket{CastNear: false, Data: relicDrop, Type: nats.ITEM_DROP}
				p.Cast()
			}

		case "main":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			countMaintenance(60)

		case "ban":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}

			userID := parts[1]
			user, err := database.FindUserByID(userID)
			if err != nil {
				return nil, err
			} else if user == nil {
				return nil, nil
			}

			hours, err := strconv.ParseInt(parts[2], 10, 64)
			if err != nil {
				return nil, err
			}

			user.UserType = 0
			user.DisabledUntil = null.NewTime(time.Now().Add(time.Hour*time.Duration(hours)), true)
			user.Update()

			database.GetSocket(userID).Conn.Close()

		case "mute":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			dumb, err := database.FindCharacterByName(parts[1])
			if err != nil {
				return nil, err
			}

			server.MutedPlayers.Set(dumb.UserID, struct{}{})

		case "unmute":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			dumb, err := database.FindCharacterByName(parts[1])
			if err != nil {
				return nil, err
			}

			server.MutedPlayers.Remove(dumb.UserID)

		case "uid":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			c, err := database.FindCharacterByName(parts[1])
			if err != nil {
				return nil, err
			} else if c == nil {
				return nil, nil
			}

			resp = messaging.InfoMessage(c.UserID)

		case "uuid":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			user, err := database.FindUserByName(parts[1])
			if err != nil {
				return nil, err
			} else if user == nil {
				return nil, nil
			}

			resp = messaging.InfoMessage(user.ID)

		case "inv":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}
			if parts[1] == "1" {
				data := database.BUFF_INFECTION
				data.Insert(utils.IntToBytes(uint64(70), 4, true), 6)     // infection id
				data.Insert(utils.IntToBytes(uint64(99999), 4, true), 11) // buff remaining time

				s.Conn.Write(data)
			} else {
				r := database.BUFF_EXPIRED
				r.Insert(utils.IntToBytes(uint64(70), 4, true), 6) // buff infection id
				r.Concat(data)

				s.Conn.Write(r)
			}
			s.Character.Invisible = parts[1] == "1"
		case "kick":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			dumb, err := database.FindCharacterByName(parts[1])
			if err != nil {
				return nil, err
			}

			database.GetSocket(dumb.UserID).Conn.Close()

		case "summon":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			c, err := database.FindCharacterByName(parts[1])
			if err != nil {
				return nil, err
			}
			return c.ChangeMap(s.Character.Map, database.ConvertPointToLocation(s.Character.Coordinate))

		case "tp":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}

			x, err := strconv.ParseFloat(parts[1], 10)
			if err != nil {
				return nil, err
			}

			y, err := strconv.ParseFloat(parts[2], 10)
			if err != nil {
				return nil, err
			}

			return s.Character.Teleport(database.ConvertPointToLocation(fmt.Sprintf("%.1f,%.1f", x, y))), nil

		case "tpp":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			c, err := database.FindCharacterByName(parts[1])
			if err != nil {
				return nil, err
			}

			return s.Character.ChangeMap(c.Map, database.ConvertPointToLocation(c.Coordinate))
		case "greatwar": // /greatwar [masodperc]
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}
			if len(parts) < 2 {
				return nil, nil
			}
			time, _ := strconv.ParseInt(parts[1], 10, 32)
			database.CanJoinWar = true
			database.StartWarTimer(int(time))

		case "factionwar": //FACTION WAR MISI CSINÁLTA
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}
			database.PrepareFactionWar()
		case "autogreatwar":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}
			c := cron.New()
			c.AddFunc("@every 5h", func() {
				database.CanJoinWar = true
				database.StartWarTimer(int(600))
			})
			c.Start()
			return messaging.InfoMessage(fmt.Sprintf("Auto Great War activated")), nil
		case "autofactionwar":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}
			c := cron.New()
			c.AddFunc("@every 5h", func() {
				database.PrepareFactionWar()
			})
			c.Start()
			return messaging.InfoMessage(fmt.Sprintf("Auto Faction War activated")), nil
		case "speed":
			if s.User.UserType < server.GM_USER {
				return nil, nil
			}

			if len(parts) < 2 {
				return nil, nil
			}

			speed, err := strconv.ParseFloat(parts[1], 10)
			if err != nil {
				return nil, err
			}

			s.Character.RunningSpeed = speed

		case "online":
			if s.User.UserType < server.GAL_USER {
				return nil, nil
			}

			characters, err := database.FindOnlineCharacters()
			if err != nil {
				return nil, err
			}

			online := funk.Values(characters).([]*database.Character)
			sort.Slice(online, func(i, j int) bool {
				return online[i].Name < online[j].Name
			})

			resp.Concat(messaging.InfoMessage(fmt.Sprintf("%d player(s) online.", len(characters))))

			for _, c := range online {
				u, _ := database.FindUserByID(c.UserID)
				if u == nil {
					continue
				}

				resp.Concat(messaging.InfoMessage(fmt.Sprintf("%s is in map %d (Dragon%d) at %s.", c.Name, c.Map, u.ConnectedServer, c.Coordinate)))
			}

		case "name":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}

			id, _ := strconv.Atoi(parts[1])
			c, err := database.FindCharacterByID(int(id))
			if err != nil {
				return nil, err
			}

			c2, err := database.FindCharacterByName(parts[2])
			if err != nil {
				return nil, err
			} else if c2 != nil {
				return nil, nil
			}

			c.Name = parts[2]
			c.Update()

		case "role":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}

			id, _ := strconv.Atoi(parts[1])
			c, err := database.FindCharacterByID(int(id))
			if err != nil {
				return nil, err
			}

			user, err := database.FindUserByID(c.UserID)
			if err != nil {
				return nil, err
			}

			role, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, err
			}

			user.UserType = int8(role)
			user.Update()
		case "skillpoint":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}

			character, err := database.FindCharacterByName(parts[1])
			if err != nil {
				return nil, err
			}
			num, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, err
			}
			character.Socket.Skills.SkillPoints += num
			s.Conn.Write(character.GetExpAndSkillPts())
		case "type":
			if s.User.UserType < server.HGM_USER {
				return nil, nil
			}

			if len(parts) < 3 {
				return nil, nil
			}

			id, _ := strconv.Atoi(parts[1])
			c, err := database.FindCharacterByID(int(id))
			if err != nil {
				return nil, err
			}

			t, err := strconv.Atoi(parts[2])
			if err != nil {
				return nil, err
			}

			c.Type = t
			c.Update()
		}

	}

	return resp, err
}

func countMaintenance(cd int) {
	msg := fmt.Sprintf("There will be maintenance after %d seconds. Please log out in order to prevent any inconvenience.", cd)
	makeAnnouncement(msg)

	if cd > 0 {
		time.AfterFunc(time.Second*10, func() {
			countMaintenance(cd - 10)
		})
	} else {
		//os.Exit(0)
	}
}

func cmdSpawnMobs(count, npcID, mapID int, coordinate string) {
	for i := 0; i < int(count); i++ {

		coordinate := database.ConvertPointToLocation(coordinate)
		randomLocX := randFloats(coordinate.X, coordinate.X+30)
		randomLocY := randFloats(coordinate.Y, coordinate.Y+30)

		minX := randomLocX
		minY := randomLocY
		maxX := randomLocX + 25
		maxY := randomLocY + 25
		MinLocation := fmt.Sprintf("%.1f,%.1f", minX, minY)
		MaxLocation := fmt.Sprintf("%.1f,%.1f", maxX, maxY)

		npcPos := &database.NpcPosition{ID: len(database.NPCPos), NPCID: int(npcID), MapID: int16(mapID), Rotation: 0, Attackable: true, IsNPC: false, RespawnTime: 30, Count: 30, MinLocation: MinLocation, MaxLocation: MaxLocation}
		database.NPCPos = append(database.NPCPos, npcPos)
		npcPos.Create()
		npc, _ := database.NPCs[npcID]

		newai := &database.AI{ID: len(database.AIs), HP: npc.MaxHp, Map: int16(mapID), PosID: npcPos.ID, RunningSpeed: 10, Server: 1, WalkingSpeed: 5, Once: false}
		newai.OnSightPlayers = make(map[int]interface{})

		loc := utils.Location{X: randomLocX, Y: randomLocY}
		npcPos.MinLocation = fmt.Sprintf("%.1f,%.1f", randomLocX, randomLocY)
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
		newai.Create()
		//ai.Init()
		if newai.WalkingSpeed > 0 {
			go newai.Handler()
		}
	}
}
func startGuildWar(sourceG, enemyG *database.Guild) []byte {
	challengerGuild := sourceG
	enemyGuild := enemyG
	makeAnnouncement(fmt.Sprintf("%s has declare war to %s.", challengerGuild.Name, enemyGuild.Name))

	return nil
}

func randFloats(min, max float64) float64 {
	return min + rand.Float64()*(max-min)
}

func RemoveIndex(a []string, index int) []string {
	a[index] = a[len(a)-1] // Copy last element to index i.
	a[len(a)-1] = ""       // Erase last element (write zero value).
	a = a[:len(a)-1]       // Truncate slice.
	return a
}
