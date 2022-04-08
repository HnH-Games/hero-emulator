package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

var (
	Pets = make(map[int64]*Pet)
)

type Pet struct {
	ID            int64  `db:"id"`
	Name          string `db:"name"`
	Evolution     int16  `db:"evolution"`
	Level         int16  `db:"level"`
	TargetLevel   int16  `db:"target_level"`
	EvolvedID     int64  `db:"evolved_id"`
	BaseSTR       int    `db:"base_str"`
	AdditionalSTR int    `db:"additional_str"`
	BaseDEX       int    `db:"base_dex"`
	AdditionalDEX int    `db:"additional_dex"`
	BaseINT       int    `db:"base_int"`
	AdditionalINT int    `db:"additional_int"`
	BaseHP        int    `db:"base_hp"`
	AdditionalHP  int    `db:"additional_hp"`
	BaseChi       int    `db:"base_chi"`
	AdditionalChi int    `db:"additional_chi"`
	SkillID       int    `db:"skill_id"`
	Combat        bool   `db:"combat"`
}

func (e *Pet) Create() error {
	return db.Insert(e)
}

func (e *Pet) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *Pet) Update() error {
	_, err := db.Update(e)
	return err
}

func (e *Pet) Delete() error {
	_, err := db.Delete(e)
	return err
}

func GetAllPets() error {

	var arr []*Pet
	query := `select * from "data".pets`

	if _, err := db.Select(&arr, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("GetAllPets: %s", err.Error())
	}

	for _, pet := range arr {
		Pets[pet.ID] = pet
	}

	return nil
}
