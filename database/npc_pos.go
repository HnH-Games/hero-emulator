package database

import (
	"database/sql"
	"fmt"

	"hero-emulator/utils"

	gorp "gopkg.in/gorp.v1"
)

var (
	NPCPos []*NpcPosition
)

type NpcPosition struct {
	ID          int     `db:"id"`
	NPCID       int     `db:"npc_id"`
	MapID       int16   `db:"map"`
	Rotation    float64 `db:"rotation"`
	MinLocation string  `db:"min_location"`
	MaxLocation string  `db:"max_location"`
	Count       int16   `db:"count"`
	RespawnTime int     `db:"respawn_time"`
	IsNPC       bool    `db:"is_npc"`
	Attackable  bool    `db:"attackable"`
	Faction     int     `db:"faction"`

	PseudoID uint16 `db:"-"`
}

func (e *NpcPosition) SetLocations(min, max *utils.Location) {
	e.MinLocation = fmt.Sprintf("%.2f,%.2f", min.X, min.Y)
	e.MaxLocation = fmt.Sprintf("%.2f,%.2f", max.X, max.Y)
}

func (e *NpcPosition) Create() error {
	return db.Insert(e)
}

func (e *NpcPosition) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *NpcPosition) Update() error {
	_, err := db.Update(e)
	return err
}

func (e *NpcPosition) Delete() error {
	_, err := db.Delete(e)
	return err
}

func GetAllNPCPos() ([]*NpcPosition, error) {

	var arr []*NpcPosition
	query := `select * from "data".npc_pos_table order by id`

	if _, err := db.Select(&arr, query); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetAllNpcPos: %s", err.Error())
	}

	return arr, nil
}

func GetAllAIPos() ([]*NpcPosition, error) {

	var arr []*NpcPosition
	query := `select * from "data".npc_pos_table WHERE is_npc = '0' order by id `

	if _, err := db.Select(&arr, query); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("GetAllNpcPos: %s", err.Error())
	}

	return arr, nil
}

func FindNPCPosByID(id int) (*NpcPosition, error) {

	var pos *NpcPosition
	query := `select * from "data".npc_pos_table where id = $1`

	if err := db.SelectOne(&pos, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindNPCPosByID: %s", err.Error())
	}

	return pos, nil
}

func FindNPCPosInMap(mapID int16) ([]*NpcPosition, error) {

	var arr []*NpcPosition
	query := `select * from "data".npc_pos_table where "map" = $1`

	if _, err := db.Select(&arr, query, mapID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindNPCPosInMap: %s", err.Error())
	}

	return arr, nil
}
