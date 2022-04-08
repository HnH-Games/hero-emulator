package player

import (
	"time"

	"hero-emulator/database"
	"hero-emulator/server"
	"hero-emulator/utils"

	"github.com/thoas/go-funk"
)

type (
	SendPartyRequestHandler    struct{}
	RespondPartyRequestHandler struct{}
	LeavePartyHandler          struct{}
	ExpelFromPartyHandler      struct{}
)

var (
	PARTY_REQUEST          = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x52, 0x01, 0x0A, 0x00, 0x00, 0x21, 0x55, 0xAA}
	PARTY_REQUEST_REJECTED = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x52, 0x02, 0x52, 0x03, 0x55, 0xAA}
	EXPEL_PARTY_MEMBER     = utils.Packet{0xAA, 0x55, 0x08, 0x00, 0x52, 0x06, 0x0A, 0x00, 0x55, 0xAA}
)

func (h *SendPartyRequestHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	pseudoID := uint16(utils.BytesToInt(data[6:8], true))
	member := server.FindCharacter(s.User.ConnectedServer, pseudoID)
	if member == nil || database.FindParty(member) != nil {
		return nil, nil
	}

	party := database.FindParty(s.Character)
	if party == nil {
		party = &database.Party{}
		party.Leader = s.Character
		s.Character.PartyID = s.Character.UserID
		party.Create()
		//s.Conn.Write(database.GetPartyMemberData(party.Leader))
	} else if len(party.GetMembers()) >= 4 {
		return nil, nil
	}

	resp := PARTY_REQUEST
	length := int16(len(s.Character.Name) + 6)
	resp.SetLength(length)

	resp[8] = byte(len(s.Character.Name))
	resp.Insert([]byte(s.Character.Name), 9)

	member.Socket.Write(resp)

	member.PartyID = s.Character.UserID
	m := &database.PartyMember{Character: member, Accepted: false}
	party.AddMember(m)

	time.AfterFunc(30*time.Second, func() {
		if !m.Accepted {
			party.RemoveMember(m)
		}
	})

	return nil, nil
}

func (h *RespondPartyRequestHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	accepted := data[6] == 1

	party := database.FindParty(s.Character)
	if party == nil {
		return nil, nil
	}

	resp := utils.Packet{}
	if accepted {
		if m := party.GetMember(s.Character.ID); m != nil {
			m.Accepted = true
		}

		members := party.GetMembers()
		members = funk.Filter(members, func(m *database.PartyMember) bool {
			return m.Accepted
		}).([]*database.PartyMember)

		r := utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x52, 0x0B, 0x00, 0x55, 0xAA} //0xAA, 0x55, 0x04, 0x00, 0x52, 0x02, 0x0A, 0x00, 0x55, 0xAA
		if len(members) == 1 {
			r.Concat(database.GetPartyMemberData(party.Leader))
			party.Leader.Socket.Write(r)
		}
		//	partyLeaderMessage := r
		party.WelcomeMember(s.Character)                       // send all party members mine data
		resp.Concat(database.GetPartyMemberData(party.Leader)) // get party leader

		for _, member := range members { // get all party members
			resp.Concat(database.GetPartyMemberData(member.Character))
		}
	} else {
		m := party.GetMember(s.Character.ID)
		party.RemoveMember(m)
		s.Character.PartyID = ""
		if len(party.GetMembers()) == 0 {
			party.Leader.PartyID = ""
		}

		r := PARTY_REQUEST_REJECTED
		party.Leader.Socket.Write(r)
	}

	return resp, nil
}

func (h *LeavePartyHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	s.Character.LeaveParty()
	return nil, nil
}

func (h *ExpelFromPartyHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	party := database.FindParty(s.Character)
	if party == nil || party.Leader.ID != s.Character.ID { // if no party or no authorization
		return nil, nil
	}

	characterID := int(utils.BytesToInt(data[6:10], true))
	character, err := database.FindCharacterByID(characterID)
	if err != nil {
		return nil, err
	} else if characterID == s.Character.ID { // expel yourself
		return nil, nil
	}

	resp := EXPEL_PARTY_MEMBER
	resp.Insert(utils.IntToBytes(uint64(characterID), 4, true), 8) // member character id

	character.Socket.Write(resp)

	member := party.GetMember(characterID)
	member.PartyID = ""
	party.RemoveMember(member)

	members := party.GetMembers()
	members = funk.Filter(members, func(m *database.PartyMember) bool {
		return m.Accepted
	}).([]*database.PartyMember)

	for _, m := range members {
		m.Socket.Write(resp)
	}

	if len(party.GetMembers()) == 0 {
		s.Character.PartyID = ""
		resp.Concat(database.PARTY_DISBANDED)
		party.Delete()
	}

	return resp, nil
}
