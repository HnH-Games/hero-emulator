package database

import (
	"math"

	"hero-emulator/utils"

	"github.com/thoas/go-funk"
)

type Drop struct {
	ID       int
	Server   int
	Map      int16
	Location utils.Location
	Item     *InventorySlot
	Claimer  *Character
}

func (drop *Drop) GenerateIDForDrop(server int, mapID int16) {
	drMutex.Lock()
	defer drMutex.Unlock()
	for {
		i := uint16(utils.RandInt(1, math.MaxUint16))
		if _, ok := DropRegister[server][mapID][i]; !ok {
			DropRegister[server][mapID][i] = drop
			drop.ID = int(i)
			return
		}
	}
}

func GetDrop(server int, mapID int16, dropID uint16) *Drop {
	drMutex.RLock()
	drop, ok := DropRegister[server][mapID][dropID]
	drMutex.RUnlock()

	if ok {
		return drop
	}

	return &Drop{}
}

func GetDropsInMap(server int, mapID int16) []*Drop {
	drMutex.RLock()
	drops := funk.Values(DropRegister[server][mapID]).([]*Drop)
	drMutex.RUnlock()

	return drops
}

func RemoveFromDropRegister(server int, mapID int16, dropID uint16) {
	drMutex.Lock()
	defer drMutex.Unlock()
	delete(DropRegister[server][mapID], dropID)
}
