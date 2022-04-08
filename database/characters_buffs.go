package database

import (
	"database/sql"
	"fmt"
	"sort"

	gorp "gopkg.in/gorp.v1"
)

type Buff struct {
	ID              int     `db:"id" json:"id"`
	CharacterID     int     `db:"character_id" json:"character_id"`
	Name            string  `db:"name" json:"name"`
	ATK             int     `db:"atk" json:"atk"`
	ATKRate         int     `db:"atk_rate" json:"atk_rate"`
	ArtsATK         int     `db:"arts_atk" json:"arts_atk"`
	ArtsATKRate     int     `db:"arts_atk_rate" json:"arts_atk_rate"`
	PoisonDEF       int     `db:"poison_def" json:"poison_def"`
	ParalysisDEF    int     `db:"paralysis_def" json:"paralysis_def"`
	ConfusionDEF    int     `db:"confusion_def" json:"confusion_def"`
	DEF             int     `db:"def" json:"def"`
	DEFRate         int     `db:"def_rate" json:"def_rate"`
	ArtsDEF         int     `db:"arts_def" json:"arts_def"`
	ArtsDEFRate     int     `db:"arts_def_rate" json:"arts_def_rate"`
	Accuracy        int     `db:"accuracy" json:"accuracy"`
	Dodge           int     `db:"dodge" json:"dodge"`
	MaxHP           int     `db:"max_hp" json:"max_hp"`
	HPRecoveryRate  int     `db:"hp_recovery_rate" json:"hp_recovery_rate"`
	MaxCHI          int     `db:"max_chi" json:"max_chi"`
	CHIRecoveryRate int     `db:"chi_recovery_rate" json:"chi_recovery_rate"`
	STR             int     `db:"str" json:"str"`
	DEX             int     `db:"dex" json:"dex"`
	INT             int     `db:"int" json:"int"`
	EXPMultiplier   int     `db:"exp_multiplier" json:"exp_multiplier"`
	DropMultiplier  int     `db:"drop_multiplier" json:"drop_multiplier"`
	RunningSpeed    float64 `db:"running_speed" json:"running_speed"`
	StartedAt       int64   `db:"started_at" json:"started_at"`
	Duration        int64   `db:"duration" json:"duration"`
	BagExpansion    bool    `db:"bag_expansion" json:"bag_expansion"`
	CanExpire       bool    `db:"canexpire" json:"canexpire"`
}

func (b *Buff) Create() error {
	return db.Insert(b)
}

func (b *Buff) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(b)
}

func (b *Buff) Delete() error {
	_, err := db.Delete(b)
	return err
}

func (b *Buff) Update() error {
	_, err := db.Update(b)
	return err
}

func FindBuffsByCharacterID(characterID int) ([]*Buff, error) {

	var buffs []*Buff
	query := `select * from hops.characters_buffs where character_id = $1`

	if _, err := db.Select(&buffs, query, characterID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindBuffsByCharacterID: %s", err.Error())
	}

	sort.Slice(buffs, func(i, j int) bool {
		return buffs[i].StartedAt+buffs[i].Duration <= buffs[j].StartedAt+buffs[j].Duration
	})

	return buffs, nil
}

func FindBuffByID(buffID, characterID int) (*Buff, error) {

	var buff *Buff
	query := `select * from hops.characters_buffs where id = $1 and character_id = $2`

	if err := db.SelectOne(&buff, query, buffID, characterID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindBuffByID: %s", err.Error())
	}

	return buff, nil
}
