package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

var (
	GamblingItems = make(map[int]*Gambling)
	rewardCounts  = map[int64]uint{17502658: 30, 17502527: 50, 17502659: 60, 17502528: 100, 17502529: 150, 17502530: 200, 17200188: 10, 17200189: 30,
		17501007: 10, 17502516: 50, 17502517: 75, 17502518: 100, 17501008: 10, 17502513: 50, 17502514: 75, 17502515: 100, 17502555: 600, 17502557: 600}
	rewardCounts2 = map[int64]map[int64]uint{
		13370000: {32049: 3},
		13370001: {1031: 2, 32240: 3, 32050: 3, 221: 3, 222: 3, 223: 3},
		13370002: {92000013: 2},
		13370003: {92000012: 2, 253: 2, 17502731: 2, 17502733: 2, 240: 2, 241: 2, 232: 2, 17200187: 2},
		13370004: {92000063: 2, 17502731: 4, 417502733: 4, 240: 4, 241: 4, 253: 4, 232: 4, 17200187: 4},
		13370005: {15700001: 2},
	}
)

type Gambling struct {
	ID     int    `db:"id"`
	Cost   uint64 `db:"cost"`
	DropID int    `db:"drop_id"`
}

func (e *Gambling) Create() error {
	return db.Insert(e)
}

func (e *Gambling) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *Gambling) Delete() error {
	_, err := db.Delete(e)
	return err
}

func getGamblingItems() error {
	var gamblings []*Gambling
	query := `select * from data.gambling`

	if _, err := db.Select(&gamblings, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getGamblingItems: %s", err.Error())
	}

	for _, g := range gamblings {
		GamblingItems[g.ID] = g
	}

	return nil
}

func RefreshGamblingItems() error {
	var gamblings []*Gambling
	query := `select * from data.gambling`

	if _, err := db.Select(&gamblings, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getGamblingItems: %s", err.Error())
	}

	for _, g := range gamblings {
		GamblingItems[g.ID] = g
	}

	return nil
}
