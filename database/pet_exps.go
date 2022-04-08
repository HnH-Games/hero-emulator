package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

var (
	PetExps = make(map[int16]*PetExpInfo)
)

type PetExpInfo struct {
	Level         int16 `db:"level"`
	ReqExpEvo1    int   `db:"req_exp_evo1"`
	ReqExpEvo2    int   `db:"req_exp_evo2"`
	ReqExpEvo3    int   `db:"req_exp_evo3"`
	ReqExpHt      int   `db:"req_exp_ht"`
	ReqExpDivEvo1 int   `db:"req_exp_div_evo1"`
	ReqExpDivEvo2 int   `db:"req_exp_div_evo2"`
	ReqExpDivEvo3 int   `db:"req_exp_div_evo3"`
}

func (p *PetExpInfo) Create() error {
	return db.Insert(p)
}

func (p *PetExpInfo) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(p)
}

func (p *PetExpInfo) Update() error {
	_, err := db.Update(p)
	return err
}

func (p *PetExpInfo) Delete() error {
	_, err := db.Delete(p)
	return err
}

func GetAllPetExps() error {

	var arr []*PetExpInfo
	query := `select * from "data".pet_exp_table`

	if _, err := db.Select(&arr, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("GetAllPetExps: %s", err.Error())
	}

	for _, petExp := range arr {
		PetExps[petExp.Level] = petExp
	}

	return nil
}
