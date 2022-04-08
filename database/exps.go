package database

import (
	"database/sql"
	"sync"

	gorp "gopkg.in/gorp.v1"
)

type ExpInfo struct {
	Level       int16 `db:"level"`
	Exp         int64 `db:"exp"`
	SkillPoints int   `db:"skill_points"`

	StatPoints   int `db:"stat_points"`
	NaturePoints int `db:"nature_points"`
}

var (
	EXPs     = make(map[int16]*ExpInfo)
	expMutex sync.Mutex
)

func (e *ExpInfo) Create() error {
	return db.Insert(e)
}

func (e *ExpInfo) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *ExpInfo) Update() error {
	_, err := db.Update(e)
	return err
}

func (e *ExpInfo) Delete() error {
	_, err := db.Delete(e)
	return err
}

func getExps() error {

	var arr []*ExpInfo
	query := `select * from data.exp_table`

	if _, err := db.Select(&arr, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return nil
	}

	for _, e := range arr {
		EXPs[e.Level] = e
	}
	return nil
}

func GetExps() error {

	var arr []*ExpInfo
	query := `select * from data.exp_table`

	if _, err := db.Select(&arr, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return nil
	}

	for _, e := range arr {
		EXPs[e.Level] = e
	}
	return nil
}
