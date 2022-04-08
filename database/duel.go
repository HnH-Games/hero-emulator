package database

import "hero-emulator/utils"

type Duel struct {
	EnemyID    int
	Coordinate utils.Location
	Started    bool
}
