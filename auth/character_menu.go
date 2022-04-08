package auth

import (
	"encoding/binary"
	"sort"

	"hero-emulator/database"
	"hero-emulator/utils"
)

type ListCharactersHandler struct {
	username string
}

var (
	CHARACTER_LIST = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x01, 0x02, 0x0A, 0x00, 0x01, 0x01, 0x00, 0x55, 0xAA}
	NO_CHARACTERS  = utils.Packet{0xAA, 0x55, 0x07, 0x00, 0x01, 0x02, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
)

func (lch *ListCharactersHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	length := int(data[6])
	lch.username = string(data[7 : length+7])
	return lch.listCharacters(s)
}

func (lch *ListCharactersHandler) listCharacters(s *database.Socket) ([]byte, error) {

	user := findUser(lch.username)
	if user == nil /*|| user.ConnectingIP == ""*/ {
		return nil, nil
	}
	if user.IsLoginedFromPanel == false {
		database.GetSocket(s.User.ID).Conn.Close()
		return nil, nil
	}
	s.User = user
	s.ClientAddr = s.User.ConnectingIP
	s.User.ConnectedIP = s.ClientAddr
	s.User.ConnectedServer = s.User.ConnectingTo
	s.User.ConnectingTo = 0
	s.User.ConnectingIP = ""
	s.Add(s.User.ID)
	database.FindCharactersByUserID(s.User.ID)
	go s.User.Update()
	return lch.showCharacterMenu(s)
}

func (lch *ListCharactersHandler) showCharacterMenu(s *database.Socket) ([]byte, error) {

	characters, err := database.FindCharactersByUserID(s.User.ID)
	if err != nil {
		return nil, err
	}

	if len(characters) == 0 {
		return NO_CHARACTERS, nil
	}
	sort.Slice(characters, func(i, j int) bool {
		if characters[i].CreatedAt.Time.Sub(characters[j].CreatedAt.Time) < 0 {
			return true
		}
		return false
	})

	resp := CHARACTER_LIST
	//return resp, nil
	length := 8
	resp[9] = byte(characters[0].Faction)
	resp[10] = byte(len(characters))

	index := 11
	for i, c := range characters {
		length += len(c.Name) + 269
		resp.Insert([]byte{byte(i)}, index) // character index
		index += 1

		id := uint64(c.ID)
		resp.Insert(utils.IntToBytes(id, 4, true), index) // character id
		index += 4

		resp.Insert([]byte{byte(len(c.Name))}, index) // character name length
		index += 1

		resp.Insert([]byte(c.Name), index) // character name
		index += len(c.Name)

		resp.Insert([]byte{byte(c.Type), byte(c.Class)}, index) // character type-class
		index += 2

		resp.Insert(utils.IntToBytes(uint64(c.Level), 4, true), index) //character level
		index += 4

		resp.Insert([]byte{0x3E}, index) //0x3E volt
		index += 1

		resp.Insert([]byte{byte(c.WeaponSlot)}, index) // character weapon slot
		index += 1

		resp.Insert([]byte{0x00, 0x00}, index)
		index += 2
		resp.Insert(utils.IntToBytes(uint64(c.HeadStyle), 4, true), index)
		index += 4
		resp.Insert(utils.IntToBytes(uint64(c.FaceStyle), 4, true), index)
		index += 4
		resp.Insert([]byte{0x00, 0x00, 0x00}, index)
		index += 3

		slots := c.GetAppearingItemSlots()
		inventory, err := c.InventorySlots()
		if err != nil {
			return nil, err
		}

		for i, s := range slots {
			slot := inventory[s]
			itemID := slot.ItemID
			if slot.Appearance != 0 {
				itemID = slot.Appearance
			} else if slot.Appearance == 0 && itemID == 0 {

			}
			resp.Insert(utils.IntToBytes(uint64(itemID), 4, true), index) // item id
			index += 4
			resp.Insert([]byte{0x00, 0x00, 0x00}, index)
			index += 3
			resp.Insert(utils.IntToBytes(uint64(i), 2, true), index) // item slot
			index += 2

			resp.Insert([]byte{byte(slot.Plus), 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, index) // item plus
			index += 13
		}
		resp.Insert([]byte{0x00, 0x26, 0x26, 0x26, 0x26, 0x26, 0x26, 0x26, 0x26, 0x26, 0x26}, index)
		index += 11
		/*resp.Insert([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, index)
		index += 10
		resp.Insert([]byte{0x00}, index)*/
	}
	resp.SetLength(int16(binary.Size(resp) - 6))
	return resp, nil
}

func findUser(username string) *database.User {

	all := database.AllUsers()
	for _, u := range all {
		if u.Username == username {
			return u
		}
	}

	return nil
}
