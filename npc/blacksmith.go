package npc

import (
	"hero-emulator/database"
	"hero-emulator/utils"
)

type (
	StrengthenHandler     struct{}
	ProductionHandler     struct{}
	AdvancedFusionHandler struct{}
	DismantleHandler      struct{}
	ExtractionHandler     struct{}
)

func (h *StrengthenHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slotID := utils.BytesToInt(data[6:8], true)
	stoneCount := data[8]

	index := 9
	var (
		stoneSlots []int64
		stones     []*database.InventorySlot
	)

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	if slots[slotID].ItemID == 0 {
		return nil, nil
	}

	for i := 0; i < int(stoneCount); i++ {
		id := utils.BytesToInt(data[index:index+2], true) // stone slot id

		if slots[id].ItemID == 0 {
			continue
		}

		stoneSlots = append(stoneSlots, id)
		stones = append(stones, slots[id])
		index += 2
	}

	var luck *database.InventorySlot
	lSlot := utils.BytesToInt(data[index:index+2], true)
	if lSlot == 0 {
		luck = nil
	} else {
		luck = slots[lSlot]
	}
	index += 2

	var protection *database.InventorySlot
	pSlot := utils.BytesToInt(data[index:index+2], true)
	if pSlot == 0 {
		protection = nil
	} else {
		protection = slots[pSlot]
	}

	resp := utils.Packet{}
	upgrageData, err := s.Character.BSUpgrade(slotID, stones, luck, protection, stoneSlots, lSlot, pSlot)
	if err != nil {
		return nil, err
	}

	resp.Concat(upgrageData)

	return resp, nil
}

func (h *ProductionHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	count := data[6]
	bookSlot := int16(utils.BytesToInt(data[7:9], true))
	book := slots[bookSlot]
	if book.ItemID == 0 {
		return nil, nil
	}

	materialCounts, materialSlots := []uint{}, []int16{}
	materials := []*database.InventorySlot{}

	index := 9
	for i := 1; i < int(count); i++ {
		materialSlot := int16(utils.BytesToInt(data[index:index+2], true))
		index += 2

		count := uint(data[index])
		index++

		if materialSlot > 0 {
			materials = append(materials, slots[materialSlot])
			materialSlots = append(materialSlots, materialSlot)
			materialCounts = append(materialCounts, count)
		}
	}

	var special *database.InventorySlot
	specialSlot := int16(utils.BytesToInt(data[index:index+2], true))
	if specialSlot == 0 {
		special = nil
	} else {
		special = slots[specialSlot]
	}
	index += 2

	prodSlot := int16(utils.BytesToInt(data[index:index+2], true))

	resp := utils.Packet{}
	prodData, err := s.Character.BSProduction(book, materials, special, prodSlot, bookSlot, specialSlot, materialSlots, materialCounts)
	if err != nil {
		return nil, err
	}

	resp.Concat(prodData)
	return resp, nil
}

func (h *AdvancedFusionHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	var slotIDs []int16
	var items []*database.InventorySlot

	index := 7
	for i := 0; i < 3; i++ {
		slot := int16(utils.BytesToInt(data[index:index+2], true))
		index += 2

		if slot > 0 {
			slotIDs = append(slotIDs, slot)
			items = append(items, slots[slot])
		}
	}

	var special *database.InventorySlot
	specialSlot := int16(utils.BytesToInt(data[index:index+2], true))
	index += 2
	if specialSlot == 0 {
		special = nil
	} else {
		special = slots[specialSlot]
	}

	prodSlot := int16(utils.BytesToInt(data[index:index+2], true))
	index += 2

	resp := utils.Packet{}
	fusionData, success, err := s.Character.AdvancedFusion(items, special, prodSlot)
	if err != nil {
		return nil, err
	}
	resp.Concat(fusionData)

	fusion := database.Fusions[items[0].ItemID]
	if success || (!success && fusion.DestroyOnFail) {
		for _, id := range slotIDs {
			itemData, err := s.Character.RemoveItem(id)
			if err != nil {
				return nil, err
			}

			resp.Concat(itemData)
		}
	}

	if special != nil {
		itemData, err := s.Character.RemoveItem(specialSlot)
		if err != nil {
			return nil, err
		}

		resp.Concat(itemData)
	}

	return resp, nil
}

func (h *DismantleHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	itemSlot := int16(utils.BytesToInt(data[6:8], true))
	if itemSlot == 0 {
		return nil, nil
	} else if slots[itemSlot].ItemID == 0 {
		return nil, nil
	}

	var special *database.InventorySlot
	specialSlot := int16(utils.BytesToInt(data[8:10], true))
	if specialSlot == 0 {
		special = nil
	} else {
		special = slots[specialSlot]
	}

	resp := utils.Packet{}
	dismantleData, _, err := s.Character.Dismantle(slots[itemSlot], special)
	if err != nil {
		return nil, err
	}

	resp.Concat(dismantleData)

	itemData, err := s.Character.RemoveItem(itemSlot)
	if err != nil {
		return nil, err
	}
	resp.Concat(itemData)

	if special != nil {
		itemData, err := s.Character.RemoveItem(specialSlot)
		if err != nil {
			return nil, err
		}
		resp.Concat(itemData)
	}

	return resp, err
}

func (h *ExtractionHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	itemSlot := int16(utils.BytesToInt(data[6:8], true))
	if itemSlot == 0 {
		return nil, nil
	} else if slots[itemSlot].ItemID == 0 {
		return nil, nil
	}

	var special *database.InventorySlot
	specialSlot := int16(utils.BytesToInt(data[8:10], true))
	if specialSlot == 0 {
		special = nil
	} else {
		special = slots[specialSlot]
	}

	resp := utils.Packet{}
	extractionData, _, err := s.Character.Extraction(slots[itemSlot], special, itemSlot)
	if err != nil {
		return nil, err
	}

	resp.Concat(extractionData)

	if special != nil {
		itemData, err := s.Character.RemoveItem(specialSlot)
		if err != nil {
			return nil, err
		}
		resp.Concat(itemData)
	}

	return resp, nil
}
