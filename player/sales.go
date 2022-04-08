package player

import (
	"hero-emulator/database"
	"hero-emulator/utils"
)

type (
	OpenBuyMenuHandler  struct{}
	OpenSaleMenuHandler struct{}
	OpenSaleHandler     struct{}
	VisitSaleHandler    struct{}
	CloseSaleHandler    struct{}
	BuySaleItemHandler  struct{}
)

var (
	OPEN_SALE_MENU = utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x55, 0x09, 0x0A, 0x00, 0x00, 0x55, 0xAA}
	OPEN_BUY_MENU  = utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x68, 0x09, 0x0A, 0x00, 0x00, 0x55, 0xAA}
)

func (h *OpenSaleMenuHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {
	if s.Character.AidMode {
		return nil, nil
	}

	sale := database.FindSale(s.Character.PseudoID)
	if sale != nil {
		return nil, nil
	}

	return OPEN_SALE_MENU, nil
}

func (h *OpenBuyMenuHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {
	if s.Character.AidMode {
		return nil, nil
	}

	sale := database.FindSale(s.Character.PseudoID)
	if sale != nil {
		return nil, nil
	}

	return OPEN_BUY_MENU, nil
}

func (h *OpenSaleHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if s.Character.AidMode {
		return nil, nil
	}

	sale := database.FindSale(s.Character.PseudoID)
	if sale != nil {
		return nil, nil
	}

	saleNameLength := data[6]
	saleName := string(data[7 : 7+saleNameLength])

	index := 7 + saleNameLength

	itemCount := int(data[index])
	index++

	slotIDs, prices := []int16{}, []uint64{}
	for i := 0; i < itemCount; i++ {
		slotID := int16(utils.BytesToInt(data[index:index+2], true))
		index += 2
		index += 2

		price := uint64(utils.BytesToInt(data[index:index+8], true))
		index += 8

		slotIDs = append(slotIDs, slotID)
		prices = append(prices, price)
	}

	return s.Character.OpenSale(saleName, slotIDs, prices)
}

func (h *VisitSaleHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if s.Character.AidMode {
		return nil, nil
	}

	sale := database.FindSale(s.Character.PseudoID)
	if sale != nil {
		return nil, nil
	}

	saleID := uint16(utils.BytesToInt(data[6:8], true))
	sale = database.FindSale(saleID)
	if sale != nil {
		s.Character.VisitedSaleID = saleID
		return sale.Data, nil
	}

	return nil, nil
}

func (h *CloseSaleHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	sale := database.FindSale(s.Character.PseudoID)
	visitors := database.FindSaleVisitors(sale.ID)
	for _, v := range visitors {
		v.Socket.Write(database.CLOSE_SALE)
		v.VisitedSaleID = 0
	}

	return s.Character.CloseSale()
}

func (h *BuySaleItemHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	if s.Character.AidMode {
		return nil, nil
	}

	sale := database.FindSale(s.Character.PseudoID)
	if sale != nil {
		return nil, nil
	}

	saleID := uint16(utils.BytesToInt(data[6:8], true))
	saleSlotID := int16(utils.BytesToInt(data[8:10], true))
	invSlotID := int16(utils.BytesToInt(data[10:12], true))

	return s.Character.BuySaleItem(saleID, saleSlotID, invSlotID)
}
