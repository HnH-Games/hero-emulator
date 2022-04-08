package database

import (
	"encoding/binary"
	"sync"

	"hero-emulator/utils"
)

type Sale struct {
	ID     uint16
	Seller *Character
	Name   string
	Items  []*SaleItem
	Data   []byte
}

type SaleItem struct {
	SlotID int16
	Price  uint64
	IsSold bool
}

var (
	Sales  = make(map[uint16]*Sale)
	sMutex sync.RWMutex
)

func (s *Sale) Create() {
	sMutex.Lock()
	defer sMutex.Unlock()
	Sales[s.ID] = s
}

func FindSale(id uint16) *Sale {
	sMutex.RLock()
	defer sMutex.RUnlock()
	return Sales[id]
}

func (s *Sale) Delete() {
	sMutex.Lock()
	defer sMutex.Unlock()
	delete(Sales, s.ID)
}

func (s *Sale) SaleData() ([]byte, error) {

	slots, err := s.Seller.InventorySlots()
	if err != nil {
		return nil, err
	}

	data := GET_SALE_ITEMS
	data[8] = byte(len(s.Items))
	index, length := 9, int16(5)
	for i := 0; i < len(s.Items); i++ {
		/*if s.Items[i].IsSold {
			continue
		}*/
		slotID := s.Items[i].SlotID
		price := s.Items[i].Price
		item := slots[slotID]

		if slotID == 0 || price == 0 || item == nil || item.ItemID == 0 {
			*item = *NewSlot()
		}

		data.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), index) // item id
		index += 4
		data.Insert([]byte{0x00}, index)
		index++
		info := Items[item.ItemID]
		itemtypeinfo := 0
		if info == nil {
			itemtypeinfo = 0
		} else {
			itemtypeinfo = info.GetType()
		}
		if itemtypeinfo == PET_TYPE { // pet
			if pet := item.Pet; pet != nil {
				data.Insert([]byte{pet.Level, pet.Loyalty, pet.Fullness}, index) // pet level, loyalty and fullness
				index += 3
				data.Insert(utils.IntToBytes(uint64(slotID), 2, true), index) // slot id
				index += 2
				data.Insert(utils.IntToBytes(uint64(pet.HP), 2, true), index) // pet hp
				index += 2
				data.Insert(utils.IntToBytes(uint64(pet.CHI), 2, true), index) // pet chi
				index += 2
				data.Insert(utils.IntToBytes(uint64(pet.Exp), 8, true), index) // pet exp
				index += 8
				data.Insert([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, index) // padding
				index += 19
			}
		} else {
			data.Insert([]byte{0xA2}, index)
			index++
			data.Insert(utils.IntToBytes(uint64(item.Quantity), 2, true), index) // item quantity
			index += 2
			data.Insert(utils.IntToBytes(uint64(slotID), 2, true), index) // slot id
			index += 2
			data.Insert(item.GetUpgrades(), index) // item upgrades
			index += 15
			data.Insert([]byte{byte(item.SocketCount)}, index) // item socket count
			index++
			data.Insert(item.GetSockets(), index) // item sockets
			index += 15
			if item.ItemType != 0 {
				data.Overwrite(utils.IntToBytes(uint64(item.ItemType), 1, true), index-6)
				if item.ItemType == 2 {
					data.Overwrite(utils.IntToBytes(uint64(item.JudgementStat), 4, true), index-5)
				}
			}
		}

		data.Insert([]byte{0x00, 0x00, 0x00}, index) // padding
		index += 3
		if item.Appearance != 0 {
			data.Overwrite(utils.IntToBytes(uint64(item.Appearance), 4, true), index-4)
		}
		data.Insert(utils.IntToBytes(uint64(i), 2, true), index) // index
		index += 2

		data.Insert(utils.IntToBytes(price, 8, true), index) // item price
		index += 8
		length += 54
	}

	data.SetLength(int16(binary.Size(data) - 6))
	return data, nil
}
