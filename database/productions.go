package database

import (
	"database/sql"
	"encoding/json"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

type Production struct {
	ID          int    `db:"id"`
	Materials   []byte `db:"materials"`
	Probability int    `db:"probability"`
	Cost        int64  `db:"cost"`
	Production  int    `db:"production"`

	materials []*ProductionMaterial `db:"-"`
}

type ProductionMaterial struct {
	ID    int  `json:"ID"`
	Count byte `json:"Count"`
}

var (
	Productions = make(map[int]*Production)
)

func (e *Production) Create() error {
	return db.Insert(e)
}

func (e *Production) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *Production) Delete() error {
	_, err := db.Delete(e)
	return err
}

func (e *Production) GetMaterials() ([]*ProductionMaterial, error) {
	if len(e.materials) > 0 {
		return e.materials, nil
	}

	err := json.Unmarshal(e.Materials, &e.materials)
	if err != nil {
		return nil, err
	}

	return e.materials, nil
}

func getProductions() error {
	var prods []*Production
	query := `select * from data.productions`

	if _, err := db.Select(&prods, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getProductions: %s", err.Error())
	}

	for _, p := range prods {
		Productions[p.ID] = p
	}

	return nil
}

func RefreshProductions() error {
	var prods []*Production
	query := `select * from data.productions`

	if _, err := db.Select(&prods, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getProductions: %s", err.Error())
	}

	for _, p := range prods {
		Productions[p.ID] = p
	}

	return nil
}
