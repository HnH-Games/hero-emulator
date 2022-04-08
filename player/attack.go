package player

import (
	"fmt"
	"log"
	"time"

	"hero-emulator/database"
	"hero-emulator/messaging"
	"hero-emulator/nats"
	"hero-emulator/server"
	"hero-emulator/utils"
)

type (
	AttackHandler        struct{}
	InstantAttackHandler struct{}
	DealDamageHandler    struct{}
	CastSkillHandler     struct{}
	CastMonkSkillHandler struct{}
	RemoveBuffHandler    struct{}
)

var (
	ATTACKED      = utils.Packet{0xAA, 0x55, 0x0C, 0x00, 0x41, 0x01, 0x0D, 0x02, 0x01, 0x00, 0x00, 0x00, 0x55, 0xAA}
	INST_ATTACKED = utils.Packet{0xAA, 0x55, 0x0C, 0x00, 0x41, 0x01, 0x0D, 0x02, 0x01, 0x00, 0x00, 0x00, 0x55, 0xAA}
)

func (h *AttackHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	c := s.Character
	if c == nil {
		return nil, nil
	}

	st := s.Stats
	if st == nil {
		return nil, nil
	}

	aiID := uint16(utils.BytesToInt(data[7:9], true))
	ai, ok := database.GetFromRegister(s.User.ConnectedServer, s.Character.Map, aiID).(*database.AI)
	if ok {
		if ai == nil || ai.HP <= 0 {
			return nil, nil
		}

		npcPos := database.NPCPos[ai.PosID]
		if npcPos == nil {
			return nil, nil
		}

		npc := database.NPCs[npcPos.NPCID]
		if npc == nil {
			return nil, nil
		}

		if npcPos.Attackable {
			ai.MovementToken = 0
			ai.IsMoving = false
			ai.TargetPlayerID = c.ID

			dmg, err := c.CalculateDamage(ai, false)
			if err != nil {
				return nil, err
			}

			if diff := int(npc.Level) - c.Level; diff > 0 {
				reqAcc := utils.SigmaFunc(float64(diff))
				if float64(st.Accuracy) < reqAcc {
					probability := float64(st.Accuracy) * 1000 / reqAcc
					if utils.RandInt(0, 1000) > int64(probability) {
						dmg = 0
					}
				}
			}

			c.Targets = append(c.Targets, &database.Target{Damage: dmg, AI: ai})
		}

	} else if enemy := server.FindCharacter(s.User.ConnectedServer, aiID); enemy != nil {
		enemy := server.FindCharacter(s.User.ConnectedServer, aiID)
		if enemy == nil || !enemy.IsActive {
			return nil, nil
		}

		dmg, err := c.CalculateDamageToPlayer(enemy, false)
		if err != nil {
			return nil, err
		}

		c.PlayerTargets = append(c.PlayerTargets, &database.PlayerTarget{Damage: dmg, Enemy: enemy})
	}

	resp := ATTACKED
	resp[4] = data[4]
	resp.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 6) // character pseudo id
	resp.Insert(utils.IntToBytes(uint64(aiID), 2, true), 9)       // ai id

	p := &nats.CastPacket{CastNear: true, CharacterID: c.ID, Data: resp, Type: nats.MOB_ATTACK}
	if err := p.Cast(); err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *InstantAttackHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	c := s.Character
	if c == nil {
		return nil, nil
	}

	st := s.Stats
	if st == nil {
		return nil, nil
	}

	aiID := uint16(utils.BytesToInt(data[7:9], true))
	ai, ok := database.GetFromRegister(s.User.ConnectedServer, s.Character.Map, aiID).(*database.AI)
	if ok {
		if ai == nil || ai.HP <= 0 {
			return nil, nil
		}

		npcPos := database.NPCPos[ai.PosID]
		if npcPos == nil {
			return nil, nil
		}

		npc := database.NPCs[npcPos.NPCID]
		if npc == nil {
			return nil, nil
		}

		if npcPos.Attackable {
			ai.MovementToken = 0
			ai.IsMoving = false
			ai.TargetPlayerID = c.ID

			dmg := int(utils.RandInt(int64(st.MinATK), int64(st.MaxATK))) - npc.DEF
			if dmg < 0 {
				dmg = 0
			} else if dmg > ai.HP {
				dmg = ai.HP
			}

			if diff := int(npc.Level) - c.Level; diff > 0 {
				reqAcc := utils.SigmaFunc(float64(diff))
				if float64(st.Accuracy) < reqAcc {
					probability := float64(st.Accuracy) * 1000 / reqAcc
					if utils.RandInt(0, 1000) > int64(probability) {
						dmg = 0
					}
				}
			}

			time.AfterFunc(time.Second/2, func() { // attack done
				go c.DealDamage(ai, dmg)
			})
		}

	} else if enemy := server.FindCharacter(s.User.ConnectedServer, aiID); enemy != nil {

		if enemy == nil || !enemy.IsActive {
			return nil, nil
		}

		dmg, err := c.CalculateDamageToPlayer(enemy, false)
		if err != nil {
			return nil, err
		}

		time.AfterFunc(time.Second/2, func() { // attack done
			if c.CanAttack(enemy) {
				go DealDamageToPlayer(s, enemy, dmg)
			}
		})
	}

	resp := INST_ATTACKED
	resp[4] = data[4]
	resp.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 6) // character pseudo id
	resp.Insert(utils.IntToBytes(uint64(aiID), 2, true), 9)       // ai id

	p := &nats.CastPacket{CastNear: true, CharacterID: c.ID, Data: resp, Type: nats.MOB_ATTACK}
	if err := p.Cast(); err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *DealDamageHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	c := s.Character
	if c == nil {
		return nil, nil
	}

	resp := utils.Packet{}
	if c.TamingAI != nil {
		ai := c.TamingAI
		pos := database.NPCPos[ai.PosID]
		npc := database.NPCs[pos.NPCID]
		petInfo := database.Pets[int64(npc.ID)]

		seed := utils.RandInt(0, 1000)
		proportion := float64(ai.HP) / float64(npc.MaxHp)
		if proportion < 0.1 && seed < 100 && petInfo != nil {
			go c.DealDamage(ai, ai.HP)

			item := &database.InventorySlot{ItemID: int64(npc.ID), Quantity: 1}
			expInfo := database.PetExps[petInfo.Level-1]
			item.Pet = &database.PetSlot{
				Fullness: 100, Loyalty: 100,
				Exp:   uint64(expInfo.ReqExpEvo1),
				HP:    petInfo.BaseHP,
				Level: byte(petInfo.Level),
				Name:  petInfo.Name,
				CHI:   petInfo.BaseChi,
			}

			r, _, err := s.Character.AddItem(item, -1, true)
			if err != nil {
				return nil, err
			}

			resp.Concat(*r)
		}

		c.TamingAI = nil
		return resp, nil
	}

	targets := c.Targets
	dealt := make(map[int]struct{})
	for _, target := range targets {
		if target == nil {
			continue
		}

		ai := target.AI
		if _, ok := dealt[ai.ID]; ok {
			continue
		}

		dmg := target.Damage

		if target.SkillId != 0 {
			c.DealInfection(ai, nil, target.SkillId)
			go c.DealDamage(ai, dmg)
		} else {
			go c.DealDamage(ai, dmg)
		}
		dealt[ai.ID] = struct{}{}
	}

	pTargets := c.PlayerTargets
	dealt = make(map[int]struct{})
	for _, target := range pTargets {
		if target == nil {
			continue
		}

		enemy := target.Enemy
		if _, ok := dealt[enemy.ID]; ok {
			continue
		}

		if c.CanAttack(enemy) {
			dmg := target.Damage
			go DealDamageToPlayer(s, enemy, dmg)
		}

		dealt[enemy.ID] = struct{}{}
	}

	c.Targets = []*database.Target{}
	c.PlayerTargets = []*database.PlayerTarget{}
	return nil, nil
}

func DealDamageToPlayer(s *database.Socket, enemy *database.Character, dmg int) {
	c := s.Character
	enemySt := enemy.Socket.Stats

	if c == nil {
		log.Println("character is nil")
		return
	} else if enemySt.HP <= 0 {
		return
	}

	if s.Character.Invisible {
		buff, _ := database.FindBuffByID(241, s.Character.ID)
		if buff != nil {
			buff.Duration = 0
			go buff.Update()
		}

		buff, _ = database.FindBuffByID(244, s.Character.ID)
		if buff != nil {
			buff.Duration = 0
			go buff.Update()
		}
	}

	enemySt.HP -= dmg
	if enemySt.HP < 0 {
		enemySt.HP = 0
	}

	r := database.MOB_DEAL_DAMAGE
	r.Insert(utils.IntToBytes(uint64(enemy.PseudoID), 2, true), 5) // character pseudo id
	r.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 7)     // mob pseudo id
	r.Insert(utils.IntToBytes(uint64(enemySt.HP), 4, true), 9)     // character hp
	r.Insert(utils.IntToBytes(uint64(enemySt.CHI), 4, true), 13)   // character chi

	r.Concat(enemy.GetHPandChi())
	p := &nats.CastPacket{CastNear: true, CharacterID: enemy.ID, Data: r, Type: nats.PLAYER_ATTACK}
	if err := p.Cast(); err != nil {
		log.Println("deal damage broadcast error:", err)
		return
	}

	if enemySt.HP <= 0 {
		enemySt.HP = 0
		enemy.Socket.Write(enemy.GetHPandChi())
		info := fmt.Sprintf("[%s] has defeated [%s]", c.Name, enemy.Name)
		r := messaging.InfoMessage(info)

		servers, _ := database.GetServers()
		if servers[int16(c.Socket.User.ConnectedServer)-1].IsPVPServer && servers[int16(c.Socket.User.ConnectedServer)-1].CanLoseEXP && servers[int16(enemy.Socket.User.ConnectedServer)-1].CanLoseEXP && !c.IsinWar && !enemy.IsinWar {
			randInt := utils.RandInt(1, 3)
			exp, _ := enemy.LosePlayerExp(int(randInt))
			different := int(enemy.Level + 20)
			if c.Level <= different {
				targetExp := database.EXPs[int16(c.Level)].Exp
				twentyPercent := (targetExp - c.Exp)
				if exp > targetExp || exp > twentyPercent {
					exp = int64(float64(twentyPercent) * 0.2)
				}
				resp, levelUp := c.AddExp(exp)
				if levelUp {
					statData, err := c.GetStats()
					if err == nil {
						c.Socket.Write(statData)
					}
				}
				c.Socket.Write(resp)
			}
		}
		p := nats.CastPacket{CastNear: true, CharacterID: c.ID, Data: r, Type: nats.PVP_FINISHED}
		p.Cast()
		if database.WarStarted && c.IsinWar && enemy.IsinWar {
			c.WarKillCount++
			if c.Faction == 1 {
				database.ShaoPoints -= 5
			} else {
				database.OrderPoints -= 5
			}
		}
		if enemy.Map == 255 && database.IsFactionWarStarted() {
			if enemy.Faction == 1 && c.Faction == 2 {
				database.AddPointsToFactionWarFaction(15, 2)
			}
			if enemy.Faction == 2 && c.Faction == 1 {
				database.AddPointsToFactionWarFaction(15, 1)
			}
		}
	}
}

func (h *CastSkillHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if len(data) < 12 {
		return nil, nil
	}

	attackCounter := int(data[6])
	skillID := int(utils.BytesToInt(data[7:11], true))
	cX := utils.BytesToFloat(data[11:15], true)
	cY := utils.BytesToFloat(data[15:19], true)
	cZ := utils.BytesToFloat(data[19:23], true)
	targetID := int(utils.BytesToInt(data[23:25], true))

	return s.Character.CastSkill(attackCounter, skillID, targetID, cX, cY, cZ)
}

func (h *CastMonkSkillHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if len(data) < 12 {
		return nil, nil
	}

	attackCounter := 0x1B
	skillID := int(utils.BytesToInt(data[6:10], true))
	cX := utils.BytesToFloat(data[10:14], true)
	cY := utils.BytesToFloat(data[14:18], true)
	cZ := utils.BytesToFloat(data[18:22], true)
	targetID := int(utils.BytesToInt(data[22:24], true))

	resp := utils.Packet{0xAA, 0x55, 0x16, 0x00, 0x49, 0x10, 0x55, 0xAA}
	resp.Insert(utils.IntToBytes(uint64(s.Character.PseudoID), 2, true), 6) // character pseudo id
	resp.Insert(utils.FloatToBytes(cX, 4, true), 8)                         // coordinate-x
	resp.Insert(utils.FloatToBytes(cY, 4, true), 12)                        // coordinate-y
	resp.Insert(utils.FloatToBytes(cZ, 4, true), 16)                        // coordinate-z
	resp.Insert(utils.IntToBytes(uint64(targetID), 2, true), 20)            // target pseudo id
	resp.Insert(utils.IntToBytes(uint64(skillID), 4, true), 22)             // skill id

	skill := database.SkillInfos[skillID]
	token := s.Character.MovementToken

	time.AfterFunc(time.Duration(skill.CastTime*1000)*time.Millisecond, func() {
		if token == s.Character.MovementToken {
			data, _ := s.Character.CastSkill(attackCounter, skillID, targetID, cX, cY, cZ)
			s.Conn.Write(data)
		}
	})

	return resp, nil
}

func (h *RemoveBuffHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	infectionID := int(utils.BytesToInt(data[6:10], true))
	buff, err := database.FindBuffByID(infectionID, s.Character.ID)
	if err != nil {
		return nil, err
	} else if buff == nil {
		return nil, nil
	}

	buff.Duration = 0
	go buff.Update()
	return nil, nil
}
