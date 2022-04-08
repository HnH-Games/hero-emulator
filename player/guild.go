package player

import (
	"log"

	"hero-emulator/database"
	"hero-emulator/gold"
	"hero-emulator/messaging"
	"hero-emulator/nats"
	"hero-emulator/server"
	"hero-emulator/utils"

	"github.com/thoas/go-funk"
)

type (
	CreateGuildHandler         struct{}
	GuildRequestHandler        struct{}
	RespondGuildRequestHandler struct{}
	ExpelFromGuildHandler      struct{}
	ChangeGuildLogoHandler     struct{}
	LeaveGuildHandler          struct{}
	ChangeRoleHandler          struct{}
)

var (
	CREATED_GUILD       = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x83, 0x01, 0x0A, 0x00, 0x00, 0x55, 0xAA}
	EXPELLED_FROM_GUILD = utils.Packet{0xAA, 0x55, 0x06, 0x00, 0x83, 0x02, 0x55, 0xAA}
	GUILD_REQUEST       = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x83, 0x03, 0x0A, 0x00, 0x05, 0x55, 0xAA}
	NEW_GUILD_MEMBER    = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x83, 0x05, 0x00, 0x01, 0x10, 0x27, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x35,
		0x02, 0x01, 0x04, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x32, 0x30, 0x31, 0x38, 0x2D, 0x30, 0x38, 0x2D, 0x32, 0x36, 0x00, 0x55, 0xAA}
	GUILD_LOGO = utils.Packet{0xAA, 0x55, 0x1B, 0x03, 0x83, 0x0B, 0x0A, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	MEMBER_EXPELLED = utils.Packet{0xAA, 0x55, 0x06, 0x00, 0x83, 0x07, 0x55, 0xAA}
)

func (h *CreateGuildHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if s.Character.Gold < 10*gold.M {
		msg := "You must have more than 10,000,000 gold to create a house."
		return messaging.InfoMessage(msg), nil
	} else if s.Character.Level < 50 {
		msg := "Your level must be more than 5Dan0Kyu to create a house."
		return messaging.InfoMessage(msg), nil
	} else if s.Character.GuildID > 0 {
		return nil, nil
	}

	faction := int16(data[7])
	nameLength := int(data[8])
	name := string(data[9 : 9+nameLength])

	index := 9 + nameLength
	logo := data[index : index+0x0300] // 16x16 logo

	g, err := database.FindGuildByName(name)
	if err != nil {
		return nil, err
	} else if g != nil {
		msg := "There is already a house with that name. Please choose another one."
		return messaging.InfoMessage(msg), nil
	}

	g = &database.Guild{
		Announcement:  "",
		Description:   "",
		Faction:       faction,
		GoldDonation:  0,
		HonorDonation: 0,
		LeaderID:      s.Character.ID,
		Logo:          logo,
		Name:          name,
		Recognition:   0,
	}

	err = g.AddMember(&database.GuildMember{ID: s.Character.ID, Role: database.GROLE_LEADER})
	if err != nil {
		return nil, err
	}

	if err = g.Create(); err != nil {
		return nil, err
	}

	s.Character.GuildID = g.ID
	resp := utils.Packet{}

	d, err := s.Character.GetGuildData()
	if err != nil {
		return nil, err
	}

	resp.Concat(d)

	s.Character.Gold -= 10 * gold.M
	resp.Concat(s.Character.GetGold())

	r := CREATED_GUILD
	r.Insert(utils.IntToBytes(uint64(g.ID), 4, true), 8) // guild id
	r[12] = byte(len(g.Name))                            // guild name length
	r.Insert([]byte(g.Name), 13)                         // guild name

	length := int16(9 + len(g.Name))
	r.SetLength(length)

	resp.Concat(utils.Packet{0xAA, 0x55, 0x49, 0x00, 0x83, 0x08, 0x56, 0x06, 0x00, 0x00, 0x1A, 0xB2, 0x01, 0x00, 0x0D, 0x4F, 0x6C, 0x64, 0x53, 0x63, 0x68, 0x6F, 0x6F, 0x6C, 0x32, 0x30, 0x30, 0x35, 0x32, 0x00, 0x00, 0x00, 0x05, 0x10, 0x27, 0x00, 0x00, 0xC8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x34, 0x15, 0x03, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x32, 0x30, 0x32, 0x30, 0x2D, 0x30, 0x33, 0x2D, 0x31, 0x32, 0x00, 0x55, 0xAA})
	//resp.Concat(utils.Packet{0xAA, 0x55, 0x0C, 0x00, 0x83, 0x0C, 0x0A, 0x00, 0x99, 0x2D, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA,
	//	0xAA, 0x55, 0x02, 0x00, 0xCC, 0x00, 0x55})

	spawnData, err := s.Character.SpawnCharacter()
	if err == nil {
		p := nats.CastPacket{CastNear: true, CharacterID: s.Character.ID, Type: nats.PLAYER_SPAWN, Data: spawnData}
		p.Cast()
		resp.Concat(spawnData)
	}

	resp.Concat(r)
	return resp, nil
}

func (h *GuildRequestHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	user, err := database.FindUserByID(s.Character.UserID)
	if err != nil {
		return nil, err
	}

	pseudoID := uint16(utils.BytesToInt(data[6:8], true))
	character := server.FindCharacter(user.ConnectedServer, pseudoID)

	guild, err := database.FindGuildByID(s.Character.GuildID)
	if err != nil {
		return nil, err
	}

	if character.Faction != s.Character.Faction || character.GuildID > 0 || s.Character.GuildID == 0 || guild == nil {
		return nil, nil
	}

	member, err := guild.GetMember(s.Character.ID)
	if err != nil {
		return nil, err
	}

	if member.Role != database.GROLE_LEADER && member.Role != database.GROLE_SOLDIER && member.Role != database.GROLE_SAGE {
		return nil, nil
	}

	leader, err := database.FindCharacterByID(guild.LeaderID)
	if err != nil {
		return nil, err
	}

	length := int16(0x0E + len(leader.Name) + len(guild.Name))
	resp := GUILD_REQUEST
	resp.SetLength(length)
	resp.Insert(utils.IntToBytes(uint64(leader.ID), 4, true), 8)          // leader id
	resp.Insert(utils.IntToBytes(uint64(leader.Faction), 1, true), 12)    // faction
	resp.Insert(utils.IntToBytes(uint64(guild.MemberCount), 2, true), 13) // member count
	resp.Insert(utils.IntToBytes(uint64(len(leader.Name)), 1, true), 15)  // leader name length
	resp.Insert([]byte(leader.Name), 16)                                  // leader name

	index := 16 + len(leader.Name)
	resp.Insert(utils.IntToBytes(uint64(len(guild.Name)), 1, true), index) // guild name length
	resp.Insert([]byte(guild.Name), index+1)                               // guild name

	character.Socket.Write(resp)
	return nil, nil
}

func (h *RespondGuildRequestHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if s.Character.GuildID > 0 {
		return nil, nil
	}

	accepted := data[6] == 1
	characterID := int(utils.BytesToInt(data[7:11], true))

	resp := utils.Packet{}
	if accepted {
		c, err := database.FindCharacterByID(characterID)
		if err != nil {
			return nil, err
		} else if c == nil {
			return nil, nil
		}

		guild, err := database.FindGuildByID(c.GuildID)
		if err != nil {
			return nil, err
		} else if guild == nil {
			return nil, nil
		}

		err = guild.AddMember(&database.GuildMember{ID: s.Character.ID, Role: database.GROLE_MEMBER})
		if err != nil {
			return nil, err
		}

		go guild.Update()
		s.Character.GuildID = guild.ID
		spawnData, err := s.Character.SpawnCharacter()
		if err == nil {
			p := nats.CastPacket{CastNear: true, CharacterID: s.Character.ID, Type: nats.PLAYER_SPAWN, Data: spawnData}
			p.Cast()
			resp.Concat(spawnData)
		}

		guildData, err := s.Character.GetGuildData()
		if err == nil {
			resp.Concat(guildData)
		}

		length := int16(0x2F + len(s.Character.Name))
		r := NEW_GUILD_MEMBER
		r.SetLength(length)
		r.Insert(utils.IntToBytes(uint64(s.Character.ID), 4, true), 6) // character id
		r[10] = byte(len(s.Character.Name))                            // character name length
		r.Insert([]byte(s.Character.Name), 11)                         //character name

		index := 11 + len(s.Character.Name)
		r.Insert(utils.IntToBytes(uint64(s.Character.Level), 4, true), index) // character level
		index += 4
		index += 20
		r.Insert(utils.IntToBytes(uint64(s.Character.Map), 1, true), index) // character map id

		members, _ := guild.GetMembers()
		members = funk.Filter(members, func(member *database.GuildMember) bool {
			c, err := database.FindCharacterByID(member.ID)
			if err != nil || c == nil {
				return false
			}

			return c.IsOnline && c.ID != s.Character.ID
		}).([]*database.GuildMember)

		for _, member := range members {
			c, err := database.FindCharacterByID(member.ID)
			if err != nil || c == nil {
				continue
			}

			memberResp := r
			guildData, err := c.GetGuildData()
			if err != nil {
				continue
			}

			memberResp.Concat(guildData)
			c.Socket.Write(memberResp)
		}

	} else {
		resp = messaging.SystemMessage(messaging.GUILD_REQUEST_REJECTED)
	}

	return resp, nil
}

func (h *ExpelFromGuildHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	characterID := int(utils.BytesToInt(data[6:10], true))
	guild, err := database.FindGuildByID(s.Character.GuildID)
	if err != nil {
		return nil, err
	} else if guild == nil {
		return nil, nil
	}

	member, err := guild.GetMember(s.Character.ID)
	if err != nil {
		return nil, err
	}

	if member.Role != database.GROLE_LEADER && member.Role != database.GROLE_SOLDIER && member.Role != database.GROLE_SAGE {
		return nil, nil
	}

	err = guild.RemoveMember(characterID)
	if err != nil {
		return nil, err
	}

	c, err := database.FindCharacterByID(characterID)
	if err != nil {
		return nil, err
	} else if c == nil {
		return nil, nil
	}

	c.GuildID = -1
	go c.Update()

	resp := MEMBER_EXPELLED
	resp.Insert(utils.IntToBytes(uint64(characterID), 4, true), 6)

	members, _ := guild.GetMembers()
	for _, member := range members {
		c, err := database.FindCharacterByID(member.ID)
		if err != nil || c == nil || !c.IsOnline {
			continue
		}

		c.Socket.Write(resp)
	}

	if c.IsOnline { // You have been EXPELLED!
		r := EXPELLED_FROM_GUILD
		r.Insert(utils.IntToBytes(uint64(guild.LeaderID), 4, true), 6)
		c.Socket.Write(resp)
	}

	resp = utils.Packet{}
	spawnData, err := c.SpawnCharacter()
	if err == nil {
		p := nats.CastPacket{CastNear: true, CharacterID: c.ID, Type: nats.PLAYER_SPAWN, Data: spawnData}
		p.Cast()
		resp.Concat(spawnData)
	}

	go guild.Update()
	return nil, nil
}

func (h *ChangeGuildLogoHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	guildID := int(utils.BytesToInt(data[6:10], true))
	logo := data[12:0x30C]

	if guildID != s.Character.GuildID {
		return nil, nil
	}

	guild, err := database.FindGuildByID(s.Character.GuildID)
	if err != nil {
		return nil, err
	} else if guild == nil {
		return nil, nil
	}

	member, err := guild.GetMember(s.Character.ID)
	if err != nil {
		return nil, err
	}

	if member.Role != database.GROLE_LEADER && member.Role != database.GROLE_SOLDIER {
		return nil, nil
	}

	guild.Logo = logo
	resp := GUILD_LOGO
	resp.Insert(utils.IntToBytes(uint64(guild.ID), 4, true), 8) // guild id
	resp[18] = byte(guild.Faction)                              // guild faction
	resp.Insert(logo, 20)                                       // guild logo

	members, _ := guild.GetMembers()
	for _, member := range members {
		c, err := database.FindCharacterByID(member.ID)
		if err != nil || c == nil || !c.IsOnline {
			continue
		}

		c.Socket.Write(resp)
	}

	go guild.Update()
	return nil, nil
}

func (h *LeaveGuildHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	guild, err := database.FindGuildByID(s.Character.GuildID)
	if err != nil {
		return nil, err
	} else if guild == nil {
		return nil, nil
	}

	resp := utils.Packet{}
	s.Character.GuildID = -1
	if guild.LeaderID == s.Character.ID { // dissolve guild
		r := EXPELLED_FROM_GUILD // FIX
		r.Insert(utils.IntToBytes(uint64(guild.LeaderID), 4, true), 6)

		resp.Concat(r)
		members, _ := guild.GetMembers()
		for _, member := range members {
			c, err := database.FindCharacterByID(member.ID)
			if err != nil || c == nil {
				continue
			}

			c.GuildID = -1
			go c.Update()

			if !c.IsOnline {
				continue
			}

			c.Socket.Write(r)

			spawnData, err := c.SpawnCharacter()
			if err == nil {
				p := nats.CastPacket{CastNear: true, CharacterID: c.ID, Type: nats.PLAYER_SPAWN, Data: spawnData}
				p.Cast()
			}
		}

		guild.Delete()
	} else { // leave guild
		err = guild.RemoveMember(s.Character.ID)
		if err != nil {
			return nil, err
		}

		go guild.Update()

		r := MEMBER_EXPELLED // FIX
		r.Insert(utils.IntToBytes(uint64(s.Character.ID), 4, true), 6)

		resp.Concat(r)
		members, _ := guild.GetMembers()
		for _, member := range members {
			c, err := database.FindCharacterByID(member.ID)
			if err != nil || c == nil || !c.IsOnline {
				continue
			}

			c.Socket.Write(r)
		}

		spawnData, err := s.Character.SpawnCharacter()
		if err == nil {
			p := nats.CastPacket{CastNear: true, CharacterID: s.Character.ID, Type: nats.PLAYER_SPAWN, Data: spawnData}
			p.Cast()
			resp.Concat(spawnData)
		}
	}

	return resp, nil
}

func (h *ChangeRoleHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	guildID := int(utils.BytesToInt(data[6:10], true))
	memberID := int(utils.BytesToInt(data[10:14], true))
	role := int(utils.BytesToInt(data[14:16], true))

	if guildID != s.Character.GuildID {
		return nil, nil
	}

	guild, err := database.FindGuildByID(s.Character.GuildID)
	if err != nil {
		return nil, err
	} else if guild == nil {
		return nil, nil
	}

	self, err := guild.GetMember(s.Character.ID)
	if err != nil {
		return nil, err
	} else if self.Role <= database.GROLE_SOLDIER {
		log.Println("Don't have permission to change member role")
		return nil, nil
	}

	member, err := guild.GetMember(memberID)
	if err != nil {
		return nil, err
	} else if member == nil {
		return nil, nil
	}

	member.Role = database.GuildRole(role)
	err = guild.SetMember(member)
	if err != nil {
		return nil, err
	}

	memberCharacter, err := database.FindCharacterByID(memberID)
	if err != nil {
		return nil, err
	}

	guild.InformMembers(memberCharacter)

	go guild.Update()
	return nil, nil
}
