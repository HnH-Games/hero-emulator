package player

import (
	"fmt"
	"log"
	"math"
	"sort"
	"time"

	"hero-emulator/database"
	"hero-emulator/messaging"
	"hero-emulator/nats"
	"hero-emulator/server"
	"hero-emulator/utils"

	"github.com/thoas/go-funk"
)

type (
	GetGoldHandler struct {
		gold uint64
	}

	GetInventoryHandler             struct{}
	ReplaceItemHandler              struct{}
	SwitchWeaponHandler             struct{}
	SwapItemsHandler                struct{}
	RemoveItemHandler               struct{}
	DestroyItemHandler              struct{}
	CombineItemsHandler             struct{}
	ArrangeInventoryHandler         struct{}
	ArrangeFunctionHandler          struct{}
	ArrangeBankHandler              struct{}
	DepositHandler                  struct{}
	WithdrawHandler                 struct{}
	OpenHTMenuHandler               struct{}
	CloseHTMenuHandler              struct{}
	BuyHTItemHandler                struct{}
	ReplaceHTItemHandler            struct{}
	DiscriminateItemHandler         struct{}
	InspectItemHandler              struct{}
	DressUpHandler                  struct{}
	SplitItemHandler                struct{}
	HolyWaterUpgradeHandler         struct{}
	UseConsumableHandler            struct{}
	OpenBoxHandler                  struct{}
	OpenBoxHandler2                 struct{}
	ActivateTimeLimitedItemHandler  struct{}
	ActivateTimeLimitedItemHandler2 struct{}
	ToggleMountPetHandler           struct{}
	TogglePetHandler                struct{}
	PetCombatModeHandler            struct{}
)

var (
	GET_GOLD       = utils.Packet{0xAA, 0x55, 0x0A, 0x00, 0x57, 0x0B, 0x55, 0xAA}
	ITEMS_COMBINED = utils.Packet{0xAA, 0x55, 0x10, 0x00, 0x59, 0x06, 0x0A, 0x00, 0x00, 0x00, 0x55, 0xAA}
	ARRANGE_ITEM   = utils.Packet{0xAA, 0x55, 0x32, 0x00, 0x78, 0x02, 0x00, 0xA2, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}

	ARRANGE_BANK_ITEM = utils.Packet{0xAA, 0x55, 0x2F, 0x00, 0x80, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}

	CLOSE_HT_MENU = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x64, 0x02, 0x0A, 0x00, 0x55, 0xAA}
	OPEN_HT_MENU  = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x64, 0x01, 0x0A, 0x00, 0x55, 0xAA}
	GET_CASH      = utils.Packet{0xAA, 0x55, 0x0C, 0x00, 0x64, 0x03, 0x0A, 0x00, 0x55, 0xAA}
	BUY_HT_ITEM   = utils.Packet{0xAA, 0x55, 0x38, 0x00, 0x64, 0x04, 0x0A, 0x00, 0x07, 0xA2, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	REPLACE_HT_ITEM = utils.Packet{0xAA, 0x55, 0x0A, 0x00, 0x59, 0x40, 0x0A, 0x00, 0x55, 0xAA}
	HT_VISIBILITY   = utils.Packet{0xAA, 0x55, 0x06, 0x00, 0x59, 0x11, 0x0A, 0x00, 0x01, 0x00, 0x55, 0xAA}
	PET_COMBAT      = utils.Packet{0xAA, 0x55, 0x06, 0x00, 0x51, 0x05, 0x0a, 0x00, 0x00, 0x55, 0xAA}
	htShopQuantites = map[int64]uint{17100004: 40, 17100005: 40, 15900001: 50}
)

func (ggh *GetGoldHandler) Handle(s *database.Socket) ([]byte, error) {

	resp := GET_GOLD
	resp.Insert(utils.IntToBytes(uint64(s.Character.Gold), 8, true), 6)
	return resp, nil
}

func (gih *GetInventoryHandler) Handle(s *database.Socket) ([]byte, error) {

	inventory, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	resp := utils.Packet{}
	for i := 0; i < len(inventory); i++ {

		slot := inventory[i]
		resp.Concat(slot.GetData(int16(i)))
	}

	if inventory[0x0A].ItemID > 0 { // pet
		resp.Concat(database.SHOW_PET_BUTTON)
	}

	if s.Character.DoesInventoryExpanded() {
		resp.Concat(database.BAG_EXPANDED)
	}

	return resp, nil
}

func (h *ReplaceItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	itemID := int(utils.BytesToInt(data[6:10], true))
	where := int16(utils.BytesToInt(data[10:12], true))
	to := int16(utils.BytesToInt(data[12:14], true))

	resp, err := s.Character.ReplaceItem(itemID, where, to)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *SwitchWeaponHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slotID := data[6]
	s.Character.WeaponSlot = int(slotID)

	itemsData, err := s.Character.ShowItems()
	if err != nil {
		return nil, err
	}

	p := nats.CastPacket{CastNear: true, CharacterID: s.Character.ID, Type: nats.SHOW_ITEMS, Data: itemsData}
	if err = p.Cast(); err != nil {
		return nil, err
	}

	gsh := &GetStatsHandler{}
	statData, err := gsh.Handle(s)
	if err != nil {
		return nil, err
	}

	resp := utils.Packet{}
	resp.Concat(itemsData)
	resp.Concat(statData)

	return resp, nil
}

func (h *SwapItemsHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	index := 11
	where := int16(utils.BytesToInt(data[index:index+2], true))

	index += 2
	to := int16(utils.BytesToInt(data[index:index+2], true))

	resp, err := s.Character.SwapItems(where, to)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *RemoveItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	index := 11
	slotID := int16(utils.BytesToInt(data[index:index+2], true)) // slot

	resp, err := s.Character.RemoveItem(slotID)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *DestroyItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	index := 10
	slotID := int16(utils.BytesToInt(data[index:index+2], true)) // slot

	resp, err := s.Character.RemoveItem(slotID)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (h *CombineItemsHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	c := s.Character
	if c == nil {
		return nil, nil
	}

	where := int16(utils.BytesToInt(data[6:8], true))
	to := int16(utils.BytesToInt(data[8:10], true))
	itemID, qty, _ := c.CombineItems(where, to)

	resp := ITEMS_COMBINED
	resp.Insert(utils.IntToBytes(uint64(itemID), 4, true), 8) // item id
	resp.Insert(utils.IntToBytes(uint64(where), 2, true), 12) // where slot
	resp.Insert(utils.IntToBytes(uint64(to), 2, true), 16)    // to slot
	resp.Insert(utils.IntToBytes(uint64(qty), 2, true), 18)   // item quantity

	return resp, nil
}
func WaitFunctionTimer(s *database.Socket) {
	time.AfterFunc(time.Second*3, func() {
		s.Character.PacketSended = false
		resp, _ := ArrangeInventory(s)
		s.Character.Socket.Write(resp)
	})
}
func (h *ArrangeFunctionHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {
	if !s.Character.PacketSended {
		WaitFunctionTimer(s)
		s.Character.PacketSended = true
		resp, err := ArrangeInventory(s)
		return resp, err
	} else {
		EMPTY_Arrange := utils.Packet{0xaa, 0x55, 0x32, 0x00, 0x78, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x55, 0xaa}
		resp := EMPTY_Arrange
		resp.Concat(messaging.InfoMessage(fmt.Sprintf("You need to wait 3 seconds.")))
		return resp, nil
	}
	return nil, nil
}
func ArrangeInventory(s *database.Socket) ([]byte, error) {
	if s.Character.TradeID != "" || database.FindSale(s.Character.PseudoID) != nil {
		return nil, nil
	}
	slots, err := s.Character.InventorySlots()
	if err != nil {
		fmt.Println("Error_:", err)
		return nil, err
	}

	newSlots := make([]database.InventorySlot, 56)
	for i := 0; i < 56; i++ {
		slotID := i + 0x0B
		newSlots[i] = *slots[slotID]
		if newSlots[i].ItemID == 0 {
			newSlots[i].ItemID = math.MaxInt64
		}
		newSlots[i].RFU = int64(slotID)
	}

	sort.Slice(newSlots, func(i, j int) bool {
		if newSlots[i].ItemID < newSlots[j].ItemID {
			return true
		}
		return false
	})

	resp := utils.Packet{}
	for i := 0; i < 56; i++ { // first page
		slot := &newSlots[i]
		r, r2 := ARRANGE_ITEM, utils.Packet{}

		if slot.ItemID == math.MaxInt64 {
			slot.ItemID = 0
		}

		slot.SlotID = int16(i + 0x0B)
		r.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), 6) // item id

		info := database.Items[slot.ItemID]
		if info != nil && slot.Activated { // using state
			if info.TimerType == 1 {
				r[10] = 3
			} else if info.TimerType == 3 {
				r[10] = 5
				r2 = database.GREEN_ITEM_COUNT
				r2.Insert(utils.IntToBytes(uint64(slot.SlotID), 2, true), 8)    // slot id
				r2.Insert(utils.IntToBytes(uint64(slot.Quantity), 4, true), 10) // item quantity
			}
		} else {
			r[10] = 0
		}

		if slot.ItemID == 0 {
			r[11] = 0
		} else if slot.Plus > 0 {
			r[11] = 0xA2
		}

		r.Insert(utils.IntToBytes(uint64(slot.Quantity), 2, true), 12) // item quantity
		r.Insert(utils.IntToBytes(uint64(slot.SlotID), 2, true), 14)   // slot id
		r.Insert(slot.GetUpgrades(), 16)                               // slot upgrades
		r[31] = byte(slot.SocketCount)                                 // socket count
		r.Insert(slot.GetSockets(), 32)                                // slot sockets
		c := 32 + 15
		if slot.ItemType != 0 {
			r.Overwrite(utils.IntToBytes(uint64(slot.ItemType), 1, true), c-6)
			if slot.ItemType == 2 {
				r.Overwrite(utils.IntToBytes(uint64(slot.JudgementStat), 4, true), c-5)
			}
		}
		if i == 55 {
			r[50] = 1
		}

		r.Insert(utils.IntToBytes(uint64(slot.RFU.(int64)), 2, true), 52) // pre slot id

		if info != nil && info.GetType() == database.PET_TYPE {
			r2.Concat(slot.GetData(int16(slot.SlotID)))
		}

		go slot.Update()
		resp.Concat(r)
		resp.Concat(r2)
	}

	for i := 0; i < 56; i++ {
		slotID := i + 0x0B
		newSlots[i].RFU = 0
		*slots[slotID] = newSlots[i]
	}

	newSlots = make([]database.InventorySlot, 56)
	for i := 0; i < 56; i++ {
		slotID := i + 0x0155
		newSlots[i] = *slots[slotID]
		if newSlots[i].ItemID == 0 {
			newSlots[i].ItemID = math.MaxInt64
		}
		newSlots[i].RFU = int64(slotID)
	}

	sort.Slice(newSlots, func(i, j int) bool {
		if newSlots[i].ItemID < newSlots[j].ItemID {
			return true
		}
		return false
	})

	for i := 0; i < 56; i++ { // second page
		slot := &newSlots[i]
		r, r2 := ARRANGE_ITEM, utils.Packet{}

		if slot.ItemID == math.MaxInt64 {
			slot.ItemID = 0
		}

		slot.SlotID = int16(i + 0x0155)
		r.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), 6) // item id

		info := database.Items[slot.ItemID]
		if info != nil && slot.Activated { // using state
			if info.TimerType == 1 {
				r[10] = 3
			} else if info.TimerType == 3 {
				r[10] = 5
				r2 = database.GREEN_ITEM_COUNT
				r2.Insert(utils.IntToBytes(uint64(slot.SlotID), 2, true), 8)    // slot id
				r2.Insert(utils.IntToBytes(uint64(slot.Quantity), 4, true), 10) // item quantity
			}
		} else {
			r[10] = 0
		}

		if slot.ItemID == 0 {
			r[11] = 0
		} else if slot.Plus > 0 {
			r[11] = 0xA2
		}

		r.Insert(utils.IntToBytes(uint64(slot.Quantity), 2, true), 12) // item quantity
		r.Insert(utils.IntToBytes(uint64(slot.SlotID), 2, true), 14)   // slot id
		r.Insert(slot.GetUpgrades(), 16)                               // slot upgrades
		r[31] = byte(slot.SocketCount)                                 // socket count
		r.Insert(slot.GetSockets(), 32)                                // slot sockets
		c := 32 + 15
		if slot.ItemType != 0 {
			resp.Overwrite(utils.IntToBytes(uint64(slot.ItemType), 1, true), c-6)
			if slot.ItemType == 2 {
				resp.Overwrite(utils.IntToBytes(uint64(slot.JudgementStat), 4, true), c-5)
			}
		}
		if i == 55 {
			r[50] = 1
		}

		r[51] = 1
		r.Insert(utils.IntToBytes(uint64(slot.RFU.(int64)), 2, true), 52) // pre slot id

		if info != nil && info.GetType() == database.PET_TYPE {
			r2.Concat(slot.GetData(int16(slot.SlotID)))
		}

		go slot.Update()
		resp.Concat(r)
		resp.Concat(r2)
	}

	for i := 0; i < 56; i++ {
		slotID := i + 0x0155
		newSlots[i].RFU = 0
		*slots[slotID] = newSlots[i]
	}
	return resp, nil
}
func (h *ArrangeInventoryHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if s.Character.TradeID != "" || database.FindSale(s.Character.PseudoID) != nil {
		return nil, nil
	}

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	newSlots := make([]database.InventorySlot, 56)
	for i := 0; i < 56; i++ {
		slotID := i + 0x0B
		newSlots[i] = *slots[slotID]
		if newSlots[i].ItemID == 0 {
			newSlots[i].ItemID = math.MaxInt64
		}
		newSlots[i].RFU = int64(slotID)
	}

	sort.Slice(newSlots, func(i, j int) bool {
		if newSlots[i].ItemID < newSlots[j].ItemID {
			return true
		}
		return false
	})

	resp := utils.Packet{}
	for i := 0; i < 56; i++ { // first page
		slot := &newSlots[i]
		r, r2 := ARRANGE_ITEM, utils.Packet{}

		if slot.ItemID == math.MaxInt64 {
			slot.ItemID = 0
		}

		slot.SlotID = int16(i + 0x0B)
		r.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), 6) // item id

		info := database.Items[slot.ItemID]
		if info != nil && slot.Activated { // using state
			if info.TimerType == 1 {
				r[10] = 3
			} else if info.TimerType == 3 {
				r[10] = 5
				r2 = database.GREEN_ITEM_COUNT
				r2.Insert(utils.IntToBytes(uint64(slot.SlotID), 2, true), 8)    // slot id
				r2.Insert(utils.IntToBytes(uint64(slot.Quantity), 4, true), 10) // item quantity
			}
		} else {
			r[10] = 0
		}

		if slot.ItemID == 0 {
			r[11] = 0
		} else if slot.Plus > 0 {
			r[11] = 0xA2
		}

		r.Insert(utils.IntToBytes(uint64(slot.Quantity), 2, true), 12) // item quantity
		r.Insert(utils.IntToBytes(uint64(slot.SlotID), 2, true), 14)   // slot id
		r.Insert(slot.GetUpgrades(), 16)                               // slot upgrades
		r[31] = byte(slot.SocketCount)                                 // socket count
		r.Insert(slot.GetSockets(), 32)                                // slot sockets
		c := 32 + 15
		if slot.ItemType != 0 {
			resp.Overwrite(utils.IntToBytes(uint64(slot.ItemType), 1, true), c-6)
			if slot.ItemType == 2 {
				resp.Overwrite(utils.IntToBytes(uint64(slot.JudgementStat), 4, true), c-5)
			}
		}
		if i == 55 {
			r[50] = 1
		}

		r.Insert(utils.IntToBytes(uint64(slot.RFU.(int64)), 2, true), 52) // pre slot id

		if info != nil && info.GetType() == database.PET_TYPE {
			r2.Concat(slot.GetData(int16(slot.SlotID)))
		}

		go slot.Update()
		resp.Concat(r)
		resp.Concat(r2)
	}

	for i := 0; i < 56; i++ {
		slotID := i + 0x0B
		newSlots[i].RFU = 0
		*slots[slotID] = newSlots[i]
	}

	newSlots = make([]database.InventorySlot, 56)
	for i := 0; i < 56; i++ {
		slotID := i + 0x0155
		newSlots[i] = *slots[slotID]
		if newSlots[i].ItemID == 0 {
			newSlots[i].ItemID = math.MaxInt64
		}
		newSlots[i].RFU = int64(slotID)
	}

	sort.Slice(newSlots, func(i, j int) bool {
		if newSlots[i].ItemID < newSlots[j].ItemID {
			return true
		}
		return false
	})

	for i := 0; i < 56; i++ { // second page
		slot := &newSlots[i]
		r, r2 := ARRANGE_ITEM, utils.Packet{}

		if slot.ItemID == math.MaxInt64 {
			slot.ItemID = 0
		}

		slot.SlotID = int16(i + 0x0155)
		r.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), 6) // item id

		info := database.Items[slot.ItemID]
		if info != nil && slot.Activated { // using state
			if info.TimerType == 1 {
				r[10] = 3
			} else if info.TimerType == 3 {
				r[10] = 5
				r2 = database.GREEN_ITEM_COUNT
				r2.Insert(utils.IntToBytes(uint64(slot.SlotID), 2, true), 8)    // slot id
				r2.Insert(utils.IntToBytes(uint64(slot.Quantity), 4, true), 10) // item quantity
			}
		} else {
			r[10] = 0
		}

		if slot.ItemID == 0 {
			r[11] = 0
		} else if slot.Plus > 0 {
			r[11] = 0xA2
		}

		r.Insert(utils.IntToBytes(uint64(slot.Quantity), 2, true), 12) // item quantity
		r.Insert(utils.IntToBytes(uint64(slot.SlotID), 2, true), 14)   // slot id
		r.Insert(slot.GetUpgrades(), 16)                               // slot upgrades
		r[31] = byte(slot.SocketCount)                                 // socket count
		r.Insert(slot.GetSockets(), 32)                                // slot sockets
		c := 32 + 15
		if slot.ItemType != 0 {
			resp.Overwrite(utils.IntToBytes(uint64(slot.ItemType), 1, true), c-6)
			if slot.ItemType == 2 {
				resp.Overwrite(utils.IntToBytes(uint64(slot.JudgementStat), 4, true), c-5)
			}
		}
		if i == 55 {
			r[50] = 1
		}

		r[51] = 1
		r.Insert(utils.IntToBytes(uint64(slot.RFU.(int64)), 2, true), 52) // pre slot id

		if info != nil && info.GetType() == database.PET_TYPE {
			r2.Concat(slot.GetData(int16(slot.SlotID)))
		}

		go slot.Update()
		resp.Concat(r)
		resp.Concat(r2)
	}

	for i := 0; i < 56; i++ {
		slotID := i + 0x0155
		newSlots[i].RFU = 0
		*slots[slotID] = newSlots[i]
	}

	return resp, nil
}

func (h *ArrangeBankHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if s.Character.TradeID != "" {
		return nil, nil
	}

	user := s.User
	if user == nil {
		return nil, nil
	}

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	slots = slots[0x43:0x133]
	resp := utils.Packet{}
	for page := 0; page < 4; page++ {

		newSlots := make([]database.InventorySlot, 60)
		for i := 0; i < 60; i++ {
			index := page*60 + i
			newSlots[i] = *slots[index]
			if newSlots[i].ItemID == 0 {
				newSlots[i].ItemID = math.MaxInt64
			}
			newSlots[i].RFU = int64(index + 0x43)
		}

		sort.Slice(newSlots, func(i, j int) bool {
			if newSlots[i].ItemID < newSlots[j].ItemID {
				return true
			}
			return false
		})

		for i := 0; i < 60; i++ {
			slot := &newSlots[i]
			r, r2 := ARRANGE_BANK_ITEM, utils.Packet{}

			if slot.ItemID == math.MaxInt64 {
				slot.ItemID = 0
			}

			slot.SlotID = int16(page*60 + i + 0x43)
			r.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), 6) // item id

			info := database.Items[slot.ItemID]
			if info != nil && slot.Activated { // using state
				if info.TimerType == 1 {
					r[10] = 3
				} else if info.TimerType == 3 {
					r[10] = 5
					r2 = database.GREEN_ITEM_COUNT
					r2.Insert(utils.IntToBytes(uint64(slot.SlotID), 2, true), 8)    // slot id
					r2.Insert(utils.IntToBytes(uint64(slot.Quantity), 4, true), 10) // item quantity
				}
			} else {
				r[10] = 0
			}

			if slot.ItemID == 0 {
				r[11] = 0
			} else if slot.Plus > 0 || slot.SocketCount > 0 {
				r[11] = 0xA2
			}

			r.Insert(utils.IntToBytes(uint64(slot.Quantity), 2, true), 12) // item quantity
			r.Insert(utils.IntToBytes(uint64(slot.SlotID), 2, true), 14)   // slot id

			if slot.Plus > 0 || slot.SocketCount > 0 {
				r.Insert(slot.GetUpgrades(), 16) // slot upgrades
				r[31] = byte(slot.SocketCount)   // socket count
				r.Insert(slot.GetSockets(), 32)  // slot sockets
				c := 32 + 15
				if slot.ItemType != 0 {
					resp.Overwrite(utils.IntToBytes(uint64(slot.ItemType), 1, true), c-6)
					if slot.ItemType == 2 {
						resp.Overwrite(utils.IntToBytes(uint64(slot.JudgementStat), 4, true), c-5)
					}
				}
				r.SetLength(0x4D)
			}

			if i == 60 {
				r[47] = 1
			}
			r[48] = byte(page)

			r.Insert(utils.IntToBytes(uint64(slot.RFU.(int64)), 2, true), 49) // pre slot id

			if info != nil && info.GetType() == database.PET_TYPE {
				r2.Concat(slot.GetData(int16(slot.SlotID)))
			}

			go slot.Update()
			resp.Concat(r)
			resp.Concat(r2)
		}

		for i := 0; i < 60; i++ {
			slotID := page*60 + i
			newSlots[i].RFU = 0
			*slots[slotID] = newSlots[i]
		}
	}

	go user.Update()
	return resp, nil
}

func (h *DepositHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	u := s.User
	if u == nil {
		return nil, nil
	}

	c := s.Character
	if c == nil {
		return nil, nil
	}

	gold := uint64(utils.BytesToInt(data[6:14], true))
	if c.Gold >= gold {
		c.LootGold(-gold)
		u.BankGold += gold

		go u.Update()
		return c.GetGold(), nil
	}

	return nil, nil
}

func (h *WithdrawHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	u := s.User
	if u == nil {
		return nil, nil
	}

	c := s.Character
	if c == nil {
		return nil, nil
	}

	gold := uint64(utils.BytesToInt(data[6:14], true))
	if u.BankGold >= gold {
		c.LootGold(gold)
		u.BankGold -= gold

		go u.Update()
		return c.GetGold(), nil
	}
	return nil, nil
}

func (h *OpenHTMenuHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	u := s.User
	if u == nil {
		return nil, nil
	}

	resp := OPEN_HT_MENU
	r := GET_CASH
	r.Insert(utils.IntToBytes(u.NCash, 8, true), 8) // user nCash

	resp.Concat(r)
	return resp, nil
}

func (h *CloseHTMenuHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	return CLOSE_HT_MENU, nil
}

func (h *BuyHTItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	itemID := int(utils.BytesToInt(data[6:10], true))
	slotID := utils.BytesToInt(data[12:14], true)

	if item, ok := database.HTItems[itemID]; ok && item.IsActive && s.User.NCash >= uint64(item.Cash) {
		s.User.NCash -= uint64(item.Cash)

		info := database.Items[int64(itemID)]
		quantity := uint(1)
		if info.Timer > 0 && info.TimerType > 0 {
			quantity = uint(info.Timer)
		} else if qty, ok := htShopQuantites[info.ID]; ok {
			quantity = qty
		}

		item := &database.InventorySlot{ItemID: int64(itemID), Quantity: quantity}
		if info.GetType() == database.PET_TYPE {
			petInfo := database.Pets[int64(itemID)]
			petExpInfo := database.PetExps[int16(petInfo.Level)]

			targetExps := []int{petExpInfo.ReqExpEvo1, petExpInfo.ReqExpEvo2, petExpInfo.ReqExpEvo3, petExpInfo.ReqExpHt, petExpInfo.ReqExpDivEvo1, petExpInfo.ReqExpDivEvo2, petExpInfo.ReqExpDivEvo3}
			item.Pet = &database.PetSlot{
				Fullness: 100, Loyalty: 100, PseudoID: 0,
				Exp:   uint64(targetExps[petInfo.Evolution-1]),
				HP:    petInfo.BaseHP,
				Level: byte(petInfo.Level),
				Name:  petInfo.Name,
				CHI:   petInfo.BaseChi,
			}
		}

		r, _, err := s.Character.AddItem(item, int16(slotID), false)
		if err != nil {
			return nil, err
		} else if r == nil {
			return nil, nil
		}

		resp := BUY_HT_ITEM
		resp.Insert(utils.IntToBytes(uint64(itemID), 4, true), 8)    // item id
		resp.Insert(utils.IntToBytes(uint64(quantity), 2, true), 14) // item quantity
		resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 16)   // slot id
		resp.Insert(utils.IntToBytes(s.User.NCash, 8, true), 52)     // user nCash

		resp.Concat(*r)

		go s.User.Update()
		return resp, nil
	}

	return nil, nil
}

func (h *DiscriminateItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		resp := messaging.SystemMessage(messaging.INSUFFICIENT_GOLD)
		return resp, nil
	}

	slotID := int(utils.BytesToInt(data[6:8], true)) //max: 754700
	item := slots[slotID]
	itemstat := database.Items[item.ItemID]
	discprice := itemstat.MinLevel * 5000
	if s.Character.Gold < uint64(discprice) {
		return nil, nil
	}
	index := 0
	seed := int64(utils.RandInt(0, 754700))
	for _, prob := range database.ItemJudgements {
		if float64(prob.Probabilities) >= float64(seed) {
			index = prob.ID
			break
		}
	}
	if index != 0 {
		s.Character.LootGold(-uint64(discprice))
		slot := slots[slotID]
		slot.ItemType = 2
		slot.JudgementStat = int64(index)
		err = slot.Update()
		database.InventoryItems.Add(slot.ID, slot)
		resp := utils.Packet{}
		resp.Concat(slot.GetData(int16(slotID)))
		resp.Concat(s.Character.GetGold())
		return resp, nil
	}
	return nil, nil
}

func (h *ReplaceHTItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	itemID := int(utils.BytesToInt(data[6:10], true))
	where := int16(utils.BytesToInt(data[10:12], true))
	to := int16(utils.BytesToInt(data[12:14], true))

	resp := REPLACE_HT_ITEM
	resp.Insert(utils.IntToBytes(uint64(itemID), 4, true), 8) // item id
	resp.Insert(utils.IntToBytes(uint64(where), 2, true), 12) // where slot id

	quantity := slots[where].Quantity

	r := database.ITEM_SLOT
	r.Insert(utils.IntToBytes(uint64(itemID), 4, true), 6)    // item id
	r.Insert(utils.IntToBytes(uint64(quantity), 2, true), 12) // item quantity
	r.Insert(utils.IntToBytes(uint64(where), 2, true), 14)    // where slot id
	resp.Concat(r)

	r, err = s.Character.ReplaceItem(itemID, where, to)
	if err != nil {
		return nil, err
	}

	resp.Concat(r)
	resp.Concat(slots[to].GetData(to))
	return resp, nil
}

func (h *InspectItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {
	if s.User.UserType < server.COMMON_USER {
		return nil, nil
	}
	resp := utils.Packet{0xaa, 0x55, 0xf1, 0x02, 0x62, 0x01, 0x0a, 0x00, 0x11, 0x55, 0xaa}
	pseudoID := uint16(utils.BytesToInt(data[6:8], true))
	index := 9
	char := database.FindCharacterByPseudoID(s.User.ConnectedServer, pseudoID)
	slots := char.GetEquipedItemSlots()
	inventory, err := char.InventorySlots()
	if err != nil {
		return nil, err
	}

	for _, s := range slots {
		slot := inventory[s]
		//id := utils.IntToBytes(uint64(s), 2, true)
		plus := 161 + slot.Plus
		resp.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), index) // item id
		index += 4
		resp.Insert(utils.IntToBytes(uint64(plus), 2, true), index)
		index += 2
		resp.Insert([]byte{0x01, 0x00}, index) //ELVILEG UGYAN AZ MINDIG
		index += 2
		resp.Insert(utils.IntToBytes(uint64(s), 2, true), index) // item slot
		index += 2
		resp.Insert(slot.GetUpgrades(), index) // item plus
		index += 15
		resp.Insert([]byte{0x00, 0x00}, index)
		index += 2
		resp.Insert(utils.IntToBytes(uint64(slot.SocketCount), 2, true), index) //SOCKET LATER MUST FIX
		index += 2
		resp.Insert(slot.GetSockets(), index) // item plus
		index += 15
		if slot.ItemType != 0 {
			resp.Overwrite(utils.IntToBytes(uint64(slot.ItemType), 1, true), index-6)
			if slot.ItemType == 2 {
				resp.Overwrite(utils.IntToBytes(uint64(slot.JudgementStat), 4, true), index-5)
			}
		}
	}

	return resp, nil
}

func (h *DressUpHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	isHT := data[6] == 1
	if isHT {
		s.Character.HTVisibility = int(data[7])
		resp := HT_VISIBILITY
		resp[9] = data[7]

		itemsData, err := s.Character.ShowItems()
		if err != nil {
			return nil, err
		}

		p := nats.CastPacket{CastNear: true, CharacterID: s.Character.ID, Type: nats.SHOW_ITEMS, Data: itemsData}
		if err = p.Cast(); err != nil {
			return nil, err
		}

		resp.Concat(itemsData)

		return resp, nil
	}

	return nil, nil
}

func (h *SplitItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	where := uint16(utils.BytesToInt(data[6:8], true))
	to := uint16(utils.BytesToInt(data[8:10], true))
	quantity := uint16(utils.BytesToInt(data[10:12], true))

	return s.Character.SplitItem(where, to, quantity)
}

func (h *HolyWaterUpgradeHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	itemSlot := int16(utils.BytesToInt(data[6:8], true))
	item := slots[itemSlot]
	if itemSlot == 0 || item.ItemID == 0 {
		return nil, nil
	}

	holyWaterSlot := int16(utils.BytesToInt(data[8:10], true))
	holyWater := slots[holyWaterSlot]
	if holyWaterSlot == 0 || holyWater.ItemID == 0 {
		return nil, nil
	}

	return s.Character.HolyWaterUpgrade(item, holyWater, itemSlot, holyWaterSlot)
}

func (h *UseConsumableHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	itemID := int64(utils.BytesToInt(data[6:10], true))
	slotID := int16(utils.BytesToInt(data[10:12], true))

	item := slots[slotID]
	if item == nil || item.ItemID != itemID {
		return nil, nil
	}

	return s.Character.UseConsumable(item, slotID)
}

func (h *OpenBoxHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	slotID := int16(utils.BytesToInt(data[6:8], true))

	item := slots[slotID]
	if item == nil {
		return nil, nil
	}

	return s.Character.UseConsumable(item, slotID)
}

func (h *OpenBoxHandler2) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	itemID := utils.BytesToInt(data[6:10], true)
	slotID := int16(utils.BytesToInt(data[10:12], true))

	item := slots[slotID]
	if item == nil || item.ItemID != itemID {
		return nil, nil
	}

	return s.Character.UseConsumable(item, slotID)
}

func (h *ActivateTimeLimitedItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	slotID := int16(utils.BytesToInt(data[6:8], true))
	item := slots[slotID]
	if item == nil {
		return nil, nil
	}

	info := database.Items[item.ItemID]
	if info == nil || info.Timer == 0 {
		return nil, nil
	}

	hasSameBuff := len(funk.Filter(slots, func(slot *database.InventorySlot) bool {
		return slot.Activated && slot.ItemID == item.ItemID
	}).([]*database.InventorySlot)) > 0

	if hasSameBuff {
		return []byte{0xAA, 0x55, 0x04, 0x00, 0x59, 0x04, 0xFA, 0x03, 0x55, 0xAA}, nil
	}

	resp := utils.Packet{}
	item.Activated = !item.Activated
	item.InUse = !item.InUse
	resp.Concat(item.GetData(slotID))

	item.Update()
	statsData, _ := s.Character.GetStats()
	resp.Concat(statsData)
	return resp, nil
}

func (h *ActivateTimeLimitedItemHandler2) Handle(s *database.Socket, data []byte) ([]byte, error) {

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, err
	}

	where := int16(utils.BytesToInt(data[6:8], true))
	itemID := utils.BytesToInt(data[8:12], true)
	to := int16(utils.BytesToInt(data[12:14], true))

	item := slots[where]
	if item == nil || item.ItemID != itemID {
		return nil, nil
	}

	info := database.Items[item.ItemID]
	if info == nil || info.Timer == 0 {
		return nil, nil
	}

	hasSameBuff := len(funk.Filter(slots, func(slot *database.InventorySlot) bool {
		return slot.Activated && slot.ItemID == item.ItemID
	}).([]*database.InventorySlot)) > 0

	if hasSameBuff {
		return []byte{0xAA, 0x55, 0x04, 0x00, 0x59, 0x04, 0xFA, 0x03, 0x55, 0xAA}, nil
	}

	resp := utils.Packet{}
	item.Activated = !item.Activated
	item.InUse = !item.InUse
	s.Conn.Write(item.GetData(where))

	var itemData utils.Packet
	if slots[to].ItemID == 0 {
		itemData, err = s.Character.ReplaceItem(int(itemID), where, to)
	} else {
		itemData, err = s.Character.SwapItems(where, to)
	}

	if err != nil {
		return nil, err
	}
	resp.Concat(itemData)

	item.Update()
	statsData, _ := s.Character.GetStats()
	resp.Concat(statsData)
	return resp, nil
}

func (h *ToggleMountPetHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	return s.Character.ToggleMountPet(), nil
}

func (h *TogglePetHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	return s.Character.TogglePet(), nil
}

func (h *PetCombatModeHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {
	CombatMode := utils.BytesToInt(data[7:8], true)
	slots, err := s.Character.InventorySlots()
	if err != nil {
		log.Println(err)
		return nil, nil
	}
	petSlot := slots[0x0A]
	pet := petSlot.Pet
	if pet == nil || petSlot.ItemID == 0 || !pet.IsOnline {
		return nil, nil
	}
	pet.PetCombatMode = int16(CombatMode)
	resp := PET_COMBAT
	resp.Insert(utils.IntToBytes(uint64(CombatMode), 1, true), 9)
	return resp, nil
}
