package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

var (
	Items         = make(map[int64]*Item)
	STRRates      = []int{1000, 800, 700, 500, 400, 100, 120, 100, 75, 50, 40, 30, 20, 15, 5}
	socketOrePlus = map[int64]byte{17402319: 1, 17402320: 2, 17402321: 3, 17402322: 4, 17402323: 5}
	haxBoxes      = []int64{92000002, 92000003, 92000004, 92000005, 92000006, 92000007, 92000008, 92000009, 92000010}
)

const (
	WEAPON_TYPE = iota
	ARMOR_TYPE
	HT_ARMOR_TYPE
	ACC_TYPE
	PENDENT_TYPE
	QUEST_TYPE
	PET_ITEM_TYPE
	SKILL_BOOK_TYPE
	PASSIVE_SKILL_BOOK_TYPE
	POTION_TYPE
	PET_TYPE
	PET_POTION_TYPE
	CHARM_OF_RETURN_TYPE
	FORTUNE_BOX_TYPE
	MARBLE_TYPE
	WRAPPER_BOX_TYPE
	NPC_SUMMONER_TYPE
	FIRE_SPIRIT
	WATER_SPIRIT
	HOLY_WATER_TYPE
	FILLER_POTION_TYPE
	SCALE_TYPE
	BAG_EXPANSION_TYPE
	MOVEMENT_SCROLL_TYPE
	SOCKET_TYPE
	INGREDIENTS_TYPE
	DEAD_SPIRIT_INCENSE_TYPE
	AFFLICTION_TYPE
	RESET_ART_TYPE
	RESET_ARTS_TYPE
	FORM_TYPE
	MASTER_HT_ACC
	UNKNOWN_TYPE
)

type Item struct {
	ID              int64   `db:"id"`
	Name            string  `db:"name"`
	UIF             string  `db:"uif"`
	Type            int16   `db:"type"`
	HtType          int16   `db:"ht_type"`
	TimerType       int16   `db:"timer_type"`
	Timer           int     `db:"timer"`
	BuyPrice        int64   `db:"buy_price"`
	SellPrice       int64   `db:"sell_price"`
	Slot            int     `db:"slot"`
	MinLevel        int     `db:"min_level"`
	MaxLevel        int     `db:"max_level"`
	BaseDef1        int     `db:"base_def1"`
	BaseDef2        int     `db:"base_def2"`
	BaseDef3        int     `db:"base_def3"`
	BaseMinAtk      int     `db:"base_min_atk"`
	BaseMaxAtk      int     `db:"base_max_atk"`
	STR             int     `db:"str"`
	DEX             int     `db:"dex"`
	INT             int     `db:"int"`
	Wind            int     `db:"wind"`
	Water           int     `db:"water"`
	Fire            int     `db:"fire"`
	MaxHp           int     `db:"max_hp"`
	MaxChi          int     `db:"max_chi"`
	MinAtk          int     `db:"min_atk"`
	MaxAtk          int     `db:"max_atk"`
	AtkRate         int     `db:"atk_rate"`
	MinArtsAtk      int     `db:"min_arts_atk"`
	MaxArtsAtk      int     `db:"max_arts_atk"`
	ArtsAtkRate     int     `db:"arts_atk_rate"`
	Def             int     `db:"def"`
	DefRate         int     `db:"def_rate"`
	ArtsDef         int     `db:"arts_def"`
	ArtsDefRate     int     `db:"arts_def_rate"`
	Accuracy        int     `db:"accuracy"`
	Dodge           int     `db:"dodge"`
	HpRecovery      int     `db:"hp_recovery"`
	ChiRecovery     int     `db:"chi_recovery"`
	HolyWaterUpg1   int     `db:"holy_water_upg1"`
	HolyWaterUpg2   int     `db:"holy_water_upg2"`
	HolyWaterUpg3   int     `db:"holy_water_upg3"`
	HolyWaterRate1  int     `db:"holy_water_rate1"`
	HolyWaterRate2  int     `db:"holy_water_rate2"`
	HolyWaterRate3  int     `db:"holy_water_rate3"`
	CharacterType   int     `db:"character_type"`
	ExpRate         float64 `db:"exp_rate"`
	DropRate        float64 `db:"drop_rate"`
	Tradable        bool    `db:"tradable"`
	MinUpgradeLevel int16   `db:"min_upgrade_level"`
	NPCID           int     `db:"npc_id"`
	SpecialItem     int64   `db:"special_item"`
	RunningSpeed    float64 `db:"running_speed"`
	ItemBuff        int     `db:"item_buff"`
	PoisonATK       int     `db:"poison_atk"`
	PoisonDEF       int     `db:"poison_def"`
	ConfusionATK    int     `db:"confusion_atk"`
	ConfusionDEF    int     `db:"confusion_def"`
	ParalysisATK    int     `db:"paralysis_atk"`
	ParalysisDEF    int     `db:"paralysis_def"`
}

func (item *Item) Create() error {
	return db.Insert(item)
}

func (item *Item) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(item)
}

func (item *Item) Delete() error {
	_, err := db.Delete(item)
	return err
}

func (item *Item) Update() error {
	_, err := db.Update(item)
	return err
}

func (item *Item) GetType() int {
	if item.Type == 51 {
		return FIRE_SPIRIT
	} else if item.Type == 52 {
		return WATER_SPIRIT
	} else if item.Type == 59 {
		return BAG_EXPANSION_TYPE
	} else if item.Type == 64 {
		return MARBLE_TYPE
	} else if (item.Type >= 70 && item.Type <= 71) || (item.Type >= 99 && item.Type <= 108) || item.Type == 44 || item.Type == 49 {
		return WEAPON_TYPE
	} else if item.Type == 80 {
		return SOCKET_TYPE
	} else if item.Type == 81 {
		return HOLY_WATER_TYPE
	} else if item.Type == 90 {
		return MASTER_HT_ACC
	} else if item.Type == 110 {
		return AFFLICTION_TYPE
	} else if item.Type == 111 {
		return RESET_ART_TYPE
	} else if item.Type == 112 {
		return RESET_ARTS_TYPE
	} else if item.Type == 115 {
		return INGREDIENTS_TYPE
	} else if item.Type >= 121 && item.Type <= 124 && (item.HtType == 0 || item.HtType == 4) {
		return ARMOR_TYPE
	} else if ((item.Type >= 121 && item.Type <= 124) || item.Type == 175) && item.HtType > 0 && item.HtType != 4 {
		return HT_ARMOR_TYPE
	} else if item.Type >= 131 && item.Type <= 134 {
		return ACC_TYPE
	} else if item.Type >= 135 && item.Type <= 137 {
		return PET_ITEM_TYPE
	} else if item.Type == 147 {
		return FILLER_POTION_TYPE
	} else if item.Type == 151 {
		return POTION_TYPE
	} else if item.Type == 152 {
		return CHARM_OF_RETURN_TYPE
	} else if item.Type == 153 {
		return MOVEMENT_SCROLL_TYPE
	} else if item.Type == 161 {
		return SKILL_BOOK_TYPE
	} else if item.Type == 162 {
		return PASSIVE_SKILL_BOOK_TYPE
	} else if item.Type == 166 {
		return SCALE_TYPE
	} else if item.Type == 168 || item.Type == 213 {
		return WRAPPER_BOX_TYPE
	} else if item.Type == 174 {
		return FORM_TYPE
	} else if item.Type == 191 {
		return PENDENT_TYPE
	} else if item.Type == 202 {
		return QUEST_TYPE
	} else if item.Type == 203 {
		return FORTUNE_BOX_TYPE
	} else if item.Type == 221 {
		return PET_TYPE
	} else if item.Type == 222 {
		return PET_POTION_TYPE
	} else if item.Type == 223 {
		return DEAD_SPIRIT_INCENSE_TYPE
	} else if item.Type == 233 {
		return NPC_SUMMONER_TYPE
	} else if item.Type == 2331 {
		return MOVEMENT_SCROLL_TYPE
	}
	return UNKNOWN_TYPE
}

func getAllItems() error {

	query := `select * from data.items`

	items := []*Item{}

	if _, err := db.Select(&items, query); err != nil {
		fmt.Println("Error:", err)
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getAllItems: %s", err.Error())
	}

	for _, item := range items {
		Items[item.ID] = item
	}

	return nil
}

func RefreshAllItems() error {

	query := `select * from data.items`

	items := []*Item{}

	if _, err := db.Select(&items, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getAllItems: %s", err.Error())
	}

	for _, item := range items {
		Items[item.ID] = item
	}

	return nil
}

// Determines if a weapon item can use an action with specified type
func (item *Item) CanUse(t byte) bool {
	if item.Type == int16(t) || t == 0 {
		return true
	} else if (item.Type == 70 || item.Type == 71) && (t == 70 || t == 71) {
		return true
	} else if (item.Type == 102 || item.Type == 103) && (t == 102 || t == 103) {
		return true
	} else if item.Type == 44 || item.Type == 49 {
		return true
	}

	return false
}
