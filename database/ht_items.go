package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

type HtItem struct {
	ID        int  `db:"id"`
	HTID      int  `db:"ht_id"`
	Cash      int  `db:"cash"`
	IsActive  bool `db:"is_active"`
	IsNew     bool `db:"is_new"`
	IsPopular bool `db:"is_popular"`
}

var (
	HTItems = make(map[int]*HtItem)
)

func (e *HtItem) Create() error {
	return db.Insert(e)
}

func (e *HtItem) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *HtItem) Update() error {
	_, err := db.Update(e)
	return err
}

func (e *HtItem) Delete() error {
	_, err := db.Delete(e)
	return err
}

func getHTItems() error {
	var htitems []*HtItem
	query := `select * from data.ht_shop`

	if _, err := db.Select(&htitems, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getHTItems: %s", err.Error())
	}

	for _, h := range htitems {
		HTItems[h.ID] = h
	}

	return nil
}

func RefreshHTItems() error {
	var htitems []*HtItem
	query := `select * from data.ht_shop`

	if _, err := db.Select(&htitems, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getHTItems: %s", err.Error())
	}

	for _, h := range htitems {
		HTItems[h.ID] = h
	}

	return nil
}
