package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	gorp "gopkg.in/gorp.v1"
)

type CraftItem struct {
	ID          int    `db:"id"`
	Materials   []byte `db:"materials"`
	Production  string `db:"production"`
	Probability string `db:"probabilities"`
	Cost        int64  `db:"cost"`

	materials []*CraftItemMaterial `db:"-"`
}

type CraftItemMaterial struct {
	ID    int  `json:"ID"`
	Count byte `json:"Count"`
}

var (
	CraftItems = make(map[int]*CraftItem)
)

func (e *CraftItem) GetItems() []int {
	items := strings.Trim(e.Production, "{}")
	sItems := strings.Split(items, ",")

	var arr []int
	for _, sItem := range sItems {
		d, _ := strconv.Atoi(sItem)
		arr = append(arr, d)
	}
	return arr
}

func (e *CraftItem) GetProbabilities() []int {
	probs := strings.Trim(e.Probability, "{}")
	sProbs := strings.Split(probs, ",")

	var arr []int
	for _, sProb := range sProbs {
		d, _ := strconv.Atoi(sProb)
		arr = append(arr, d)
	}
	return arr
}

func (e *CraftItem) Create() error {
	return db.Insert(e)
}

func (e *CraftItem) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *CraftItem) Delete() error {
	_, err := db.Delete(e)
	return err
}

func (e *CraftItem) GetMaterials() ([]*CraftItemMaterial, error) {
	if len(e.materials) > 0 {
		return e.materials, nil
	}

	err := json.Unmarshal(e.Materials, &e.materials)
	if err != nil {
		return nil, err
	}

	return e.materials, nil
}

func getCraftItem() error {
	var crafts []*CraftItem
	query := `select * from data.craft_items`

	if _, err := db.Select(&crafts, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getCraftItems: %s", err.Error())
	}

	for _, cr := range crafts {
		CraftItems[cr.ID] = cr
	}

	return nil
}

func RefreshCraftItem() error {
	var crafts []*CraftItem
	query := `select * from data.craft_items`

	if _, err := db.Select(&crafts, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getCraftItems: %s", err.Error())
	}

	for _, cr := range crafts {
		CraftItems[cr.ID] = cr
	}

	return nil
}
