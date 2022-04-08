package database

import (
	"database/sql"
	"fmt"
	"sort"

	gorp "gopkg.in/gorp.v1"
)

var (
	SkillInfos       = make(map[int]*SkillInfo)
	SkillInfosByBook = make(map[int64][]*SkillInfo)
	SkillPoints      = make([]uint64, 12000)
)

type SkillInfo struct {
	ID                      int     `db:"id"`
	BookID                  int64   `db:"book_id"`
	Name                    string  `db:"name"`
	Target                  int8    `db:"target"`
	PassiveType             uint8   `db:"passive_type"`
	Type                    uint8   `db:"type"`
	MaxPlus                 int8    `db:"max_plus"`
	Slot                    int     `db:"slot"`
	BaseTime                int     `db:"base_duration"`
	AdditionalTime          int     `db:"additional_duration"`
	CastTime                float64 `db:"cast_time"`
	BaseChi                 int     `db:"base_chi"`
	AdditionalChi           int     `db:"additional_chi"`
	BaseMinMultiplier       int     `db:"base_min_multiplier"`
	AdditionalMinMultiplier int     `db:"additional_min_multiplier"`
	BaseMaxMultiplier       int     `db:"base_max_multiplier"`
	AdditionalMaxMultiplier int     `db:"additional_max_multiplier"`
	BaseRadius              float64 `db:"base_radius"`
	AdditionalRadius        float64 `db:"additional_radius"`
	Passive                 bool    `db:"passive"`
	BasePassive             int     `db:"base_passive"`
	AdditionalPassive       int     `db:"additional_passive"`
	InfectionID             int     `db:"infection_id"`
	AreaCenter              int     `db:"area_center"`
	Cooldown                float64 `db:"cooldown"`
}

func (e *SkillInfo) Create() error {
	return db.Insert(e)
}

func (e *SkillInfo) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *SkillInfo) Delete() error {
	_, err := db.Delete(e)
	return err
}

func (e *SkillInfo) Update() error {
	_, err := db.Update(e)
	return err
}

func getSkillInfos() error {
	var skills []*SkillInfo
	query := `select * from data.skills`

	if _, err := db.Select(&skills, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getSkillInfos: %s", err.Error())
	}

	for _, s := range skills {
		SkillInfos[s.ID] = s
		if s.Slot > 0 {
			SkillInfosByBook[s.BookID] = append(SkillInfosByBook[s.BookID], s)
		}
	}

	for _, b := range SkillInfosByBook {
		sort.Slice(b, func(i, j int) bool {
			return b[i].Slot < b[j].Slot
		})
	}

	for i := uint64(0); i < uint64(len(SkillPoints)); i++ {
		SkillPoints[i] = 20000 * i * i
	}

	return nil
}
