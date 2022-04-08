package npc

import (
	"hero-emulator/database"
	"hero-emulator/utils"
)

type (
	CreateSocketHandler  struct{}
	UpgradeSocketHandler struct{}
	CoProductionHandler  struct{}
)

var (
	CREATED_SOCKET = utils.Packet{}
)

func (h *CreateSocketHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {
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

	return s.Character.CreateSocket(slots[itemSlot], special, itemSlot, specialSlot)
}

func (h *UpgradeSocketHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

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

	socketSlot := int16(utils.BytesToInt(data[8:10], true))
	if socketSlot == 0 {
		resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x54, 0x17, 0x0E, 0xCF, 0x55, 0xAA}
		return resp, nil
	} else if slots[socketSlot].ItemID == 0 {
		resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x54, 0x17, 0x0E, 0xCF, 0x55, 0xAA}
		return resp, nil
	}

	var special *database.InventorySlot
	specialSlot := int16(utils.BytesToInt(data[10:12], true))
	if specialSlot == 0 {
		special = nil
	} else {
		special = slots[specialSlot]
	}

	var edit *database.InventorySlot
	editSlot := int16(utils.BytesToInt(data[12:14], true))
	if editSlot == 0 {
		edit = nil
	} else {
		edit = slots[editSlot]
	}

	index, locks := 14, make([]bool, 5)
	if edit != nil {
		for i := 0; i < 5; i++ {
			locks[i] = data[index] == 1
			index++
		}
	}

	return s.Character.UpgradeSocket(slots[itemSlot], slots[socketSlot], special, edit, itemSlot, socketSlot, specialSlot, editSlot, locks)
}

func (h *CoProductionHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {
	resp := utils.Packet{}
	craftID := int(utils.BytesToInt(data[6:10], true))
	bFinished := int(data[10])
	canCraft := true
	production := database.CraftItems[int(craftID)]
	prodMaterials, err := production.GetMaterials()
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(prodMaterials); i++ {
		matItemId := int64(prodMaterials[i].ID)
		matCount := uint(prodMaterials[i].Count)
		slotID, _, _ := s.Character.FindItemInInventory(nil, matItemId)
		slots, err := s.Character.InventorySlots()
		if err != nil {
			return nil, err
		}
		item := slots[slotID]
		if item.Quantity < matCount {
			canCraft = false
		}
	}
	if canCraft {
		prod, _ := s.Character.CoProduction(craftID, bFinished)
		resp.Concat(prod)
	}
	return resp, nil
}
