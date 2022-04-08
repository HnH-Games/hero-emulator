package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/thoas/go-funk"
)

var (
	ShopItems = make(map[int]*ShopItem)
)

type ShopItem struct {
	Type  int    `db:"type"`
	Items string `db:"items"`
}

func (e *ShopItem) Create() error {
	return db.Insert(e)
}

func (e *ShopItem) New(t int, items []int) {
	e.Type = t
	items = funk.FilterInt(items, func(t int) bool {
		return t > 0
	})

	e.Items = strings.Trim(strings.Join(strings.Fields(fmt.Sprint(items)), ","), "[]")
	e.Items = fmt.Sprintf("{%s}", e.Items)
}

func (e *ShopItem) GetItems() []int {
	items := strings.Trim(e.Items, "{}")
	sItems := strings.Split(items, ",")

	var arr []int
	for _, sItem := range sItems {
		d, _ := strconv.Atoi(sItem)
		arr = append(arr, d)
	}
	return arr
}

func getAllShopItems() error {
	var shopItems []*ShopItem
	query := `select * from data.shop_items`

	if _, err := db.Select(&shopItems, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getAllShopItems: %s", err.Error())
	}

	for _, s := range shopItems {
		ShopItems[s.Type] = s
	}

	return nil
}

func GetAllShopItems() error {
	var shopItems []*ShopItem
	query := `select * from data.shop_items`

	if _, err := db.Select(&shopItems, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getAllShopItems: %s", err.Error())
	}

	for _, s := range shopItems {
		ShopItems[s.Type] = s
	}

	return nil
}
