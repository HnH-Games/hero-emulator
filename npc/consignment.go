package npc

import (
	"encoding/json"
	"math"

	"hero-emulator/database"
	"hero-emulator/utils"
)

type (
	OpenConsignmentHandler      struct{}
	RegisterItemHandler         struct{}
	ClaimMenuHandler            struct{}
	BuyConsignmentItemHandler   struct{}
	ClaimConsignmentItemHandler struct{}
)

var (
	GET_CONS_ITEMS = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x3D, 0x00, 0x0A, 0x00, 0x55, 0xAA}
)

func (h *OpenConsignmentHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	category := int(data[6])
	itemNameLength := data[7]
	itemSearch := string(data[8 : 8+itemNameLength])

	index := 8 + itemNameLength

	minUpgLevel := int(data[index+3])
	maxUpgLevel := int(data[index+4])
	minPrice := uint64(utils.BytesToInt(data[index+6:index+14], true))
	maxPrice := uint64(utils.BytesToInt(data[index+14:index+21], true))
	orderBy := int(data[index+21])
	page := int(data[index+22]) + 1

	items, count, err := database.GetConsignmentItems(page, category, minUpgLevel, maxUpgLevel, orderBy, minPrice, maxPrice, itemSearch)
	if err != nil {
		return nil, err
	}

	totalPage := int(math.Ceil(float64(count) / 20))
	if len(items) > 20 {
		items = items[:20]
	}

	resp := GET_CONS_ITEMS

	resp.Insert(utils.IntToBytes(uint64(len(items)), 2, true), 8) // item count

	c, counter := 10, 0
	for i := 0; i < len(items); i++ {
		consignmentItem := items[i]

		item, err := database.FindInventorySlotByID(consignmentItem.ID)
		if err != nil {
			return nil, err
		}

		info := database.Items[int64(item.ItemID)]

		seller, err := database.FindCharacterByID(int(item.CharacterID.Int64))
		if err != nil || seller == nil {
			continue
		}

		resp.Insert(utils.IntToBytes(uint64(totalPage), 4, true), c) // page count
		c += 4

		resp.Insert(utils.IntToBytes(uint64(item.ID), 4, true), c) // consignment item id
		c += 4

		resp.Insert([]byte{0x15, 0x14, 0x00, 0x00}, c)
		c += 4

		resp.Insert([]byte(seller.Name), c) // seller name
		c += len(seller.Name)

		for j := len(seller.Name); j < 20; j++ {
			resp.Insert([]byte{0x00}, c)
			c++
		}

		resp.Insert(utils.IntToBytes(consignmentItem.Price, 8, true), c) // item price
		c += 8

		time := consignmentItem.ExpiresAt.Time.Format("2006-01-02 15:04:05") // expires at
		resp.Insert([]byte(time), c)
		c += 19

		resp.Insert([]byte{0x00, 0x01, 0x00, 0x00, 0x00, 0xAE, 0x21, 0xF5, 0x00}, c)
		c += 9

		resp.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), c) // item id
		c += 4

		resp.Insert([]byte{0x00, 0xA1}, c)
		c += 2

		resp.Insert(utils.IntToBytes(uint64(item.Quantity), 2, true), c) // item count
		c += 2

		pet := &database.PetSlot{}
		isPet := info.GetType() == database.PET_TYPE
		if isPet {
			json.Unmarshal(item.PetInfo, pet)
			resp[c-3] = pet.Level
			resp[c-2] = pet.Loyalty
			resp[c-1] = pet.Fullness
		}

		upgrades := item.GetUpgrades()
		resp.Insert(upgrades, c) // item upgrades
		c += 15

		if isPet {
			resp.Overwrite(utils.IntToBytes(uint64(pet.HP), 2, true), c-15)
			resp.Overwrite(utils.IntToBytes(uint64(pet.CHI), 2, true), c-13)
			resp.Overwrite(utils.IntToBytes(uint64(pet.Exp), 8, true), c-11)
		}

		sockets := item.GetSockets()
		/*socketCount := len(funk.Filter(sockets, func(s byte) bool {
			return s > 0
		}).([]byte))*/

		resp.Insert([]byte{byte(item.SocketCount)}, c) // socket count
		if isPet {
			resp[c] = 0
		}
		c++

		resp.Insert(sockets, c)
		c += 15
		if item.ItemType != 0 {
			resp.Overwrite(utils.IntToBytes(uint64(item.ItemType), 1, true), c-6)
			if item.ItemType == 2 {
				resp.Overwrite(utils.IntToBytes(uint64(item.JudgementStat), 4, true), c-5)
			}
		}
		resp.Insert([]byte{0x00, 0x00, 0x00}, c)
		c += 3
		counter++
	}

	length := int16(0x6E*counter + 6)
	resp.SetLength(length)
	return resp, nil
}

func (h *RegisterItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	slotID := int16(utils.BytesToInt(data[6:8], true))
	item := slots[slotID]

	if slotID == 0 || item == nil {
		return nil, nil
	}

	price := uint64(utils.BytesToInt(data[8:16], true))
	return s.Character.RegisterItem(item, price, slotID)
}

func (h *ClaimMenuHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	return s.Character.ClaimMenu()
}

func (h *BuyConsignmentItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	consignmentID := int(utils.BytesToInt(data[6:10], true))
	return s.Character.BuyConsignmentItem(consignmentID)
}

func (h *ClaimConsignmentItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	isCancel := data[7] == 0
	consignmentID := int(utils.BytesToInt(data[8:12], true))

	return s.Character.ClaimConsignmentItem(consignmentID, isCancel)
}
