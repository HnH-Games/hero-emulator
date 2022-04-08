package npc

import (
	"hero-emulator/database"
	"hero-emulator/utils"
)

type (
	AppearanceHandler struct{}
)

func (h *AppearanceHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {
	resp := utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x54, 0x22, 0x0a, 0x00, 0x00, 0x55, 0xAA}
	itemSlot := int(utils.BytesToInt(data[6:8], true))
	newitemSlot := int(utils.BytesToInt(data[8:10], true))
	matitemSlot := int(utils.BytesToInt(data[10:12], true))
	inventory, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}
	weapon := inventory[itemSlot]
	newWeapon := inventory[newitemSlot]
	weapon.Appearance = newWeapon.ItemID
	//(item.Type >= 70 && item.Type <= 71) || (item.Type >= 99 && item.Type <= 108) WEAPONS TYPE
	//item := database.Items[newWeapon.ItemID]
	//if (item.Type >= 70 && item.Type <= 71) || (item.Type >= 99 && item.Type <= 108) {
	s.Character.Socket.Write(resp)
	matData := s.Character.DecrementItem(int16(matitemSlot), 1)
	s.Character.Socket.Write(*matData)
	weapData := s.Character.DecrementItem(int16(newitemSlot), 1)
	s.Character.Socket.Write(*weapData)
	return resp, nil
	//}

	return nil, nil
}
