package database

import (
	"database/sql"
	"fmt"

	"hero-emulator/utils"

	gorp "gopkg.in/gorp.v1"
)

type Gate struct {
	ID        int    `db:"id"`
	TargetMap uint8  `db:"target_map"`
	Point     string `db:"point"`
}

var (
	Gates = make(map[int]*Gate)
)

func (e *Gate) SetPoint(point *utils.Location) {
	e.Point = fmt.Sprintf("%.2f,%.2f", point.X, point.Y)
}

func (e *Gate) Create() error {
	return db.Insert(e)
}

func (e *Gate) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *Gate) Delete() error {
	_, err := db.Delete(e)
	return err
}

func getGates() error {
	var gates []*Gate
	query := `select * from data.gates`

	if _, err := db.Select(&gates, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getGates: %s", err.Error())
	}

	for _, g := range gates {
		Gates[g.ID] = g
	}

	return nil
}
