package player

import (
	"hero-emulator/database"
	"hero-emulator/utils"
)

type (
	UpgradeSkillHandler          struct{}
	UpgradePassiveSkillHandler   struct{}
	DowngradeSkillHandler        struct{}
	DowngradePassiveSkillHandler struct{}
	RemoveSkillHandler           struct{}
	DivineUpgradeSkillHandler    struct{}
	RemovePassiveSkillHandler    struct{}
)

func (h *DivineUpgradeSkillHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {
	skillIndex := data[6]
	slotIndex := data[7]
	bookID := utils.BytesToInt(data[8:12], true)
	return s.Character.DivineUpgradeSkills(int(skillIndex), int(slotIndex), bookID)
}

func (h *UpgradeSkillHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slotIndex := data[6]
	skillIndex := data[7]

	return s.Character.UpgradeSkill(slotIndex, skillIndex)
}

func (h *DowngradeSkillHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slotIndex := data[6]
	skillIndex := data[7]

	return s.Character.DowngradeSkill(slotIndex, skillIndex)
}

func (h *RemoveSkillHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slotIndex := data[6]
	bookID := utils.BytesToInt(data[7:11], true)

	slotID, slot, err := s.Character.FindItemInInventory(nil, 15400001)
	if err != nil {
		return nil, err
	} else if slot == nil {
		return nil, nil
	}

	skillData, err := s.Character.RemoveSkill(slotIndex, bookID)
	if err != nil {
		return nil, err
	}

	resp := *s.Character.DecrementItem(slotID, 1)
	resp.Concat(skillData)
	return resp, nil
}

func (h *UpgradePassiveSkillHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slotIndex := data[6]
	skillIndex := byte(0)
	if slotIndex == 0 {
		skillIndex = 5
	} else if slotIndex == 1 {
		skillIndex = 6
	} else if slotIndex == 8 {
		skillIndex = 7
	}

	return s.Character.UpgradePassiveSkill(slotIndex, skillIndex)
}

func (h *DowngradePassiveSkillHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slotIndex := data[6]
	skillIndex := byte(0)
	if slotIndex == 0 {
		skillIndex = 5
	} else if slotIndex == 1 {
		skillIndex = 6
	} else if slotIndex == 8 {
		skillIndex = 7
	}

	return s.Character.DowngradePassiveSkill(slotIndex, skillIndex)
}

func (h *RemovePassiveSkillHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	bookID := utils.BytesToInt(data[6:10], true)
	slotIndex := data[10]

	slotID, slot, err := s.Character.FindItemInInventory(nil, 15400001)
	if err != nil {
		return nil, err
	} else if slot == nil {
		return nil, nil
	}

	skillIndex := byte(0)
	if slotIndex == 0 {
		skillIndex = 5
	} else if slotIndex == 1 {
		skillIndex = 6
	} else if slotIndex == 8 {
		skillIndex = 7
	}

	skillData, err := s.Character.RemovePassiveSkill(slotIndex, skillIndex, bookID)
	if err != nil {
		return nil, err
	}

	resp := *s.Character.DecrementItem(slotID, 1)
	resp.Concat(skillData)

	return resp, nil
}
