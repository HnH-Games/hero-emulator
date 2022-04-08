package npc

import (
	"encoding/binary"

	"hero-emulator/database"
	"hero-emulator/utils"
)

type BuyItemHandler struct {
}

type SellItemHandler struct {
}

var ()

func (h *BuyItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	c := s.Character
	if c == nil {
		return nil, nil
	}
	item_buyed := utils.Packet{0xaa, 0x55, 0x3c, 0x00, 0x58, 0x01, 0x0a, 0x00, 0x15, 0x64, 0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x20, 0x1c, 0x00, 0x00, 0x55, 0xaa}
	itemindex := 8
	itemID := utils.BytesToInt(data[6:10], true)
	quantity := utils.BytesToInt(data[10:12], true)
	slotID := int16(utils.BytesToInt(data[16:18], true))

	npcID := int(utils.BytesToInt(data[18:22], true))
	shopID, ok := shops[npcID]
	if !ok {
		shopID = 25
	}

	shop, ok := database.Shops[shopID]
	if !ok {
		return nil, nil
	}

	canPurchase := shop.IsPurchasable(int(itemID))
	if !canPurchase {
		return nil, nil
	}

	slots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	info := database.Items[itemID]
	if info.SpecialItem != 0 {
		canChange := true
		reqCoinCount := uint(info.BuyPrice) * uint(quantity)
		slotIDitem, _, _ := s.Character.FindItemInInventory(nil, info.SpecialItem)
		slots, err := s.Character.InventorySlots()
		if err != nil {
			return nil, err
		}
		items := slots[slotIDitem]
		if items.Quantity < reqCoinCount {
			canChange = false
			return nil, nil
		}
		if canChange {
			if info.Timer != 0 {
				quantity = int64(info.Timer)
			}
			item := &database.InventorySlot{ItemID: itemID, Quantity: uint(quantity)}
			if info.GetType() == database.PET_TYPE {
				petInfo := database.Pets[item.ItemID]
				expInfo := database.PetExps[petInfo.Level-1]

				item.Pet = &database.PetSlot{
					Fullness: 100, Loyalty: 100,
					Exp:   uint64(expInfo.ReqExpEvo1),
					HP:    petInfo.BaseHP,
					Level: byte(petInfo.Level),
					Name:  petInfo.Name,
					CHI:   petInfo.BaseChi}
			}
			resp, _, err := c.AddItem(item, slotID, false)
			if err != nil {
				return nil, err
			} else if resp == nil {
				return nil, nil
			}
			itemData := c.DecrementItem(slotIDitem, reqCoinCount)
			resp.Concat(*itemData)
			resp.Concat(item_buyed)
			return *resp, nil
		}
		return nil, nil
	}
	cost := uint64(info.BuyPrice) * uint64(quantity)
	if slots[slotID].ItemID == 0 && cost <= c.Gold && quantity > 0 && info.SpecialItem == 0 { // slot is empty, player can afford and quantity is positive
		c.LootGold(-cost)
		if info.Timer != 0 {
			quantity = int64(info.Timer)
		}
		item := &database.InventorySlot{ItemID: itemID, Quantity: uint(quantity)}

		if info.GetType() == database.PET_TYPE {
			petInfo := database.Pets[item.ItemID]
			expInfo := database.PetExps[petInfo.Level-1]

			item.Pet = &database.PetSlot{
				Fullness: 100, Loyalty: 100,
				Exp:   uint64(expInfo.ReqExpEvo1),
				HP:    petInfo.BaseHP,
				Level: byte(petInfo.Level),
				Name:  petInfo.Name,
				CHI:   petInfo.BaseChi}
		}

		resp, _, err := c.AddItem(item, slotID, false)
		if err != nil {
			return nil, err
		} else if resp == nil {
			return nil, nil
		}
		item_buyed.Insert(utils.IntToBytes(uint64(itemID), 4, true), itemindex)
		itemindex += 8
		item_buyed.Insert(utils.IntToBytes(uint64(slotID), 2, true), itemindex)
		itemindex += 2
		item_buyed.Insert([]byte{0x00, 0x00, 0x00, 0x00}, itemindex) //PET HP
		item_buyed.Insert([]byte{0x00, 0x00, 0x00, 0x00}, itemindex) //PET EXP
		itemindex += 8
		item_buyed.Insert([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, itemindex)
		itemindex += 26
		item_buyed.Insert(utils.IntToBytes(uint64(c.Gold), 4, true), itemindex)
		item_buyed.SetLength(int16(binary.Size(item_buyed) - 6))
		resp.Concat(item_buyed)
		return *resp, nil
	}

	return nil, nil
}

func (h *SellItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	c := s.Character
	if c == nil {
		return nil, nil
	}

	slots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	c.Looting.Lock()
	defer c.Looting.Unlock()

	itemID := utils.BytesToInt(data[6:10], true)
	quantity := int(utils.BytesToInt(data[10:12], true))
	slotID := int16(utils.BytesToInt(data[12:14], true))

	item := database.Items[itemID]
	slot := slots[slotID]

	if !item.Tradable {
		return nil, nil
	}

	multiplier := 0
	if slot.ItemID == itemID && quantity > 0 && uint(quantity) <= slot.Quantity {
		upgs := slot.GetUpgrades()
		for i := uint8(0); i < slot.Plus; i++ {
			upg := upgs[i]
			if code, ok := database.HaxCodes[int(upg)]; ok {
				multiplier += code.SaleMultiplier
			}
		}

		multiplier /= 1000
		if multiplier == 0 {
			multiplier = 1
		}

		unitPrice := uint64(item.SellPrice) * uint64(multiplier)
		if slot.Plus > 0 {
			unitPrice *= uint64(slot.Plus)
		}

		return c.SellItem(int(itemID), int(slotID), int(quantity), unitPrice)
	}

	return nil, nil
}
