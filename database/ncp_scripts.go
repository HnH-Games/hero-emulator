package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

type NPCScript struct {
	ID     int    `db:"id"`
	Script []byte `db:"script"`
}

var (
	NPCScripts = make(map[int]*NPCScript)
)

func (e *NPCScript) Create() error {
	return db.Insert(e)
}

func (e *NPCScript) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *NPCScript) Update() error {
	_, err := db.Update(e)
	return err
}

func (e *NPCScript) Delete() error {
	_, err := db.Delete(e)
	return err
}

func getScripts() error {
	var scripts []*NPCScript
	query := `select * from data.npc_scripts`

	if _, err := db.Select(&scripts, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getScripts: %s", err.Error())
	}

	for _, s := range scripts {
		NPCScripts[s.ID] = s
	}

	return nil
}

func RefreshScripts() error {
	var scripts []*NPCScript
	query := `select * from data.npc_scripts`

	if _, err := db.Select(&scripts, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getScripts: %s", err.Error())
	}

	for _, s := range scripts {
		NPCScripts[s.ID] = s
	}

	return nil
}
