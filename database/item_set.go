package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

var (
	ItemSets = make(map[int]*ItemSet)
)

type ItemSet struct {
	ID           int    `db:"id"`
	SetItemCount int    `db:"itemcount"`
	SetItemsID   string `db:"itemsid"`
	SetBonusID   string `db:"bonusid"`
}

func (e *ItemSet) Create() error {
	return db.Insert(e)
}

func getItemSet() error {
	var sets []*ItemSet
	query := `select * from data.item_set`

	if _, err := db.Select(&sets, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getItemSet: %s", err.Error())
	}

	for _, cr := range sets {
		ItemSets[cr.ID] = cr
	}

	return nil
}

func (e *ItemSet) GetSetItems() []int64 {
	items := strings.Trim(e.SetItemsID, "{}")
	sItems := strings.Split(items, ",")
	var arr []int64
	for _, sItem := range sItems {
		d, _ := strconv.ParseInt(sItem, 10, 64)
		//d, _ := strconv.Atoi(sItem)
		arr = append(arr, d)
	}
	return arr
}

func (e *ItemSet) GetSetBonus() []int64 {
	items := strings.Trim(e.SetBonusID, "{}")
	sItems := strings.Split(items, ",")

	var arr []int64
	for _, sItem := range sItems {
		d, _ := strconv.ParseInt(sItem, 10, 64)
		//d, _ := strconv.Atoi(sItem)
		arr = append(arr, d)
	}
	return arr
}
