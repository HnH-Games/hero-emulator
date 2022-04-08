package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/thoas/go-funk"
)

var (
	Shops = make(map[int]*Shop)
)

type Shop struct {
	ID    int    `db:"id"`
	Name  string `db:"name"`
	Types string `db:"types"`
}

func (e *Shop) Create() error {
	return db.Insert(e)
}

func (e *Shop) New(id int, name string, types []int) {
	e.ID = id
	types = funk.FilterInt(types, func(t int) bool {
		return t > 0
	})

	e.Types = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(types)), ","), "[]")
	e.Types = fmt.Sprintf("{%s}", e.Types)
}

func (e *Shop) GetTypes() []int {
	types := strings.Trim(e.Types, "{}")
	sTypes := strings.Split(types, ",")

	var arr []int
	for _, sType := range sTypes {
		d, _ := strconv.Atoi(sType)
		arr = append(arr, d)
	}
	return arr
}

func getAllShops() error {
	var shops []*Shop
	query := `select * from data.shop_table`

	if _, err := db.Select(&shops, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getAllShops: %s", err.Error())
	}

	for _, s := range shops {
		Shops[s.ID] = s
	}

	return nil
}

func (e *Shop) IsPurchasable(itemID int) bool {

	types := e.GetTypes()
	for _, t := range types {
		shopItems := ShopItems[t]
		items := shopItems.GetItems()

		if funk.Contains(items, itemID) {
			return true
		}
	}

	return false
}
