package player

import (
	"fmt"
	"net"
	"time"

	"hero-emulator/database"
	"hero-emulator/logging"
	"hero-emulator/messaging"
	"hero-emulator/server"
	"hero-emulator/utils"

	"github.com/google/uuid"
	"gopkg.in/guregu/null.v3"
)

type (
	SendTradeRequestHandler    struct{}
	RespondTradeRequestHandler struct{}
	CancelTradeHandler         struct{}
	AddTradeItemHandler        struct{}
	AddTradeGoldHandler        struct{}
	RemoveTradeItemHandler     struct{}
	AcceptTradeHandler         struct{}
)

var (
	SEND_TRADE_REQUEST     = utils.Packet{0xAA, 0x55, 0x06, 0x00, 0x53, 0x01, 0x0A, 0x00, 0x55, 0xAA}
	TRADE_REQUEST_ACCEPTED = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x53, 0x02, 0x0A, 0x00, 0x55, 0xAA}
	TRADE_ITEM_ADDED       = utils.Packet{0xAA, 0x55, 0x33, 0x00, 0x53, 0x04, 0x0A, 0x00, 0x00, 0x00, 0xA2, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	TRADE_GOLD_ADDED       = utils.Packet{0xAA, 0x55, 0x0E, 0x00, 0x53, 0x06, 0x0A, 0x00, 0x55, 0xAA}
	TRADE_ITEM_REMOVED     = utils.Packet{0xAA, 0x55, 0x07, 0x00, 0x53, 0x08, 0x0A, 0x00, 0x00, 0x55, 0xAA}
	TRADE_ACCEPTED         = utils.Packet{0xAA, 0x55, 0x07, 0x00, 0x53, 0x09, 0x0A, 0x00, 0x01, 0x55, 0xAA}
	TRADE_REJECTED         = utils.Packet{0xAA, 0x55, 0x07, 0x00, 0x53, 0x09, 0x0A, 0x00, 0x00, 0x55, 0xAA}
	TRADE_COMPLETED        = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x53, 0x10, 0x0A, 0x00, 0x00, 0x55, 0xAA}

	logger = logging.Logger
)

func (h *SendTradeRequestHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if s.Character.TradeID != "" || !s.Character.IsActive {
		return messaging.SystemMessage(messaging.INVALID_TRADE_REQUEST), nil
	}

	user, err := database.FindUserByID(s.Character.UserID)
	if err != nil {
		return nil, err
	}

	pseudoID := uint16(utils.BytesToInt(data[6:8], true))
	receiver := server.FindCharacter(user.ConnectedServer, pseudoID)
	if receiver == nil {
		return database.TRADE_CANCELLED, nil

	} else if !receiver.IsActive {
		return messaging.SystemMessage(messaging.INVALID_TRADE_REQUEST), nil
	}

	sock := database.GetSocket(receiver.UserID)
	if sock == nil {
		return nil, nil
	}

	resp := SEND_TRADE_REQUEST
	resp.Insert(utils.IntToBytes(uint64(s.Character.PseudoID), 2, true), 8) // sender pseudo id

	tradeID := uuid.New().String()
	s.Character.TradeID = tradeID
	sock.Conn.Write(resp)

	time.AfterFunc(time.Second*12, func() {
		trade := database.FindTrade(s.Character)
		if trade == nil && s.Character.TradeID == tradeID {
			s.Character.TradeID = ""

			r := messaging.SystemMessage(messaging.TRADE_REQUEST_REJECTED)
			s.Character.Socket.Write(r)
		}
	})
	return nil, nil
}

func (h *RespondTradeRequestHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if s.Character.TradeID != "" {
		return messaging.SystemMessage(messaging.INVALID_TRADE_REQUEST), nil
	}

	user, err := database.FindUserByID(s.Character.UserID)
	if err != nil {
		return nil, err
	}

	accepted := data[6] == 1
	pseudoID := uint16(utils.BytesToInt(data[7:9], true))

	sender := server.FindCharacter(user.ConnectedServer, pseudoID)
	if sender == nil || sender.TradeID == "" {
		return database.TRADE_CANCELLED, nil
	}

	t := database.FindTrade(sender)
	if t != nil {
		return database.TRADE_CANCELLED, nil
	}

	sock := database.GetSocket(sender.UserID)
	if sock == nil {
		return nil, nil
	}

	resp, r := utils.Packet{}, utils.Packet{}
	if accepted && sender.IsOnline {
		resp = TRADE_REQUEST_ACCEPTED

		trade := database.Trade{}
		trade.New(sender, s.Character)

		s.Character.TradeID = sender.TradeID
		r = resp

	} else {
		sender.TradeID = ""
		resp = database.TRADE_CANCELLED
		r = messaging.SystemMessage(messaging.TRADE_REQUEST_REJECTED)
	}

	sock.Conn.Write(r)
	return resp, nil
}

func (h *CancelTradeHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	s.Character.CancelTrade()
	return nil, nil
}

func (h *AddTradeItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	trade := database.FindTrade(s.Character)
	if trade == nil {
		return nil, nil
	}

	snd, rcv := trade.Sender.Accepted, trade.Receiver.Accepted
	trade.Sender.Accepted = false
	trade.Receiver.Accepted = false

	slots, err := s.Character.InventorySlots()
	if err != nil {
		return nil, nil
	}

	slotID := int16(utils.BytesToInt(data[6:8], true))
	//count := utils.BytesToInt(data[8:10], true)
	tradeSlotID := int16(utils.BytesToInt(data[10:12], true))

	item := slots[slotID]
	if item == nil {
		return nil, nil
	}

	info := database.Items[item.ItemID]
	if !info.Tradable {
		return nil, nil
	}

	resp := TRADE_ITEM_ADDED
	resp.Insert(utils.IntToBytes(uint64(s.Character.PseudoID), 2, true), 8) // character pseudo id
	resp[10] = byte(tradeSlotID)                                            // trade slot id
	resp.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 11)         // item id

	if info.GetType() == database.PET_TYPE {
		pet := item.Pet
		resp[16] = pet.Level
		resp.Insert([]byte{pet.Loyalty, pet.Fullness}, 17)                   // loyalty and fullness
		resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 19)           // slot id
		resp.Insert(utils.IntToBytes(uint64(pet.HP), 2, true), 21)           // pet hp
		resp.Insert(utils.IntToBytes(uint64(pet.CHI), 2, true), 23)          // pet chi
		resp.Insert(utils.IntToBytes(uint64(pet.Exp), 8, true), 25)          // pet exp
		resp.Insert([]byte{0, 0, 0}, 33)                                     // padding
		resp[36] = 0                                                         // padding
		resp.Insert([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, 37) // padding

	} else {
		resp[16] = 0xA2
		resp.Insert(utils.IntToBytes(uint64(item.Quantity), 2, true), 17) // item quantity
		resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 19)        // slot id
		resp.Insert(item.GetUpgrades(), 21)                               // item upgrades
		resp[36] = byte(item.SocketCount)                                 // socket count
		resp.Insert(item.GetSockets(), 37)                                // item sockets
		c := 37 + 15
		if item.ItemType != 0 {
			resp.Overwrite(utils.IntToBytes(uint64(item.ItemType), 1, true), c-6)
			if item.ItemType == 2 {
				resp.Overwrite(utils.IntToBytes(uint64(item.JudgementStat), 4, true), c-5)
			}
		}
	}

	if snd {
		r := TRADE_REJECTED
		r.Insert(utils.IntToBytes(uint64(trade.Sender.Character.PseudoID), 2, true), 8) // character pseudo id
		resp.Concat(r)
	}
	if rcv {
		r := TRADE_REJECTED
		r.Insert(utils.IntToBytes(uint64(trade.Receiver.Character.PseudoID), 2, true), 8) // character pseudo id
		resp.Concat(r)
	}

	isSender := trade.Sender.Character.UserID == s.Character.UserID
	if isSender {
		trade.Receiver.Character.Socket.Write(resp)
		trade.Sender.Slots[tradeSlotID] = slotID
	} else {
		trade.Sender.Character.Socket.Write(resp)
		trade.Receiver.Slots[tradeSlotID] = slotID
	}

	return resp, nil
}

func (h *AddTradeGoldHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	trade := database.FindTrade(s.Character)
	if trade == nil {
		return nil, nil
	}

	snd, rcv := trade.Sender.Accepted, trade.Receiver.Accepted
	trade.Sender.Accepted = false
	trade.Receiver.Accepted = false

	gold := utils.BytesToInt(data[6:14], true)
	if s.Character.Gold < uint64(gold) || gold < 0 {
		return nil, nil
	}

	resp := TRADE_GOLD_ADDED
	resp.Insert(utils.IntToBytes(uint64(s.Character.PseudoID), 2, true), 8) // character pseudo id
	resp.Insert(utils.IntToBytes(uint64(gold), 8, true), 10)                // gold

	if snd {
		r := TRADE_REJECTED
		r.Insert(utils.IntToBytes(uint64(trade.Sender.Character.PseudoID), 2, true), 8) // character pseudo id
		resp.Concat(r)
	}
	if rcv {
		r := TRADE_REJECTED
		r.Insert(utils.IntToBytes(uint64(trade.Receiver.Character.PseudoID), 2, true), 8) // character pseudo id
		resp.Concat(r)
	}

	isSender := trade.Sender.Character.UserID == s.Character.UserID
	if isSender {
		trade.Receiver.Character.Socket.Write(resp)
		trade.Sender.Gold = uint64(gold)
	} else {
		trade.Sender.Character.Socket.Write(resp)
		trade.Receiver.Gold = uint64(gold)
	}

	return resp, nil
}

func (h *RemoveTradeItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	trade := database.FindTrade(s.Character)
	if trade == nil {
		return nil, nil
	}

	snd, rcv := trade.Sender.Accepted, trade.Receiver.Accepted
	trade.Sender.Accepted = false
	trade.Receiver.Accepted = false

	tradeSlotID := int16(utils.BytesToInt(data[10:12], true))

	resp := TRADE_ITEM_REMOVED
	resp.Insert(utils.IntToBytes(uint64(s.Character.PseudoID), 2, true), 8) // character pseudo id
	resp[10] = byte(tradeSlotID)

	if snd {
		r := TRADE_REJECTED
		r.Insert(utils.IntToBytes(uint64(trade.Sender.Character.PseudoID), 2, true), 8) // character pseudo id
		resp.Concat(r)
	}
	if rcv {
		r := TRADE_REJECTED
		r.Insert(utils.IntToBytes(uint64(trade.Receiver.Character.PseudoID), 2, true), 8) // character pseudo id
		resp.Concat(r)
	}

	isSender := trade.Sender.Character.UserID == s.Character.UserID
	if isSender {
		trade.Receiver.Character.Socket.Write(resp)
		delete(trade.Sender.Slots, tradeSlotID)
	} else {
		trade.Sender.Character.Socket.Write(resp)
		delete(trade.Receiver.Slots, tradeSlotID)
	}

	return resp, nil
}

func (h *AcceptTradeHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	trade := database.FindTrade(s.Character)
	if trade == nil {
		return nil, nil
	}

	if trade.Completing {
		return nil, nil
	}

	accepted := data[6] == 1

	var conn net.Conn
	isSender := trade.Sender.Character.UserID == s.Character.UserID
	if isSender {
		conn = trade.Receiver.Character.Socket.Conn
		trade.Sender.Accepted = accepted
		if !accepted && trade.Receiver.Accepted {
			trade.Receiver.Accepted = false
		}

	} else {
		conn = trade.Sender.Character.Socket.Conn
		trade.Receiver.Accepted = accepted
		if !accepted && trade.Sender.Accepted {
			trade.Sender.Accepted = false
		}
	}

	resp := utils.Packet{}
	if accepted {
		resp = TRADE_ACCEPTED
		resp.Insert(utils.IntToBytes(uint64(s.Character.PseudoID), 2, true), 8) // character pseudo id

	} else {
		r := TRADE_REJECTED
		r.Insert(utils.IntToBytes(uint64(trade.Sender.Character.PseudoID), 2, true), 8) // character pseudo id
		resp.Concat(r)

		r = TRADE_REJECTED
		r.Insert(utils.IntToBytes(uint64(trade.Receiver.Character.PseudoID), 2, true), 8) // character pseudo id
		resp.Concat(r)
	}

	if conn != nil {
		conn.Write(resp)
	}

	revertCBs := []func(){}
	rGold, sGold := uint64(0), uint64(0)
	receiverResp, senderResp := utils.Packet{}, utils.Packet{}
	if trade.Sender.Accepted && trade.Receiver.Accepted && !trade.Completing {

		trade.Completing = true

		senderSlots, err := trade.Sender.Character.InventorySlots()
		if err != nil {
			return nil, nil
		}

		receiverSlots, err := trade.Receiver.Character.InventorySlots()
		if err != nil {
			return nil, nil
		}

		success := true
		if success {
			sGold += trade.Receiver.Gold
			sGold -= trade.Sender.Gold

			rGold += trade.Sender.Gold
			rGold -= trade.Receiver.Gold
		}

		recvItemIDs, senderItemIDs := []int{}, []int{}
		itemArray := []*database.InventorySlot{}
		if success {
			r, r2 := TRADE_COMPLETED, utils.Packet{}
			r.Insert(utils.IntToBytes(trade.Receiver.Character.Gold+rGold, 8, true), 8) // receiver character gold
			r[16] = byte(len(trade.Sender.Slots))

			index, length := 17, int16(13)
			for _, slotID := range trade.Sender.Slots {
				item := *senderSlots[slotID]
				senderItemIDs = append(senderItemIDs, item.ID)

				freeSlot, err := trade.Receiver.Character.FindFreeSlot()
				if err != nil {
					success = false
					break
				}

				//receiverSlots[freeSlot] = &item

				r.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), index) // item id
				index += 4
				r.Insert([]byte{0x00, 0xA2}, index)
				index += 2
				r.Insert(utils.IntToBytes(uint64(item.Quantity), 2, true), index) // item quantity
				index += 2
				r.Insert(utils.IntToBytes(uint64(freeSlot), 2, true), index) // item quantity
				index += 2
				r.Insert(item.GetUpgrades(), index) // item upgrades
				index += 15
				r.Insert([]byte{byte(item.SocketCount)}, index) // socket count
				index++
				r.Insert(item.GetSockets(), index) // item upgrades
				index += 15

				if item.ItemType != 0 {
					resp.Overwrite(utils.IntToBytes(uint64(item.ItemType), 1, true), index-6)
					if item.ItemType == 2 {
						resp.Overwrite(utils.IntToBytes(uint64(item.JudgementStat), 4, true), index-5)
					}
				}
				r.Insert([]byte{0x00, 0x00, 0x00}, index)
				index += 3
				length += 44

				if item.Pet != nil {
					r2.Concat(item.GetData(freeSlot))
				}

				item.UserID = null.StringFrom(trade.Receiver.Character.UserID)
				item.CharacterID = null.IntFrom(int64(trade.Receiver.Character.ID))
				item.SlotID = freeSlot

				*receiverSlots[freeSlot] = item
				*senderSlots[slotID] = *database.NewSlot()

				cb := func() {
					item.UserID = null.StringFrom(trade.Sender.Character.UserID)
					item.CharacterID = null.IntFrom(int64(trade.Sender.Character.ID))
					item.SlotID = slotID

					*receiverSlots[freeSlot] = *database.NewSlot()
					*senderSlots[slotID] = item
				}

				revertCBs = append(revertCBs, cb)
				itemArray = append(itemArray, receiverSlots[freeSlot])

				senderResp.Concat(senderSlots[slotID].GetData(slotID))
			}

			r.SetLength(length)
			receiverResp.Concat(r)
			receiverResp.Concat(r2)
		}

		if success {
			r, r2 := TRADE_COMPLETED, utils.Packet{}
			r.Insert(utils.IntToBytes(trade.Sender.Character.Gold+sGold, 8, true), 8) // sender character gold
			r[16] = byte(len(trade.Receiver.Slots))

			index, length := 17, int16(13)
			for _, slotID := range trade.Receiver.Slots {
				item := *receiverSlots[slotID]
				recvItemIDs = append(recvItemIDs, item.ID)

				freeSlot, err := trade.Sender.Character.FindFreeSlot()
				if err != nil {
					success = false
					break
				}

				//senderSlots[freeSlot] = &item

				r.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), index) // item id
				index += 4
				r.Insert([]byte{0x00, 0xA2}, index)
				index += 2
				r.Insert(utils.IntToBytes(uint64(item.Quantity), 2, true), index) // item quantity
				index += 2
				r.Insert(utils.IntToBytes(uint64(freeSlot), 2, true), index) // item quantity
				index += 2
				r.Insert(item.GetUpgrades(), index) // item upgrades
				index += 15
				r.Insert([]byte{byte(item.SocketCount)}, index) // socket count
				index++
				r.Insert(item.GetSockets(), index) // item upgrades
				index += 15
				if item.ItemType != 0 {
					resp.Overwrite(utils.IntToBytes(uint64(item.ItemType), 1, true), index-6)
					if item.ItemType == 2 {
						resp.Overwrite(utils.IntToBytes(uint64(item.JudgementStat), 4, true), index-5)
					}
				}
				r.Insert([]byte{0x00, 0x00, 0x00}, index)
				index += 3
				length += 44

				if item.Pet != nil {
					r2.Concat(item.GetData(freeSlot))
				}

				item.UserID = null.StringFrom(trade.Sender.Character.UserID)
				item.CharacterID = null.IntFrom(int64(trade.Sender.Character.ID))
				item.SlotID = freeSlot

				*senderSlots[freeSlot] = item
				*receiverSlots[slotID] = *database.NewSlot()

				cb := func() {
					item.UserID = null.StringFrom(trade.Receiver.Character.UserID)
					item.CharacterID = null.IntFrom(int64(trade.Receiver.Character.ID))
					item.SlotID = slotID

					*senderSlots[freeSlot] = *database.NewSlot()
					*receiverSlots[slotID] = item
				}

				revertCBs = append(revertCBs, cb)
				itemArray = append(itemArray, senderSlots[freeSlot])

				receiverResp.Concat(receiverSlots[slotID].GetData(slotID))
			}

			r.SetLength(length)
			senderResp.Concat(r)
			senderResp.Concat(r2)
		}

		if !success { // trade failed
			for _, cb := range revertCBs {
				cb()
			}

			resp := database.TRADE_CANCELLED
			if conn != nil {
				conn.Write(resp)
			}

			logger.Log(logging.ACTION_TRADE, trade.Sender.Character.ID, fmt.Sprintf("Trade failed with (%d)", trade.Receiver.Character.ID), trade.Sender.Character.UserID)
			logger.Log(logging.ACTION_TRADE, trade.Receiver.Character.ID, fmt.Sprintf("Trade failed with (%d)", trade.Sender.Character.ID), trade.Receiver.Character.UserID)
			return resp, nil
		}

		for _, i := range itemArray {
			i.Update()
			database.InventoryItems.Add(i.ID, i)
		}

		trade.Sender.Character.LootGold(sGold)
		trade.Receiver.Character.LootGold(rGold)

		if isSender {
			resp.Concat(senderResp)
			if conn != nil {
				conn.Write(receiverResp)
			}
		} else {
			resp.Concat(receiverResp)
			if conn != nil {
				conn.Write(senderResp)
			}
		}

		logger.Log(logging.ACTION_TRADE, trade.Sender.Character.ID,
			fmt.Sprintf("Trade success with (%d), Gold: %d, Items: %+v", trade.Receiver.Character.ID, trade.Receiver.Gold, recvItemIDs),
			trade.Sender.Character.UserID)

		logger.Log(logging.ACTION_TRADE, trade.Receiver.Character.ID,
			fmt.Sprintf("Trade success with (%d), Gold: %d, Items: %+v", trade.Sender.Character.ID, trade.Sender.Gold, senderItemIDs),
			trade.Receiver.Character.UserID)

		trade.Delete()
	}

	return resp, nil
}
