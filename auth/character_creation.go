package auth

import (
	"hero-emulator/database"
	"hero-emulator/logging"
	"hero-emulator/messaging"
	"hero-emulator/utils"
)

type CancelCharacterCreationHandler struct {
}

type CharacterCreationHandler struct {
	characterType int
	faction       int
	height        int
	name          string
	headstyle     int64
	facestyle     int64
	Buff          int
	itemid        int
}

var (
	CHARACTER_CREATED = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x01, 0x03, 0x0A, 0x00, 0x00, 0x00, 0x55, 0xAA}
)

func (ccch *CancelCharacterCreationHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	lch := &ListCharactersHandler{}
	return lch.showCharacterMenu(s)
}

func (cch *CharacterCreationHandler) Handle(s *database.Socket, data []byte) ([]byte, error) {

	index := 7
	length := int(data[index])
	index += 1

	cch.name = string(data[8 : length+8])
	index += len(cch.name)

	cch.characterType = int(data[index])
	index += 1

	characters, err := database.FindCharactersByUserID(s.User.ID)
	if err != nil {
		return nil, err
	}

	if len(characters) > 0 {
		cch.faction = characters[0].Faction
	} else {
		cch.faction = int(data[index])
	}
	index += 1

	cch.height = int(data[index])
	headint := utils.BytesToInt(data[index:index+4], true)
	cch.headstyle = headint
	index += 4
	faceint := utils.BytesToInt(data[index:index+4], true)
	cch.facestyle = faceint

	return cch.createCharacter(s)
}

func (cch *CharacterCreationHandler) createCharacter(s *database.Socket) ([]byte, error) {

	ok, err := database.IsValidUsername(cch.name)
	if err != nil {
		return nil, err
	} else if !ok {
		return messaging.SystemMessage(messaging.INVALID_NAME), nil
	} else if cch.faction == 0 {
		return messaging.SystemMessage(messaging.EMPTY_FACTION), nil
	}

	coordinate := database.SavePoints[70]
	if err != nil {
		return nil, err
	}

	character := &database.Character{
		Type:           cch.characterType,
		UserID:         s.User.ID,
		Name:           cch.name,
		Epoch:          0,
		Faction:        cch.faction,
		Height:         cch.height,
		Level:          1,
		Class:          0,
		IsOnline:       false,
		IsActive:       false,
		Gold:           210000,
		Map:            70,
		Exp:            0,
		HTVisibility:   0,
		WeaponSlot:     3,
		RunningSpeed:   5.6,
		GuildID:        -1,
		ExpMultiplier:  1,
		DropMultiplier: 1,
		Slotbar:        []byte{},
		Coordinate:     coordinate.Point,
		AidTime:        120,
		HeadStyle:      cch.headstyle,
		FaceStyle:      cch.facestyle,
	}

	err = character.Create()
	if err != nil {
		return nil, err
	}

	character.AddItem(&database.InventorySlot{ItemID: 17200576, Quantity: 1}, -1, false)
	if character.Type == 50 {
		buff := database.Buff{ID: int(277),
			CharacterID: character.ID, Name: "Great Tiger Statue", DEFRate: 5, ArtsDEFRate: 5, ATKRate: 4, ArtsATKRate: 4, StartedAt: character.Epoch, Duration: 99946, CanExpire: false}
		err = buff.Create()
	}
	if character.Type == 51 {
		buff := database.Buff{ID: int(280),
			CharacterID: character.ID, Name: "Youngblood Tiger Statue", DEFRate: 5, ArtsDEFRate: 5, ATKRate: 4, ArtsATKRate: 4, StartedAt: character.Epoch, Duration: 99946, CanExpire: false}
		err = buff.Create()

	}
	character.Update()

	stat := &database.Stat{}
	err = stat.Create(character)
	if err != nil {
		return nil, err
	}

	skills := &database.Skills{}
	err = skills.Create(character)
	if err != nil {
		return nil, err
	}

	stat, err = database.FindStatByID(character.ID)
	if err != nil {
		return nil, err
	}

	err = stat.Calculate()
	if err != nil {
		return nil, err
	}

	resp := CHARACTER_CREATED
	length := int16(len(cch.name)) + 10
	resp.SetLength(length)

	resp.Insert(utils.IntToBytes(uint64(character.ID), 4, true), 9) // character id

	resp[13] = byte(len(cch.name)) // character name length

	resp.Insert([]byte(cch.name), 14) // character name

	lch := &ListCharactersHandler{}
	data, err := lch.showCharacterMenu(s)
	if err != nil {
		return nil, err
	}

	logger.Log(logging.ACTION_CREATE_CHARACTER, character.ID, "Character created", s.User.ID)
	resp.Concat(data)
	return resp, nil

}
