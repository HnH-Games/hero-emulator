package database

import (
	"database/sql"
	"fmt"
	"sort"

	"gopkg.in/gorp.v1"
)

type AiBuff struct {
	ID              int     `db:"id" json:"id"`
	AiID            int     `db:"ai_id" json:"ai_id"`
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
	CharacterID     int     `db:"character_id" json:"character_id"`
	DropMultiplier  int     `db:"drop_multiplier" json:"drop_multiplier"`
	RunningSpeed    float64 `db:"running_speed" json:"running_speed"`
	StartedAt       int64   `db:"started_at" json:"started_at"`
	Duration        int64   `db:"duration" json:"duration"`
	SkillPlus       int     `db:"skill_plus" json:"skill_plus"`
}

func (b *AiBuff) Create() error {
	return db.Insert(b)
}

func (b *AiBuff) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(b)
}

func (b *AiBuff) Delete() error {
	_, err := db.Delete(b)
	return err
}

func (b *AiBuff) Update() error {
	_, err := db.Update(b)
	return err
}
func FindBuffByAIID(buffID, aiID int) (*AiBuff, error) {

	var buff *AiBuff
	query := `select * from hops.ai_buffs where id = $1 and ai_id = $2`

	if err := db.SelectOne(&buff, query, buffID, aiID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindBuffByID: %s", err.Error())
	}

	return buff, nil
}
func FindBuffsByAiPseudoID(aiID uint16) ([]*AiBuff, error) {

	var buffs []*AiBuff
	query := `select * from hops.ai_buffs where ai_id = $1`

	if _, err := db.Select(&buffs, query, aiID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindBuffsByAiPseudoID: %s", err.Error())
	}

	sort.Slice(buffs, func(i, j int) bool {
		return buffs[i].StartedAt+buffs[i].Duration <= buffs[j].StartedAt+buffs[j].Duration
	})

	return buffs, nil
}

func DeleteBuffsByAiPseudoID(aiID uint16) ([]*AiBuff, error) {
	query := `DELETE from hops.ai_buffs where ai_id = $1`

	_, err := db.Exec(query, aiID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func DeleteBuffByAiPseudoID(aiID uint16, buffID int) ([]*AiBuff, error) {
	query := `DELETE from hops.ai_buffs where id = $1 and ai_id = $2`

	_, err := db.Exec(query, buffID, aiID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
