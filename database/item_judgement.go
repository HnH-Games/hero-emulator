package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

var (
	ItemJudgements = make(map[int]*ItemJudgement)
)

type ItemJudgement struct {
	ID               int     `db:"id"`
	Name             string  `db:"name"`
	AttackPlus       int     `db:"attack_plus"`
	AccuracyPlus     int     `db:"accuracy_plus"`
	StrPlus          int     `db:"str_plus"`
	DexPlus          int     `db:"dex_plus"`
	IntPlus          int     `db:"int_plus"`
	ExtraDef         int     `db:"extra_def"`
	MaxHP            int     `db:"Max_HP"`
	MaxChi           int     `db:"Max_CHI"`
	ExtraArtsDef     int     `db:"extra_arts_def"`
	ExtraDodge       int     `db:"extra_dodge"`
	ExtraAttackSpeed int     `db:"extra_attackspeed"`
	WindPlus         int     `db:"wind_plus"`
	WaterPlus        int     `db:"water_plus"`
	FirePlus         int     `db:"fire_plus"`
	ExtraArtsRange   float64 `db:"extra_arts_range"`
	Unknown          int     `db:"ismeretlen"`
	Probabilities    int64   `db:"probabilities"`
}

func (e *ItemJudgement) Create() error {
	return db.Insert(e)
}

func (e *ItemJudgement) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *ItemJudgement) Delete() error {
	_, err := db.Delete(e)
	return err
}

func (e *ItemJudgement) Update() error {
	_, err := db.Update(e)
	return err
}

func getItemJudgements() error {
	var Itemjudgements []*ItemJudgement
	query := `select * from data.item_judgement`

	if _, err := db.Select(&Itemjudgements, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getitem_judgement: %s", err.Error())
	}

	for _, b := range Itemjudgements {
		ItemJudgements[b.ID] = b
	}

	return nil
}

func RefresItemJudgements() error {
	var itemJudgements []*ItemJudgement
	query := `select * from data.item_judgement`

	if _, err := db.Select(&itemJudgements, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("refreshItemJudgements: %s", err.Error())
	}

	for _, b := range itemJudgements {
		ItemJudgements[b.ID] = b
	}

	return nil
}
