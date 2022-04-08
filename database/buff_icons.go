package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

var (
	BuffIcons = make(map[int]*BuffIcon)
)

type BuffIcon struct {
	SkillID int `db:"skill_id"`
	IconID  int `db:"icon_id"`
}

func (e *BuffIcon) Create() error {
	return db.Insert(e)
}

func (e *BuffIcon) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *BuffIcon) Delete() error {
	_, err := db.Delete(e)
	return err
}

func getBuffIcons() error {
	var buffIcons []*BuffIcon
	query := `select * from data.buff_icons`

	if _, err := db.Select(&buffIcons, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getBuffIcons: %s", err.Error())
	}

	for _, b := range buffIcons {
		BuffIcons[b.SkillID] = b
	}

	return nil
}
