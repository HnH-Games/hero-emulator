package database

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"hero-emulator/utils"

	gorp "gopkg.in/gorp.v1"
)

var (
	SavePoints = make(map[uint8]*SavePoint)
)

type SavePoint struct {
	ID    uint8  `db:"id"`
	Point string `db:"point"`
}

func (e *SavePoint) SetPoint(point *utils.Location) {
	e.Point = fmt.Sprintf("%.2f,%.2f", point.X, point.Y)
}

func (e *SavePoint) Create() error {
	return db.Insert(e)
}

func (e *SavePoint) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *SavePoint) Delete() error {
	_, err := db.Delete(e)
	return err
}

func getAllSavePoints() error {

	query := `select * from data.save_points`

	t := []*SavePoint{}

	if _, err := db.Select(&t, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getAllSavePoints: %s", err.Error())
	}

	for _, s := range t {
		SavePoints[s.ID] = s
	}

	return nil
}

func ConvertPointToLocation(point string) *utils.Location {

	location := &utils.Location{0, 0}
	parts := strings.Split(strings.Trim(point, "()"), ",")
	location.X, _ = strconv.ParseFloat(parts[0], 64)
	location.Y, _ = strconv.ParseFloat(parts[1], 64)
	return location
}
