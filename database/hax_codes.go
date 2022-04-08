package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

type HaxCode struct {
	ID                   int    `db:"id"`
	Code                 string `db:"code"`
	SaleMultiplier       int    `db:"sale_multiplier"`
	ExtractionMultiplier int    `db:"extraction_multiplier"`
	ExtractedItem        int    `db:"extracted_item"`
}

var (
	HaxCodes = make(map[int]*HaxCode)
)

func (e *HaxCode) Create() error {
	return db.Insert(e)
}

func (e *HaxCode) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *HaxCode) Delete() error {
	_, err := db.Delete(e)
	return err
}

func getHaxCodes() error {

	var haxcodes []*HaxCode
	query := `select * from data.hax_codes`

	if _, err := db.Select(&haxcodes, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getHaxCodes: %s", err.Error())
	}

	for _, h := range haxcodes {
		HaxCodes[h.ID] = h
	}

	return nil
}
