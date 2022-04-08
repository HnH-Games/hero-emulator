package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

type Fusion struct {
	Item1            int64 `db:"item1"`
	Item2            int64 `db:"item2"`
	Count2           int16 `db:"count2"`
	Item3            int64 `db:"item3"`
	Count3           int16 `db:"count3"`
	SpecialItem      int64 `db:"special_item"`
	SpecialItemCount int16 `db:"special_item_count"`
	Probability      int   `db:"probability"`
	Cost             int64 `db:"cost"`
	Production       int64 `db:"production"`
	DestroyOnFail    bool  `db:"destroy_on_fail"`
}

var (
	Fusions = make(map[int64]*Fusion)
)

func (e *Fusion) Create() error {
	return db.Insert(e)
}

func (e *Fusion) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *Fusion) Update() error {
	_, err := db.Update(e)
	return err
}

func (e *Fusion) Delete() error {
	_, err := db.Delete(e)
	return err
}

func getAdvancedFusions() error {
	var fusions []*Fusion
	query := `select * from data.advanced_fusion`

	if _, err := db.Select(&fusions, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getProductions: %s", err.Error())
	}

	for _, f := range fusions {
		Fusions[f.Item1] = f
	}

	return nil
}

func RefreshAdvancedFusions() error {
	var fusions []*Fusion
	query := `select * from data.advanced_fusion`

	if _, err := db.Select(&fusions, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getProductions: %s", err.Error())
	}

	for _, f := range fusions {
		Fusions[f.Item1] = f
	}

	return nil
}
