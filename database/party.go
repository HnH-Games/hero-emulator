package database

import (
	"sync"

	"hero-emulator/utils"

	"github.com/thoas/go-funk"
)

var (
	Parties = make(map[string]*Party)
	pMutex  sync.RWMutex
)

type Party struct {
	Leader  *Character
	Members map[int]*PartyMember
	mutex   sync.RWMutex
}

type PartyMember struct {
	*Character
	Accepted bool
}

var (
	PARTY_DATA      = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x52, 0x04, 0x0A, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x04, 0xFF, 0xFF, 0xFF, 0xFF, 0x03, 0x55, 0xAA}
	PARTY_STATUS    = utils.Packet{0xAA, 0x55, 0x2A, 0x00, 0x52, 0x07, 0x01, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0xFF, 0xFF, 0x55, 0xAA}
	LEFT_PARTY      = utils.Packet{0xAA, 0x55, 0x09, 0x00, 0x52, 0x03, 0x0A, 0x00, 0x00, 0x55, 0xAA}
	PARTY_DISBANDED = utils.Packet{0xAA, 0x55, 0x02, 0x00, 0x52, 0x05, 0x55, 0xAA}
)

func (p *Party) Create() {
	p.Members = make(map[int]*PartyMember)
	pMutex.Lock()
	Parties[p.Leader.UserID] = p
	pMutex.Unlock()
}

func (p *Party) Delete() {
	pMutex.Lock()
	defer pMutex.Unlock()
	delete(Parties, p.Leader.UserID)
}

func (p *Party) AddMember(m *PartyMember) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.Members[m.ID] = m
}

func (p *Party) GetMember(id int) *PartyMember {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return p.Members[id]
}

func (p *Party) GetMembers() []*PartyMember {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return funk.Values(p.Members).([]*PartyMember)
}

func (p *Party) RemoveMember(m *PartyMember) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if m != nil {
		delete(p.Members, m.ID)
	}
}

func GetPartyMemberData(c *Character) []byte {

	user := c.Socket.User
	stat := c.Socket.Stats

	coordinate := ConvertPointToLocation(c.Coordinate)

	resp := PARTY_DATA
	length := int16(len(c.Name) + 47)
	resp.SetLength(length)

	resp.Insert(utils.IntToBytes(uint64(c.ID), 4, true), 9) // character id
	resp[13] = byte(len(c.Name))                            // character name length
	resp.Insert([]byte(c.Name), 14)                         // character name

	index := len(c.Name) + 14
	resp.Insert(utils.IntToBytes(uint64(stat.HP), 4, true), index) // character hp
	index += 4
	resp.Insert(utils.IntToBytes(uint64(stat.MaxHP), 4, true), index) // character max hp
	index += 4
	index++
	resp.Insert(utils.FloatToBytes(float64(coordinate.X), 4, true), index) // coordinate-x
	index += 4
	resp.Insert(utils.FloatToBytes(float64(coordinate.Y), 4, true), index) // coordinate-y
	index += 4
	resp.Insert(utils.IntToBytes(uint64(stat.CHI), 4, true), index) // character chi
	index += 4
	resp.Insert(utils.IntToBytes(uint64(stat.MaxCHI), 4, true), index) // character max chi
	index += 4
	resp.Insert(utils.IntToBytes(uint64(c.Level), 4, true), index) // character level
	index += 4
	resp[index] = byte(c.Type) // character type
	index += 2
	resp[index] = byte(user.ConnectedServer - 1) // connected server id
	return resp
}

func (p *Party) WelcomeMember(c *Character) {

	resp := GetPartyMemberData(c)
	if resp == nil {
		return
	}

	p.Leader.Socket.Write(resp)
	members := p.GetMembers()
	for _, member := range members {
		if member.ID == c.ID || !member.Accepted {
			continue
		}

		member.Socket.Write(resp)
	}
}

func FindParty(c *Character) *Party {
	pMutex.RLock()
	p := Parties[c.PartyID]
	pMutex.RUnlock()

	return p
}
