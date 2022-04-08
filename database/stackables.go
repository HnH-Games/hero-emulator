package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/thoas/go-funk"
	gorp "gopkg.in/gorp.v1"
)

type Stackable struct {
	ID  int    `db:"id"`
	UIF string `db:"uif"`
}

var (
	Stackables = make(map[int]*Stackable)
)

func (e *Stackable) Create() error {
	return db.Insert(e)
}

func (e *Stackable) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *Stackable) Delete() error {
	_, err := db.Delete(e)
	return err
}

func FindStackableByUIF(uif string) *Stackable {

	if stackable, ok := funk.Find(funk.Values(Stackables), func(stackable *Stackable) bool {
		return strings.ToLower(stackable.UIF) == strings.ToLower(uif)
	}).(*Stackable); ok {
		return stackable
	}

	return nil
}

func getStackables() error {
	var stackables []*Stackable
	query := `select * from data.stackables`

	if _, err := db.Select(&stackables, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getStackables: %s", err.Error())
	}

	for _, s := range stackables {
		Stackables[s.ID] = s
	}

	return nil
}
