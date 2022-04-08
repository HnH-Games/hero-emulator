package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/thoas/go-funk"
	gorp "gopkg.in/gorp.v1"
)

var (
	Drops = make(map[int]*DropInfo)
)

type DropInfo struct {
	ID            int    `db:"id"`
	Items         string `db:"items"`
	Probabilities string `db:"probabilities"`
}

func (e *DropInfo) New(id int, items, probabilities []int) {
	e.ID = id
	items = funk.FilterInt(items, func(item int) bool {
		return item > 0
	})
	probabilities = funk.FilterInt(probabilities, func(prob int) bool {
		return prob > 0
	})
	e.Items = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(items)), ","), "[]")
	e.Items = fmt.Sprintf("{%s}", e.Items)
	e.Probabilities = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(probabilities)), ","), "[]")
	e.Probabilities = fmt.Sprintf("{%s}", e.Probabilities)
}

func (e *DropInfo) GetItems() []int {
	items := strings.Trim(e.Items, "{}")
	sItems := strings.Split(items, ",")

	var arr []int
	for _, sItem := range sItems {
		d, _ := strconv.Atoi(sItem)
		arr = append(arr, d)
	}
	return arr
}

func (e *DropInfo) GetProbabilities() []int {
	probs := strings.Trim(e.Probabilities, "{}")
	sProbs := strings.Split(probs, ",")

	var arr []int
	for _, sProb := range sProbs {
		d, _ := strconv.Atoi(sProb)
		arr = append(arr, d)
	}
	return arr
}

func (e *DropInfo) Create() error {
	return db.Insert(e)
}

func (e *DropInfo) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *DropInfo) Update() error {
	_, err := db.Update(e)
	return err
}

func (e *DropInfo) Delete() error {
	_, err := db.Delete(e)
	return err
}

func getAllDrops() error {
	var drops []*DropInfo
	query := `select * from data.drops`

	if _, err := db.Select(&drops, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getAllDrops: %s", err.Error())
	}

	for _, d := range drops {
		Drops[d.ID] = d
	}

	return nil
}
func RefreshAllDrops() error {
	var drops []*DropInfo
	query := `select * from data.drops`

	if _, err := db.Select(&drops, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getAllDrops: %s", err.Error())
	}

	for _, d := range drops {
		Drops[d.ID] = d
	}

	return nil
}
