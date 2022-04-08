package database

import (
	"database/sql"
	"fmt"

	null "gopkg.in/guregu/null.v3"
)

var (
	FiveClans = make(map[int]*FiveClan)
)

type FiveClan struct {
	AreaID    int       `db:"id"`
	ClanID    int       `db:"clanid"`
	ExpiresAt null.Time `db:"expires_at" json:"expires_at"`
}

func (b *FiveClan) Update() error {
	_, err := db.Update(b)
	return err
}

func getFiveAreas() error {
	var areas []*FiveClan
	query := `select * from data.fiveclan_war`

	if _, err := db.Select(&areas, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getFiveAreas: %s", err.Error())
	}

	for _, cr := range areas {
		FiveClans[cr.AreaID] = cr
	}

	return nil
}

func CaptureFive(number int, char *Character) {
	switch number {
	case 1:
		guild, err := FindGuildByID(char.GuildID)
		if err != nil {
			return
		}
		allmembers, _ := guild.GetMembers()
		for _, m := range allmembers {
			c, err := FindCharacterByID(m.ID)
			if err != nil || c == nil {
				continue
			}
			infection := BuffInfections[60006]
			makeAnnouncement("[" + guild.Name + "] captured [Southern Wood Temple]")
			buff := &Buff{ID: infection.ID, CharacterID: c.ID, Name: infection.Name, EXPMultiplier: 200, StartedAt: c.Epoch, Duration: 7200, CanExpire: true}
			err = buff.Create()
			if err != nil {
				continue
			}
		}
	case 2:
		guild, err := FindGuildByID(char.GuildID)
		if err != nil {
			return
		}
		allmembers, _ := guild.GetMembers()
		for _, m := range allmembers {
			c, err := FindCharacterByID(m.ID)
			if err != nil || c == nil {
				continue
			}
			infection := BuffInfections[60011]
			buff := &Buff{ID: infection.ID, CharacterID: c.ID, Name: infection.Name, EXPMultiplier: 200, StartedAt: c.Epoch, Duration: 7200, CanExpire: true}
			makeAnnouncement("[" + guild.Name + "] captured [Lightning Hill Temple]")
			err = buff.Create()
			if err != nil {
				continue
			}
		}
	case 3:
		guild, err := FindGuildByID(char.GuildID)
		if err != nil {
			return
		}
		allmembers, _ := guild.GetMembers()
		for _, m := range allmembers {
			c, err := FindCharacterByID(m.ID)
			if err != nil || c == nil {
				continue
			}
			infection := BuffInfections[60016]
			buff := &Buff{ID: infection.ID, CharacterID: c.ID, Name: infection.Name, EXPMultiplier: 200, StartedAt: c.Epoch, Duration: 7200, CanExpire: true}
			makeAnnouncement("[" + guild.Name + "] captured [Ocean Army Temple]")
			err = buff.Create()
			if err != nil {
				continue
			}
		}
	case 4:
		guild, err := FindGuildByID(char.GuildID)
		if err != nil {
			return
		}
		allmembers, _ := guild.GetMembers()
		for _, m := range allmembers {
			c, err := FindCharacterByID(m.ID)
			if err != nil || c == nil {
				continue
			}
			infection := BuffInfections[60021]
			buff := &Buff{ID: infection.ID, CharacterID: c.ID, Name: infection.Name, EXPMultiplier: 200, StartedAt: c.Epoch, Duration: 7200, CanExpire: true}
			makeAnnouncement("[" + guild.Name + "] captured [Flame Wolf Temple]")
			err = buff.Create()
			if err != nil {
				continue
			}
		}
	case 5:
		guild, err := FindGuildByID(char.GuildID)
		if err != nil {
			return
		}
		allmembers, _ := guild.GetMembers()
		for _, m := range allmembers {
			c, err := FindCharacterByID(m.ID)
			if err != nil || c == nil {
				continue
			}
			infection := BuffInfections[60026]
			buff := &Buff{ID: infection.ID, CharacterID: c.ID, Name: infection.Name, EXPMultiplier: 200, StartedAt: c.Epoch, Duration: 7200, CanExpire: true}
			makeAnnouncement("[" + guild.Name + "] captured [Western Land Temple]")
			err = buff.Create()
			if err != nil {
				continue
			}
		}
	}
}
