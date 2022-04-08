package database

import (
	"database/sql"
	"fmt"

	null "gopkg.in/guregu/null.v3"
)

var (
	RelicsLog = make(map[int]*RelicLog)
)

type RelicLog struct {
	CharID   int       `db:"id"`
	ItemID   int64     `db:"item_id"`
	DropTime null.Time `db:"drop_time"`
}

func (e *RelicLog) Create() error {
	return db.Insert(e)
}

func getRelicLog() error {
	var relics []*RelicLog
	query := `SELECT * FROM data.relic_log WHERE drop_time >= now()::date + interval '0 hour';`

	if _, err := db.Select(&relics, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getRelicLog: %s", err.Error())
	}

	for _, r := range relics {
		RelicsLog[len(RelicsLog)] = r
	}

	return nil
}
