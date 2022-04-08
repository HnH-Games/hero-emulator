package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"

	"hero-emulator/utils"

	gorp "gopkg.in/gorp.v1"
)

var (
	allSkills = make(map[int]*Skills)
	skMutex   sync.RWMutex

	miningSkills = []int{41004, 41104, 41204, 41304, 41404, 41504, 41604, 41704, 2503, 2604, 2705, 2805,
		23001, 23005, 23009}

	SkillPTS = map[string]map[int]int{
		"fjp": map[int]int{0: 1, 1: 1, 2: 1, 3: 2, 4: 3, 5: 4, 6: 6, 7: 9, 8: 14, 9: 21, 10: 31, 11: 47},
		"wd":  map[int]int{0: 3, 1: 2, 2: 3, 3: 4, 4: 6, 5: 9, 6: 14, 7: 21, 8: 31, 9: 47, 10: 70, 11: 105},
		"sjp": map[int]int{
			0: 1, 1: 2, 2: 3, 3: 3, 4: 4, 5: 5, 6: 6, 7: 7, 8: 8, 9: 10,
			10: 12, 11: 14, 12: 17, 13: 21, 14: 25, 15: 30, 16: 36, 17: 43, 18: 52, 19: 62,
			20: 62, 21: 62, 22: 62, 23: 62, 24: 62, 25: 62, 26: 62, 27: 62, 28: 62, 29: 62,
			30: 94, 31: 94, 32: 94, 33: 94, 34: 94, 35: 94, 36: 94, 37: 94, 38: 94, 39: 94,
			40: 116, 41: 116, 42: 116, 43: 116, 44: 116, 45: 116, 46: 116, 47: 116, 48: 116, 49: 116,
		},
		"dsjp": map[int]int{
			0: 1, 1: 2, 2: 3, 3: 3, 4: 4, 5: 5, 6: 6, 7: 7, 8: 8, 9: 10,
			10: 12, 11: 14, 12: 17, 13: 21, 14: 25, 15: 30, 16: 36, 17: 43, 18: 52, 19: 62,
			20: 62, 21: 62, 22: 62, 23: 62, 24: 62, 25: 62, 26: 62, 27: 62, 28: 62, 29: 62,
			30: 94, 31: 94, 32: 94, 33: 94, 34: 94, 35: 94, 36: 94, 37: 94, 38: 94, 39: 94,
			40: 116, 41: 116, 42: 116, 43: 116, 44: 116, 45: 116, 46: 116, 47: 116, 48: 116, 49: 116,
			50: 138, 51: 138, 52: 138, 53: 138, 54: 138, 55: 138, 56: 138, 57: 138, 58: 138, 59: 138,
			60: 160, 61: 160, 62: 160, 63: 160, 64: 160, 65: 160, 66: 160, 67: 160, 68: 160, 69: 160,
			70: 198, 71: 198, 72: 198, 73: 198, 74: 198, 75: 198, 76: 198, 77: 198, 78: 198, 79: 198,
			80: 240, 81: 240, 82: 240, 83: 240, 84: 240, 85: 240, 86: 240, 87: 240, 88: 240, 89: 240,
			90: 295, 91: 295, 92: 295, 93: 295, 94: 295, 95: 295, 96: 295, 97: 295, 98: 295, 99: 295,
		},
	}

	COMBAT_SKILL_BOOK  = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x81, 0x01, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	PASSIVE_SKILL_BOOK = utils.Packet{0xAA, 0x55, 0x07, 0x00, 0x82, 0x01, 0x00, 0x55, 0xAA}
	//EZ a 0x0a,0x00 után jön az id
	DIVINE_SKILL_BOOk = utils.Packet{0xAA, 0x55, 0x0b, 0x00, 0x81, 0x05, 0x0a, 0x00, 0x00, 0x55, 0xAA}
)

type Skills struct {
	ID          int             `db:"id" json:"id"`
	SkillPoints int             `db:"skill_points" json:"skill_points"`
	Skills      json.RawMessage `db:"skills" json:"skill"`
}

type SkillSlots struct {
	Slots []*SkillSet `json:"slots"`
}

type SkillSet struct {
	BookID       int64          `json:"book"`
	DivinePoints []*DivineTuple `json:"divinepoints"`
	Skills       []*SkillTuple  `json:"skills"`
}

type SkillTuple struct {
	SkillID int `json:"skill_id"`
	Plus    int `json:"plus"`
}

type DivineTuple struct {
	DivineID   int `json:"divine_id"`
	DivinePlus int `json:"value"`
}

func (e *Skills) Create(c *Character) error {
	e.ID = c.ID
	e.Skills = json.RawMessage(`{"slots": [{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}]}`)
	return db.Insert(e)
}

func (e *Skills) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *Skills) Update() error {
	_, err := db.Update(e)
	return err
}

func (e *Skills) Delete() error {
	skMutex.Lock()
	delete(allSkills, e.ID)
	skMutex.Unlock()

	_, err := db.Delete(e)
	return err
}

func (e *Skills) GetSkills() (*SkillSlots, error) {
	slots := &SkillSlots{}
	err := json.Unmarshal([]byte(e.Skills), &slots)
	if err != nil {
		return nil, err
	}
	return slots, nil
}

func (e *Skills) SetSkills(slots *SkillSlots) error {
	data, err := json.Marshal(slots)
	if err != nil {
		return err
	}

	e.Skills = json.RawMessage(data)
	return nil
}

func FindSkillsByID(id int) (*Skills, error) {

	skMutex.RLock()
	s, ok := allSkills[id]
	skMutex.RUnlock()

	if ok {
		return s, nil
	}

	query := `select * from hops.skills where id = $1`
	skills := &Skills{}
	if err := db.SelectOne(&skills, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindSkillsByID: %s", err.Error())
	}

	skMutex.Lock()
	allSkills[id] = skills
	skMutex.Unlock()

	return skills, nil
}

func (s *Skills) GetSkillsData() ([]byte, error) {
	skillSlots, err := s.GetSkills()
	if err != nil {
		return nil, err
	}

	resp := utils.Packet{}
	for i, slot := range skillSlots.Slots {
		if slot.BookID == 0 {
			continue
		}

		if i < 5 { // Combat book
			r := COMBAT_SKILL_BOOK
			r[6] = byte(i)                                              // book index
			r.Insert(utils.IntToBytes(uint64(slot.BookID), 4, true), 7) // book id
			c, index, length := 1, 14, int16(10)
			for _, skill := range slot.Skills {
				if skill.SkillID == 0 {
					continue
				}

				info := SkillInfos[skill.SkillID]
				for ; c < info.Slot; c++ {
					r.Insert([]byte{0x00, 0x00, 0x00, 0x00, 0x00}, index) // empty slot
					index += 5
					length += 5
				}

				r.Insert(utils.IntToBytes(uint64(skill.SkillID), 4, true), index) // skill id
				index += 4

				r.Insert([]byte{byte(skill.Plus)}, index) // skill plus
				index++

				c++
				length += 5
			}
			r.SetLength(length)
			resp.Concat(r)
			for _, points := range slot.DivinePoints {
				newr := DIVINE_SKILL_BOOk
				newr[8] = byte(i)
				index = 9
				newr.Insert([]byte{byte(points.DivineID)}, index) // divine plus // divine id
				index++
				newr.Insert(utils.IntToBytes(uint64(slot.BookID), 4, true), index) // book id
				index += 4
				newr.Insert([]byte{byte(points.DivinePlus)}, index) // divine plus
				index++
				resp.Concat(newr)
			}
			//log.Printf("Divine ID: %d", newr)
		} else { // Passive book
			r := PASSIVE_SKILL_BOOK
			r[6] = byte(i - 5)
			r.Insert(utils.IntToBytes(uint64(slot.BookID), 4, true), 7)
			resp.Concat(r)

			r = PASSIVE_SKILL_UGRADED
			r[8] = byte(i - 5)
			r[9] = byte(slot.Skills[0].Plus)
			resp.Concat(r)
		}
	}

	return resp, nil
}

func (s *Skills) GetPlus(skillID int) (byte, error) {
	skillSlots, err := s.GetSkills()
	if err != nil {
		return 0, err
	}

	for _, slot := range skillSlots.Slots {
		for _, skill := range slot.Skills {
			if skill.SkillID == skillID {
				return byte(skill.Plus), nil
			}
		}
	}

	return 0, nil
}

func (s *Skills) GetDivinePlus(BookID int64) (int, error) {
	skillSlots, err := s.GetSkills()
	if err != nil {
		return 0, err
	}
	usedPoints := 0
	for _, slot := range skillSlots.Slots {
		if BookID == slot.BookID {
			for _, points := range slot.DivinePoints {
				usedPoints += points.DivinePlus
			}
		}
		return usedPoints, nil
	}

	return 0, nil
}
