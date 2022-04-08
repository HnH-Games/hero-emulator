package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	gorp "gopkg.in/gorp.v1"
)

type ItemMelting struct {
	ID                 int     `db:"id"`
	MeltedItems        string  `db:"melted_items"`
	ItemCounts         string  `db:"item_counts"`
	ProfitMultiplier   float64 `db:"profit_multiplier"`
	Probability        int     `db:"probability"`
	Cost               int64   `db:"cost"`
	SpecialItem        int     `db:"special_item"`
	SpecialProbability int     `db:"special_probability"`

	meltedItems []int `db:"-"`
	itemCounts  []int `db:"-"`
}

var (
	Meltings = make(map[int]*ItemMelting)
)

func (e *ItemMelting) Create() error {
	return db.Insert(e)
}

func (e *ItemMelting) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *ItemMelting) Delete() error {
	_, err := db.Delete(e)
	return err
}

func (e *ItemMelting) GetMeltedItems() ([]int, error) {

	if len(e.meltedItems) > 0 {
		return e.meltedItems, nil
	}

	data := fmt.Sprintf("[%s]", strings.Trim(e.MeltedItems, "{}"))
	err := json.Unmarshal([]byte(data), &e.meltedItems)
	if err != nil {
		return nil, err
	}

	return e.meltedItems, nil
}

func (e *ItemMelting) GetItemCounts() ([]int, error) {

	if len(e.itemCounts) > 0 {
		return e.itemCounts, nil
	}

	data := fmt.Sprintf("[%s]", strings.Trim(e.ItemCounts, "{}"))
	err := json.Unmarshal([]byte(data), &e.itemCounts)
	if err != nil {
		return nil, err
	}

	return e.itemCounts, nil
}

func getItemMeltings() error {
	var meltings []*ItemMelting
	query := `select * from data.item_meltings`

	if _, err := db.Select(&meltings, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getItemMeltings: %s", err.Error())
	}

	for _, m := range meltings {
		Meltings[m.ID] = m
	}

	return nil
}
