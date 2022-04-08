package database

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	dbg "runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"hero-emulator/logging"
	"hero-emulator/messaging"
	"hero-emulator/nats"
	"hero-emulator/utils"

	"github.com/thoas/go-funk"
	gorp "gopkg.in/gorp.v1"
	null "gopkg.in/guregu/null.v3"
)

const (
	MONK                  = 0x34
	MALE_BLADE            = 0x35
	FEMALE_BLADE          = 0x36
	AXE                   = 0x38
	FEMALE_ROD            = 0x39
	DUAL_BLADE            = 0x3B
	BEAST_KING            = 0x32
	EMPRESS               = 0x33
	DIVINE_MONK           = 0x3E
	DIVINE_MALE_BLADE     = 0x3F
	DIVINE_FEMALE_BLADE   = 0x40
	DIVINE_AXE            = 0x42
	DIVINE_FEMALE_ROD     = 0x43
	DIVINE_DUAL_BLADE     = 0x45
	DIVINE_BEAST_KING     = 0x3c
	DIVINE_EMPRESS        = 0x3D
	DARKNESS_MONK         = 0x48
	DARKNESS_MALE_BLADE   = 0x49
	DARKNESS_FEMALE_BLADE = 0x4A
	DARKNESS_AXE          = 0x4C
	DARKNESS_FEMALE_ROD   = 0x4D
	DARKNESS_DUAL_BLADE   = 0x4F
	DARKNESS_BEAST_KING   = 0x46
	DARKNESS_EMPRESS      = 0x47
)

var (
	characters            = make(map[int]*Character)
	characterMutex        sync.RWMutex
	GenerateID            func(*Character) error
	GeneratePetID         func(*Character, *PetSlot)
	challengerGuild       = &Guild{}
	enemyGuild            = &Guild{}
	Beast_King_Infections = []int16{277, 307, 368, 283, 319, 382, 291, 333, 398, 297, 351, 418}
	Empress_Infections    = []int16{280, 313, 375, 287, 326, 390, 294, 342, 408, 302, 359, 429}
	DEAL_DAMAGE           = utils.Packet{0xAA, 0x55, 0x18, 0x00, 0x16, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	BAG_EXPANDED          = utils.Packet{0xAA, 0x55, 0x17, 0x00, 0xA3, 0x02, 0x01, 0x32, 0x30, 0x32, 0x31, 0x2D, 0x30, 0x35, 0x2D, 0x31, 0x33, 0x20, 0x31, 0x30, 0x3A, 0x32, 0x39, 0x3A, 0x3, 0x31, 0x00, 0x55, 0xAA}
	BANK_ITEMS            = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x57, 0x05, 0x01, 0x02, 0x55, 0xAA}
	CHARACTER_DIED        = utils.Packet{0xAA, 0x55, 0x02, 0x00, 0x12, 0x01, 0x55, 0xAA}
	CHARACTER_SPAWNED     = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x21, 0x01, 0xD7, 0xEF, 0xE6, 0x00, 0x03, 0x01, 0x00, 0x00, 0x00, 0x00, 0xC9, 0x00, 0x00, 0x00,
		0x49, 0x2A, 0xFE, 0x00, 0x20, 0x1C, 0x00, 0x00, 0x02, 0xD2, 0x7E, 0x7F, 0xBF, 0xCD, 0x1A, 0x86, 0x3D, 0x33, 0x33, 0x6B, 0x41, 0xFF, 0xFF, 0x10, 0x27,
		0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0xC4, 0x0E, 0x00, 0x00, 0xC8, 0xBB, 0x30, 0x00, 0x00, 0x03, 0xF3, 0x03, 0x01, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00,
		0x00, 0x10, 0x27, 0x00, 0x00, 0x49, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64, 0xFF, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x55, 0xAA}

	EXP_SKILL_PT_CHANGED = utils.Packet{0xAA, 0x55, 0x0D, 0x00, 0x13, 0x55, 0xAA}

	HP_CHI = utils.Packet{0xAA, 0x55, 0x28, 0x00, 0x16, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x27, 0x00, 0x00, 0x55, 0xAA}

	RESPAWN_COUNTER = utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x12, 0x02, 0x01, 0x00, 0x00, 0x55, 0xAA}
	SHOW_ITEMS      = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x59, 0x05, 0x0A, 0x00, 0x00, 0x00, 0x55, 0xAA}

	TELEPORT_PLAYER  = utils.Packet{0xAA, 0x55, 0x09, 0x00, 0x24, 0x55, 0xAA}
	ITEM_COUNT       = utils.Packet{0xAA, 0x55, 0x0C, 0x00, 0x59, 0x04, 0x0A, 0x00, 0x55, 0xAA}
	GREEN_ITEM_COUNT = utils.Packet{0xAA, 0x55, 0x0A, 0x00, 0x59, 0x19, 0x0A, 0x00, 0x55, 0xAA}
	ITEM_EXPIRED     = utils.Packet{0xAA, 0x55, 0x06, 0x00, 0x69, 0x03, 0x55, 0xAA}
	ITEM_ADDED       = utils.Packet{0xaa, 0x55, 0x2e, 0x00, 0x57, 0x0a, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x83, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xaa}
	ITEM_LOOTED      = utils.Packet{0xAA, 0x55, 0x33, 0x00, 0x59, 0x01, 0x0A, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x21, 0x11, 0x55, 0xAA}

	PTS_CHANGED = utils.Packet{0xAA, 0x55, 0x06, 0x00, 0xA2, 0x04, 0x55, 0xAA}
	GOLD_LOOTED = utils.Packet{0xAA, 0x55, 0x0D, 0x00, 0x59, 0x01, 0x0A, 0x00, 0x02, 0x55, 0xAA}
	GET_GOLD    = utils.Packet{0xAA, 0x55, 0x12, 0x00, 0x63, 0x01, 0x55, 0xAA}

	MAP_CHANGED = utils.Packet{0xAA, 0x55, 0x02, 0x00, 0x2B, 0x01, 0x55, 0xAA, 0xAA, 0x55, 0x0E, 0x00, 0x73, 0x00, 0x00, 0x00, 0x7A, 0x44, 0x55, 0xAA,
		0xAA, 0x55, 0x07, 0x00, 0x01, 0xB9, 0x0A, 0x00, 0x00, 0x01, 0x00, 0x55, 0xAA, 0xAA, 0x55, 0x09, 0x00, 0x24, 0x55, 0xAA,
		0xAA, 0x55, 0x03, 0x00, 0xA6, 0x00, 0x00, 0x55, 0xAA, 0xAA, 0x55, 0x02, 0x00, 0xAD, 0x01, 0x55, 0xAA}

	ITEM_REMOVED = utils.Packet{0xAA, 0x55, 0x0B, 0x00, 0x59, 0x02, 0x0A, 0x00, 0x01, 0x55, 0xAA}
	SELL_ITEM    = utils.Packet{0xAA, 0x55, 0x16, 0x00, 0x58, 0x02, 0x0A, 0x00, 0x20, 0x1C, 0x00, 0x00, 0x55, 0xAA}

	GET_STATS = utils.Packet{0xAA, 0x55, 0xDE, 0x00, 0x14, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x66, 0x66, 0xC6, 0x40, 0xF3,
		0x03, 0x00, 0x00, 0x00, 0x40, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x0C, 0x00, 0x00, 0x00, 0x20, 0x00, 0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x30, 0x30, 0x31, 0x2D, 0x30, 0x31, 0x2D, 0x30,
		0x31, 0x20, 0x30, 0x30, 0x3A, 0x30, 0x30, 0x3A, 0x30, 0x30, 0x20, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x03, 0x00, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x64, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x80, 0x3F, 0x10, 0x27, 0x80, 0x3F, 0x55, 0xAA}

	ITEM_REPLACEMENT   = utils.Packet{0xAA, 0x55, 0x0C, 0x00, 0x59, 0x03, 0x0A, 0x00, 0x55, 0xAA}
	ITEM_SWAP          = utils.Packet{0xAA, 0x55, 0x15, 0x00, 0x59, 0x07, 0x0A, 0x00, 0x00, 0x55, 0xAA}
	HT_UPG_FAILED      = utils.Packet{0xAA, 0x55, 0x31, 0x00, 0x54, 0x02, 0xA7, 0x0F, 0x01, 0x00, 0xA3, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	UPG_FAILED         = utils.Packet{0xAA, 0x55, 0x09, 0x00, 0x54, 0x02, 0xA2, 0x0F, 0x00, 0x55, 0xAA}
	PRODUCTION_SUCCESS = utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x54, 0x04, 0x08, 0x10, 0x01, 0x55, 0xAA}
	PRODUCTION_FAILED  = utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x54, 0x04, 0x09, 0x10, 0x00, 0x55, 0xAA}
	FUSION_SUCCESS     = utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x54, 0x09, 0x08, 0x10, 0x01, 0x55, 0xAA}
	FUSION_FAILED      = utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x54, 0x09, 0x09, 0x10, 0x00, 0x55, 0xAA}
	DISMANTLE_SUCCESS  = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x54, 0x05, 0x68, 0x10, 0x01, 0x00, 0x55, 0xAA}
	EXTRACTION_SUCCESS = utils.Packet{0xAA, 0x55, 0xB7, 0x00, 0x54, 0x06, 0xCC, 0x10, 0x01, 0x00, 0xA2, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	HOLYWATER_FAILED   = utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x54, 0x10, 0x32, 0x11, 0x00, 0x55, 0xAA}
	HOLYWATER_SUCCESS  = utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x54, 0x10, 0x31, 0x11, 0x01, 0x55, 0xAA}
	ITEM_REGISTERED    = utils.Packet{0xAA, 0x55, 0x43, 0x00, 0x3D, 0x01, 0x0A, 0x00, 0x00, 0x80, 0x1A, 0x06, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0D, 0x00,
		0x00, 0x00, 0x63, 0x99, 0xEA, 0x00, 0x00, 0xA1, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	CLAIM_MENU              = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x3D, 0x03, 0x0A, 0x00, 0x55, 0xAA}
	CONSIGMENT_ITEM_BOUGHT  = utils.Packet{0xAA, 0x55, 0x08, 0x00, 0x3D, 0x02, 0x0A, 0x00, 0x55, 0xAA}
	CONSIGMENT_ITEM_SOLD    = utils.Packet{0xAA, 0x55, 0x02, 0x00, 0x3F, 0x00, 0x55, 0xAA}
	CONSIGMENT_ITEM_CLAIMED = utils.Packet{0xAA, 0x55, 0x0A, 0x00, 0x3D, 0x04, 0x0A, 0x00, 0x01, 0x00, 0x55, 0xAA}
	SKILL_UPGRADED          = utils.Packet{0xAA, 0x55, 0x0B, 0x00, 0x81, 0x02, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	SKILL_DOWNGRADED        = utils.Packet{0xAA, 0x55, 0x0E, 0x00, 0x81, 0x03, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	SKILL_REMOVED           = utils.Packet{0xAA, 0x55, 0x09, 0x00, 0x81, 0x06, 0x0A, 0x00, 0x00, 0x55, 0xAA}
	PASSIVE_SKILL_UGRADED   = utils.Packet{0xAA, 0x55, 0x06, 0x00, 0x82, 0x02, 0x0A, 0x00, 0x00, 0x00, 0x55, 0xAA}
	PASSIVE_SKILL_REMOVED   = utils.Packet{0xAA, 0x55, 0x09, 0x00, 0x82, 0x04, 0x0A, 0x00, 0x00, 0x55, 0xAA}
	SKILL_CASTED            = utils.Packet{0xAA, 0x55, 0x1D, 0x00, 0x42, 0x0A, 0x00, 0x00, 0x00, 0x01, 0x01, 0x55, 0xAA}
	TRADE_CANCELLED         = utils.Packet{0xAA, 0x55, 0x06, 0x00, 0x53, 0x03, 0xD5, 0x07, 0x7E, 0x02, 0x55, 0xAA}
	SKILL_BOOK_EXISTS       = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x59, 0x04, 0xFA, 0x03, 0x55, 0xAA}
	INVALID_CHARACTER_TYPE  = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x59, 0x04, 0xF2, 0x03, 0x55, 0xAA}
	CHANGE_RANK             = utils.Packet{0xAA, 0x55, 0x09, 0x00, 0x2F, 0xF1, 0x36, 0x55, 0xAA, 0xAA, 0x55, 0x03, 0x00, 0x2F, 0xf2, 0x00, 0x55, 0xAA}
	NO_SLOTS_FOR_SKILL_BOOK = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x59, 0x04, 0xF3, 0x03, 0x55, 0xAA}
	OPEN_SALE               = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x55, 0x01, 0x0A, 0x00, 0x55, 0xAA}
	DEAL_BUFF_AI            = utils.Packet{0xaa, 0x55, 0x1e, 0x00, 0x16, 0x00, 0x00, 0x00, 0x00, 0x01, 0x01, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xaa}
	GET_SALE_ITEMS          = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x55, 0x03, 0x0A, 0x00, 0x00, 0x55, 0xAA}
	CLOSE_SALE              = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x55, 0x02, 0x0A, 0x00, 0x55, 0xAA}
	BOUGHT_SALE_ITEM        = utils.Packet{0xAA, 0x55, 0x39, 0x00, 0x53, 0x10, 0x0A, 0x00, 0x01, 0x00, 0xA2, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
	SOLD_SALE_ITEM          = utils.Packet{0xAA, 0x55, 0x10, 0x00, 0x55, 0x07, 0x0A, 0x00, 0x55, 0xAA}
	BUFF_INFECTION          = utils.Packet{0xAA, 0x55, 0x0C, 0x00, 0x4D, 0x02, 0x0A, 0x01, 0x55, 0xAA}
	BUFF_EXPIRED            = utils.Packet{0xAA, 0x55, 0x06, 0x00, 0x4D, 0x03, 0x55, 0xAA}

	SPLIT_ITEM = utils.Packet{0xAA, 0x55, 0x5C, 0x00, 0x59, 0x09, 0x0A, 0x00, 0x00, 0xA1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xA1, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}

	RELIC_DROP       = utils.Packet{0xAA, 0x55, 0x00, 0x00, 0x71, 0x10, 0x00, 0x55, 0xAA}
	PVP_FINISHED     = utils.Packet{0xAA, 0x55, 0x02, 0x00, 0x2A, 0x05, 0x55, 0xAA}
	FORM_ACTIVATED   = utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x37, 0x55, 0xAA}
	FORM_DEACTIVATED = utils.Packet{0xAA, 0x55, 0x01, 0x00, 0x38, 0x55, 0xAA}
	HONOR_RANKS      = []int{0, 1, 2, 14, 30, 50, 4}
)

type Target struct {
	Damage  int `db:"-" json:"damage"`
	SkillId int `db:"-" json:"skillid"`
	AI      *AI `db:"-" json:"ai"`
}

type PlayerTarget struct {
	Damage int        `db:"-" json:"damage"`
	Enemy  *Character `db:"-" json:"ai"`
}

type Character struct {
	ID                       int        `db:"id" json:"id"`
	UserID                   string     `db:"user_id" json:"user_id"`
	Name                     string     `db:"name" json:"name"`
	Epoch                    int64      `db:"epoch" json:"epoch"`
	Type                     int        `db:"type" json:"type"`
	Faction                  int        `db:"faction" json:"faction"`
	Height                   int        `db:"height" json:"height"`
	Level                    int        `db:"level" json:"level"`
	Class                    int        `db:"class" json:"class"`
	IsOnline                 bool       `db:"is_online" json:"is_online"`
	IsActive                 bool       `db:"is_active" json:"is_active"`
	Gold                     uint64     `db:"gold" json:"gold"`
	Coordinate               string     `db:"coordinate" json:"coordinate"`
	Map                      int16      `db:"map" json:"map"`
	Exp                      int64      `db:"exp" json:"exp"`
	HTVisibility             int        `db:"ht_visibility" json:"ht_visibility"`
	WeaponSlot               int        `db:"weapon_slot" json:"weapon_slot"`
	RunningSpeed             float64    `db:"running_speed" json:"running_speed"`
	GuildID                  int        `db:"guild_id" json:"guild_id"`
	ExpMultiplier            float64    `db:"exp_multiplier" json:"exp_multiplier"`
	DropMultiplier           float64    `db:"drop_multiplier" json:"drop_multiplier"`
	Slotbar                  []byte     `db:"slotbar" json:"slotbar"`
	CreatedAt                null.Time  `db:"created_at" json:"created_at"`
	AdditionalExpMultiplier  float64    `db:"additional_exp_multiplier" json:"additional_exp_multiplier"`
	AdditionalDropMultiplier float64    `db:"additional_drop_multiplier" json:"additional_drop_multiplier"`
	AidMode                  bool       `db:"aid_mode" json:"aid_mode"`
	AidTime                  uint32     `db:"aid_time" json:"aid_time"`
	HonorRank                int64      `db:"rank" json:"rank"`
	HeadStyle                int64      `db:"headstyle" json:"headstyle"`
	FaceStyle                int64      `db:"facestyle" json:"facestyle"`
	AddingExp                sync.Mutex `db:"-" json:"-"`
	AddingGold               sync.Mutex `db:"-" json:"-"`
	Looting                  sync.Mutex `db:"-" json:"-"`
	AdditionalRunningSpeed   float64    `db:"-" json:"-"`
	InvMutex                 sync.Mutex `db:"-"`
	Socket                   *Socket    `db:"-" json:"-"`
	ExploreWorld             func()     `db:"-" json:"-"`
	HasLot                   bool       `db:"-" json:"-"`
	LastRoar                 time.Time  `db:"-" json:"-"`
	Meditating               bool       `db:"-"`
	MovementToken            int64      `db:"-" json:"-"`
	PseudoID                 uint16     `db:"-" json:"pseudo_id"`
	PTS                      int        `db:"-" json:"pts"`
	OnSight                  struct {
		Drops       map[int]interface{} `db:"-" json:"drops"`
		DropsMutex  sync.RWMutex
		Mobs        map[int]interface{} `db:"-" json:"mobs"`
		MobMutex    sync.RWMutex        `db:"-"`
		NPCs        map[int]interface{} `db:"-" json:"npcs"`
		NpcMutex    sync.RWMutex        `db:"-"`
		Pets        map[int]interface{} `db:"-" json:"pets"`
		PetsMutex   sync.RWMutex        `db:"-"`
		Players     map[int]interface{} `db:"-" json:"players"`
		PlayerMutex sync.RWMutex        `db:"-"`
	} `db:"-" json:"on_sight"`
	PartyID         string          `db:"-"`
	Selection       int             `db:"-" json:"selection"`
	Targets         []*Target       `db:"-" json:"target"`
	TamingAI        *AI             `db:"-" json:"-"`
	PlayerTargets   []*PlayerTarget `db:"-" json:"player_targets"`
	TradeID         string          `db:"-" json:"trade_id"`
	Invisible       bool            `db:"-" json:"-"`
	DetectionMode   bool            `db:"-" json:"-"`
	VisitedSaleID   uint16          `db:"-" json:"-"`
	DuelID          int             `db:"-" json:"-"`
	DuelStarted     bool            `db:"-" json:"-"`
	IsinWar         bool            `db:"-" json:"-"`
	WarKillCount    int             `db:"-" json:"-"`
	WarContribution int             `db:"-" json:"-"`
	IsAcceptedWar   bool            `db:"-" json:"-"`
	Respawning      bool            `db:"-" json:"-"`
	SkillHistory    utils.SMap      `db:"-" json:"-"`
	Morphed         bool            `db:"-" json:"-"`
	IsDungeon       bool            `db:"-" json:"-"`
	DungeonLevel    int16           `db:"-" json:"-"`
	CanTip          int16           `db:"-" json:"-"`
	PacketSended    bool            `db:"-" json:"-"`
	GeneratedNumber int             `db:"-" json:"-"`
	HandlerCB       func()          `db:"-"`
	PetHandlerCB    func()          `db:"-"`

	inventory []*InventorySlot `db:"-" json:"-"`
}

func (t *Character) PreInsert(s gorp.SqlExecutor) error {
	now := time.Now().UTC()
	t.CreatedAt = null.TimeFrom(now)
	return nil
}

func (t *Character) SetCoordinate(coordinate *utils.Location) {
	t.Coordinate = fmt.Sprintf("(%.1f,%.1f)", coordinate.X, coordinate.Y)
}

func (t *Character) Create() error {
	return db.Insert(t)
}

func (t *Character) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(t)
}

func (t *Character) PreUpdate(s gorp.SqlExecutor) error {
	if int64(t.Gold) < 0 {
		t.Gold = 0
	}
	return nil
}

func (t *Character) Update() error {
	_, err := db.Update(t)
	if err != nil {
		log.Println(err)
	}
	return err
}

func (t *Character) Delete() error {
	characterMutex.Lock()
	defer characterMutex.Unlock()

	delete(characters, t.ID)
	_, err := db.Delete(t)
	return err
}
func FindCharactersInMap(mapid int16) map[int]*Character {

	characterMutex.RLock()
	allChars := funk.Values(characters).([]*Character)
	characterMutex.RUnlock()

	allChars = funk.Filter(allChars, func(c *Character) bool {
		if c.Socket == nil {
			return false
		}

		return c.Map == mapid && c.IsOnline
	}).([]*Character)

	candidates := make(map[int]*Character)
	for _, c := range allChars {
		candidates[c.ID] = c
	}

	return candidates
}

func (t *Character) InventorySlots() ([]*InventorySlot, error) {

	if len(t.inventory) > 0 {
		return t.inventory, nil
	}

	inventory := make([]*InventorySlot, 450)

	for i := range inventory {
		inventory[i] = NewSlot()
	}

	slots, err := FindInventorySlotsByCharacterID(t.ID)
	if err != nil {
		return nil, err
	}

	bankSlots, err := FindBankSlotsByUserID(t.UserID)
	if err != nil {
		return nil, err
	}

	for _, s := range slots {
		inventory[s.SlotID] = s
	}

	for _, s := range bankSlots {
		inventory[s.SlotID] = s
	}

	t.inventory = inventory
	return inventory, nil
}

func (t *Character) SetInventorySlots(slots []*InventorySlot) { // FIX HERE
	t.inventory = slots
}

func (t *Character) CopyInventorySlots() []*InventorySlot {
	slots := []*InventorySlot{}
	for _, s := range t.inventory {
		copySlot := *s
		slots = append(slots, &copySlot)
	}

	return slots
}

func RefreshAIDs() error {
	query := `update hops.characters SET aid_time = 18000`
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	characterMutex.RLock()
	allChars := funk.Values(characters).([]*Character)
	characterMutex.RUnlock()
	for _, c := range allChars {
		c.AidTime = 18000
	}

	return err
}

func FindCharactersByUserID(userID string) ([]*Character, error) {

	charMap := make(map[int]*Character)
	for _, c := range characters {
		if c.UserID == userID {
			charMap[c.ID] = c
		}
	}

	var arr []*Character
	query := `select * from hops.characters where user_id = $1`

	if _, err := db.Select(&arr, query, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindCharactersByUserID: %s", err.Error())
	}

	characterMutex.Lock()
	defer characterMutex.Unlock()

	var chars []*Character
	for _, c := range arr {
		char, ok := charMap[c.ID]
		if ok {
			chars = append(chars, char)
		} else {
			characters[c.ID] = c
			chars = append(chars, c)
		}
	}

	return chars, nil
}

func IsValidUsername(name string) (bool, error) {

	var (
		count int64
		err   error
		query string
	)

	re := regexp.MustCompile("^[a-zA-Z0-9]{4,18}$")
	if !re.MatchString(name) {
		return false, nil
	}

	query = `select count(*) from hops.characters where lower(name) = $1`

	if count, err = db.SelectInt(query, strings.ToLower(name)); err != nil {
		return false, fmt.Errorf("IsValidUsername: %s", err.Error())
	}

	return count == 0, nil
}

func FindCharacterByName(name string) (*Character, error) {

	for _, c := range characters {
		if c.Name == name {
			return c, nil
		}
	}

	character := &Character{}
	query := `select * from hops.characters where name = $1`

	if err := db.SelectOne(&character, query, name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindCharacterByName: %s", err.Error())
	}

	characterMutex.Lock()
	defer characterMutex.Unlock()
	characters[character.ID] = character

	return character, nil
}

func FindAllCharacter() ([]*Character, error) {

	charMap := make(map[int]*Character)

	var arr []*Character
	query := `select * from hops.characters`

	if _, err := db.Select(&arr, query); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindAllCharacter: %s", err.Error())
	}

	characterMutex.Lock()
	defer characterMutex.Unlock()

	var chars []*Character
	for _, c := range arr {
		char, ok := charMap[c.ID]
		if ok {
			chars = append(chars, char)
		} else {
			characters[c.ID] = c
			chars = append(chars, c)
		}
	}

	return chars, nil
}

func FindCharacterByID(id int) (*Character, error) {

	characterMutex.RLock()
	c, ok := characters[id]
	characterMutex.RUnlock()

	if ok {
		return c, nil
	}

	character := &Character{}
	query := `select * from hops.characters where id = $1`

	if err := db.SelectOne(&character, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindCharacterByID: %s", err.Error())
	}

	characterMutex.Lock()
	defer characterMutex.Unlock()
	characters[character.ID] = character

	return character, nil
}

func (c *Character) GetAppearingItemSlots() []int {

	helmSlot := 0
	if c.HTVisibility&0x01 != 0 {
		helmSlot = 0x0133
	}

	maskSlot := 1
	if c.HTVisibility&0x02 != 0 {
		maskSlot = 0x0134
	}

	armorSlot := 2
	if c.HTVisibility&0x04 != 0 {
		armorSlot = 0x0135
	}

	bootsSlot := 9
	if c.HTVisibility&0x10 != 0 {
		bootsSlot = 0x0136
	}

	armorSlot2 := 2
	if c.HTVisibility&0x08 != 0 {
		armorSlot2 = 0x0137
	}

	if armorSlot2 != 2 {
		armorSlot = armorSlot2
	}

	return []int{helmSlot, maskSlot, armorSlot, 3, 4, 5, 6, 7, 8, bootsSlot, 10}
}

func (c *Character) GetEquipedItemSlots() []int {
	return []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 307, 309, 310, 312, 313, 314, 315}
}

func (c *Character) Logout() {
	c.IsOnline = false
	c.IsActive = false
	c.IsDungeon = false
	c.OnSight.Drops = map[int]interface{}{}
	c.OnSight.Mobs = map[int]interface{}{}
	c.OnSight.NPCs = map[int]interface{}{}
	c.OnSight.Pets = map[int]interface{}{}
	c.OnSight.Players = map[int]interface{}{}
	c.ExploreWorld = nil
	c.HandlerCB = nil
	c.PetHandlerCB = nil
	c.PTS = 0
	c.TradeID = ""
	c.LeaveParty()
	c.EndPvP()
	sale := FindSale(c.PseudoID)
	if sale != nil {
		sale.Delete()
	}

	if trade := FindTrade(c); trade != nil {
		c.CancelTrade()
	}

	if c.GuildID > 0 {
		guild, err := FindGuildByID(c.GuildID)
		if err == nil && guild != nil {
			guild.InformMembers(c)
		}
	}

	RemoveFromRegister(c)
	RemovePetFromRegister(c)
	//DeleteCharacterFromCache(c.ID)
	//DeleteStatFromCache(c.ID)
}

func (c *Character) EndPvP() {
	if c.DuelID > 0 {
		op, _ := FindCharacterByID(c.DuelID)
		if op != nil {
			op.Socket.Write(PVP_FINISHED)
			op.DuelID = 0
			op.DuelStarted = false
		}
		c.DuelID = 0
		c.DuelStarted = false
		c.Socket.Write(PVP_FINISHED)
	}
}

func DeleteCharacterFromCache(id int) {
	delete(characters, id)
}

func (c *Character) GetNearbyCharacters() ([]*Character, error) {

	var (
		distance = float64(50)
	)

	u, err := FindUserByID(c.UserID)
	if err != nil {
		return nil, err
	}

	myCoordinate := ConvertPointToLocation(c.Coordinate)
	characterMutex.RLock()
	allChars := funk.Values(characters)
	characterMutex.RUnlock()
	characters := funk.Filter(allChars, func(character *Character) bool {

		user, err := FindUserByID(character.UserID)
		if err != nil || user == nil {
			return false
		}

		characterCoordinate := ConvertPointToLocation(character.Coordinate)

		return character.IsOnline && user.ConnectedServer == u.ConnectedServer && character.Map == c.Map &&
			(!character.Invisible || c.DetectionMode) && utils.CalculateDistance(characterCoordinate, myCoordinate) <= distance
	}).([]*Character)

	return characters, nil
}

func (c *Character) GetNearbyAIIDs() ([]int, error) {

	var (
		distance = 64.0
		ids      []int
	)
	if funk.Contains(DungeonZones, c.Map) {
		distance = 150.0
	}
	if c.IsinWar {
		distance = 25.0
	}

	user, err := FindUserByID(c.UserID)
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, nil
	}

	candidates := AIsByMap[user.ConnectedServer][c.Map]
	filtered := funk.Filter(candidates, func(ai *AI) bool {

		characterCoordinate := ConvertPointToLocation(c.Coordinate)
		aiCoordinate := ConvertPointToLocation(ai.Coordinate)

		return utils.CalculateDistance(characterCoordinate, aiCoordinate) <= distance
	})

	for _, ai := range filtered.([]*AI) {
		ids = append(ids, ai.ID)
	}

	return ids, nil
}

func (c *Character) GetNearbyNPCIDs() ([]int, error) {

	var (
		distance = 50.0
		ids      []int
	)

	user, err := FindUserByID(c.UserID)
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, nil
	}

	filtered := funk.Filter(NPCPos, func(pos *NpcPosition) bool {

		characterCoordinate := ConvertPointToLocation(c.Coordinate)
		minLocation := ConvertPointToLocation(pos.MinLocation)
		maxLocation := ConvertPointToLocation(pos.MaxLocation)

		npcCoordinate := &utils.Location{X: (minLocation.X + maxLocation.X) / 2, Y: (minLocation.Y + maxLocation.Y) / 2}

		return c.Map == pos.MapID && utils.CalculateDistance(characterCoordinate, npcCoordinate) <= distance && pos.IsNPC && !pos.Attackable
	})

	for _, pos := range filtered.([]*NpcPosition) {
		ids = append(ids, pos.ID)
	}

	return ids, nil
}

func (c *Character) GetNearbyDrops() ([]int, error) {

	var (
		distance = 50.0
		ids      []int
	)

	user, err := FindUserByID(c.UserID)
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, nil
	}

	allDrops := GetDropsInMap(user.ConnectedServer, c.Map)
	filtered := funk.Filter(allDrops, func(drop *Drop) bool {

		characterCoordinate := ConvertPointToLocation(c.Coordinate)

		return utils.CalculateDistance(characterCoordinate, &drop.Location) <= distance
	})

	for _, d := range filtered.([]*Drop) {
		ids = append(ids, d.ID)
	}

	return ids, nil
}

func (c *Character) SpawnCharacter() ([]byte, error) {

	if c == nil {
		return nil, nil
	}

	resp := CHARACTER_SPAWNED
	resp.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 6) // character pseudo id
	if c.IsActive {
		resp[12] = 3
	} else {
		resp[12] = 4
	}

	/*
		if c.DuelID > 0 {
			resp.Overwrite(utils.IntToBytes(500, 2, true), 13) // duel state
		}
	*/

	resp[17] = byte(len(c.Name))    // character name length
	resp.Insert([]byte(c.Name), 18) // character name

	index := len(c.Name) + 18 + 4
	resp[index] = byte(c.Type) // character type
	index += 1

	index += 8

	coordinate := ConvertPointToLocation(c.Coordinate)
	resp.Insert(utils.FloatToBytes(coordinate.X, 4, true), index) // coordinate-x
	index += 4

	resp.Insert(utils.FloatToBytes(coordinate.Y, 4, true), index) // coordinate-y
	index += 8

	resp.Insert(utils.FloatToBytes(coordinate.X, 4, true), index) // coordinate-x
	index += 4

	resp.Insert(utils.FloatToBytes(coordinate.Y, 4, true), index) // coordinate-y
	index += 4
	index += 18

	resp.Overwrite(utils.IntToBytes(uint64(c.Socket.Stats.HP), 4, true), index) // hp
	index += 9

	resp[index] = byte(c.WeaponSlot) // weapon slot
	index += 16

	resp.Insert(utils.IntToBytes(uint64(c.GuildID), 4, true), index) // guild id
	index += 8

	resp[index] = byte(c.Faction) // character faction
	index += 10

	items, err := c.ShowItems()
	if err != nil {
		return nil, err
	}

	itemsData := items[11 : len(items)-2]
	sale := FindSale(c.PseudoID)
	if sale != nil {
		itemsData = []byte{0x05, 0xAA, 0x45, 0xF1, 0x00, 0x00, 0x00, 0xA1, 0x00, 0x00, 0x00, 0x00, 0xB4, 0x6C, 0xF1, 0x00, 0x01, 0x00, 0xA1, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x09, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	}

	resp.Insert(itemsData, index)
	index += len(itemsData)

	length := int16(len(itemsData) + len(c.Name) + 111)

	if sale != nil {
		resp.Insert([]byte{0x02}, index) // sale indicator
		index++

		resp.Insert([]byte{byte(len(sale.Name))}, index) // sale name length
		index++

		resp.Insert([]byte(sale.Name), index) // sale name
		index += len(sale.Name)

		resp.Insert([]byte{0x00}, index)
		index++
		length += int16(len(sale.Name) + 3)
	}

	resp.SetLength(length)
	resp.Concat(items) // FIX => workaround for weapon slot

	if c.GuildID > 0 {
		guild, err := FindGuildByID(c.GuildID)
		if err == nil && guild != nil {
			resp.Concat(guild.GetInfo())
		}
	}

	return resp, nil
}

func (c *Character) ShowItems() ([]byte, error) {

	if c == nil {
		return nil, nil
	}

	slots := c.GetAppearingItemSlots()
	inventory, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	helm := inventory[slots[0]]
	mask := inventory[slots[1]]
	armor := inventory[slots[2]]
	weapon1 := inventory[slots[3]]
	weapon2 := inventory[slots[4]]
	boots := inventory[slots[9]]
	pet := inventory[slots[10]].Pet

	count := byte(4)
	if weapon1.ItemID > 0 {
		count++
	}
	if weapon2.ItemID > 0 {
		count++
	}
	if pet != nil && pet.IsOnline {
		count++
	}
	weapon1ID := weapon1.ItemID
	if weapon1.Appearance != 0 {
		weapon1ID = weapon1.Appearance
	}
	weapon2ID := weapon2.ItemID
	if weapon2.Appearance != 0 {
		weapon2ID = weapon2.Appearance
	}
	helmID := helm.ItemID
	if slots[0] == 0 && helm.Appearance != 0 {
		helmID = helm.Appearance
	}
	maskID := mask.ItemID
	if slots[1] == 1 && mask.Appearance != 0 {
		maskID = mask.Appearance
	}
	armorID := armor.ItemID
	if slots[2] == 2 && armor.Appearance != 0 {
		armorID = armor.Appearance
	}
	bootsID := boots.ItemID
	if slots[9] == 9 && boots.Appearance != 0 {
		bootsID = boots.Appearance
	}
	resp := SHOW_ITEMS
	resp.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 8) // character pseudo id
	resp[10] = byte(c.WeaponSlot)                                 // character weapon slot
	resp[11] = count

	index := 12
	resp.Insert(utils.IntToBytes(uint64(helm.ItemID), 4, true), index) // helm id
	index += 4

	resp.Insert(utils.IntToBytes(uint64(slots[0]), 2, true), index) // helm slot
	resp.Insert([]byte{0xA2}, index+2)
	index += 3

	resp.Insert(utils.IntToBytes(uint64(helm.Plus), 1, true), index) // helm plus
	resp.Insert(utils.IntToBytes(uint64(helmID), 4, true), index+1)
	index += 5

	resp.Insert(utils.IntToBytes(uint64(mask.ItemID), 4, true), index) // mask id
	index += 4

	resp.Insert(utils.IntToBytes(uint64(slots[1]), 2, true), index) // mask slot
	resp.Insert([]byte{0xA2}, index+2)
	index += 3

	resp.Insert(utils.IntToBytes(uint64(mask.Plus), 1, true), index) // mask plus
	resp.Insert(utils.IntToBytes(uint64(maskID), 4, true), index+1)
	index += 5

	resp.Insert(utils.IntToBytes(uint64(armor.ItemID), 4, true), index) // armor id
	index += 4

	resp.Insert(utils.IntToBytes(uint64(slots[2]), 2, true), index) // armor slot
	resp.Insert([]byte{0xA2}, index+2)
	index += 3

	resp.Insert(utils.IntToBytes(uint64(armor.Plus), 1, true), index) // armor plus
	resp.Insert(utils.IntToBytes(uint64(armorID), 4, true), index+1)
	index += 5

	if weapon1.ItemID > 0 {
		resp.Insert(utils.IntToBytes(uint64(weapon1.ItemID), 4, true), index) // weapon1 id
		index += 4

		resp.Insert([]byte{0x03, 0x00}, index) // weapon1 slot
		resp.Insert([]byte{0xA2}, index+2)
		index += 3

		resp.Insert(utils.IntToBytes(uint64(weapon1.Plus), 1, true), index) // weapon1 plus
		resp.Insert(utils.IntToBytes(uint64(weapon1ID), 4, true), index+1)
		index += 5
	}

	if weapon2.ItemID > 0 {
		resp.Insert(utils.IntToBytes(uint64(weapon2.ItemID), 4, true), index) // weapon2 id
		index += 4

		resp.Insert([]byte{0x04, 0x00}, index) // weapon2 slot
		resp.Insert([]byte{0xA2}, index+2)
		index += 3

		resp.Insert(utils.IntToBytes(uint64(weapon2.Plus), 1, true), index) // weapon2 plus
		resp.Insert(utils.IntToBytes(uint64(weapon2ID), 4, true), index+1)
		index += 5
	}

	resp.Insert(utils.IntToBytes(uint64(boots.ItemID), 4, true), index) // boots id
	index += 4

	resp.Insert(utils.IntToBytes(uint64(slots[9]), 2, true), index) // boots slot
	resp.Insert([]byte{0xA2}, index+2)
	index += 3

	resp.Insert(utils.IntToBytes(uint64(boots.Plus), 1, true), index) // boots plus
	resp.Insert(utils.IntToBytes(uint64(bootsID), 4, true), index+1)
	index += 5

	if pet != nil && pet.IsOnline {
		resp.Insert(utils.IntToBytes(uint64(inventory[10].ItemID), 4, true), index) // pet id
		index += 4

		resp.Insert(utils.IntToBytes(uint64(slots[10]), 2, true), index) // pet slot
		resp.Insert([]byte{pet.Level}, index+2)
		index += 3

		resp.Insert(utils.IntToBytes(4, 1, true), index) // pet plus ?
		resp.Insert([]byte{0x00, 0x00, 0x00, 0x00}, index+1)
		index += 5
	}

	resp.SetLength(int16(count*12) + 8) // packet length
	return resp, nil
}

func FindOnlineCharacterByUserID(userID string) (*Character, error) {

	var id int
	query := `select id from hops.characters where user_id = $1 and is_online = true`

	if err := db.SelectOne(&id, query, userID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindOnlineCharacterByUserID: %s", err.Error())
	}

	return FindCharacterByID(id)
}

func FindCharactersInServer(server int) (map[int]*Character, error) {

	characterMutex.RLock()
	allChars := funk.Values(characters).([]*Character)
	characterMutex.RUnlock()

	allChars = funk.Filter(allChars, func(c *Character) bool {
		if c.Socket == nil {
			return false
		}
		user := c.Socket.User
		if user == nil {
			return false
		}

		return user.ConnectedServer == server && c.IsOnline
	}).([]*Character)

	candidates := make(map[int]*Character)
	for _, c := range allChars {
		candidates[c.ID] = c
	}

	return candidates, nil
}

func FindOnlineCharacters() (map[int]*Character, error) {

	characters := make(map[int]*Character)
	users := AllUsers()
	users = funk.Filter(users, func(u *User) bool {
		return u.ConnectedIP != "" && u.ConnectedServer > 0
	}).([]*User)

	for _, u := range users {
		c, _ := FindOnlineCharacterByUserID(u.ID)
		if c == nil {
			continue
		}

		characters[c.ID] = c
	}

	return characters, nil
}
func (c *Character) FindItemInInventoryByPlus(callback func(*InventorySlot) bool, itemPlus uint8, itemIDs ...int64) (int16, *InventorySlot, error) {

	slots, err := c.InventorySlots()
	if err != nil {
		return -1, nil, err
	}

	for index, slot := range slots {
		ok, _ := utils.Contains(itemIDs, slot.ItemID)
		ok2, _ := utils.Contains(itemPlus, slot.Plus)
		if ok && ok2 {
			if index >= 0x43 && index <= 0x132 {
				continue
			}

			if callback == nil || callback(slot) {
				return int16(index), slot, nil
			}
		}
	}

	return -1, nil, nil
}

func (c *Character) FindItemInInventory(callback func(*InventorySlot) bool, itemIDs ...int64) (int16, *InventorySlot, error) {

	slots, err := c.InventorySlots()
	if err != nil {
		return -1, nil, err
	}

	for index, slot := range slots {
		if ok, _ := utils.Contains(itemIDs, slot.ItemID); ok {
			if index >= 0x43 && index <= 0x132 {
				continue
			}

			if callback == nil || callback(slot) {
				return int16(index), slot, nil
			}
		}
	}

	return -1, nil, nil
}

func (c *Character) DecrementItem(slotID int16, amount uint) *utils.Packet {

	slots, err := c.InventorySlots()
	if err != nil {
		return nil
	}

	slot := slots[slotID]
	if slot == nil || slot.ItemID == 0 || slot.Quantity < amount {
		return nil
	}

	slot.Quantity -= amount

	info := Items[slot.ItemID]
	resp := utils.Packet{}

	if info.TimerType == 3 {
		resp = GREEN_ITEM_COUNT
		resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 8)         // slot id
		resp.Insert(utils.IntToBytes(uint64(slot.Quantity), 4, true), 10) // item quantity
	} else {
		resp = ITEM_COUNT
		resp.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), 8)    // item id
		resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 12)        // slot id
		resp.Insert(utils.IntToBytes(uint64(slot.Quantity), 2, true), 14) // item quantity
	}

	if slot.Quantity == 0 {
		err = slot.Delete()
		if err != nil {
			log.Print(err)
		}
		*slot = *NewSlot()
	} else {
		err = slot.Update()
		if err != nil {
			log.Print(err)
		}
	}

	return &resp
}

func (c *Character) FindFreeSlot() (int16, error) {

	slotID := 11
	slots, err := c.InventorySlots()
	if err != nil {
		return -1, err
	}

	for ; slotID <= 66; slotID++ {
		slot := slots[slotID]
		if slot.ItemID == 0 {
			return int16(slotID), nil
		}
	}

	if c.DoesInventoryExpanded() {
		slotID = 341
		for ; slotID <= 396; slotID++ {
			slot := slots[slotID]
			if slot.ItemID == 0 {
				return int16(slotID), nil
			}
		}
	}

	return -1, nil
}

func (c *Character) FindFreeSlots(count int) ([]int16, error) {

	var slotIDs []int16
	slots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	for slotID := int16(11); slotID <= 66; slotID++ {
		slot := slots[slotID]
		if slot.ItemID == 0 {
			slotIDs = append(slotIDs, slotID)
		}
		if len(slotIDs) == count {
			return slotIDs, nil
		}
	}

	if c.DoesInventoryExpanded() {
		for slotID := int16(341); slotID <= 396; slotID++ {
			slot := slots[slotID]
			if slot.ItemID == 0 {
				slotIDs = append(slotIDs, slotID)
			}
			if len(slotIDs) == count {
				return slotIDs, nil
			}
		}
	}

	return nil, fmt.Errorf("not enough inventory space")
}

func (c *Character) DoesInventoryExpanded() bool {
	buffs, err := FindBuffsByCharacterID(c.ID)
	if err != nil || len(buffs) == 0 {
		return false
	}

	buffs = funk.Filter(buffs, func(b *Buff) bool {
		return b.BagExpansion
	}).([]*Buff)

	return len(buffs) > 0
}

func (c *Character) AddItem(itemToAdd *InventorySlot, slotID int16, lootingDrop bool) (*utils.Packet, int16, error) {
	var (
		item *InventorySlot
	)

	if itemToAdd == nil {
		return nil, -1, nil
	}

	itemToAdd.CharacterID = null.IntFrom(int64(c.ID))
	itemToAdd.UserID = null.StringFrom(c.UserID)

	i := Items[itemToAdd.ItemID]
	stackable := FindStackableByUIF(i.UIF)

	slots, err := c.InventorySlots()
	if err != nil {
		return nil, -1, err
	}

	stacking := false
	resp := utils.Packet{}
	if slotID == -1 {
		if stackable != nil { // stackable item
			if itemToAdd.Plus > 0 {
				slotID, item, err = c.FindItemInInventoryByPlus(nil, itemToAdd.Plus, itemToAdd.ItemID)
			} else {
				slotID, item, err = c.FindItemInInventory(nil, itemToAdd.ItemID)
			}
			if err != nil {
				return nil, -1, err
			} else if slotID == -1 { // no same item found => find free slot
				slotID, err = c.FindFreeSlot()
				if err != nil {
					return nil, -1, err
				} else if slotID == -1 { // no free slot
					return nil, -1, nil
				}
				stacking = false
			} else if item.ItemID != itemToAdd.ItemID { // slot is not available
				return nil, -1, nil
			} else if item != nil { // can be stacked
				itemToAdd.Quantity += item.Quantity
				stacking = true
			}
		} else { // not stackable item => find free slot
			slotID, err = c.FindFreeSlot()
			if err != nil {
				return nil, -1, err
			} else if slotID == -1 {
				return nil, -1, nil
			}
		}
	}

	itemToAdd.SlotID = slotID
	slot := slots[slotID]
	id := slot.ID
	*slot = *itemToAdd
	slot.ID = id

	if !stacking && stackable == nil {
		//for j := 0; j < int(itemToAdd.Quantity); j++ {
		if lootingDrop {
			r := ITEM_LOOTED
			r.Insert(utils.IntToBytes(uint64(itemToAdd.ItemID), 4, true), 9) // item id
			r[14] = 0xA1
			if itemToAdd.Plus > 0 || itemToAdd.SocketCount > 0 {
				r[14] = 0xA2
			}

			r.Insert(utils.IntToBytes(uint64(slot.Quantity), 2, true), 15) // item count
			r.Insert(utils.IntToBytes(uint64(slotID), 2, true), 17)        // slot id
			r.Insert(itemToAdd.GetUpgrades(), 19)                          // item upgrades
			resp.Concat(r)
		} else {
			resp = ITEM_ADDED
			resp.Insert(utils.IntToBytes(uint64(itemToAdd.ItemID), 4, true), 6) // item id
			resp.Insert(utils.IntToBytes(uint64(slot.Quantity), 2, true), 12)   // item quantity
			resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 14)          // slot id
			resp.Insert(utils.IntToBytes(uint64(0), 8, true), 20)               // ures
			resp.Overwrite(utils.IntToBytes(uint64(itemToAdd.ItemType), 1, true), 41)
			resp.Overwrite(utils.IntToBytes(uint64(itemToAdd.JudgementStat), 4, true), 42)
		}
		resp.Concat(c.GetGold())
		/*
			slotID, err = c.FindFreeSlot()
			if err != nil || slotID == -1 {
				break
			}

			slot = slots[slotID]
		*/
		//}
	} else {
		if lootingDrop {
			r := ITEM_LOOTED
			r.Insert(utils.IntToBytes(uint64(itemToAdd.ItemID), 4, true), 9) // item id
			r[14] = 0xA1
			r.Insert(utils.IntToBytes(uint64(slot.Quantity), 2, true), 15) // item count
			r.Insert(utils.IntToBytes(uint64(slotID), 2, true), 17)        // slot id
			r.Insert(itemToAdd.GetUpgrades(), 19)                          // item upgrades
			resp.Concat(r)

		} else if stacking {
			resp = ITEM_COUNT
			resp.Insert(utils.IntToBytes(uint64(itemToAdd.ItemID), 4, true), 8)    // item id
			resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 12)             // slot id
			resp.Insert(utils.IntToBytes(uint64(itemToAdd.Quantity), 2, true), 14) // item quantity

		} else if !stacking {
			slot := slots[slotID]
			slot.ItemID = itemToAdd.ItemID
			slot.Quantity = itemToAdd.Quantity
			slot.Plus = itemToAdd.Plus
			slot.UpgradeArr = itemToAdd.UpgradeArr

			resp = ITEM_ADDED
			resp.Insert(utils.IntToBytes(uint64(itemToAdd.ItemID), 4, true), 6)    // item id
			resp.Insert(utils.IntToBytes(uint64(itemToAdd.Quantity), 2, true), 12) // item quantity
			resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 14)             // slot id
			resp.Insert(utils.IntToBytes(uint64(0), 8, true), 20)                  // gold
			resp.Overwrite(utils.IntToBytes(uint64(itemToAdd.ItemType), 1, true), 41)
			resp.Overwrite(utils.IntToBytes(uint64(itemToAdd.JudgementStat), 4, true), 42)
		}
		resp.Concat(c.GetGold())
	}

	if slot.ID > 0 {
		err = slot.Update()
	} else {
		err = slot.Insert()
	}
	if err != nil {
		*slot = *NewSlot()
		resp = utils.Packet{}
		resp.Concat(slot.GetData(slotID))
		return &resp, -1, nil
	}

	InventoryItems.Add(slot.ID, slot)
	resp.Concat(slot.GetData(slotID))
	return &resp, slotID, nil
}

func (c *Character) ReplaceItem(itemID int, where, to int16) ([]byte, error) {

	sale := FindSale(c.PseudoID)
	if sale != nil {
		return nil, fmt.Errorf("cannot replace item on sale")
	} else if c.TradeID != "" {
		return nil, fmt.Errorf("cannot replace item on trade")
	}

	c.InvMutex.Lock()
	defer c.InvMutex.Unlock()

	invSlots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	whereItem := invSlots[where]
	if whereItem.ItemID == 0 {
		return nil, nil
	}

	toItem := invSlots[to]

	if (where >= 0x0043 && where <= 0x132) && (to >= 0x0043 && to <= 0x132) && toItem.ItemID == 0 { // From: Bank, To: Bank
		whereItem.SlotID = to
		*toItem = *whereItem
		*whereItem = *NewSlot()

	} else if (where >= 0x0043 && where <= 0x132) && (to < 0x0043 || to > 0x132) && toItem.ItemID == 0 { // From: Bank, To: Inventory
		whereItem.SlotID = to
		whereItem.CharacterID = null.IntFrom(int64(c.ID))
		*toItem = *whereItem
		*whereItem = *NewSlot()

	} else if (to >= 0x0043 && to <= 0x132) && (where < 0x0043 || where > 0x132) && toItem.ItemID == 0 &&
		!whereItem.Activated && !whereItem.InUse { // From: Inventory, To: Bank

		whereItem.SlotID = to
		whereItem.CharacterID = null.IntFromPtr(nil)
		*toItem = *whereItem
		*whereItem = *NewSlot()

	} else if ((to < 0x0043 || to > 0x132) && (where < 0x0043 || where > 0x132)) && toItem.ItemID == 0 { // From: Inventory, To: Inventory
		whereItem.SlotID = to
		*toItem = *whereItem
		*whereItem = *NewSlot()

	} else {
		return nil, nil
	}

	toItem.Update()
	InventoryItems.Add(toItem.ID, toItem)

	resp := ITEM_REPLACEMENT
	resp.Insert(utils.IntToBytes(uint64(itemID), 4, true), 8) // item id
	resp.Insert(utils.IntToBytes(uint64(where), 2, true), 12) // where slot id
	resp.Insert(utils.IntToBytes(uint64(to), 2, true), 14)    // to slot id

	whereAffects, toAffects := DoesSlotAffectStats(where), DoesSlotAffectStats(to)

	info := Items[int64(itemID)]
	if whereAffects {
		if info != nil && info.Timer > 0 {
			toItem.InUse = false
		}
	}
	if toItem.ItemID != 0 {
		toIteminfo := Items[int64(toItem.ItemID)]
		if toIteminfo.ItemBuff != 0 {
			buff, err := FindBuffByID(int(toIteminfo.ItemBuff), c.ID)
			if err != nil {
				return nil, err
			}
			if buff != nil {
				buff.CanExpire = true
				buff.Update()
			}

		}
	}
	if toAffects {
		if info != nil && info.Timer > 0 {
			toItem.InUse = true
		}
	}

	if whereAffects || toAffects {
		statData, err := c.GetStats()
		if err != nil {
			return nil, err
		}

		resp.Concat(statData)

		itemsData, err := c.ShowItems()
		if err != nil {
			return nil, err
		}

		p := nats.CastPacket{CastNear: true, CharacterID: c.ID, Type: nats.SHOW_ITEMS, Data: itemsData}
		if err = p.Cast(); err != nil {
			return nil, err
		}

		resp.Concat(itemsData)
	}

	if to == 0x0A {
		resp.Concat(invSlots[to].GetPetStats(c))
		resp.Concat(SHOW_PET_BUTTON)
	} else if where == 0x0A {
		resp.Concat(DISMISS_PET)
	}

	if (where >= 317 && where <= 319) || (to >= 317 && to <= 319) {
		resp.Concat(c.GetPetStats())
	}

	return resp, nil
}

//Discirmination_Items   = utils.Packet{ 0xaa,0x55,0x2e,0x00,0x57,0x0a,0x81,0x25,0xe8,0x05,0x00,0xa2,0x01,0x00,0x00,0x00,0x00,0x02,0xe6,0x02,0x00,0x00,0x00,0x00,0x00,0x00,0x55,0xAA}
func (c *Character) SwapItems(where, to int16) ([]byte, error) {

	sale := FindSale(c.PseudoID)
	if sale != nil {
		return nil, fmt.Errorf("cannot swap items on sale")
	} else if c.TradeID != "" {
		return nil, fmt.Errorf("cannot swap item on trade")
	}

	c.InvMutex.Lock()
	defer c.InvMutex.Unlock()

	invSlots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	whereItem := invSlots[where]
	toItem := invSlots[to]

	if whereItem.ItemID == 0 || toItem.ItemID == 0 {
		return nil, nil
	}
	if where == 317 || where == 318 || where == 319 {
		return nil, nil
	}

	if (where >= 0x0043 && where <= 0x132) && (to >= 0x0043 && to <= 0x132) { // From: Bank, To: Bank
		temp := *toItem
		*toItem = *whereItem
		*whereItem = temp
		toItem.SlotID = to
		whereItem.SlotID = where

	} else if (where >= 0x0043 && where <= 0x132) && (to < 0x0043 || to > 0x132) &&
		!toItem.Activated && !toItem.InUse { // From: Bank, To: Inventory

		temp := *toItem
		*toItem = *whereItem
		*whereItem = temp
		toItem.SlotID = to
		whereItem.SlotID = where

	} else if (to >= 0x0043 && to <= 0x132) && (where < 0x0043 || where > 0x132) &&
		!whereItem.Activated && !whereItem.InUse { // From: Inventory, To: Bank

		temp := *toItem
		*toItem = *whereItem
		*whereItem = temp
		toItem.SlotID = to
		whereItem.SlotID = where

	} else if (to < 0x0043 || to > 0x132) && (where < 0x0043 || where > 0x132) { // From: Inventory, To: Inventory
		temp := *toItem
		*toItem = *whereItem
		*whereItem = temp
		toItem.SlotID = to
		whereItem.SlotID = where

	} else {
		return nil, nil
	}

	whereItem.Update()
	toItem.Update()
	InventoryItems.Add(whereItem.ID, whereItem)
	InventoryItems.Add(toItem.ID, toItem)

	resp := ITEM_SWAP
	resp.Insert(utils.IntToBytes(uint64(where), 4, true), 9)  // where slot
	resp.Insert(utils.IntToBytes(uint64(where), 2, true), 13) // where slot
	resp.Insert(utils.IntToBytes(uint64(to), 2, true), 15)    // to slot
	resp.Insert(utils.IntToBytes(uint64(to), 4, true), 17)    // to slot
	resp.Insert(utils.IntToBytes(uint64(to), 2, true), 21)    // to slot
	resp.Insert(utils.IntToBytes(uint64(where), 2, true), 23) // where slot

	whereAffects, toAffects := DoesSlotAffectStats(where), DoesSlotAffectStats(to)
	//whereItemInfo := Items[whereItem.ItemID]
	if toItem.ItemID != 0 {
		toIteminfo := Items[int64(toItem.ItemID)]
		if toIteminfo.ItemBuff != 0 {
			buff, err := FindBuffByID(int(toIteminfo.ItemBuff), c.ID)
			if err != nil {
				return nil, err
			}
			if buff != nil {
				buff.CanExpire = true
				buff.Update()
			}

		}
	}
	if whereAffects {
		item := whereItem // new item
		info := Items[int64(item.ItemID)]
		if info != nil && info.Timer > 0 {
			item.InUse = true
		}

		item = toItem // old item
		info = Items[int64(item.ItemID)]
		if info != nil && info.Timer > 0 {
			item.InUse = false
		}
	}

	if toAffects {
		item := whereItem // old item
		info := Items[int64(item.ItemID)]
		if info != nil && info.Timer > 0 {
			item.InUse = false
		}

		item = toItem // new item
		info = Items[int64(item.ItemID)]
		if info != nil && info.Timer > 0 {
			item.InUse = true
		}
	}

	if whereAffects || toAffects {

		statData, err := c.GetStats()
		if err != nil {
			return nil, err
		}

		resp.Concat(statData)

		itemsData, err := c.ShowItems()
		if err != nil {
			return nil, err
		}

		p := nats.CastPacket{CastNear: true, CharacterID: c.ID, Data: itemsData, Type: nats.SHOW_ITEMS}
		if err = p.Cast(); err != nil {
			return nil, err
		}

		resp.Concat(itemsData)
	}

	if to == 0x0A {
		resp.Concat(invSlots[to].GetPetStats(c))
		resp.Concat(SHOW_PET_BUTTON)
	}

	if (where >= 317 && where <= 319) || (to >= 317 && to <= 319) {
		resp.Concat(c.GetPetStats())
	}

	return resp, nil
}

func (c *Character) SplitItem(where, to, quantity uint16) ([]byte, error) {

	sale := FindSale(c.PseudoID)
	if sale != nil {
		return nil, fmt.Errorf("cannot split item on sale")
	} else if c.TradeID != "" {
		return nil, fmt.Errorf("cannot split item on trade")
	}

	c.InvMutex.Lock()
	defer c.InvMutex.Unlock()

	slots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	whereItem := slots[where]
	toItem := slots[to]

	if quantity > 0 {

		if whereItem.Quantity >= uint(quantity) {
			*toItem = *whereItem
			toItem.SlotID = int16(to)
			toItem.Quantity = uint(quantity)
			c.DecrementItem(int16(where), uint(quantity))

		} else {
			return nil, nil
		}

		toItem.Insert()
		InventoryItems.Add(toItem.ID, toItem)

		resp := SPLIT_ITEM
		resp.Insert(utils.IntToBytes(uint64(toItem.ItemID), 4, true), 8)       // item id
		resp.Insert(utils.IntToBytes(uint64(whereItem.Quantity), 2, true), 14) // remaining quantity
		resp.Insert(utils.IntToBytes(uint64(where), 2, true), 16)              // where slot id

		resp.Insert(utils.IntToBytes(uint64(toItem.ItemID), 4, true), 52) // item id
		resp.Insert(utils.IntToBytes(uint64(quantity), 2, true), 58)      // new quantity
		resp.Insert(utils.IntToBytes(uint64(to), 2, true), 60)            // to slot id
		resp.Concat(toItem.GetData(int16(to)))
		return resp, nil
	}

	return nil, nil
}
func (c *Character) GetHPandChi() []byte {
	hpChi := HP_CHI
	stat := c.Socket.Stats

	hpChi.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 5)
	hpChi.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 7)
	hpChi.Insert(utils.IntToBytes(uint64(stat.HP), 4, true), 9)
	hpChi.Insert(utils.IntToBytes(uint64(stat.CHI), 4, true), 13)

	count := 0
	buffs, _ := FindBuffsByCharacterID(c.ID)
	for _, buff := range buffs {

		_, ok := BuffInfections[buff.ID]
		if !ok {
			continue
		}

		hpChi.Insert(utils.IntToBytes(uint64(buff.ID), 4, true), 22)
		hpChi.Insert([]byte{0x01, 0x01}, 26)
		count++
	}

	if c.AidMode {
		hpChi.Insert(utils.IntToBytes(11121, 4, true), 22)
		hpChi.Insert([]byte{0x00, 0x00}, 26)
		count++
	}

	hpChi[21] = byte(count) // buff count
	hpChi.SetLength(int16(0x28 + count*6))

	return hpChi
}

func (c *Character) Handler() {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			log.Printf("handler error: %+v", string(dbg.Stack()))
			c.HandlerCB = nil
			c.Socket.Conn.Close()
		}
	}()

	st := c.Socket.Stats
	c.Epoch++

	if st.HP > 0 && c.Epoch%2 == 0 {
		hp, chi := st.HP, st.CHI
		if st.HP += st.HPRecoveryRate; st.HP > st.MaxHP {
			st.HP = st.MaxHP
		}

		if st.CHI += st.CHIRecoveryRate; st.CHI > st.MaxCHI {
			st.CHI = st.MaxCHI
		}

		if c.Meditating {
			if st.HP += st.HPRecoveryRate; st.HP > st.MaxHP {
				st.HP = st.MaxHP
			}

			if st.CHI += st.CHIRecoveryRate; st.CHI > st.MaxCHI {
				st.CHI = st.MaxCHI
			}
		}

		if hp != st.HP || chi != st.CHI {
			c.Socket.Write(c.GetHPandChi()) // hp-chi packet
		}

	} else if st.HP <= 0 && !c.Respawning { // dead
		c.Respawning = true
		st.HP = 0
		c.Socket.Write(c.GetHPandChi())
		c.Socket.Write(CHARACTER_DIED)
		go c.RespawnCounter(10)

		if c.DuelID > 0 { // lost pvp
			opponent, _ := FindCharacterByID(c.DuelID)

			c.DuelID = 0
			c.DuelStarted = false
			c.Socket.Write(PVP_FINISHED)

			opponent.DuelID = 0
			opponent.DuelStarted = false
			opponent.Socket.Write(PVP_FINISHED)

			//info := fmt.Sprintf("[%s] has defeated [%s]", opponent.Name, c.Name)
			//r := messaging.InfoMessage(info)

			//p := nats.CastPacket{CastNear: true, CharacterID: c.ID, Data: r, Type: nats.PVP_FINISHED}
			//p.Cast()
		}
	}

	if c.AidTime <= 0 && c.AidMode {

		c.AidTime = 0
		c.AidMode = false
		c.Socket.Write(c.AidStatus())

		tpData, _ := c.ChangeMap(c.Map, nil)
		c.Socket.Write(tpData)
	}

	if c.AidMode && !c.HasAidBuff() {
		c.AidTime--
		if c.AidTime%60 == 0 {
			stData, _ := c.GetStats()
			c.Socket.Write(stData)
		}
	}

	if !c.AidMode && c.Epoch%2 == 0 && c.AidTime < 7200 {
		c.AidTime++
		if c.AidTime%60 == 0 {
			stData, _ := c.GetStats()
			c.Socket.Write(stData)
		}
	}

	if c.PartyID != "" {
		c.UpdatePartyStatus()
	}

	c.HandleBuffs()
	c.HandleLimitedItems()

	go c.Update()
	go st.Update()
	time.AfterFunc(time.Second, func() {
		if c.HandlerCB != nil {
			c.HandlerCB()
		}
	})
}

func (c *Character) PetHandler() {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			log.Printf("%+v", string(dbg.Stack()))
		}
	}()

	{
		slots, err := c.InventorySlots()
		if err != nil {
			log.Println(err)
			goto OUT
		}

		petSlot := slots[0x0A]
		pet := petSlot.Pet
		if pet == nil || petSlot.ItemID == 0 || !pet.IsOnline {
			return
		}

		petInfo, ok := Pets[petSlot.ItemID]
		if !ok {
			return
		}

		if pet.HP <= 0 {
			resp := utils.Packet{}
			resp.Concat(c.GetPetStats())
			resp.Concat(DISMISS_PET)
			c.Socket.Write(resp)

			pet.IsOnline = false
			return
		}
		if c.AidMode {

		}
		if pet.Target == 0 && pet.Loyalty >= 10 {
			pet.Target, err = pet.FindTargetMobID(c) // 75% chance to trigger
			if err != nil {
				log.Println("AIHandler error:", err)
			}
		}

		if pet.Target > 0 {
			pet.IsMoving = false
		}

		if c.Epoch%60 == 0 {
			if pet.Fullness > 1 {
				pet.Fullness--
			}
			if pet.Fullness < 25 && pet.Loyalty > 1 {
				pet.Loyalty--
			} else if pet.Fullness >= 25 && pet.Loyalty < 100 {
				pet.Loyalty++
			}
		}
		cPetLevel := int(pet.Level)
		if c.Epoch%20 == 0 {
			if pet.HP < pet.MaxHP {
				pet.HP = int(math.Min(float64(pet.HP+cPetLevel*3), float64(pet.MaxHP)))
			}
			pet.RefreshStats = true
		}

		if pet.RefreshStats {
			pet.RefreshStats = false
			c.Socket.Write(c.GetPetStats())
		}

		if pet.IsMoving || pet.Casting {
			goto OUT
		}

		if pet.Loyalty < 10 {
			pet.Target = 0
		}

	BEGIN:
		if pet.PetCombatMode == 2 {
			pet.Target = c.Selection
		}
		if pet.Target == 0 { // Idle mode

			ownerPos := ConvertPointToLocation(c.Coordinate)
			distance := utils.CalculateDistance(ownerPos, &pet.Coordinate)

			if distance > 10 { // Pet is so far from his owner
				pet.IsMoving = true
				targetX := utils.RandFloat(ownerPos.X-5, ownerPos.X+5)
				targetY := utils.RandFloat(ownerPos.Y-5, ownerPos.Y+5)

				target := utils.Location{X: targetX, Y: targetY}
				pet.TargetLocation = target
				speed := float64(10.0)

				token := pet.MovementToken
				for token == pet.MovementToken {
					pet.MovementToken = utils.RandInt(1, math.MaxInt64)
				}

				go pet.MovementHandler(pet.MovementToken, &pet.Coordinate, &target, speed)
			}

		} else { // Target mode
			if pet.PetCombatMode == 2 {
				mob := FindCharacterByPseudoID(c.Socket.User.ConnectedServer, uint16(pet.Target))
				if mob.Socket.Stats.HP <= 0 || !c.CanAttack(mob) {
					pet.Target = 0
					time.Sleep(time.Second)
					goto BEGIN
				}
				aiCoordinate := ConvertPointToLocation(mob.Coordinate)
				distance := utils.CalculateDistance(&pet.Coordinate, aiCoordinate)

				if distance <= 3 && pet.LastHit%2 == 0 { // attack
					seed := utils.RandInt(1, 1000)
					r := utils.Packet{}
					skill, ok := SkillInfos[petInfo.SkillID]
					if seed < 500 && ok && pet.CHI >= skill.BaseChi {
						r.Concat(pet.CastSkill(c))
					} else {
						r.Concat(pet.PlayerAttack(c))
					}

					p := nats.CastPacket{CastNear: true, PetID: pet.PseudoID, Data: r, Type: nats.MOB_ATTACK}
					p.Cast()
					pet.LastHit++

				} else if distance > 3 && distance <= 50 { // chase
					pet.IsMoving = true
					target := GeneratePoint(aiCoordinate)
					pet.TargetLocation = target
					speed := float64(10.0)

					token := pet.MovementToken
					for token == pet.MovementToken {
						pet.MovementToken = utils.RandInt(1, math.MaxInt64)
					}

					go pet.MovementHandler(pet.MovementToken, &pet.Coordinate, &target, speed)
					pet.LastHit = 0

				} else {
					pet.LastHit++
				}
			} else {
				mob, ok := GetFromRegister(c.Socket.User.ConnectedServer, c.Map, uint16(pet.Target)).(*AI)
				if !ok || mob == nil {
					pet.Target = 0
					goto OUT

				} else if mob.HP <= 0 {
					pet.Target = 0
					time.Sleep(time.Second)
					goto BEGIN
				}
				aiCoordinate := ConvertPointToLocation(mob.Coordinate)
				distance := utils.CalculateDistance(&pet.Coordinate, aiCoordinate)

				if distance <= 3 && pet.LastHit%2 == 0 { // attack
					seed := utils.RandInt(1, 1000)
					r := utils.Packet{}
					skill, ok := SkillInfos[petInfo.SkillID]
					if seed < 500 && ok && pet.CHI >= skill.BaseChi {
						r.Concat(pet.CastSkill(c))
					} else {
						r.Concat(pet.Attack(c))
					}

					p := nats.CastPacket{CastNear: true, PetID: pet.PseudoID, Data: r, Type: nats.MOB_ATTACK}
					p.Cast()
					pet.LastHit++

				} else if distance > 3 && distance <= 50 { // chase
					pet.IsMoving = true
					target := GeneratePoint(aiCoordinate)
					pet.TargetLocation = target
					speed := float64(10.0)

					token := pet.MovementToken
					for token == pet.MovementToken {
						pet.MovementToken = utils.RandInt(1, math.MaxInt64)
					}

					go pet.MovementHandler(pet.MovementToken, &pet.Coordinate, &target, speed)
					pet.LastHit = 0

				} else {
					pet.LastHit++
				}
			}

		}

		petSlot.Update()
	}

OUT:
	time.AfterFunc(time.Second, func() {
		if c.PetHandlerCB != nil {
			c.PetHandlerCB()
		}
	})
}

func (c *Character) HandleBuffs() {
	buffs, err := FindBuffsByCharacterID(c.ID)
	if err != nil || len(buffs) == 0 {
		return
	}

	stat := c.Socket.Stats
	if buff := buffs[0]; buff.StartedAt+buff.Duration <= c.Epoch && buff.CanExpire { // buff expired
		stat.MinATK -= buff.ATK
		stat.MaxATK -= buff.ATK
		stat.ATKRate -= buff.ATKRate
		stat.Accuracy -= buff.Accuracy
		stat.MinArtsATK -= buff.ArtsATK
		stat.MaxArtsATK -= buff.ArtsATK
		stat.ArtsATKRate -= buff.ArtsATKRate
		stat.ArtsDEF -= buff.ArtsDEF
		stat.ArtsDEFRate -= buff.ArtsDEFRate
		stat.CHIRecoveryRate -= buff.CHIRecoveryRate
		stat.ConfusionDEF -= buff.ConfusionDEF
		stat.DEF -= buff.DEF
		stat.DefRate -= buff.DEFRate
		stat.DEXBuff -= buff.DEX
		stat.Dodge -= buff.Dodge
		stat.HPRecoveryRate -= buff.HPRecoveryRate
		stat.INTBuff -= buff.INT
		stat.MaxCHI -= buff.MaxCHI
		stat.MaxHP -= buff.MaxHP
		stat.ParalysisDEF -= buff.ParalysisDEF
		stat.PoisonDEF -= buff.PoisonDEF
		stat.STRBuff -= buff.STR
		c.ExpMultiplier -= float64(buff.EXPMultiplier) / 100
		c.DropMultiplier -= float64(buff.DropMultiplier) / 100
		c.RunningSpeed -= buff.RunningSpeed

		data, _ := c.GetStats()

		r := BUFF_EXPIRED
		r.Insert(utils.IntToBytes(uint64(buff.ID), 4, true), 6) // buff infection id
		r.Concat(data)

		c.Socket.Write(r)
		buff.Delete()

		p := &nats.CastPacket{CastNear: true, CharacterID: c.ID, Data: c.GetHPandChi()}
		p.Cast()

		if buff.ID == 241 || buff.ID == 244 { // invisibility
			c.Invisible = false
			if c.DuelID > 0 {
				opponent, _ := FindCharacterByID(c.DuelID)
				sock := opponent.Socket
				if sock != nil {
					time.AfterFunc(time.Second*1, func() {
						sock.Write(opponent.OnDuelStarted())
					})
				}
			}

		} else if buff.ID == 242 || buff.ID == 245 { // detection arts
			c.DetectionMode = true
		}

		if len(buffs) == 1 {
			buffs = []*Buff{}
		} else {
			buffs = buffs[1:]
		}
	}

	for _, buff := range buffs {
		mapping := map[int]int{19000018: 10100, 19000019: 10098}
		id := buff.ID
		if d, ok := mapping[buff.ID]; ok {
			id = d
		}

		infection, ok := BuffInfections[id]
		if !ok {
			continue
		}

		remainingTime := int64(0)
		if buff.CanExpire {
			remainingTime = buff.StartedAt + buff.Duration - c.Epoch
		} else {
			remainingTime = 0
		}
		data := BUFF_INFECTION
		data.Insert(utils.IntToBytes(uint64(infection.ID), 4, true), 6)   // infection id
		data.Insert(utils.IntToBytes(uint64(remainingTime), 4, true), 11) // buff remaining time

		c.Socket.Write(data)
	}
}

func (c *Character) HandleLimitedItems() {

	invSlots, err := c.InventorySlots()
	if err != nil {
		return
	}

	slotIDs := []int16{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0133, 0x0134, 0x0135, 0x0136, 0x0137, 0x0138, 0x0139, 0x013A, 0x013B}

	for _, slotID := range slotIDs {
		slot := invSlots[slotID]
		item := Items[slot.ItemID]
		if item != nil && (item.TimerType == 1 || item.TimerType == 3) { // time limited item
			if c.Epoch%60 == 0 {
				data := c.DecrementItem(slotID, 1)
				c.Socket.Write(*data)
			}
			if slot.Quantity == 0 {
				data := ITEM_EXPIRED
				data.Insert(utils.IntToBytes(uint64(item.ID), 4, true), 6)

				removeData, _ := c.RemoveItem(slotID)
				data.Concat(removeData)

				statData, _ := c.GetStats()
				data.Concat(statData)
				c.Socket.Write(data)
			}
		}
	}

	starts, ends := []int16{0x0B, 0x0155}, []int16{0x043, 0x018D}
	for j := 0; j < 2; j++ {
		start, end := starts[j], ends[j]
		for slotID := start; slotID <= end; slotID++ {
			slot := invSlots[slotID]
			item := Items[slot.ItemID]
			if slot.Activated {
				if c.Epoch%60 == 0 {
					data := c.DecrementItem(slotID, 1)
					c.Socket.Write(*data)
				}
				if slot.Quantity == 0 { // item expired
					data := ITEM_EXPIRED
					data.Insert(utils.IntToBytes(uint64(item.ID), 4, true), 6)

					c.RemoveItem(slotID)
					data.Concat(slot.GetData(slotID))

					statData, _ := c.GetStats()
					data.Concat(statData)
					c.Socket.Write(data)

					if slot.ItemID == 100080008 { // eyeball of divine
						c.DetectionMode = false
					}

					if item.GetType() == FORM_TYPE {
						c.Morphed = false
						c.Socket.Write(FORM_DEACTIVATED)
					}

				} else { // item not expired
					if slot.ItemID == 100080008 && !c.DetectionMode { // eyeball of divine
						c.DetectionMode = true
					} else if item.GetType() == FORM_TYPE && !c.Morphed {
						c.Morphed = true
						r := FORM_ACTIVATED
						r.Insert(utils.IntToBytes(uint64(item.NPCID), 4, true), 5) // form npc id
						c.Socket.Write(r)
					}
				}
			}
		}
	}
}

func (c *Character) makeCharacterMorphed(npcID uint64, activateState bool) []byte {

	resp := FORM_ACTIVATED
	resp.Insert(utils.IntToBytes(uint64(npcID), 4, true), 5) // form npc id
	c.Socket.Write(resp)

	return resp
}

func (c *Character) RespawnCounter(seconds byte) {

	resp := RESPAWN_COUNTER
	resp[7] = seconds
	c.Socket.Write(resp)

	if seconds > 0 {
		time.AfterFunc(time.Second, func() {
			c.RespawnCounter(seconds - 1)
		})
	}
}

func (c *Character) Teleport(coordinate *utils.Location) []byte {

	c.SetCoordinate(coordinate)

	resp := TELEPORT_PLAYER
	resp.Insert(utils.FloatToBytes(coordinate.X, 4, true), 5) // coordinate-x
	resp.Insert(utils.FloatToBytes(coordinate.Y, 4, true), 9) // coordinate-x

	return resp
}

func (c *Character) ActivityStatus(remainingTime int) {

	var msg string
	if c.IsActive || remainingTime == 0 {
		msg = "Your character has been activated."
		c.IsActive = true

		data, err := c.SpawnCharacter()
		if err != nil {
			return
		}

		p := &nats.CastPacket{CastNear: true, CharacterID: c.ID, Data: data, Type: nats.PLAYER_SPAWN}
		if err = p.Cast(); err != nil {
			return
		}

	} else {
		msg = fmt.Sprintf("Your character will be activated %d seconds later.", remainingTime)

		if c.IsOnline {
			time.AfterFunc(time.Second, func() {
				if !c.IsActive {
					c.ActivityStatus(remainingTime - 1)
				}
			})
		}
	}

	info := messaging.InfoMessage(msg)
	if c.Socket != nil {
		c.Socket.Write(info)
	}
}

func contains(v int64, a []int64) bool {
	for _, i := range a {
		if i == v {
			return true
		}
	}
	return false
}

func unorderedEqual(first, second []int64, count int) bool {
	exists := make(map[int64]bool)
	match := 0
	for _, value := range first {
		exists[value] = true
	}
	for _, value := range second {
		if match >= count {
			return true
		}
		if !exists[value] {
			return false
		}
		match++
	}
	return true
}

func BonusActive(first, second []int64) bool {
	exists := make(map[int64]bool)
	for _, value := range first {
		exists[value] = true
	}
	for _, value := range second {
		if !exists[value] {
			return false
		}
	}
	return true
}

func (c *Character) ItemSetEffects(indexes []int16) []int64 {
	//	log.Printf("Bejön ide")
	slots, _ := c.InventorySlots()
	playerItems := []int64{}
	for _, i := range indexes {
		if (i == 3 && c.WeaponSlot == 4) || (i == 4 && c.WeaponSlot == 3) {
			continue
		}
		item := slots[i]
		playerItems = append(playerItems, item.ItemID)
	}
	setEffect := []int64{}
	for _, i := range playerItems {
		for _, sets := range ItemSets {
			if contains(i, sets.GetSetItems()) {
				if unorderedEqual(playerItems, sets.GetSetItems(), sets.SetItemCount) {
					buffEffects := sets.GetSetBonus()
					for _, effect := range buffEffects {
						if effect == 0 || effect == 616 || effect == 617 {
							continue
						}
						if !contains(effect, setEffect) {
							setEffect = append(setEffect, effect)
						}
					}
				}
			}
		}
	}
	return setEffect
}
func (c *Character) applySetEffect(bonuses []int64, st *Stat) {
	additionalDropMultiplier, additionalExpMultiplier, additionalRunningSpeed := float64(0), float64(0), float64(0)
	for _, id := range bonuses {
		item := Items[id]
		if item == nil {
			continue
		}

		st.STRBuff += item.STR
		st.DEXBuff += item.DEX
		st.INTBuff += item.INT
		st.WindBuff += item.Wind
		st.WaterBuff += item.Water
		st.FireBuff += item.Fire

		st.DEF += item.Def + ((item.BaseDef1 + item.BaseDef2 + item.BaseDef3) / 3)
		st.DefRate += item.DefRate

		st.ArtsDEF += item.ArtsDef
		st.ArtsDEFRate += item.ArtsDefRate

		st.MaxHP += item.MaxHp
		st.MaxCHI += item.MaxChi

		st.Accuracy += item.Accuracy
		st.Dodge += item.Dodge

		st.MinATK += item.BaseMinAtk + item.MinAtk
		st.MaxATK += item.BaseMaxAtk + item.MaxAtk
		st.ATKRate += item.AtkRate

		st.MinArtsATK += item.MinArtsAtk
		st.MaxArtsATK += item.MaxArtsAtk
		st.ArtsATKRate += item.ArtsAtkRate
		additionalExpMultiplier += item.ExpRate
		additionalDropMultiplier += item.DropRate
		additionalRunningSpeed += item.RunningSpeed

		st.ConfusionATK += item.ConfusionATK
		st.ConfusionDEF += item.ConfusionDEF
		st.PoisonATK += item.PoisonATK
		st.PoisonDEF += item.PoisonDEF
		st.ConfusionATK += item.ConfusionATK
		st.ConfusionDEF += item.ConfusionDEF
	}
	c.AdditionalExpMultiplier += additionalExpMultiplier
	c.AdditionalDropMultiplier += additionalDropMultiplier
	c.AdditionalRunningSpeed += additionalRunningSpeed
}
func (c *Character) NewBuffInfection(infectionID int64, seconds int) {

	character := c
	infection := BuffInfections[int(infectionID)]
	duration := 0
	plus := 1

	buff, err := FindBuffByID(int(infectionID), character.ID)

	if buff != nil && err == nil {
		id := buff.ID
		buff = &Buff{ID: id, CharacterID: character.ID, StartedAt: character.Epoch, Duration: int64(duration), Name: infection.Name,
			ATK: infection.BaseATK + infection.AdditionalATK*int(plus), ArtsATK: infection.BaseArtsATK + infection.AdditionalArtsATK*int(plus),
			ArtsDEF: infection.ArtsDEF + infection.AdditionalArtsDEF*int(plus), ConfusionDEF: infection.ConfusionDef,
			DEF: infection.BaseDef + infection.AdditionalDEF*int(plus), DEX: infection.DEX, HPRecoveryRate: infection.HPRecoveryRate, INT: infection.INT,
			MaxHP: infection.MaxHP, ParalysisDEF: infection.ParalysisDef, PoisonDEF: infection.ParalysisDef, STR: infection.STR, Accuracy: infection.Accuracy * int(plus), Dodge: infection.DodgeRate * int(plus), CanExpire: false}
		buff.Update()

	} else if buff == nil && err == nil {
		buff = &Buff{ID: int(infectionID), CharacterID: character.ID, StartedAt: character.Epoch, Duration: int64(duration), Name: infection.Name,
			ATK: infection.BaseATK + infection.AdditionalATK*int(plus), ArtsATK: infection.BaseArtsATK + infection.AdditionalArtsATK*int(plus),
			ArtsDEF: infection.ArtsDEF + infection.AdditionalArtsDEF*int(plus), ConfusionDEF: infection.ConfusionDef,
			DEF: infection.BaseDef + infection.AdditionalDEF*int(plus), DEX: infection.DEX, HPRecoveryRate: infection.HPRecoveryRate, INT: infection.INT,
			MaxHP: infection.MaxHP, ParalysisDEF: infection.ParalysisDef, PoisonDEF: infection.ParalysisDef, STR: infection.STR, Accuracy: infection.Accuracy * int(plus), Dodge: infection.DodgeRate * int(plus), CanExpire: false}
		err := buff.Create()
		if err != nil {
			fmt.Println("BUFF ADD ERROR: ", err)
			//	return nil, err
		}
		data, _ := c.GetStats()
		c.Socket.Write(data)
	}
}
func (c *Character) applyJudgementEffect(inventoryItem *InventorySlot, st *Stat) {
	bonusID := inventoryItem.JudgementStat
	item := ItemJudgements[int(bonusID)]
	itemInfo := Items[inventoryItem.ItemID]
	if itemInfo.STR > 0 {
		st.STRBuff += item.StrPlus
	}
	if itemInfo.DEX > 0 {
		st.DEXBuff += item.DexPlus
	}
	if itemInfo.INT > 0 {
		st.INTBuff += item.IntPlus
	}
	if itemInfo.Wind > 0 {
		st.WindBuff += item.WindPlus
	}
	if itemInfo.Water > 0 {
		st.WaterBuff += item.WaterPlus
	}
	if itemInfo.Fire > 0 {
		st.FireBuff += item.FirePlus
	}

	if itemInfo.Def > 0 {
		st.DEF += item.ExtraDef
	}
	if itemInfo.ArtsDef > 0 {
		st.ArtsDEF += item.ExtraArtsDef
	}
	if itemInfo.MaxHp > 0 {
		st.MaxHP += item.MaxHP
	}
	if itemInfo.MaxChi > 0 {
		st.MaxCHI += item.MaxChi
	}
	if itemInfo.Accuracy > 0 {
		st.Accuracy += item.AccuracyPlus
	}
	if itemInfo.Dodge > 0 {
		st.Dodge += item.ExtraDodge
	}
	if itemInfo.MinAtk > 0 {
		st.MinATK += item.AttackPlus
		st.MaxATK += item.AttackPlus
	}
	//st.ATKRate += item.AtkRate

	//st.MinArtsATK += item.MinArtsAtk
	//st.MaxArtsATK += item.MaxArtsAtk
	//st.ArtsATKRate += item.ArtsAtkRate
}
func (c *Character) ItemEffects(st *Stat, start, end int16) error {

	slots, err := c.InventorySlots()
	if err != nil {
		return err
	}

	indexes := []int16{}

	for i := start; i <= end; i++ {
		slot := slots[i]
		if start == 0x0B || start == 0x155 {
			if slot != nil && slot.Activated && slot.InUse {
				indexes = append(indexes, i)
			}
		} else {
			indexes = append(indexes, i)
		}
	}
	// Handle Items Equpped in wrong slot
	if start == 0 && end == 9 || start == 307 && end == 315 || start == 397 && end == 399 || start == 400 && end == 4001 {
		for i := start; i <= end; i++ {
			sl := slots[i]
			item := Items[sl.ItemID]

			if item != nil && int16(item.Slot) != i && sl.SlotID != 0 && item.Slot != 3 {

				f, err := os.OpenFile("AntiSlotExploit/"+c.Name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) //log file
				if err != nil {
					log.Fatalf("error opening file: %v", err)
				}
				fi, err := os.OpenFile("Log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) //log file
				if err != nil {
					log.Fatalf("error opening file: %v", err)
				}
				defer f.Close()
				log.SetOutput(f)
				log.Print(sl)

				//delete item
				data := c.DecrementItem(i, sl.Quantity)
				c.Socket.Write(*data)
				c.Update()
				sl.Update()
				log.SetOutput(fi)
			}
		}
	}
	// Handle Items Equpped in wrong slot
	start2 := start
	start3 := start
	start2 = 317
	start3 = 319
	for i := start2; i <= start3; i++ {
		sl := slots[i]
		item := Items[sl.ItemID]

		if item != nil && int16(item.Slot) != i && sl.SlotID != 0 && item.Slot != 3 {

			f, err := os.OpenFile("AntiSlotExploit/"+c.Name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) //log file
			if err != nil {
				log.Fatalf("error opening file: %v", err)
			}
			fi, err := os.OpenFile("Log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666) //log file
			if err != nil {
				log.Fatalf("error opening file: %v", err)
			}
			defer f.Close()
			log.SetOutput(f)
			log.Print(sl)

			//delete item
			data := c.DecrementItem(i, sl.Quantity)
			c.Socket.Write(*data)
			c.Update()
			sl.Update()
			log.SetOutput(fi)
		}
	}

	additionalDropMultiplier, additionalExpMultiplier, additionalRunningSpeed := float64(0), float64(0), float64(0)
	setEffects := c.ItemSetEffects(indexes)
	c.applySetEffect(setEffects, st)
	for _, i := range indexes {
		if (i == 3 && c.WeaponSlot == 4) || (i == 4 && c.WeaponSlot == 3) {
			continue
		}

		item := slots[i]

		if item.ItemID != 0 {

			info := Items[item.ItemID]
			slotId := i
			if slotId == 4 {
				slotId = 3
			}

			if (info == nil || slotId != int16(info.Slot) || c.Level < info.MinLevel || (info.MaxLevel > 0 && c.Level > info.MaxLevel)) &&
				!(start == 0x0B || start == 0x155) {
				continue
			}
			if Items[item.ItemID].ItemBuff != 0 { // item buff (aura)
				//if c.HTVisibility&0x04 != 0 {
				c.NewBuffInfection(int64(Items[item.ItemID].ItemBuff), 2)
				//}

				/*time.AfterFunc(time.Second*1, func() {
					c.ItemEffects(st, start, end)
				})*/
			}

			ids := []int64{item.ItemID}
			if item.ItemType != 0 {
				c.applyJudgementEffect(item, st)
			}
			for _, u := range item.GetUpgrades() {
				if u == 0 {
					break
				}
				ids = append(ids, int64(u))
			}

			for _, s := range item.GetSockets() {
				if s == 0 {
					break
				}
				ids = append(ids, int64(s))
			}
			for _, id := range ids {
				item := Items[id]
				if item == nil {
					continue
				}

				st.STRBuff += item.STR
				st.DEXBuff += item.DEX
				st.INTBuff += item.INT
				st.WindBuff += item.Wind
				st.WaterBuff += item.Water
				st.FireBuff += item.Fire

				st.DEF += item.Def + ((item.BaseDef1 + item.BaseDef2 + item.BaseDef3) / 3)
				st.DefRate += item.DefRate

				st.ArtsDEF += item.ArtsDef
				st.ArtsDEFRate += item.ArtsDefRate

				st.MaxHP += item.MaxHp
				st.MaxCHI += item.MaxChi

				st.Accuracy += item.Accuracy
				st.Dodge += item.Dodge

				st.MinATK += item.BaseMinAtk + item.MinAtk
				st.MaxATK += item.BaseMaxAtk + item.MaxAtk
				st.ATKRate += item.AtkRate

				st.MinArtsATK += item.MinArtsAtk
				st.MaxArtsATK += item.MaxArtsAtk
				st.ArtsATKRate += item.ArtsAtkRate
				additionalExpMultiplier += item.ExpRate
				additionalDropMultiplier += item.DropRate
				additionalRunningSpeed += item.RunningSpeed

				st.ConfusionATK += item.ConfusionATK
				st.ConfusionDEF += item.ConfusionDEF
				st.PoisonATK += item.PoisonATK
				st.PoisonDEF += item.PoisonDEF
				st.ConfusionATK += item.ConfusionATK
				st.ConfusionDEF += item.ConfusionDEF
			}
		}
	}

	c.AdditionalExpMultiplier += additionalExpMultiplier
	c.AdditionalDropMultiplier += additionalDropMultiplier
	c.AdditionalRunningSpeed += additionalRunningSpeed
	return nil
}

func (c *Character) GetExpAndSkillPts() []byte {

	resp := EXP_SKILL_PT_CHANGED
	resp.Insert(utils.IntToBytes(uint64(c.Exp), 8, true), 5)                        // character exp
	resp.Insert(utils.IntToBytes(uint64(c.Socket.Skills.SkillPoints), 4, true), 13) // character skill points
	return resp
}

func (c *Character) GetPTS() []byte {

	resp := PTS_CHANGED
	resp.Insert(utils.IntToBytes(uint64(c.PTS), 4, true), 6) // character pts
	return resp
}

func (c *Character) LootGold(amount uint64) []byte {

	c.AddingGold.Lock()
	defer c.AddingGold.Unlock()

	c.Gold += amount
	resp := GOLD_LOOTED
	resp.Insert(utils.IntToBytes(uint64(c.Gold), 8, true), 9) // character gold

	return resp
}

func (c *Character) AddExp(amount int64) ([]byte, bool) {

	c.AddingExp.Lock()
	defer c.AddingExp.Unlock()
	c.Socket.Skills.SkillPoints = 300000
	expMultipler := c.ExpMultiplier + c.AdditionalExpMultiplier
	exp := c.Exp + int64(float64(amount)*(expMultipler*EXP_RATE))
	spIndex := utils.SearchUInt64(SkillPoints, uint64(c.Exp))
	canLevelUp := true
	if exp > 233332051410 && c.Level <= 100 {
		exp = 233332051410
	}
	if exp > 544951059310 && c.Level <= 200 {
		exp = 544951059310
	}
	c.Exp = exp
	spIndex2 := utils.SearchUInt64(SkillPoints, uint64(c.Exp))

	//resp := c.GetExpAndSkillPts()

	st := c.Socket.Stats
	if st == nil {
		return nil, false
	}

	levelUp := false
	level := int16(c.Level)
	targetExp := EXPs[level].Exp
	skPts, sp := 3000, 0
	np := 0                                             //nature pts
	for exp >= targetExp && level < 300 && canLevelUp { // Levelling up && level < 100

		if c.Type <= 59 && level >= 100 {
			level = 100
			canLevelUp = false
		} else if c.Type <= 69 && level >= 200 {
			level = 200
			canLevelUp = false
		} else {
			level++
			st.HP = st.MaxHP
			if EXPs[level] != nil {
				sp += EXPs[level].StatPoints
				np += EXPs[level].NaturePoints
			}
			targetExp = EXPs[level].Exp
			levelUp = true
		}

	}
	c.Level = int(level)
	resp := EXP_SKILL_PT_CHANGED

	if level < 101 {
		skPts = spIndex2 - spIndex
		c.Socket.Skills.SkillPoints += skPts
	}
	//skPts = spIndex2 - spIndex

	if levelUp {
		if level > 100 {
			skPts = EXPs[level].SkillPoints
			c.Socket.Skills.SkillPoints = 1000
		}

		st.StatPoints += sp
		st.NaturePoints += np
		resp.Insert(utils.IntToBytes(uint64(exp), 8, true), 5)                          // character exp
		resp.Insert(utils.IntToBytes(uint64(c.Socket.Skills.SkillPoints), 4, true), 13) // character skill points
		if c.GuildID > 0 {
			guild, err := FindGuildByID(c.GuildID)
			if err == nil && guild != nil {
				guild.InformMembers(c)
			}
		}

		resp.Concat(messaging.SystemMessage(messaging.LEVEL_UP))
		resp.Concat(messaging.SystemMessage(messaging.LEVEL_UP_SP))
		resp.Concat(messaging.InfoMessage(c.GetLevelText()))

		spawnData, err := c.SpawnCharacter()
		if err == nil {
			c.Socket.Write(spawnData)
			c.Update()
			//resp.Concat(spawnData)
		}
	} else {
		resp.Insert(utils.IntToBytes(uint64(exp), 8, true), 5)                          // character exp
		resp.Insert(utils.IntToBytes(uint64(c.Socket.Skills.SkillPoints), 4, true), 13) // character skill points
	}
	go c.Socket.Skills.Update()
	return resp, levelUp
}
func (c *Character) CombineItems(where, to int16) (int64, int16, error) {

	invSlots, err := c.InventorySlots()
	if err != nil {
		return 0, 0, err
	}

	c.InvMutex.Lock()
	defer c.InvMutex.Unlock()

	whereItem := invSlots[where]
	toItem := invSlots[to]

	if toItem.ItemID == whereItem.ItemID {
		toItem.Quantity += whereItem.Quantity

		go toItem.Update()
		whereItem.Delete()
		*whereItem = *NewSlot()

	} else {
		return 0, 0, nil
	}

	return toItem.ItemID, int16(toItem.Quantity), nil
}

func (c *Character) BankItems() []byte {

	bankSlots, err := c.InventorySlots()
	if err != nil {
		return nil
	}

	bankSlots = bankSlots[0x43:0x133]
	resp := BANK_ITEMS

	index, length := 8, int16(4)
	for i, slot := range bankSlots {
		if slot.ItemID == 0 {
			continue
		}

		resp.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), index) // item id
		index += 4

		resp.Insert([]byte{0x00, 0xA1, 0x01, 0x00}, index)
		index += 4

		test := 67 + i
		resp.Insert(utils.IntToBytes(uint64(test), 2, true), index) // slot id
		index += 2

		resp.Insert([]byte{0x00, 0x00, 0x00, 0x00}, index)
		index += 4
		length += 14
	}

	resp.SetLength(length)
	return resp
}

func (c *Character) GetGold() []byte {

	user, err := FindUserByID(c.UserID)
	if err != nil || user == nil {
		return nil
	}

	resp := GET_GOLD
	resp.Insert(utils.IntToBytes(uint64(c.Gold), 8, true), 6)         // gold
	resp.Insert(utils.IntToBytes(uint64(user.BankGold), 8, true), 14) // bank gold

	return resp
}

func (c *Character) ChangeMap(mapID int16, coordinate *utils.Location, args ...interface{}) ([]byte, error) {

	if !funk.Contains(unlockedMaps, mapID) {
		return nil, nil
	}

	resp, r := MAP_CHANGED, utils.Packet{}
	c.Map = mapID
	c.EndPvP()
	if !c.IsinWar {
		if coordinate == nil { // if no coordinate then teleport home
			d := SavePoints[uint8(mapID)]
			if d == nil {
				d = &SavePoint{Point: "(100.0,100.0)"}
			}
			coordinate = ConvertPointToLocation(d.Point)
		}
	} else {
		if c.Faction == 1 && c.Map != 230 {
			delete(OrderCharacters, c.ID)
			c.IsinWar = false
		} else if c.Faction == 2 && c.Map != 230 {
			delete(ShaoCharacters, c.ID)
			c.IsinWar = false
		}
	}
	if coordinate == nil { // if no coordinate then teleport home
		d := SavePoints[uint8(mapID)]
		if d == nil {
			d = &SavePoint{Point: "(100.0,100.0)"}
		}
		coordinate = ConvertPointToLocation(d.Point)
	}
	servers, _ := GetServers()
	if funk.Contains(sharedMaps, mapID) && !servers[int16(c.Socket.User.ConnectedServer)-1].IsPVPServer { // shared map
		c.Socket.User.ConnectedServer = 1
	}

	if c.GuildID > 0 {
		guild, err := FindGuildByID(c.GuildID)
		if err == nil && guild != nil {
			guild.InformMembers(c)
		}
	}

	consItems, _ := FindConsignmentItemsBySellerID(c.ID)
	consItems = (funk.Filter(consItems, func(item *ConsignmentItem) bool {
		return item.IsSold
	}).([]*ConsignmentItem))
	if len(consItems) > 0 {
		r.Concat(CONSIGMENT_ITEM_SOLD)
	}

	slots, err := c.InventorySlots()
	if err == nil {
		pet := slots[0x0A].Pet
		if pet != nil && pet.IsOnline {
			pet.IsOnline = false
			r.Concat(DISMISS_PET)
		}
	}

	if c.AidMode {
		c.AidMode = false
		r.Concat(c.AidStatus())
	}

	RemovePetFromRegister(c)
	//RemoveFromRegister(c)
	//GenerateID(c)

	c.SetCoordinate(coordinate)

	if len(args) == 0 { // not logging in
		c.OnSight.DropsMutex.Lock()
		c.OnSight.Drops = map[int]interface{}{}
		c.OnSight.DropsMutex.Unlock()

		c.OnSight.MobMutex.Lock()
		c.OnSight.Mobs = map[int]interface{}{}
		c.OnSight.MobMutex.Unlock()

		c.OnSight.NpcMutex.Lock()
		c.OnSight.NPCs = map[int]interface{}{}
		c.OnSight.NpcMutex.Unlock()

		c.OnSight.PetsMutex.Lock()
		c.OnSight.Pets = map[int]interface{}{}
		c.OnSight.PetsMutex.Unlock()
	}

	resp[13] = byte(mapID)                                     // map id
	resp.Insert(utils.FloatToBytes(coordinate.X, 4, true), 14) // coordinate-x
	resp.Insert(utils.FloatToBytes(coordinate.Y, 4, true), 18) // coordinate-y
	resp[36] = byte(mapID)                                     // map id
	resp.Insert(utils.FloatToBytes(coordinate.X, 4, true), 46) // coordinate-x
	resp.Insert(utils.FloatToBytes(coordinate.Y, 4, true), 50) // coordinate-y
	resp[61] = byte(mapID)                                     // map id

	spawnData, _ := c.SpawnCharacter()
	r.Concat(spawnData)
	resp.Concat(r)
	resp.Concat(c.Socket.User.GetTime())
	return resp, nil
}
func (c *Character) LosePlayerExp(percent int) (int64, error) {
	level := int16(c.Level)
	expminus := int64(0)
	if level >= 10 {
		oldExp := EXPs[level-1].Exp
		resp := EXP_SKILL_PT_CHANGED
		if oldExp <= c.Exp {
			per := float64(percent) / 100
			expLose := float64(c.Exp) * float64(1-per)
			if int64(expLose) >= oldExp {
				exp := c.Exp - int64(expLose)
				expminus = int64(float64(exp) * float64(1-0.30))
				c.Exp = int64(expLose)
			} else {
				exp := c.Exp - oldExp
				expminus = int64(float64(exp) * float64(1-0.30))
				c.Exp = oldExp
			}
		}
		resp.Insert(utils.IntToBytes(uint64(c.Exp), 8, true), 5)                        // character exp
		resp.Insert(utils.IntToBytes(uint64(c.Socket.Skills.SkillPoints), 4, true), 13) // character skill points
		go c.Socket.Skills.Update()
		c.Socket.Write(resp)
	}
	return expminus, nil
}
func DoesSlotAffectStats(slotNo int16) bool {
	return slotNo < 0x0B || (slotNo >= 0x0133 && slotNo <= 0x013B) || (slotNo >= 0x18D && slotNo <= 0x192)
}

func (c *Character) RemoveItem(slotID int16) ([]byte, error) {

	slots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	item := slots[slotID]

	resp := ITEM_REMOVED
	resp.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 9) // item id
	resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 13)     // slot id

	affects, activated := DoesSlotAffectStats(slotID), item.Activated
	if affects || activated {
		item.Activated = false
		item.InUse = false

		statData, err := c.GetStats()
		if err != nil {
			return nil, err
		}

		resp.Concat(statData)
	}

	info := Items[item.ItemID]
	if activated {
		if item.ItemID == 100080008 { // eyeball of divine
			c.DetectionMode = false
		}

		if info != nil && info.GetType() == FORM_TYPE {
			c.Morphed = false
			resp.Concat(FORM_DEACTIVATED)
		}

		data := ITEM_EXPIRED
		data.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 6)
		resp.Concat(data)
	}

	if affects {
		itemsData, err := c.ShowItems()
		if err != nil {
			return nil, err
		}

		p := nats.CastPacket{CastNear: true, CharacterID: c.ID, Data: itemsData, Type: nats.SHOW_ITEMS}
		if err = p.Cast(); err != nil {
			return nil, err
		}

		resp.Concat(itemsData)
	}

	consItem, _ := FindConsignmentItemByID(item.ID)
	if consItem == nil { // item is not in consigment table
		err = item.Delete()
		if err != nil {
			return nil, err
		}

	} else { // seller did not claim the consigment item
		newItem := NewSlot()
		*newItem = *item
		newItem.UserID = null.StringFromPtr(nil)
		newItem.CharacterID = null.IntFromPtr(nil)
		newItem.Update()
		InventoryItems.Add(consItem.ID, newItem)
	}

	*item = *NewSlot()
	return resp, nil
}

func (c *Character) SellItem(itemID, slot, quantity int, unitPrice uint64) ([]byte, error) {

	c.LootGold(unitPrice * uint64(quantity))
	_, err := c.RemoveItem(int16(slot))
	if err != nil {
		return nil, err
	}

	resp := SELL_ITEM
	resp.Insert(utils.IntToBytes(uint64(itemID), 4, true), 8)  // item id
	resp.Insert(utils.IntToBytes(uint64(slot), 2, true), 12)   // slot id
	resp.Insert(utils.IntToBytes(uint64(c.Gold), 8, true), 14) // character gold

	return resp, nil
}

func (c *Character) GetStats() ([]byte, error) {

	if c == nil {
		log.Println("c is nil")
		return nil, nil

	} else if c.Socket == nil {
		log.Println("socket is nil")
		return nil, nil
	}

	st := c.Socket.Stats
	if st == nil {
		return nil, nil
	}

	err := st.Calculate()
	if err != nil {
		return nil, err
	}

	resp := GET_STATS

	index := 5
	resp.Insert(utils.IntToBytes(uint64(c.Level), 4, true), index) // character level
	index += 4

	duelState := 0
	if c.DuelID > 0 && c.DuelStarted {
		duelState = 500
	}

	resp.Insert(utils.IntToBytes(uint64(duelState), 2, true), index) // duel state
	index += 4

	resp.Insert(utils.IntToBytes(uint64(st.StatPoints), 2, true), index) // stat points
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.NaturePoints), 2, true), index) // divine stat points
	index += 2

	resp.Insert(utils.IntToBytes(uint64(c.Socket.Skills.SkillPoints), 4, true), index) // character skill points
	index += 6

	resp.Insert(utils.IntToBytes(uint64(c.Exp), 8, true), index) // character experience
	index += 8

	resp.Insert(utils.IntToBytes(uint64(c.AidTime), 4, true), index) // remaining aid
	index += 4
	index++

	targetExp := EXPs[int16(c.Level)].Exp
	resp.Insert(utils.IntToBytes(uint64(targetExp), 8, true), index) // character target experience
	index += 8

	resp.Insert(utils.IntToBytes(uint64(st.STR), 2, true), index) // character str
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.STR+st.STRBuff), 2, true), index) // character str buff
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.DEX), 2, true), index) // character dex
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.DEX+st.DEXBuff), 2, true), index) // character dex buff
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.INT), 2, true), index) // character int
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.INT+st.INTBuff), 2, true), index) // character int buff
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.Wind), 2, true), index) // character wind
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.Wind+st.WindBuff), 2, true), index) // character wind buff
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.Water), 2, true), index) // character water
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.Water+st.WaterBuff), 2, true), index) // character water buff
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.Fire), 2, true), index) // character fire
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.Fire+st.FireBuff), 2, true), index) // character fire buff
	index += 7

	resp.Insert(utils.FloatToBytes(c.RunningSpeed+c.AdditionalRunningSpeed, 4, true), index) // character running speed
	index += 10

	resp.Insert(utils.IntToBytes(uint64(st.MaxHP), 4, true), index) // character max hp
	index += 4

	resp.Insert(utils.IntToBytes(uint64(st.MaxCHI), 4, true), index) // character max chi
	index += 4

	resp.Insert(utils.IntToBytes(uint64(st.MinATK), 2, true), index) // character min atk
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.MaxATK), 2, true), index) // character max atk
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.DEF), 4, true), index) // character def
	index += 4
	resp.Insert(utils.IntToBytes(uint64(st.DEF), 4, true), index) // character def
	index += 4
	resp.Insert(utils.IntToBytes(uint64(st.DEF), 4, true), index) // character def
	index += 4

	resp.Insert(utils.IntToBytes(uint64(st.MinArtsATK), 4, true), index) // character min arts atk
	index += 4

	resp.Insert(utils.IntToBytes(uint64(st.MaxArtsATK), 4, true), index) // character max arts atk
	index += 4

	resp.Insert(utils.IntToBytes(uint64(st.ArtsDEF), 4, true), index) // character arts def
	index += 4

	resp.Insert(utils.IntToBytes(uint64(st.Accuracy), 2, true), index) // character accuracy
	index += 2

	resp.Insert(utils.IntToBytes(uint64(st.Dodge), 2, true), index) // character dodge
	index += 2

	resp.Concat(c.GetHPandChi()) // hp and chi

	return resp, nil
}

func (c *Character) BSUpgrade(slotID int64, stones []*InventorySlot, luck, protection *InventorySlot, stoneSlots []int64, luckSlot, protectionSlot int64) ([]byte, error) {

	slots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	item := slots[slotID]
	if item.Plus >= 15 { // cannot be upgraded more
		resp := utils.Packet{0xAA, 0x55, 0x31, 0x00, 0x54, 0x02, 0xA6, 0x0F, 0x01, 0x00, 0xA3, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x55, 0xAA}
		resp.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 9) // item id
		resp.Insert(utils.IntToBytes(uint64(slotID), 2, true), 17)     // slot id
		resp.Insert(item.GetUpgrades(), 19)                            // item upgrades
		resp[34] = byte(item.SocketCount)                              // socket count
		resp.Insert(item.GetSockets(), 35)                             // item sockets
		c := 35 + 15
		if item.ItemType != 0 {
			resp.Overwrite(utils.IntToBytes(uint64(item.ItemType), 1, true), c-6)
			if item.ItemType == 2 {
				resp.Overwrite(utils.IntToBytes(uint64(item.JudgementStat), 4, true), c-5)
			}
		}

		return resp, nil
	}

	info := Items[item.ItemID]
	cost := (info.BuyPrice / 10) * int64(item.Plus+1) * int64(math.Pow(2, float64(len(stones)-1)))

	if uint64(cost) > c.Gold {
		resp := messaging.SystemMessage(messaging.INSUFFICIENT_GOLD)
		return resp, nil

	} else if len(stones) == 0 {
		resp := messaging.SystemMessage(messaging.INCORRECT_GEM_QTY)
		return resp, nil
	}

	stone := stones[0]
	stoneInfo := Items[stone.ItemID]

	if int16(item.Plus) < stoneInfo.MinUpgradeLevel || stoneInfo.ID > 255 {
		resp := messaging.SystemMessage(messaging.INCORRECT_GEM)
		return resp, nil
	}

	itemType := info.GetType()
	typeMatch := (stoneInfo.Type == 190 && itemType == PET_ITEM_TYPE) || (stoneInfo.Type == 191 && itemType == HT_ARMOR_TYPE) ||
		(stoneInfo.Type == 192 && (itemType == ACC_TYPE || itemType == MASTER_HT_ACC)) && item.ItemType == 0 || (stoneInfo.Type == 194 && itemType == WEAPON_TYPE && item.ItemType == 0) || (stoneInfo.Type == 195 && itemType == ARMOR_TYPE && item.ItemType == 0) ||
		//DISC ITEMS
		(stoneInfo.Type == 229 && stoneInfo.HtType == 36 && itemType == WEAPON_TYPE && item.ItemType == 2) || (stoneInfo.Type == 229 && stoneInfo.HtType == 37 && itemType == ARMOR_TYPE && item.ItemType == 2) || (stoneInfo.Type == 229 && stoneInfo.HtType == 38 && (itemType == ACC_TYPE || itemType == MASTER_HT_ACC) && item.ItemType == 2)

	if !typeMatch {

		resp := utils.Packet{0xAA, 0x55, 0x09, 0x00, 0x54, 0x02, 0xA4, 0x0F, 0x00, 0x55, 0xAA}
		resp.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 9) // item id
		return resp, nil
	}

	rate := float64(STRRates[item.Plus] * len(stones))
	plus := item.Plus + 1

	if stone.Plus > 0 { // Precious Pendent or Ghost Dagger or Dragon Scale
		for i := 0; i < len(stones); i++ {
			for j := i; j < len(stones); j++ {
				if stones[i].Plus != stones[j].Plus { // mismatch stone plus
					resp := utils.Packet{0xAA, 0x55, 0x09, 0x00, 0x54, 0x02, 0xA4, 0x0F, 0x00, 0x55, 0xAA}
					resp.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 9) // item id
					return resp, nil
				}
			}
		}

		plus = item.Plus + stone.Plus
		if plus > 15 {
			plus = 15
		}

		rate = float64(STRRates[plus-1] * len(stones))
	}

	if luck != nil {
		luckInfo := Items[luck.ItemID]
		if luckInfo.Type == 164 { // charm of luck
			k := float64(luckInfo.SellPrice) / 100
			rate += rate * k / float64(len(stones))

		} else if luckInfo.Type == 219 { // bagua
			if byte(luckInfo.SellPrice) != item.Plus { // item plus not matching with bagua
				resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x54, 0x02, 0xB6, 0x0F, 0x55, 0xAA}
				return resp, nil

			} else if len(stones) < 3 {
				resp := messaging.SystemMessage(messaging.INCORRECT_GEM_QTY)
				return resp, nil
			}

			rate = 1000
			bagRates := []int{luckInfo.HolyWaterUpg3, luckInfo.HolyWaterRate1, luckInfo.HolyWaterRate2, luckInfo.HolyWaterRate3}
			seed := utils.RandInt(0, 100)

			for i := 0; i < len(bagRates); i++ {
				if int(seed) > bagRates[i] {
					plus++
				}
			}
		}
	}

	protectionInfo := &Item{}
	if protection != nil {
		protectionInfo = Items[protection.ItemID]
	}

	resp := utils.Packet{}
	c.LootGold(-uint64(cost))
	resp.Concat(c.GetGold())

	seed := int(utils.RandInt(0, 1000))
	if float64(seed) < rate { // upgrade successful
		var codes []byte
		for i := item.Plus; i < plus; i++ {
			codes = append(codes, byte(stone.ItemID))
		}

		before := item.GetUpgrades()
		resp.Concat(item.Upgrade(int16(slotID), codes...))
		logger.Log(logging.ACTION_UPGRADE_ITEM, c.ID, fmt.Sprintf("Item (%d) upgraded: %s -> %s", item.ID, before, item.GetUpgrades()), c.UserID)

	} else if itemType == HT_ARMOR_TYPE || itemType == PET_ITEM_TYPE ||
		(protection != nil && protectionInfo.GetType() == SCALE_TYPE) { // ht or pet item failed or got protection

		if protectionInfo.GetType() == SCALE_TYPE { // if scale
			if item.Plus < uint8(protectionInfo.SellPrice) {
				item.Plus = 0
			} else {
				item.Plus -= uint8(protectionInfo.SellPrice)
			}
		} else {
			if item.Plus < stone.Plus {
				item.Plus = 0
			} else {
				item.Plus -= stone.Plus
			}
		}

		upgs := item.GetUpgrades()
		for i := int(item.Plus); i < len(upgs); i++ {
			item.SetUpgrade(i, 0)
		}

		r := HT_UPG_FAILED
		r.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 9) // item id
		r.Insert(utils.IntToBytes(uint64(slotID), 2, true), 17)     // slot id
		r.Insert(item.GetUpgrades(), 19)                            // item upgrades
		r[34] = byte(item.SocketCount)                              // socket count
		r.Insert(item.GetSockets(), 35)                             // item sockets

		resp.Concat(r)
		logger.Log(logging.ACTION_UPGRADE_ITEM, c.ID, fmt.Sprintf("Item (%d) upgrade failed but not vanished", item.ID), c.UserID)

	} else { // casual item failed so destroy it
		r := UPG_FAILED
		r.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 9) // item id
		resp.Concat(r)

		itemsData, err := c.RemoveItem(int16(slotID))
		if err != nil {
			return nil, err
		}

		resp.Concat(itemsData)
		logger.Log(logging.ACTION_UPGRADE_ITEM, c.ID, fmt.Sprintf("Item (%d) upgrade failed and destroyed", item.ID), c.UserID)
	}

	for _, slot := range stoneSlots {
		resp.Concat(*c.DecrementItem(int16(slot), 1))
	}

	if luck != nil {
		resp.Concat(*c.DecrementItem(int16(luckSlot), 1))
	}

	if protection != nil {
		resp.Concat(*c.DecrementItem(int16(protectionSlot), 1))
	}

	err = item.Update()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Character) BSProduction(book *InventorySlot, materials []*InventorySlot, special *InventorySlot, prodSlot int16, bookSlot, specialSlot int16, materialSlots []int16, materialCounts []uint) ([]byte, error) {

	production := Productions[int(book.ItemID)]
	prodMaterials, err := production.GetMaterials()
	if err != nil {
		return nil, err
	}

	canProduce := true

	for i := 0; i < len(materials); i++ {
		if materials[i].Quantity < uint(prodMaterials[i].Count) || int(materials[i].ItemID) != prodMaterials[i].ID {
			canProduce = false
			break
		}
	}

	if prodMaterials[2].ID > 0 && (special.Quantity < uint(prodMaterials[2].Count) || int(special.ItemID) != prodMaterials[2].ID) {
		canProduce = false
	}

	cost := uint64(production.Cost)
	if cost > c.Gold || !canProduce {
		resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x54, 0x04, 0x07, 0x10, 0x55, 0xAA}
		return resp, nil
	}

	c.LootGold(-cost)
	luckRate := float64(1)
	if special != nil {
		specialInfo := Items[special.ItemID]
		luckRate = float64(specialInfo.SellPrice+100) / 100
	}

	resp := &utils.Packet{}
	seed := int(utils.RandInt(0, 1000))
	if float64(seed) < float64(production.Probability)*luckRate { // Success
		itemInfo := Items[int64(production.Production)]
		quantity := 1
		if itemInfo.Timer > 0 {
			quantity = itemInfo.Timer
		}
		resp, _, err = c.AddItem(&InventorySlot{ItemID: int64(production.Production), Quantity: uint(quantity)}, prodSlot, false)
		if err != nil {
			return nil, err
		} else if resp == nil {
			return nil, nil
		}

		resp.Concat(PRODUCTION_SUCCESS)
		logger.Log(logging.ACTION_PRODUCTION, c.ID, fmt.Sprintf("Production (%d) success", book.ItemID), c.UserID)

	} else { // Failed
		resp.Concat(PRODUCTION_FAILED)
		resp.Concat(c.GetGold())
		logger.Log(logging.ACTION_PRODUCTION, c.ID, fmt.Sprintf("Production (%d) failed", book.ItemID), c.UserID)
	}

	resp.Concat(*c.DecrementItem(int16(bookSlot), 1))

	for i := 0; i < len(materialSlots); i++ {
		resp.Concat(*c.DecrementItem(int16(materialSlots[i]), uint(materialCounts[i])))
	}

	if special != nil {
		resp.Concat(*c.DecrementItem(int16(specialSlot), 1))
	}

	return *resp, nil
}

func (c *Character) AdvancedFusion(items []*InventorySlot, special *InventorySlot, prodSlot int16) ([]byte, bool, error) {

	if len(items) < 3 {
		return nil, false, nil
	}

	fusion := Fusions[items[0].ItemID]
	seed := int(utils.RandInt(0, 1000))

	cost := uint64(fusion.Cost)
	if c.Gold < cost {
		return FUSION_FAILED, false, nil
	}

	if items[0].ItemID != fusion.Item1 || items[1].ItemID != fusion.Item2 || items[2].ItemID != fusion.Item3 {
		return FUSION_FAILED, false, nil
	}

	c.LootGold(-cost)
	rate := float64(fusion.Probability)
	if special != nil {
		info := Items[special.ItemID]
		rate *= float64(info.SellPrice+100) / 100
	}

	if float64(seed) < rate { // Success
		resp := utils.Packet{}
		quantity := 1
		iteminfo := Items[fusion.Production]
		if iteminfo.Timer > 0 {
			quantity = iteminfo.Timer
		}
		itemData, _, err := c.AddItem(&InventorySlot{ItemID: fusion.Production, Quantity: uint(quantity)}, prodSlot, false)
		if err != nil {
			return nil, false, err
		} else if itemData == nil {
			return nil, false, nil
		}

		resp.Concat(*itemData)
		resp.Concat(FUSION_SUCCESS)
		logger.Log(logging.ACTION_ADVANCED_FUSION, c.ID, fmt.Sprintf("Advanced fusion (%d) success", items[0].ItemID), c.UserID)
		return resp, true, nil

	} else { // Failed
		resp := FUSION_FAILED
		resp.Concat(c.GetGold())
		logger.Log(logging.ACTION_ADVANCED_FUSION, c.ID, fmt.Sprintf("Advanced fusion (%d) failed", items[0].ItemID), c.UserID)
		return resp, false, nil
	}
}

func (c *Character) Dismantle(item, special *InventorySlot) ([]byte, bool, error) {

	melting := Meltings[int(item.ItemID)]
	cost := uint64(melting.Cost)

	if c.Gold < cost {
		return nil, false, nil
	}

	meltedItems, err := melting.GetMeltedItems()
	if err != nil {
		return nil, false, err
	}

	itemCounts, err := melting.GetItemCounts()
	if err != nil {
		return nil, false, err
	}

	c.LootGold(-cost)

	info := Items[item.ItemID]

	profit := utils.RandFloat(1, melting.ProfitMultiplier) * float64(info.BuyPrice*2)
	c.LootGold(uint64(profit))

	resp := utils.Packet{}
	r := DISMANTLE_SUCCESS
	r.Insert(utils.IntToBytes(uint64(profit), 8, true), 9) // profit

	count, index := 0, 18
	for i := 0; i < 3; i++ {
		id := meltedItems[i]
		if id == 0 {
			continue
		}

		maxCount := int64(itemCounts[i])
		meltedCount := utils.RandInt(0, maxCount+1)
		if meltedCount == 0 {
			continue
		}

		count++
		r.Insert(utils.IntToBytes(uint64(id), 4, true), index) // melted item id
		index += 4

		r.Insert([]byte{0x00, 0xA2}, index)
		index += 2

		r.Insert(utils.IntToBytes(uint64(meltedCount), 2, true), index) // melted item count
		index += 2

		freeSlot, err := c.FindFreeSlot()
		if err != nil {
			return nil, false, err
		}

		r.Insert(utils.IntToBytes(uint64(freeSlot), 2, true), index) // free slot id
		index += 2

		r.Insert([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, index) // upgrades
		index += 34

		itemData, _, err := c.AddItem(&InventorySlot{ItemID: int64(id), Quantity: uint(meltedCount)}, freeSlot, false)
		if err != nil {
			return nil, false, err
		} else if itemData == nil {
			return nil, false, nil
		}

		resp.Concat(*itemData)
	}

	r[17] = byte(count)
	length := int16(44*count) + 14

	if melting.SpecialItem > 0 {
		seed := int(utils.RandInt(0, 1000))

		if seed < melting.SpecialProbability {

			freeSlot, err := c.FindFreeSlot()
			if err != nil {
				return nil, false, err
			}

			r.Insert([]byte{0x01}, index)
			index++

			r.Insert(utils.IntToBytes(uint64(melting.SpecialItem), 4, true), index) // special item id
			index += 4

			r.Insert([]byte{0x00, 0xA2, 0x01, 0x00}, index)
			index += 4

			r.Insert(utils.IntToBytes(uint64(freeSlot), 2, true), index) // free slot id
			index += 2

			r.Insert([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, index) // upgrades
			index += 34

			itemData, _, err := c.AddItem(&InventorySlot{ItemID: int64(melting.SpecialItem), Quantity: 1}, freeSlot, false)
			if err != nil {
				return nil, false, err
			} else if itemData == nil {
				return nil, false, nil
			}

			resp.Concat(*itemData)
			length += 45
		}
	}

	r.SetLength(length)
	resp.Concat(r)
	resp.Concat(c.GetGold())
	logger.Log(logging.ACTION_DISMANTLE, c.ID, fmt.Sprintf("Dismantle (%d) success with %d gold", item.ID, c.Gold), c.UserID)
	return resp, true, nil
}

func (c *Character) Extraction(item, special *InventorySlot, itemSlot int16) ([]byte, bool, error) {

	info := Items[item.ItemID]
	code := int(item.GetUpgrades()[item.Plus-1])
	cost := uint64(info.SellPrice) * uint64(HaxCodes[code].ExtractionMultiplier) / 1000

	if c.Gold < cost {
		return nil, false, nil
	}

	c.LootGold(-cost)
	item.Plus--
	item.SetUpgrade(int(item.Plus), 0)

	resp := utils.Packet{}
	r := EXTRACTION_SUCCESS
	r.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 9)    // item id
	r.Insert(utils.IntToBytes(uint64(item.Quantity), 2, true), 15) // item quantity
	r.Insert(utils.IntToBytes(uint64(itemSlot), 2, true), 17)      // item slot
	r.Insert(item.GetUpgrades(), 19)                               // item upgrades
	r[34] = byte(item.SocketCount)                                 // item socket count
	r.Insert(item.GetUpgrades(), 35)                               // item sockets

	count := 1          //int(utils.RandInt(1, 4))
	r[53] = byte(count) // stone count

	index, length := 54, int16(51)
	for i := 0; i < count; i++ {

		freeSlot, err := c.FindFreeSlot()
		if err != nil {
			return nil, false, err
		}

		id := int64(HaxCodes[code].ExtractedItem)
		itemData, _, err := c.AddItem(&InventorySlot{ItemID: id, Quantity: 1}, freeSlot, false)
		if err != nil {
			return nil, false, err
		} else if itemData == nil {
			return nil, false, nil
		}

		resp.Concat(*itemData)

		r.Insert(utils.IntToBytes(uint64(id), 4, true), index) // extracted item id
		index += 4

		r.Insert([]byte{0x00, 0xA2, 0x01, 0x00}, index)
		index += 4

		r.Insert(utils.IntToBytes(uint64(freeSlot), 2, true), index) // free slot id
		index += 2

		r.Insert([]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
			0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, index) // upgrades
		index += 34

		length += 44
	}

	r.SetLength(length)
	resp.Concat(r)
	resp.Concat(c.GetGold())

	err := item.Update()
	if err != nil {
		return nil, false, err
	}

	logger.Log(logging.ACTION_EXTRACTION, c.ID, fmt.Sprintf("Extraction success for item (%d)", item.ID), c.UserID)
	return resp, true, nil
}

func (c *Character) CreateSocket(item, special *InventorySlot, itemSlot, specialSlot int16) ([]byte, error) {

	info := Items[item.ItemID]

	cost := uint64(info.SellPrice * 164)
	if c.Gold < cost {
		return nil, nil
	}

	if item.SocketCount > 0 && special != nil && special.ItemID == 17200186 { // socket init
		resp := c.DecrementItem(specialSlot, 1)
		resp.Concat(item.CreateSocket(itemSlot, 0))
		return *resp, nil

	} else if item.SocketCount > 0 { // item already has socket
		resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x54, 0x16, 0x0B, 0xCF, 0x55, 0xAA}
		return resp, nil

	} else if item.SocketCount == 0 && special != nil && special.ItemID == 17200186 { // socket init with no sockets
		resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x54, 0x16, 0x0A, 0xCF, 0x55, 0xAA}
		return resp, nil
	}

	seed := utils.RandInt(0, 1000)
	socketCount := int8(1)
	if seed >= 850 {
		socketCount = 4
	} else if seed >= 650 {
		socketCount = 3
	} else if seed >= 350 {
		socketCount = 2
	}

	c.LootGold(-cost)
	resp := utils.Packet{}
	if special != nil {
		if special.ItemID == 17200185 { // +1 miled stone
			socketCount++

		} else if special.ItemID == 15710239 { // +2 miled stone
			socketCount += 2
			if socketCount > 5 {
				socketCount = 5
			}

		}

		resp.Concat(*c.DecrementItem(specialSlot, 1))
	}
	item.SocketCount = socketCount
	item.Update()
	resp.Concat(item.CreateSocket(itemSlot, socketCount))
	resp.Concat(c.GetGold())
	return resp, nil
}

func (c *Character) UpgradeSocket(item, socket, special, edit *InventorySlot, itemSlot, socketSlot, specialSlot, editSlot int16, locks []bool) ([]byte, error) {

	info := Items[item.ItemID]
	cost := uint64(info.SellPrice * 164)
	if c.Gold < cost {
		return nil, nil
	}

	if item.SocketCount == 0 { // No socket on item
		resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x54, 0x16, 0x10, 0xCF, 0x55, 0xAA}
		return resp, nil
	}

	if socket.Plus < uint8(item.SocketCount) { // Insufficient socket
		resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x54, 0x17, 0x0D, 0xCF, 0x55, 0xAA}
		return resp, nil
	}

	stabilize := special != nil && special.ItemID == 17200187

	if edit != nil {
		if edit.ItemID < 17503030 && edit.ItemID > 17503032 {
			return nil, nil
		}
	}

	upgradesArray := bytes.Join([][]byte{ArmorUpgrades, WeaponUpgrades, AccUpgrades}, []byte{})
	sockets := make([]byte, item.SocketCount)
	socks := item.GetSockets()
	for i := int8(0); i < item.SocketCount; i++ {
		if locks[i] {
			sockets[i] = socks[i]
			continue
		}

		seed := utils.RandInt(0, int64(len(upgradesArray)+1))
		code := upgradesArray[seed]
		if stabilize && code%5 > 0 {
			code++
		} else if !stabilize && code%5 == 0 {
			code--
		}

		sockets[i] = code
	}

	c.LootGold(-cost)
	resp := utils.Packet{}
	resp.Concat(item.UpgradeSocket(itemSlot, sockets))
	resp.Concat(c.GetGold())
	resp.Concat(*c.DecrementItem(socketSlot, 1))

	if special != nil {
		resp.Concat(*c.DecrementItem(specialSlot, 1))
	}

	if edit != nil {
		resp.Concat(*c.DecrementItem(editSlot, 1))
	}

	return resp, nil
}

func (c *Character) CoProduction(craftID, bFinished int) ([]byte, error) {
	resp := utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x54, 0x20, 0x0a, 0x00, 0x00, 0x55, 0xAA}
	resp.Concat(utils.Packet{0xAA, 0x55, 0x28, 0x00, 0x16, 0x2d, 0x03, 0x2d, 0x03, 0xfe, 0x9a, 0x00, 0x00, 0x83, 0x1b, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x54, 0x83, 0x05, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x10, 0x27, 0x00, 0x00, 0x55, 0xAA})
	if bFinished == 1 {
		production := CraftItems[int(craftID)]
		prodMaterials, err := production.GetMaterials()
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(prodMaterials); i++ {
			if int64(prodMaterials[i].ID) == 0 {
				break
			} else {
				slotID, _, _ := c.FindItemInInventory(nil, int64(prodMaterials[i].ID))
				matCount := uint(prodMaterials[i].Count)
				if _, ok := Relics[prodMaterials[i].ID]; !ok { // relic drop
					itemData := c.DecrementItem(slotID, matCount)
					c.Socket.Write(*itemData)
				}
			}
		}
		cost := uint64(production.Cost)
		c.Gold -= cost
		resp.Concat(c.GetGold())
		//log.Printf("Item deleted")
		slots, err := c.InventorySlots()
		if err != nil {
		}
		index := 0
		seed := int(utils.RandInt(0, 1000))
		probabilities := production.GetProbabilities()
		for _, prob := range probabilities {
			if float64(seed) > float64(prob) {
				index++
				continue
			}
			break
		}
		craftedItems := production.GetItems()
		reward := NewSlot()
		reward.ItemID = int64(craftedItems[index])
		reward.Quantity = 1
		_, slot, _ := c.AddItem(reward, -1, true)
		resp.Concat(slots[slot].GetData(slot))
	}
	return resp, nil
}

func (c *Character) HolyWaterUpgrade(item, holyWater *InventorySlot, itemSlot, holyWaterSlot int16) ([]byte, error) {

	itemInfo := Items[item.ItemID]
	hwInfo := Items[holyWater.ItemID]

	if (itemInfo.GetType() == WEAPON_TYPE && (hwInfo.HolyWaterUpg1 < 66 || hwInfo.HolyWaterUpg1 > 105)) ||
		(itemInfo.GetType() == ARMOR_TYPE && (hwInfo.HolyWaterUpg1 < 41 || hwInfo.HolyWaterUpg1 > 65)) ||
		(itemInfo.GetType() == ACC_TYPE && hwInfo.HolyWaterUpg1 > 40) { // Mismatch type

		resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x54, 0x10, 0x36, 0x11, 0x55, 0xAA}
		return resp, nil
	}

	if item.Plus == 0 {
		resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x54, 0x10, 0x37, 0x11, 0x55, 0xAA}
		return resp, nil
	}

	resp := utils.Packet{}
	seed, upgrade := int(utils.RandInt(0, 60)), 0
	if seed < hwInfo.HolyWaterRate1 {
		upgrade = hwInfo.HolyWaterUpg1
	} else if seed < hwInfo.HolyWaterRate2 {
		upgrade = hwInfo.HolyWaterUpg2
	} else if seed < hwInfo.HolyWaterRate3 {
		upgrade = hwInfo.HolyWaterUpg3
	} else {
		resp = HOLYWATER_FAILED
	}

	if upgrade > 0 {
		randSlot := utils.RandInt(0, int64(item.Plus))
		preUpgrade := item.GetUpgrades()[randSlot]
		item.SetUpgrade(int(randSlot), byte(upgrade))

		if preUpgrade == byte(upgrade) {
			resp = HOLYWATER_FAILED
		} else {
			resp = HOLYWATER_SUCCESS

			r := ITEM_UPGRADED
			r.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 9) // item id
			r.Insert(utils.IntToBytes(uint64(itemSlot), 2, true), 17)   // slot id
			r.Insert(item.GetUpgrades(), 19)                            // item upgrades
			r[34] = byte(item.SocketCount)                              // socket count
			r.Insert(item.GetSockets(), 35)                             // item sockets
			resp.Concat(r)

			new := funk.Map(item.GetUpgrades()[:item.Plus], func(upg byte) string {
				return HaxCodes[int(upg)].Code
			}).([]string)

			old := make([]string, len(new))
			copy(old, new)
			old[randSlot] = HaxCodes[int(preUpgrade)].Code

			msg := fmt.Sprintf("[%s] has been upgraded from [%s] to [%s].", itemInfo.Name, strings.Join(old, ""), strings.Join(new, ""))
			msgData := messaging.InfoMessage(msg)
			resp.Concat(msgData)
		}
	}

	itemData, _ := c.RemoveItem(holyWaterSlot)
	resp.Concat(itemData)

	err := item.Update()
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Character) RegisterItem(item *InventorySlot, price uint64, itemSlot int16) ([]byte, error) {

	items, err := FindConsignmentItemsBySellerID(c.ID)
	if err != nil {
		return nil, err
	}

	if len(items) >= 10 {
		return nil, nil
	}

	commision := uint64(math.Min(float64(price/100), 50000000))
	if c.Gold < commision {
		return nil, nil
	}

	info, ok := Items[item.ItemID]
	if !ok {
		return nil, nil
	}

	consItem := &ConsignmentItem{
		ID:       item.ID,
		SellerID: c.ID,
		ItemName: info.Name,
		Quantity: int(item.Quantity),
		IsSold:   false,
		Price:    price,
	}

	if err := consItem.Create(); err != nil {
		return nil, err
	}

	c.LootGold(-commision)
	resp := ITEM_REGISTERED
	resp.Insert(utils.IntToBytes(uint64(consItem.ID), 4, true), 9)  // consignment item id
	resp.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 29) // item id

	if item.Pet != nil {
		resp[34] = byte(item.SocketCount)
	}

	resp.Insert(utils.IntToBytes(uint64(item.Quantity), 2, true), 35) // item count
	resp.Insert(item.GetUpgrades(), 37)                               // item upgrades

	if item.Pet != nil {
		resp[42] = 0 // item socket count
	} else {
		resp[42] = byte(item.SocketCount) // item socket count
	}

	resp.Insert(item.GetSockets(), 43) // item sockets

	newItem := NewSlot()
	*newItem = *item
	newItem.SlotID = -1
	newItem.Consignment = true
	newItem.Update()
	InventoryItems.Add(newItem.ID, newItem)

	*item = *NewSlot()
	resp.Concat(c.GetGold())
	resp.Concat(item.GetData(itemSlot))

	claimData, err := c.ClaimMenu()
	if err != nil {
		return nil, err
	}
	resp.Concat(claimData)

	return resp, nil
}

func (c *Character) ClaimMenu() ([]byte, error) {
	items, err := FindConsignmentItemsBySellerID(c.ID)
	if err != nil {
		return nil, err
	}

	resp := CLAIM_MENU
	resp.SetLength(int16(len(items)*0x6B + 6))
	resp.Insert(utils.IntToBytes(uint64(len(items)), 2, true), 8) // items count

	index := 10
	for _, item := range items {

		slot, err := FindInventorySlotByID(item.ID) // FIX: Buyer can destroy the item..
		if err != nil {
			continue
		}
		if slot == nil {
			slot = NewSlot()
			slot.ItemID = 17502455
			slot.Quantity = 1
		}

		info := Items[int64(slot.ItemID)]

		if item.IsSold {
			resp.Insert([]byte{0x01}, index)
		} else {
			resp.Insert([]byte{0x00}, index)
		}
		index++

		resp.Insert(utils.IntToBytes(uint64(item.ID), 4, true), index) // consignment item id
		index += 4

		resp.Insert([]byte{0x5E, 0x15, 0x01, 0x00}, index)
		index += 4

		resp.Insert([]byte(c.Name), index) // seller name
		index += len(c.Name)

		for j := len(c.Name); j < 20; j++ {
			resp.Insert([]byte{0x00}, index)
			index++
		}

		resp.Insert(utils.IntToBytes(item.Price, 8, true), index) // item price
		index += 8

		time := item.ExpiresAt.Time.Format("2006-01-02 15:04:05") // expires at
		resp.Insert([]byte(time), index)
		index += 19

		resp.Insert([]byte{0x00, 0x09, 0x00, 0x00, 0x00, 0x99, 0x31, 0xF5, 0x00}, index)
		index += 9

		resp.Insert(utils.IntToBytes(uint64(slot.ItemID), 4, true), index) // item id
		index += 4

		resp.Insert([]byte{0x00, 0xA1}, index)
		index += 2

		if info.GetType() == PET_TYPE {
			resp[index-1] = byte(slot.SocketCount)
		}

		resp.Insert(utils.IntToBytes(uint64(slot.Quantity), 2, true), index) // item count
		index += 2

		resp.Insert(slot.GetUpgrades(), index) // item upgrades
		index += 15

		resp.Insert([]byte{byte(slot.SocketCount)}, index) // socket count
		index++

		resp.Insert(slot.GetSockets(), index)
		index += 15

		resp.Insert([]byte{0x00, 0x00, 0x00}, index)
		index += 3
	}

	return resp, nil
}

func (c *Character) BuyConsignmentItem(consignmentID int) ([]byte, error) {

	consignmentItem, err := FindConsignmentItemByID(consignmentID)
	if err != nil || consignmentItem == nil || consignmentItem.IsSold {
		return nil, err
	}

	slots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	slot, err := FindInventorySlotByID(consignmentItem.ID)
	if err != nil {
		return nil, err
	}

	if c.Gold < consignmentItem.Price {
		return nil, nil
	}

	seller, err := FindCharacterByID(int(slot.CharacterID.Int64))
	if err != nil {
		return nil, err
	}

	resp := CONSIGMENT_ITEM_BOUGHT
	resp.Insert(utils.IntToBytes(uint64(consignmentID), 4, true), 8) // consignment item id

	slotID, err := c.FindFreeSlot()
	if err != nil {
		return nil, nil
	}

	newItem := NewSlot()
	*newItem = *slot
	newItem.Consignment = false
	newItem.UserID = null.StringFrom(c.UserID)
	newItem.CharacterID = null.IntFrom(int64(c.ID))
	newItem.SlotID = slotID

	err = newItem.Update()
	if err != nil {
		return nil, err
	}

	*slots[slotID] = *newItem
	InventoryItems.Add(newItem.ID, slots[slotID])
	c.LootGold(-consignmentItem.Price)

	resp.Concat(newItem.GetData(slotID))
	resp.Concat(c.GetGold())

	s, ok := Sockets[seller.UserID]
	if ok {
		s.Write(CONSIGMENT_ITEM_SOLD)
	}

	logger.Log(logging.ACTION_BUY_CONS_ITEM, c.ID, fmt.Sprintf("Bought consignment item (%d) with %d gold from (%d)", newItem.ID, consignmentItem.Price, seller.ID), c.UserID)
	consignmentItem.IsSold = true
	go consignmentItem.Update()
	return resp, nil
}

func (c *Character) ClaimConsignmentItem(consignmentID int, isCancel bool) ([]byte, error) {

	consignmentItem, err := FindConsignmentItemByID(consignmentID)
	if err != nil || consignmentItem == nil {
		return nil, err
	}

	resp := CONSIGMENT_ITEM_CLAIMED
	resp.Insert(utils.IntToBytes(uint64(consignmentID), 4, true), 10) // consignment item id

	if isCancel {
		if consignmentItem.IsSold {
			return nil, nil
		}

		slots, err := c.InventorySlots()
		if err != nil {
			return nil, err
		}

		slotID, err := c.FindFreeSlot()
		if err != nil {
			return nil, err
		}

		slot, err := FindInventorySlotByID(consignmentItem.ID)
		if err != nil {
			return nil, err
		}

		newItem := NewSlot()
		*newItem = *slot
		newItem.Consignment = false
		newItem.SlotID = slotID

		err = newItem.Update()
		if err != nil {
			return nil, err
		}

		*slots[slotID] = *newItem
		InventoryItems.Add(newItem.ID, slots[slotID])

		resp.Concat(slot.GetData(slotID))

	} else {
		if !consignmentItem.IsSold {
			return nil, nil
		}

		logger.Log(logging.ACTION_BUY_CONS_ITEM, c.ID, fmt.Sprintf("Claimed consignment item (consid:%d) with %d gold", consignmentID, consignmentItem.Price), c.UserID)

		c.LootGold(consignmentItem.Price)
		resp.Concat(c.GetGold())
	}

	s, _ := FindInventorySlotByID(consignmentItem.ID)
	if s != nil && !s.UserID.Valid && !s.CharacterID.Valid {
		s.Delete()
	}

	go consignmentItem.Delete()
	return resp, nil
}

func (c *Character) UseConsumable(item *InventorySlot, slotID int16) ([]byte, error) {

	defer func() {
		if err := recover(); err != nil {
			log.Println(err)
			log.Printf("%+v", string(dbg.Stack()))

			r := utils.Packet{}
			r.Concat(*c.DecrementItem(slotID, 0))
			c.Socket.Write(r)
		}
	}()

	stat := c.Socket.Stats
	if stat.HP <= 0 {
		return *c.DecrementItem(slotID, 0), nil
	}

	info := Items[item.ItemID]
	if info == nil {
		return nil, nil
	} else if info.MinLevel > c.Level || (info.MaxLevel > 0 && info.MaxLevel < c.Level) {
		resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x59, 0x04, 0xF0, 0x03, 0x55, 0xAA} // inappropriate level
		return resp, nil
	}

	resp := utils.Packet{}
	canUse := c.CanUse(info.CharacterType)
	switch info.GetType() {
	case AFFLICTION_TYPE:
		err := stat.Reset()
		if err != nil {
			return nil, err
		}

		statData, _ := c.GetStats()
		resp.Concat(statData)

	case CHARM_OF_RETURN_TYPE:
		d := SavePoints[uint8(c.Map)]
		coordinate := ConvertPointToLocation(d.Point)
		resp.Concat(c.Teleport(coordinate))

		slots, err := c.InventorySlots()
		if err == nil {
			pet := slots[0x0A].Pet
			if pet != nil && pet.IsOnline {
				pet.IsOnline = false
				resp.Concat(DISMISS_PET)
			}
		}

	case DEAD_SPIRIT_INCENSE_TYPE:
		slots, err := c.InventorySlots()
		if err != nil {
			return nil, err
		}

		pet := slots[0x0A].Pet
		if pet != nil && !pet.IsOnline && pet.HP <= 0 {
			pet.HP = pet.MaxHP / 10
			resp.Concat(c.GetPetStats())
			resp.Concat(c.TogglePet())
		} else {
			goto FALLBACK
		}

	case MOVEMENT_SCROLL_TYPE:
		mapID := int16(info.SellPrice)
		data, _ := c.ChangeMap(mapID, nil)
		resp.Concat(data)

	case BAG_EXPANSION_TYPE:
		buff, err := FindBuffByID(int(item.ItemID), c.ID)
		if err != nil {
			return nil, err
		} else if buff != nil {
			return []byte{0xAA, 0x55, 0x04, 0x00, 0x59, 0x04, 0xFA, 0x03, 0x55, 0xAA}, nil // FIX: Already Exists => Already have the same effect
		}

		buff = &Buff{ID: int(item.ItemID), CharacterID: c.ID, Name: info.Name, BagExpansion: true, StartedAt: c.Epoch, Duration: int64(info.Timer) * 60, CanExpire: true}
		err = buff.Create()
		if err != nil {
			return nil, err
		}

		resp = BAG_EXPANDED

	case FIRE_SPIRIT:
		buff, err := FindBuffByID(int(item.ItemID), c.ID)
		if err != nil {
			return nil, err
		} else if buff != nil {
			return []byte{0xAA, 0x55, 0x04, 0x00, 0x59, 0x04, 0xFA, 0x03, 0x55, 0xAA}, nil // FIX: Already Exists => Already have the same effect
		}

		buff, err = FindBuffByID(19000019, c.ID) // check for water spirit
		if err != nil {
			return nil, err
		} else if buff != nil {
			return []byte{0xAA, 0x55, 0x04, 0x00, 0x59, 0x04, 0xFA, 0x03, 0x55, 0xAA}, nil // FIX: Already Exists => Already have the same effect
		}

		buff = &Buff{ID: int(item.ItemID), CharacterID: c.ID, Name: info.Name, EXPMultiplier: 30, DropMultiplier: 5, DEFRate: 5, ArtsDEFRate: 5,
			ATKRate: 4, ArtsATKRate: 4, StartedAt: c.Epoch, Duration: 2592000, CanExpire: true}
		err = buff.Create()
		if err != nil {
			return nil, err
		}

		c.ExpMultiplier += 0.3
		c.DropMultiplier += 0.05
		itemData, _, _ := c.AddItem(&InventorySlot{ItemID: 17502645, Quantity: 1}, -1, false)
		resp.Concat(*itemData)

		data, _ := c.GetStats()
		resp.Concat(data)

	case WATER_SPIRIT:
		buff, err := FindBuffByID(int(item.ItemID), c.ID)
		if err != nil {
			return nil, err
		} else if buff != nil {
			return []byte{0xAA, 0x55, 0x04, 0x00, 0x59, 0x04, 0xFA, 0x03, 0x55, 0xAA}, nil // FIX: Already Exists => Already have the same effect
		}

		buff, err = FindBuffByID(19000018, c.ID) // check for fire spirit
		if err != nil {
			return nil, err
		} else if buff != nil {
			return []byte{0xAA, 0x55, 0x04, 0x00, 0x59, 0x04, 0xFA, 0x03, 0x55, 0xAA}, nil // FIX: Already Exists => Already have the same effect
		}

		buff = &Buff{ID: int(item.ItemID), CharacterID: c.ID, Name: info.Name, EXPMultiplier: 60, DropMultiplier: 10, DEFRate: 15, ArtsDEFRate: 15,
			ATKRate: 8, ArtsATKRate: 8, StartedAt: c.Epoch, Duration: 2592000, CanExpire: true}
		err = buff.Create()
		if err != nil {
			return nil, err
		}

		c.ExpMultiplier += 0.6
		c.DropMultiplier += 0.1
		itemData, _, _ := c.AddItem(&InventorySlot{ItemID: 17502646, Quantity: 1}, -1, false)
		resp.Concat(*itemData)

		data, _ := c.GetStats()
		resp.Concat(data)

	case FORTUNE_BOX_TYPE:

		c.InvMutex.Lock()
		defer c.InvMutex.Unlock()
		_, err := c.FindFreeSlots(1)
		if err != nil {
			goto FALLBACK
		}
		gambling := GamblingItems[int(item.ItemID)]
		if gambling == nil || gambling.Cost > c.Gold { // FIX Gambling null
			resp := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x59, 0x08, 0xF9, 0x03, 0x55, 0xAA} // not enough gold
			return resp, nil
		}

		slots, err := c.InventorySlots()
		if err != nil {
			return nil, err
		}

		c.LootGold(-gambling.Cost)
		resp.Concat(c.GetGold())

		drop, ok := Drops[gambling.DropID]
		if drop == nil || !ok {
			goto FALLBACK
		}

		var itemID int
		for ok {
			index := 0
			seed := int(utils.RandInt(0, 1000))
			items := drop.GetItems()
			probabilities := drop.GetProbabilities()

			for _, prob := range probabilities {
				if float64(seed) > float64(prob) {
					index++
					continue
				}
				break
			}

			if index >= len(items) {
				break
			}

			itemID = items[index]
			drop, ok = Drops[itemID]
		}

		plus, quantity, upgs := uint8(0), uint(1), []byte{}
		rewardInfo := Items[int64(itemID)]
		if rewardInfo != nil {
			if rewardInfo.ID == 235 || rewardInfo.ID == 242 || rewardInfo.ID == 254 || rewardInfo.ID == 255 { // Socket-PP-Ghost Dagger-Dragon Scale
				var rates []int
				if rewardInfo.ID == 235 { // Socket
					rates = []int{300, 550, 750, 900, 1000}
				} else {
					rates = []int{500, 900, 950, 975, 990, 995, 998, 100}
				}

				seed := int(utils.RandInt(0, 1000))
				for ; seed > rates[plus]; plus++ {
				}
				plus++

				upgs = utils.CreateBytes(byte(rewardInfo.ID), int(plus), 15)

			} else if rewardInfo.GetType() == MARBLE_TYPE { // Marble
				rates := []int{200, 300, 500, 750, 950, 1000}
				seed := int(utils.RandInt(0, 1000))
				for i := 0; seed > rates[i]; i++ {
					itemID++
				}

				rewardInfo = Items[int64(itemID)]

			} else if funk.Contains(haxBoxes, item.ItemID) { // Hax Box
				seed := utils.RandInt(0, 1000)
				plus = uint8(sort.SearchInts(plusRates, int(seed)) + 1)

				upgradesArray := []byte{}
				rewardType := rewardInfo.GetType()
				if rewardType == WEAPON_TYPE {
					upgradesArray = WeaponUpgrades
				} else if rewardType == ARMOR_TYPE {
					upgradesArray = ArmorUpgrades
				} else if rewardType == ACC_TYPE {
					upgradesArray = AccUpgrades
				}

				index := utils.RandInt(0, int64(len(upgradesArray)))
				code := upgradesArray[index]
				if (code-1)%5 == 3 {
					code--
				} else if (code-1)%5 == 4 {
					code -= 2
				}

				upgs = utils.CreateBytes(byte(code), int(plus), 15)
			}

			if q, ok := rewardCounts[item.ItemID]; ok {
				quantity = q
			}

			if box, ok := rewardCounts2[item.ItemID]; ok {
				if q, ok := box[rewardInfo.ID]; ok {
					quantity = q
				}
			}

			if int(rewardInfo.TimerType) > 0 || rewardInfo.Timer > 0 {
				quantity = uint(rewardInfo.Timer)
			}
			item := &InventorySlot{ItemID: rewardInfo.ID, Plus: uint8(plus), Quantity: quantity}
			item.SetUpgrades(upgs)

			if rewardInfo.GetType() == PET_TYPE {
				petInfo := Pets[int64(rewardInfo.ID)]
				petExpInfo := PetExps[int16(petInfo.Level)]

				targetExps := []int{petExpInfo.ReqExpEvo1, petExpInfo.ReqExpEvo2, petExpInfo.ReqExpEvo3, petExpInfo.ReqExpHt, petExpInfo.ReqExpDivEvo1, petExpInfo.ReqExpDivEvo2, petExpInfo.ReqExpDivEvo3}
				item.Pet = &PetSlot{
					Fullness: 100, Loyalty: 100,
					Exp:   uint64(targetExps[petInfo.Evolution-1]),
					HP:    petInfo.BaseHP,
					Level: byte(petInfo.Level),
					Name:  petInfo.Name,
					CHI:   petInfo.BaseChi,
				}
			}

			_, slot, err := c.AddItem(item, -1, true)
			if err != nil {
				return nil, err
			}

			resp.Concat(slots[slot].GetData(slot))
		}

	case NPC_SUMMONER_TYPE:
		if item.ItemID == 17502966 || item.ItemID == 17100004 { // Tavern
			r := utils.Packet{0xAA, 0x55, 0x07, 0x00, 0x57, 0x03, 0x01, 0x06, 0x00, 0x00, 0x00, 0x55, 0xAA}
			resp.Concat(r)
		} else if item.ItemID == 17502967 || item.ItemID == 17100005 { // Bank
			resp.Concat(c.BankItems())
		}

	case PASSIVE_SKILL_BOOK_TYPE:
		/*if !canUse { // invalid character type
			return INVALID_CHARACTER_TYPE, nil
		}*/

		skills, err := FindSkillsByID(c.ID)
		if err != nil {
			return nil, err
		}

		skillSlots, err := skills.GetSkills()
		if err != nil {
			return nil, err
		}

		i := -1
		if info.Name == "Wind Drift Arts" {
			i = 7
			if skillSlots.Slots[i].BookID > 0 {
				return SKILL_BOOK_EXISTS, nil
			}

		} else {
			for j := 5; j < 7; j++ {
				if skillSlots.Slots[j].BookID == 0 {
					i = j
					break
				} else if skillSlots.Slots[j].BookID == item.ItemID { // skill book exists
					return SKILL_BOOK_EXISTS, nil
				}
			}
		}

		if i == -1 {
			return NO_SLOTS_FOR_SKILL_BOOK, nil // FIX resp
		}

		set := &SkillSet{BookID: item.ItemID}
		set.Skills = append(set.Skills, &SkillTuple{SkillID: int(info.ID), Plus: 0})
		skillSlots.Slots[i] = set
		skills.SetSkills(skillSlots)

		go skills.Update()

		skillsData, err := skills.GetSkillsData()
		if err != nil {
			return nil, err
		}

		resp.Concat(skillsData)

	case PET_POTION_TYPE:
		slots, err := c.InventorySlots()
		if err != nil {
			return nil, err
		}

		petSlot := slots[0x0A]
		pet := petSlot.Pet

		if pet == nil || !pet.IsOnline {
			goto FALLBACK
		}

		pet.HP = int(math.Min(float64(pet.HP+info.HpRecovery), float64(pet.MaxHP)))
		pet.CHI = int(math.Min(float64(pet.CHI+info.ChiRecovery), float64(pet.MaxCHI)))
		pet.Fullness = byte(math.Min(float64(pet.Fullness+5), float64(100)))
		resp.Concat(c.GetPetStats())

	case POTION_TYPE:
		hpRec := info.HpRecovery
		chiRec := info.ChiRecovery
		if hpRec == 0 && chiRec == 0 {
			hpRec = 50000
			chiRec = 50000
		}

		stat.HP = int(math.Min(float64(stat.HP+hpRec), float64(stat.MaxHP)))
		stat.CHI = int(math.Min(float64(stat.CHI+chiRec), float64(stat.MaxCHI)))
		resp.Concat(c.GetHPandChi())

	case FILLER_POTION_TYPE:
		hpRecovery, chiRecovery := math.Min(float64(stat.MaxHP-stat.HP), 50000), float64(0)
		if hpRecovery > float64(item.Quantity) {
			hpRecovery = float64(item.Quantity)
		} else {
			chiRecovery = math.Min(float64(stat.MaxCHI-stat.CHI), 50000)
			if chiRecovery+hpRecovery > float64(item.Quantity) {
				chiRecovery = float64(item.Quantity) - hpRecovery
			}
		}

		stat.HP = int(math.Min(float64(stat.HP)+hpRecovery, float64(stat.MaxHP)))
		stat.CHI = int(math.Min(float64(stat.CHI)+chiRecovery, float64(stat.MaxCHI)))
		resp.Concat(c.GetHPandChi())
		resp.Concat(*c.DecrementItem(slotID, uint(hpRecovery+chiRecovery)))
		resp.Concat(item.GetData(slotID))
		return resp, nil

	case SKILL_BOOK_TYPE:
		if !canUse { // invalid character type
			return INVALID_CHARACTER_TYPE, nil
		}

		skills, err := FindSkillsByID(c.ID)
		if err != nil {
			return nil, err
		}

		skillSlots, err := skills.GetSkills()
		if err != nil {
			return nil, err
		}

		i := -1
		for j := 0; j < 5; j++ {
			if skillSlots.Slots[j].BookID == 0 {
				i = j
				break
			} else if skillSlots.Slots[j].BookID == item.ItemID { // skill book exists
				return SKILL_BOOK_EXISTS, nil
			}
		}

		if i == -1 {
			return NO_SLOTS_FOR_SKILL_BOOK, nil // FIX resp
		}

		skillInfos := SkillInfosByBook[item.ItemID]
		set := &SkillSet{BookID: item.ItemID}
		c := 0
		for i := 1; i <= 24; i++ { // there should be 24 skills with empty ones

			if len(skillInfos) <= c {
				set.Skills = append(set.Skills, &SkillTuple{SkillID: 0, Plus: 0})
			} else if si := skillInfos[c]; si.Slot == i {
				tuple := &SkillTuple{SkillID: si.ID, Plus: 0}
				set.Skills = append(set.Skills, tuple)

				c++
			} else {
				set.Skills = append(set.Skills, &SkillTuple{SkillID: 0, Plus: 0})
			}
		}

		skillSlots.Slots[i] = set
		divtuple := &DivineTuple{DivineID: 0, DivinePlus: 0}
		div2tuple := &DivineTuple{DivineID: 1, DivinePlus: 0}
		div3tuple := &DivineTuple{DivineID: 2, DivinePlus: 0}
		set.DivinePoints = append(set.DivinePoints, divtuple, div2tuple, div3tuple)
		skills.SetSkills(skillSlots)

		go skills.Update()

		skillsData, err := skills.GetSkillsData()
		if err != nil {
			return nil, err
		}
		resp.Concat(skillsData)

	case WRAPPER_BOX_TYPE:

		c.InvMutex.Lock()
		defer c.InvMutex.Unlock()
		if item.ItemID == 90000304 {
			if c.Exp >= 544951059310 && c.Level == 200 {
				c.Type += 10
				c.Update()
				c.Socket.Skills.Delete()
				c.Socket.Skills.Create(c)
				c.Socket.Skills.SkillPoints = 300000
				c.Update()
				data, levelUp := c.AddExp(10)
				if levelUp {
					skillsData, err := c.Socket.Skills.GetSkillsData()
					resp.Concat(skillsData)
					if err == nil && c.Socket != nil {
						c.Socket.Write(skillsData)
					}
					statData, err := c.GetStats()
					if err == nil && c.Socket != nil {
						c.Socket.Write(statData)
					}
				}
				if c.Socket != nil {
					c.Socket.Write(data)
				}
				goto DARKBACK
			}
			goto FALLBACK
		}
		if item.ItemID == 13000015 || item.ItemID == 13000037 || item.ItemID == 13000060 {
			c.AidTime += 7200 //120 perc
			stData, _ := c.GetStats()
			resp.Concat(stData)
			resp.Concat(*c.DecrementItem(slotID, 1))
			return resp, nil
		}
		if item.ItemID == 13000074 {
			c.AidTime += 14400 //240 perc
			stData, _ := c.GetStats()
			resp.Concat(stData)
			resp.Concat(*c.DecrementItem(slotID, 1))
			return resp, nil
		}
		if item.ItemID == 13000011 {
			c.AidTime += 7200 //360 perc
			stData, _ := c.GetStats()
			resp.Concat(stData)
			resp.Concat(*c.DecrementItem(slotID, 1))
			return resp, nil
		}
		if item.ItemID == 13000012 || item.ItemID == 14000008 {
			c.AidTime += 86400 //360 perc
			stData, _ := c.GetStats()
			resp.Concat(stData)
			resp.Concat(*c.DecrementItem(slotID, 1))
			return resp, nil
		}
		gambling := GamblingItems[int(item.ItemID)]
		d := Drops[gambling.DropID]
		items := d.GetItems()
		_, err := c.FindFreeSlots(len(items))
		if err != nil {
			goto FALLBACK
		}

		slots, err := c.InventorySlots()
		if err != nil {
			goto FALLBACK
		}

		i := 0
		for _, itemID := range items {
			info := Items[int64(itemID)]
			reward := NewSlot()
			reward.ItemID = int64(itemID)
			reward.Quantity = 1

			i++

			itemType := info.GetType()
			if info.Timer > 0 && itemType != BAG_EXPANSION_TYPE {
				reward.Quantity = uint(info.Timer)
			} else if q, ok := rewardCounts[item.ItemID]; ok {
				reward.Quantity = q
			} else if itemType == FILLER_POTION_TYPE {
				reward.Quantity = uint(info.SellPrice)
			}

			plus, upgs := uint8(0), []byte{}
			if info.ID == 235 || info.ID == 242 || info.ID == 254 || info.ID == 255 { // Socket-PP-Ghost Dagger-Dragon Scale
				var rates []int
				if info.ID == 235 { // Socket
					rates = []int{300, 550, 750, 900, 1000}
				} else {
					rates = []int{500, 900, 950, 975, 990, 995, 998, 100}
				}

				seed := int(utils.RandInt(0, 1000))
				for ; seed > rates[plus]; plus++ {
				}
				plus++

				upgs = utils.CreateBytes(byte(info.ID), int(plus), 15)
			}

			reward.Plus = plus
			reward.SetUpgrades(upgs)

			_, slot, _ := c.AddItem(reward, -1, true)
			resp.Concat(slots[slot].GetData(slot))
		}

	case HOLY_WATER_TYPE:
		goto FALLBACK

	case FORM_TYPE:

		info, ok := Items[int64(item.ItemID)]
		if !ok || item.Activated != c.Morphed {
			goto FALLBACK
		}

		item.Activated = !item.Activated
		item.InUse = !item.InUse
		c.Morphed = item.Activated

		if item.Activated {
			r := FORM_ACTIVATED
			r.Insert(utils.IntToBytes(uint64(info.NPCID), 4, true), 5) // form npc id
			resp.Concat(r)
		} else {
			resp.Concat(FORM_DEACTIVATED)
		}

		go item.Update()

		statData, err := c.GetStats()
		if err != nil {
			return nil, err
		}

		resp.Concat(statData)
		goto FALLBACK

	default:
		if info.Timer > 0 {
			item.Activated = !item.Activated
			item.InUse = !item.InUse
			resp.Concat(item.GetData(slotID))

			statsData, _ := c.GetStats()
			resp.Concat(statsData)
			goto FALLBACK
		} else {
			goto FALLBACK
		}
	}

	resp.Concat(*c.DecrementItem(slotID, 1))
	return resp, nil

FALLBACK:
	resp.Concat(*c.DecrementItem(slotID, 0))
	return resp, nil
DARKBACK:
	resp.Concat(*c.DecrementItem(item.SlotID, 1))
	c.Socket.Write(resp)

	ATARAXIA := utils.Packet{0xAA, 0x55, 0x05, 0x00, 0x57, 0x21, 0x0a, 0x00, 0x01, 0x55, 0xAA}
	c.Socket.Write(ATARAXIA)

	CHARACTER_MENU := utils.Packet{0xAA, 0x55, 0x03, 0x00, 0x09, 0x09, 0x00, 0x55, 0xAA}
	resp = CHARACTER_MENU

	time.AfterFunc(time.Duration(100*time.Second), func() {
		CharacterSelect := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x01, 0x05, 0x0A, 0x00, 0x55, 0xAA}

		resp.Concat(CharacterSelect)
	})
	return resp, nil
}

func (c *Character) CanUse(t int) bool {
	if c.Type == 0x34 && t == 0x34 { // Monk
		return true
	} else if c.Type == 0x35 && (t == 0x35 || t == 0x37) { //MALE_BLADE
		return true
	} else if c.Type == 0x36 && (t == 0x36 || t == 0x37) { //FEMALE_BLADE
		return true
	} else if c.Type == 0x38 && (t == 0x38 || t == 0x3A) { //AXE
		return true
	} else if c.Type == 0x39 && (t == 0x39 || t == 0x3A) { //FEMALE_ROD
		return true
	} else if c.Type == 0x32 && (t == 0x32 || t == 0x33) { //Beast
		return true
	} else if c.Type == 0x33 && (t == 0x33 || t == 0x32) { //empress
		return true
	} else if c.Type == 0x3D && (t == 0x3D || t == 0x3C) { //Divine_empress
		return true
	} else if c.Type == 0x3C && (t == 0x3C || t == 0x3D) { //Divine_Beast
		return true
	} else if c.Type == 0x46 && (t == 0x46 || t == 0x47) { //Darkness_Beast
		return true
	} else if c.Type == 0x47 && (t == 0x47 || t == 0x46) { //Darkness_Empress
		return true
	} else if c.Type == 0x3B && t == 0x3B { //DUAL_BLADE
		return true
	} else if c.Type == 0x3E && t == 0x3E { //DIVINE MONK
		return true
	} else if c.Type == 0x3F && (t == 0x3F || t == 0x41) { //DIVINE MALE_BLADE
		return true
	} else if c.Type == 0x40 && (t == 0x40 || t == 0x41) { //DIVINE FEMALE_BLADE
		return true
	} else if c.Type == 0x42 && (t == 0x42 || t == 0x44) { //DIVINE MALE_AXE
		return true
	} else if c.Type == 0x43 && (t == 0x43 || t == 0x44) { //DIVINE FEMALE_ROD
		return true
	} else if c.Type == 0x45 && t == 0x45 { //DIVINE Dual Sword
		return true
	} else if c.Type == 0x48 && t == 0x48 { //DARK LORD MONK
		return true
	} else if c.Type == 0x49 && (t == 0x49 || t == 0x4B) { //DARK LORD MALE_BLADE
		return true
	} else if c.Type == 0x4A && (t == 0x4A || t == 0x4B) { //DARK LORD FEMALE_BLADE
		return true
	} else if c.Type == 0x4C && (t == 0x4C || t == 0x4E) { //DARK LORD MALE_AXE
		return true
	} else if c.Type == 0x4D && (t == 0x4D || t == 0x4E) { //DARK LORD FEMALE_ROD
		return true
	} else if c.Type == 0x4F && t == 0x4F { //DARK LORD Dual Sword
		return true
	} else if t == 0x0A || t == 0x00 || t == 0x34 || t == 0x20 || t == 0x14 { //All character Type
		return true
	}

	return false
}

func (c *Character) UpgradeSkill(slotIndex, skillIndex byte) ([]byte, error) {
	skills, err := FindSkillsByID(c.ID)
	if err != nil {
		return nil, err
	}

	skillSlots, err := skills.GetSkills()
	if err != nil {
		return nil, err
	}

	set := skillSlots.Slots[slotIndex]
	skill := set.Skills[skillIndex]

	info := SkillInfos[skill.SkillID]
	if int8(skill.Plus) >= info.MaxPlus {
		return nil, nil
	}

	requiredSP := 1
	if info.ID >= 28000 && info.ID <= 28007 || info.ID >= 28100 && info.ID <= 28107 { // 2nd job passives (non-divine)
		requiredSP = SkillPTS["sjp"][skill.Plus]
	} else if info.ID >= 29000 && info.ID <= 29007 || info.ID >= 29100 && info.ID <= 29107 { // 2nd job passives (divine)
		requiredSP = SkillPTS["dsjp"][skill.Plus]
	} else if info.ID >= 20193 && info.ID <= 20217 || info.ID >= 20293 && info.ID <= 20317 { // 3nd job passives (darkness)
		requiredSP = SkillPTS["dsjp"][skill.Plus]
	}

	if skills.SkillPoints < requiredSP {
		return nil, nil
	}

	skills.SkillPoints -= requiredSP
	skill.Plus++
	resp := SKILL_UPGRADED
	resp[8] = slotIndex
	resp[9] = skillIndex
	resp.Insert(utils.IntToBytes(uint64(skill.SkillID), 4, true), 10) // skill id
	resp[14] = byte(skill.Plus)

	skills.SetSkills(skillSlots)
	skills.Update()

	if info.Passive {
		statData, err := c.GetStats()
		if err == nil {
			resp.Concat(statData)
		}
	}

	resp.Concat(c.GetExpAndSkillPts())
	return resp, nil
}

func (c *Character) DowngradeSkill(slotIndex, skillIndex byte) ([]byte, error) {
	skills, err := FindSkillsByID(c.ID)
	if err != nil {
		return nil, err
	}

	skillSlots, err := skills.GetSkills()
	if err != nil {
		return nil, err
	}

	set := skillSlots.Slots[slotIndex]
	skill := set.Skills[skillIndex]

	info := SkillInfos[skill.SkillID]
	if int8(skill.Plus) <= 0 {
		return nil, nil
	}

	requiredSP := 1
	if info.ID >= 28000 && info.ID <= 28007 { // 2nd job passives (non-divine)
		requiredSP = SkillPTS["sjp"][skill.Plus]
	} else if info.ID >= 29000 && info.ID <= 29007 { // 2nd job passives (divine)
		requiredSP = SkillPTS["dsjp"][skill.Plus]
	} else if info.ID >= 20193 && info.ID <= 20217 { // 3nd job passives (darkness)
		requiredSP = SkillPTS["dsjp"][skill.Plus]
	}

	skills.SkillPoints += requiredSP
	skill.Plus--
	resp := SKILL_DOWNGRADED
	resp[8] = slotIndex
	resp[9] = skillIndex
	resp.Insert(utils.IntToBytes(uint64(skill.SkillID), 4, true), 10) // skill id
	resp[14] = byte(skill.Plus)
	resp.Insert([]byte{0, 0, 0}, 15) //

	skills.SetSkills(skillSlots)
	skills.Update()

	if info.Passive {
		statData, err := c.GetStats()
		if err == nil {
			resp.Concat(statData)
		}
	}

	resp.Concat(c.GetExpAndSkillPts())
	return resp, nil
}
func (c *Character) DivineUpgradeSkills(skillIndex, slot int, bookID int64) ([]byte, error) {
	skills, err := FindSkillsByID(c.ID)
	if err != nil {
		return nil, err
	}

	skillSlots, err := skills.GetSkills()
	if err != nil {
		return nil, err
	}
	resp := utils.Packet{}
	//divineID := 0
	bonusPlus := 0
	usedPoints := 0
	for _, skill := range skillSlots.Slots {
		if skill.BookID == bookID {
			if len(skill.DivinePoints) == 0 {
				divtuple := &DivineTuple{DivineID: 0, DivinePlus: 0}
				div2tuple := &DivineTuple{DivineID: 1, DivinePlus: 0}
				div3tuple := &DivineTuple{DivineID: 2, DivinePlus: 0}
				skill.DivinePoints = append(skill.DivinePoints, divtuple, div2tuple, div3tuple)
				skills.SetSkills(skillSlots)
				skills.Update()
			}
			for _, point := range skill.DivinePoints {
				usedPoints += point.DivinePlus
				//if point.DivineID == slot {
				if usedPoints >= 10 {
					return nil, nil
				}
				//	divineID = point.DivineID
				if point.DivineID == slot {
					bonusPlus = point.DivinePlus
				}
			}
			skill.DivinePoints[slot].DivinePlus++
		}
	}
	bonusPlus++
	resp = DIVINE_SKILL_BOOk
	resp[8] = byte(skillIndex)
	index := 9
	resp.Insert([]byte{byte(slot)}, index) // divine id
	index++
	resp.Insert(utils.IntToBytes(uint64(bookID), 4, true), index) // book id
	index += 4
	resp.Insert([]byte{byte(bonusPlus)}, index) // divine plus
	index++
	skills.SetSkills(skillSlots)
	skills.Update()
	return resp, nil
}

func (c *Character) RemoveSkill(slotIndex byte, bookID int64) ([]byte, error) {
	skills, err := FindSkillsByID(c.ID)
	if err != nil {
		return nil, err
	}

	skillSlots, err := skills.GetSkills()
	if err != nil {
		return nil, err
	}

	set := skillSlots.Slots[slotIndex]
	if set.BookID != bookID {
		return nil, fmt.Errorf("RemoveSkill: skill book not found")
	}

	skillSlots.Slots[slotIndex] = &SkillSet{}
	skills.SetSkills(skillSlots)
	skills.Update()

	resp := SKILL_REMOVED
	resp[8] = slotIndex
	resp.Insert(utils.IntToBytes(uint64(bookID), 4, true), 9) // book id

	return resp, nil
}

func (c *Character) UpgradePassiveSkill(slotIndex, skillIndex byte) ([]byte, error) {
	skills, err := FindSkillsByID(c.ID)
	if err != nil {
		return nil, err
	}

	skillSlots, err := skills.GetSkills()
	if err != nil {
		return nil, err
	}

	set := skillSlots.Slots[skillIndex]
	if len(set.Skills) == 0 || set.Skills[0].Plus >= 12 {
		return nil, nil
	}

	if skillIndex == 5 || skillIndex == 6 { // 1st job skill
		requiredSP := SkillPTS["fjp"][set.Skills[0].Plus]
		if skills.SkillPoints < requiredSP {
			return nil, nil
		}

		skills.SkillPoints -= requiredSP

	} else if skillIndex == 7 { // running
		requiredSP := SkillPTS["wd"][set.Skills[0].Plus]
		if skills.SkillPoints < requiredSP {
			return nil, nil
		}

		skills.SkillPoints -= requiredSP
		c.RunningSpeed = 10.0 + (float64(set.Skills[0].Plus) * 0.2)
	}

	set.Skills[0].Plus++

	skills.SetSkills(skillSlots)
	skills.Update()

	resp := PASSIVE_SKILL_UGRADED
	resp[8] = slotIndex
	resp[9] = byte(set.Skills[0].Plus)

	statData, err := c.GetStats()
	if err != nil {
		return nil, err
	}

	resp.Concat(statData)
	resp.Concat(c.GetExpAndSkillPts())
	return resp, nil
}

func (c *Character) DowngradePassiveSkill(slotIndex, skillIndex byte) ([]byte, error) {
	skills, err := FindSkillsByID(c.ID)
	if err != nil {
		return nil, err
	}

	skillSlots, err := skills.GetSkills()
	if err != nil {
		return nil, err
	}

	set := skillSlots.Slots[skillIndex]
	if len(set.Skills) == 0 || set.Skills[0].Plus <= 0 {
		return nil, nil
	}

	if skillIndex == 5 && set.Skills[0].Plus > 0 { // 1st job skill
		//requiredSP := SkillPTS["fjp"][set.Skills[0].Plus-1]

		//skills.SkillPoints += requiredSP

	} else if skillIndex == 7 && set.Skills[0].Plus > 0 { // running
		//requiredSP := SkillPTS["wd"][set.Skills[0].Plus]

		//skills.SkillPoints += requiredSP
		c.RunningSpeed = 10.0 + (float64(set.Skills[0].Plus-1) * 0.2)
	}

	set.Skills[0].Plus--

	skills.SetSkills(skillSlots)
	skills.Update()

	resp := PASSIVE_SKILL_UGRADED
	resp[8] = slotIndex
	resp[9] = byte(set.Skills[0].Plus)

	statData, err := c.GetStats()
	if err != nil {
		return nil, err
	}

	resp.Concat(statData)
	resp.Concat(c.GetExpAndSkillPts())
	return resp, nil
}

func (c *Character) RemovePassiveSkill(slotIndex, skillIndex byte, bookID int64) ([]byte, error) {
	skills, err := FindSkillsByID(c.ID)
	if err != nil {
		return nil, err
	}

	skillSlots, err := skills.GetSkills()
	if err != nil {
		return nil, err
	}

	set := skillSlots.Slots[skillIndex]
	fmt.Println(set, bookID)
	if set.BookID != bookID {
		return nil, fmt.Errorf("RemovePassiveSkill: skill book not found")
	}

	skillSlots.Slots[skillIndex] = &SkillSet{}
	skills.SetSkills(skillSlots)
	skills.Update()

	resp := PASSIVE_SKILL_REMOVED
	resp.Insert(utils.IntToBytes(uint64(bookID), 4, true), 8) // book id
	resp[12] = slotIndex

	return resp, nil
}

func (c *Character) CastSkill(attackCounter, skillID, targetID int, cX, cY, cZ float64) ([]byte, error) {

	slots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	petSlot := slots[0x0A]
	pet := petSlot.Pet
	petInfo, ok := Pets[petSlot.ItemID]
	if pet != nil && ok && pet.IsOnline && !petInfo.Combat {
		return nil, nil
	}

	stat := c.Socket.Stats
	user := c.Socket.User
	skills := c.Socket.Skills

	canCast := true
	skillInfo := SkillInfos[skillID]
	weapon := slots[c.WeaponSlot]
	if weapon.ItemID == 0 { // there are some skills which can be casted without weapon such as monk skills
		if c.Type == MONK || c.Type == DIVINE_MONK || c.Type == DARKNESS_MONK {
			canCast = true
		}
		if weapon.ItemID == 0 { // there are some skills which can be casted without weapon such as beast king skills
			if c.Type == BEAST_KING || c.Type == DIVINE_BEAST_KING || c.Type == DARKNESS_BEAST_KING {
				canCast = true
			}

			if weapon.ItemID == 0 { // there are some skills which can be casted without weapon such as empress skills
				if c.Type == EMPRESS || c.Type == DIVINE_EMPRESS || c.Type == DARKNESS_EMPRESS {
					canCast = true
				}
			}
		}

		if c.Type == BEAST_KING || c.Type == DIVINE_BEAST_KING || c.Type == DARKNESS_BEAST_KING {
			ok := funk.Contains(Beast_King_Infections, int16(skillInfo.InfectionID))
			if ok {
				for _, buffid := range Beast_King_Infections {
					buff, err := FindBuffByID(int(buffid), c.ID)
					if buff != nil && err == nil {
						buff.Delete()
					}
				}

			}
		}
		if c.Type == EMPRESS || c.Type == DIVINE_EMPRESS || c.Type == DARKNESS_EMPRESS {
			ok := funk.Contains(Empress_Infections, int16(skillInfo.InfectionID))
			if ok {
				for _, buffid := range Empress_Infections {
					buff, err := FindBuffByID(int(buffid), c.ID)
					if buff != nil && err == nil {
						buff.Delete()
					}
				}

			}
		}
	} else {
		weaponInfo := Items[weapon.ItemID]
		canCast = weaponInfo.CanUse(skillInfo.Type)
	}
	if !canCast {
		return nil, nil
	}

	plus, err := skills.GetPlus(skillID)
	if err != nil {
		return nil, err
	}
	skillSlots, err := c.Socket.Skills.GetSkills()
	if err != nil {
		return nil, err
	}
	plusCooldown := 0
	plusChiCost := 0
	divinePlus := 0
	for _, slot := range skillSlots.Slots {
		if slot.BookID == skillInfo.BookID {
			for _, points := range slot.DivinePoints {
				if points.DivineID == 0 && points.DivinePlus > 0 {
					divinePlus = points.DivinePlus
					plusChiCost = 50
				}
				if points.DivinePlus == 2 && points.DivinePlus > 0 {
					plusCooldown = 100
				}
			}
		}
	}
	t := c.SkillHistory.Get(skillID)
	if t != nil {
		castedAt := t.(time.Time)
		cooldown := time.Duration(skillInfo.Cooldown*1000) * time.Millisecond
		cooldown -= time.Duration(plusCooldown * divinePlus) //plusCooldown * divinePlus
		if time.Now().Sub(castedAt) < cooldown {
			return nil, nil
		}
	}

	c.SkillHistory.Add(skillID, time.Now())

	addChiCost := float64(skillInfo.AdditionalChi*int(plus)) * 2.2 / 3 // some bad words here
	chiCost := skillInfo.BaseChi + int(addChiCost) - (plusChiCost * divinePlus)
	if stat.CHI < chiCost {
		return nil, nil
	}

	stat.CHI -= chiCost

	resp := utils.Packet{}
	if target := skillInfo.Target; target == 0 || target == 2 { // buff skill
		character := c
		if target == 2 {
			ch := FindCharacterByPseudoID(c.Socket.User.ConnectedServer, uint16(c.Selection))
			if ch != nil {
				character = ch
			}
		}

		infection := BuffInfections[skillInfo.InfectionID]
		duration := (skillInfo.BaseTime + skillInfo.AdditionalTime*int(plus)) / 10

		if character != c && !c.CanAttack(character) { //target other player
			c.DealInfection(nil, character, skillID)
		} else if character != c && c.CanAttack(character) { //target other player but is in pvp
			c.DealInfection(nil, c, skillID)
		}
		expire := true
		if skillInfo.InfectionID != 0 && duration == 0 {
			expire = false
		}
		buff, err := FindBuffByID(infection.ID, character.ID)
		if err != nil {
			return nil, err
		} else if buff != nil {
			buff.Delete()
			stat, err = FindStatByID(character.ID)
			if err != nil {
				return nil, err
			}

			err = stat.Calculate()
			if err != nil {
				return nil, err
			}
			c.HandleBuffs()
			if infection.IsPercent == false {
				buff = &Buff{ID: infection.ID, CharacterID: character.ID, StartedAt: character.Epoch, Duration: int64(duration), Name: skillInfo.Name,
					ATK: infection.BaseATK + infection.AdditionalATK*int(plus), ArtsATK: infection.BaseArtsATK + infection.AdditionalArtsATK*int(plus),
					ArtsDEF: infection.ArtsDEF + infection.AdditionalArtsDEF*int(plus), ConfusionDEF: infection.ConfusionDef + infection.AdditionalConfusionDef*int(plus),
					DEF: infection.BaseDef + infection.AdditionalDEF*int(plus), DEX: infection.DEX + infection.AdditionalDEX*int(plus), HPRecoveryRate: infection.HPRecoveryRate + infection.AdditionalHPRecovery*int(plus), INT: infection.INT + infection.AdditionalINT*int(plus),
					MaxHP: infection.MaxHP + infection.AdditionalHP*int(plus), ParalysisDEF: infection.ParalysisDef + infection.AdditionalParalysisDef*int(plus), PoisonDEF: infection.ParalysisDef + infection.AdditionalPoisonDef*int(plus), STR: infection.STR + infection.AdditionalSTR*int(plus),
					Accuracy: infection.Accuracy + infection.AdditionalAccuracy*int(plus), Dodge: infection.DodgeRate + infection.AdditionalDodgeRate*int(plus), RunningSpeed: infection.MovSpeed + infection.AdditionalMovSpeed*float64(plus), CanExpire: expire}
				buff.Create()
			} else {
				percentArtsDEF := int(float64(character.Socket.Stats.ArtsDEF) * (float64(infection.ArtsDEF+infection.AdditionalArtsDEF*int(plus)) / 1000))
				percentDEF := int(float64(character.Socket.Stats.DEF) * (float64(infection.BaseDef+infection.AdditionalDEF*int(plus)) / 1000))
				percentATK := int(float64(character.Socket.Stats.MinATK) * (float64(infection.BaseATK+infection.AdditionalATK*int(plus)) / 1000))
				percentArtsATK := int(float64(character.Socket.Stats.MinArtsATK) * (float64(infection.BaseArtsATK+infection.AdditionalArtsATK*int(plus)) / 1000))
				buff = &Buff{ID: infection.ID, CharacterID: character.ID, StartedAt: character.Epoch, Duration: int64(duration), Name: skillInfo.Name,
					ATK: percentATK, ArtsATK: percentArtsATK,
					ArtsDEF: percentArtsDEF, ConfusionDEF: infection.ConfusionDef + infection.AdditionalConfusionDef*int(plus),
					DEF: percentDEF, DEX: infection.DEX + infection.AdditionalDEX*int(plus), HPRecoveryRate: infection.HPRecoveryRate + infection.AdditionalHPRecovery*int(plus), INT: infection.INT + infection.AdditionalINT*int(plus),
					MaxHP: infection.MaxHP + infection.AdditionalHP*int(plus), ParalysisDEF: infection.ParalysisDef + infection.AdditionalParalysisDef*int(plus), PoisonDEF: infection.ParalysisDef + infection.AdditionalPoisonDef*int(plus), STR: infection.STR + infection.AdditionalSTR*int(plus),
					Accuracy: infection.Accuracy + infection.AdditionalAccuracy*int(plus), Dodge: infection.DodgeRate + infection.AdditionalDodgeRate*int(plus), CanExpire: expire}
				buff.Create()
			}
			buff.Update()

		} else if buff == nil {
			c.HandleBuffs()
			if infection.IsPercent == false {
				buff = &Buff{ID: infection.ID, CharacterID: character.ID, StartedAt: character.Epoch, Duration: int64(duration), Name: skillInfo.Name,
					ATK: infection.BaseATK + infection.AdditionalATK*int(plus), ArtsATK: infection.BaseArtsATK + infection.AdditionalArtsATK*int(plus),
					ArtsDEF: infection.ArtsDEF + infection.AdditionalArtsDEF*int(plus), ConfusionDEF: infection.ConfusionDef + infection.AdditionalConfusionDef*int(plus),
					DEF: infection.BaseDef + infection.AdditionalDEF*int(plus), DEX: infection.DEX + infection.AdditionalDEX*int(plus), HPRecoveryRate: infection.HPRecoveryRate + infection.AdditionalHPRecovery*int(plus), INT: infection.INT + infection.AdditionalINT*int(plus),
					MaxHP: infection.MaxHP + infection.AdditionalHP*int(plus), ParalysisDEF: infection.ParalysisDef + infection.AdditionalParalysisDef*int(plus), PoisonDEF: infection.ParalysisDef + infection.AdditionalPoisonDef*int(plus), STR: infection.STR + infection.AdditionalSTR*int(plus),
					Accuracy: infection.Accuracy + infection.AdditionalAccuracy*int(plus), Dodge: infection.DodgeRate + infection.AdditionalDodgeRate*int(plus), RunningSpeed: infection.MovSpeed + infection.AdditionalMovSpeed*float64(plus), CanExpire: expire}
			} else {
				percentArtsDEF := int(float64(character.Socket.Stats.ArtsDEF) * (float64(infection.ArtsDEF+infection.AdditionalArtsDEF*int(plus)) / 1000))
				percentDEF := int(float64(character.Socket.Stats.DEF) * (float64(infection.BaseDef+infection.AdditionalDEF*int(plus)) / 1000))
				percentATK := int(float64(character.Socket.Stats.MinATK) * (float64(infection.BaseATK+infection.AdditionalATK*int(plus)) / 1000))
				percentArtsATK := int(float64(character.Socket.Stats.MinArtsATK) * (float64(infection.BaseArtsATK+infection.AdditionalArtsATK*int(plus)) / 1000))
				buff = &Buff{ID: infection.ID, CharacterID: character.ID, StartedAt: character.Epoch, Duration: int64(duration), Name: skillInfo.Name,
					ATK: percentATK, ArtsATK: percentArtsATK,
					ArtsDEF: percentArtsDEF, ConfusionDEF: infection.ConfusionDef + infection.AdditionalConfusionDef*int(plus),
					DEF: percentDEF, DEX: infection.DEX + infection.AdditionalDEX*int(plus), HPRecoveryRate: infection.HPRecoveryRate + infection.AdditionalHPRecovery*int(plus), INT: infection.INT + infection.AdditionalINT*int(plus),
					MaxHP: infection.MaxHP + infection.AdditionalHP*int(plus), ParalysisDEF: infection.ParalysisDef + infection.AdditionalParalysisDef*int(plus), PoisonDEF: infection.ParalysisDef + infection.AdditionalPoisonDef*int(plus), STR: infection.STR + infection.AdditionalSTR*int(plus),
					Accuracy: infection.Accuracy + infection.AdditionalAccuracy*int(plus), Dodge: infection.DodgeRate + infection.AdditionalDodgeRate*int(plus), CanExpire: expire}
			}
			err := buff.Create()
			if err != nil {
				fmt.Println("Buff error: ", err)
			}
		}

		if buff.ID == 241 || buff.ID == 244 { // invisibility
			time.AfterFunc(time.Second*1, func() {
				if character != nil {
					character.Invisible = true
				}
			})
		} else if buff.ID == 242 || buff.ID == 245 { // detection arts
			character.DetectionMode = true
		}

		statData, _ := character.GetStats()
		character.Socket.Write(statData)

		p := &nats.CastPacket{CastNear: true, CharacterID: character.ID, Data: character.GetHPandChi()}
		p.Cast()

	} else { // combat skill
		target := GetFromRegister(user.ConnectedServer, c.Map, uint16(targetID))
		if ai, ok := target.(*AI); ok { // attacked to ai

			pos := NPCPos[ai.PosID]
			npc := NPCs[pos.NPCID]

			for _, factionNPC := range ZhuangFactionMobs {
				if factionNPC == npc.ID && c.Faction == 1 {
					goto OUT
				}
			}
			for _, factionNPC := range ShaoFactionMobs {
				if factionNPC == npc.ID && c.Faction == 2 {
					goto OUT
				}
			}

			if skillID == 41201 || skillID == 41301 { // howl of tame
				c.TamingAI = ai
				goto OUT
			}

			if pos.Attackable { // target is attackable
				castLocation := ConvertPointToLocation(c.Coordinate)
				if skillInfo.AreaCenter == 1 || skillInfo.AreaCenter == 2 {
					castLocation = ConvertPointToLocation(ai.Coordinate)
				}
				skillSlots, err := c.Socket.Skills.GetSkills()
				if err != nil {
					return nil, err
				}
				plusRange := 0.0
				divinePlus := 0
				plusDamage := 0
				for _, slot := range skillSlots.Slots {
					if slot.BookID == skillInfo.BookID {
						for _, points := range slot.DivinePoints {
							if points.DivineID == 2 && points.DivinePlus > 0 {
								divinePlus = points.DivinePlus
								plusRange = 0.2
							}
							if points.DivineID == 1 && points.DivinePlus > 0 {
								divinePlus = points.DivinePlus
								plusDamage = 100
							}
						}
					}
				}
				castRange := skillInfo.BaseRadius + skillInfo.AdditionalRadius*float64(plus+0) + (float64(plusRange) * float64(divinePlus))
				candidates := AIsByMap[ai.Server][ai.Map]

				candidates = funk.Filter(candidates, func(cand *AI) bool {
					nPos := NPCPos[cand.PosID]
					if nPos == nil {
						return false
					}

					aiCoordinate := ConvertPointToLocation(cand.Coordinate)
					return (cand.PseudoID == ai.PseudoID || (utils.CalculateDistance(aiCoordinate, castLocation) < castRange)) && cand.HP > 0 && nPos.Attackable
				}).([]*AI)

				for _, mob := range candidates {
					dmg, _ := c.CalculateDamage(mob, true)
					dmg += plusDamage * divinePlus
					if skillInfo.InfectionID != 0 && skillInfo.Target == 1 {
						c.Targets = append(c.Targets, &Target{Damage: dmg, AI: mob, SkillId: skillID})
					}
					c.Targets = append(c.Targets, &Target{Damage: dmg, AI: mob, SkillId: skillID})
				}

			} else { // target is not attackable
				if funk.Contains(miningSkills, skillID) { // mining skill
					c.Targets = []*Target{{Damage: 10, AI: ai}}
				}
			}

		} else { // FIX => attacked to player
			enemy := FindCharacterByPseudoID(user.ConnectedServer, uint16(targetID))
			if enemy != nil && enemy.IsActive {
				dmg, _ := c.CalculateDamageToPlayer(enemy, true)
				c.PlayerTargets = append(c.PlayerTargets, &PlayerTarget{Damage: dmg, Enemy: enemy})
			}
		}
	}

OUT:
	r := SKILL_CASTED
	r.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 7) // character pseudo id
	r[9] = byte(attackCounter)
	r.Insert(utils.IntToBytes(uint64(skillID), 4, true), 10)  // skill id
	r.Insert(utils.FloatToBytes(cX, 4, true), 14)             // coordinate-x
	r.Insert(utils.FloatToBytes(cY, 4, true), 18)             // coordinate-y
	r.Insert(utils.FloatToBytes(cZ, 4, true), 22)             // coordinate-z
	r.Insert(utils.IntToBytes(uint64(targetID), 2, true), 27) // target id
	r.Insert(utils.IntToBytes(uint64(targetID), 2, true), 30) // target id

	p := nats.CastPacket{CastNear: true, CharacterID: c.ID, Type: nats.CAST_SKILL, Data: r}
	if err = p.Cast(); err != nil {
		return nil, err
	}

	resp.Concat(r)
	resp.Concat(c.GetHPandChi())

	//go stat.Update()
	return resp, nil

}

func (c *Character) ApplyBuffEffectToAi(ai *AI, buffid int, dps int, duration int64) {

	s := c.Socket
	buff, err := FindBuffByAIID(buffid, int(ai.PseudoID))
	if buff == nil && err == nil {
		now := time.Now()
		secs := now.Unix()
		infection := BuffInfections[buffid]
		ai, _ := GetFromRegister(s.User.ConnectedServer, s.Character.Map, uint16(s.Character.Selection)).(*AI)
		buff := &AiBuff{ID: buffid, AiID: int(ai.PseudoID), Name: infection.Name, HPRecoveryRate: dps, StartedAt: secs, CharacterID: s.Character.ID, Duration: duration}
		err = buff.Create()
		if err != nil {
			fmt.Println(fmt.Sprintf("Error: %s", err.Error()))
			return
		}

		index := 5
		r := DEAL_BUFF_AI
		r.Insert(utils.IntToBytes(uint64(ai.PseudoID), 2, true), index) // ai pseudo id
		index += 2
		r.Insert(utils.IntToBytes(uint64(ai.PseudoID), 2, true), index) // ai pseudo id
		index += 2
		r.Insert(utils.IntToBytes(uint64(ai.HP), 4, true), index) // ai current hp
		index += 4
		r.Insert(utils.IntToBytes(uint64(ai.CHI), 4, true), index) // ai current chi
		r.Insert(utils.IntToBytes(uint64(buffid), 4, true), 22)    //BUFF ID
		c.Socket.Write(r)
	}
}

func (c *Character) DealInfection(ai *AI, character *Character, skillID int) {
	skillInfo := SkillInfos[skillID]
	if skillInfo.InfectionID == 0 {
		return
	}
	infection := BuffInfections[skillInfo.InfectionID]

	skills := c.Socket.Skills
	plus, err := skills.GetPlus(skillID)
	if err != nil {
		return
	}

	duration := (skillInfo.BaseTime + skillInfo.AdditionalTime*int(plus)) / 10

	if ai != nil { //AI BUFF
		c.ApplyBuffEffectToAi(ai, int(infection.ID), 10, int64(duration))
	} else if character != nil { //PLAYER BUFF ADD

		if infection.ID == 66 {
			statEnemy := character.Socket.Stats
			r := MOB_DEAL_DAMAGE
			r.Insert(utils.IntToBytes(uint64(character.PseudoID), 2, true), 5) // character pseudo id
			r.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 7)         // mob pseudo id
			r.Insert(utils.IntToBytes(uint64(statEnemy.HP), 4, true), 9)       // character hp
			r.Insert(utils.IntToBytes(uint64(statEnemy.CHI), 4, true), 13)     // character chi

			r.Concat(character.GetHPandChi())
			p := &nats.CastPacket{CastNear: true, CharacterID: character.ID, Data: r, Type: nats.PLAYER_ATTACK}
			if err := p.Cast(); err != nil {
				log.Println("deal damage broadcast error:", err)
				return
			}
			return

		}

		buff, err := FindBuffByID(infection.ID, character.ID)
		if err != nil {
			return
		} else if buff != nil {
			buff = &Buff{ID: infection.ID, CharacterID: character.ID, StartedAt: character.Epoch, Duration: int64(duration), Name: skillInfo.Name,
				ATK: infection.BaseATK + infection.AdditionalATK*int(plus), ArtsATK: infection.BaseArtsATK + infection.AdditionalArtsATK*int(plus),
				ArtsDEF: infection.ArtsDEF + infection.AdditionalArtsDEF*int(plus), ConfusionDEF: infection.ConfusionDef,
				DEF: infection.BaseDef + infection.AdditionalDEF*int(plus), DEX: infection.DEX, HPRecoveryRate: infection.HPRecoveryRate, INT: infection.INT,
				MaxHP: infection.MaxHP, ParalysisDEF: infection.ParalysisDef, PoisonDEF: infection.ParalysisDef, STR: infection.STR, Accuracy: infection.Accuracy * int(plus), Dodge: infection.DodgeRate * int(plus), CanExpire: true}
			buff.Update()

		} else if buff == nil {

			buff = &Buff{ID: infection.ID, CharacterID: character.ID, StartedAt: character.Epoch, Duration: int64(duration), Name: skillInfo.Name,
				ATK: infection.BaseATK + infection.AdditionalATK*int(plus), ArtsATK: infection.BaseArtsATK + infection.AdditionalArtsATK*int(plus),
				ArtsDEF: infection.ArtsDEF + infection.AdditionalArtsDEF*int(plus), ConfusionDEF: infection.ConfusionDef,
				DEF: infection.BaseDef + infection.AdditionalDEF*int(plus), DEX: infection.DEX, HPRecoveryRate: infection.HPRecoveryRate, INT: infection.INT,
				MaxHP: infection.MaxHP, ParalysisDEF: infection.ParalysisDef, PoisonDEF: infection.ParalysisDef, STR: infection.STR, Accuracy: infection.Accuracy * int(plus), Dodge: infection.DodgeRate * int(plus), CanExpire: true}
			buff.Create()
		}
		if buff.ID == 70 || buff.ID == 73 || buff.ID == 58 { // invisibility
			time.AfterFunc(time.Second*1, func() {
				if character != nil {
					character.Invisible = true
				}
			})
		} else if buff.ID == 242 || buff.ID == 245 { // detection arts
			character.DetectionMode = true
		}

		statData, _ := character.GetStats()
		character.Socket.Write(statData)

		p := &nats.CastPacket{CastNear: true, CharacterID: character.ID, Data: character.GetHPandChi()}
		p.Cast()
	}
}

func (c *Character) CalculateDamage(ai *AI, isSkill bool) (int, error) {

	st := c.Socket.Stats

	npcPos := NPCPos[ai.PosID]
	npc := NPCs[npcPos.NPCID]

	def, min, max := npc.DEF, st.MinATK, st.MaxATK
	if isSkill {
		def, min, max = npc.ArtsDEF, st.MinArtsATK, st.MaxArtsATK
	}

	dmg := int(utils.RandInt(int64(min), int64(max))) - def
	if dmg < 3 {
		dmg = 3
	} else if dmg > ai.HP {
		dmg = ai.HP
	}

	if diff := int(npc.Level) - c.Level; diff > 0 {
		reqAcc := utils.SigmaFunc(float64(diff))
		if float64(st.Accuracy) < reqAcc {
			probability := float64(st.Accuracy) * 1000 / reqAcc
			if utils.RandInt(0, 1000) > int64(probability) {
				dmg = 0
			}
		}
	}

	return dmg, nil
}

func (c *Character) CalculateDamageToPlayer(enemy *Character, isSkill bool) (int, error) {
	st := c.Socket.Stats
	enemySt := enemy.Socket.Stats

	def, min, max := enemySt.DEF, st.MinATK, st.MaxATK
	if isSkill {
		def, min, max = enemySt.ArtsDEF, st.MinArtsATK, st.MaxArtsATK
	}

	def = utils.PvPFunc(def)

	dmg := int(utils.RandInt(int64(min), int64(max))) - def
	if dmg < 0 {
		dmg = 3
	} else if dmg > enemySt.HP {
		dmg = enemySt.HP
	}

	reqAcc := float64(enemySt.Dodge) - float64(st.Accuracy) + float64(c.Level-int(enemy.Level))*10
	if utils.RandInt(0, 2000) < int64(reqAcc) {
		dmg = 0
	}

	return dmg, nil
}

func (c *Character) CancelTrade() {

	trade := FindTrade(c)
	if trade == nil {
		return
	}

	receiver, sender := trade.Receiver.Character, trade.Sender.Character
	trade.Delete()

	resp := TRADE_CANCELLED
	sender.Socket.Write(resp)
	receiver.Socket.Write(resp)
}

func (c *Character) OpenSale(name string, slotIDs []int16, prices []uint64) ([]byte, error) {

	slots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	sale := &Sale{ID: c.PseudoID, Seller: c, Name: name}
	for i := 0; i < len(slotIDs); i++ {
		slotID := slotIDs[i]
		price := prices[i]
		item := slots[slotID]

		info := Items[item.ItemID]

		if slotID == 0 || price == 0 || item == nil || item.ItemID == 0 || !info.Tradable {
			continue
		}

		saleItem := &SaleItem{SlotID: slotID, Price: price, IsSold: false}
		sale.Items = append(sale.Items, saleItem)
	}

	sale.Data, err = sale.SaleData()
	if err != nil {
		return nil, err
	}

	sale.Create()

	resp := OPEN_SALE
	spawnData, err := c.SpawnCharacter()
	if err == nil {
		p := nats.CastPacket{CastNear: true, CharacterID: c.ID, Type: nats.PLAYER_SPAWN, Data: spawnData}
		p.Cast()
		resp.Concat(spawnData)
	}

	return resp, nil

}

func FindSaleVisitors(saleID uint16) []*Character {

	characterMutex.RLock()
	allChars := funk.Values(characters)
	characterMutex.RUnlock()

	return funk.Filter(allChars, func(c *Character) bool {
		return c.IsOnline && c.VisitedSaleID == saleID
	}).([]*Character)
}

func (c *Character) CloseSale() ([]byte, error) {
	sale := FindSale(c.PseudoID)
	if sale != nil {
		sale.Delete()
		resp := CLOSE_SALE

		spawnData, err := c.SpawnCharacter()
		if err == nil {
			p := nats.CastPacket{CastNear: true, CharacterID: c.ID, Type: nats.PLAYER_SPAWN, Data: spawnData}
			p.Cast()
			resp.Concat(spawnData)
		}

		return resp, nil
	}

	return nil, nil
}

func (c *Character) BuySaleItem(saleID uint16, saleSlotID, inventorySlotID int16) ([]byte, error) {
	sale := FindSale(saleID)
	if sale == nil {
		return nil, nil
	}

	mySlots, err := c.InventorySlots()
	if err != nil {
		return nil, err
	}

	seller := sale.Seller
	slots, err := seller.InventorySlots()
	if err != nil {
		return nil, err
	}

	saleItem := sale.Items[saleSlotID]
	if saleItem == nil || saleItem.IsSold {
		return nil, nil
	}

	item := slots[saleItem.SlotID]
	if item == nil || item.ItemID == 0 || c.Gold < saleItem.Price {
		return nil, nil
	}

	c.LootGold(-saleItem.Price)
	seller.Gold += saleItem.Price

	resp := BOUGHT_SALE_ITEM
	resp.Insert(utils.IntToBytes(c.Gold, 8, true), 8)                   // buyer gold
	resp.Insert(utils.IntToBytes(uint64(item.ItemID), 4, true), 17)     // sale item id
	resp.Insert(utils.IntToBytes(uint64(item.Quantity), 2, true), 23)   // sale item quantity
	resp.Insert(utils.IntToBytes(uint64(inventorySlotID), 2, true), 25) // inventory slot id
	resp.Insert(item.GetUpgrades(), 27)                                 // sale item upgrades
	resp[42] = byte(item.SocketCount)                                   // item socket count
	resp.Insert(item.GetSockets(), 43)                                  // sale item sockets

	myItem := NewSlot()
	*myItem = *item
	myItem.CharacterID = null.IntFrom(int64(c.ID))
	myItem.UserID = null.StringFrom(c.UserID)
	myItem.SlotID = int16(inventorySlotID)
	mySlots[inventorySlotID] = myItem
	myItem.Update()
	InventoryItems.Add(myItem.ID, myItem)

	resp.Concat(item.GetData(inventorySlotID))
	c.Socket.Write(resp)
	logger.Log(logging.ACTION_BUY_SALE_ITEM, c.ID, fmt.Sprintf("Bought sale item (%d) with %d gold from seller (%d)", myItem.ID, saleItem.Price, seller.ID), c.UserID)
	saleItem.IsSold = true

	sellerResp := SOLD_SALE_ITEM
	sellerResp.Insert(utils.IntToBytes(uint64(saleSlotID), 2, true), 8)  // sale slot id
	sellerResp.Insert(utils.IntToBytes(seller.Gold, 8, true), 10)        // seller gold
	sellerResp.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 18) // buyer pseudo id

	*item = *NewSlot()
	sellerResp.Concat(item.GetData(saleItem.SlotID))

	remainingCount := len(funk.Filter(sale.Items, func(i *SaleItem) bool {
		return i.IsSold == false
	}).([]*SaleItem))
	seller.Socket.Write(sellerResp)
	if remainingCount > 0 {
		sale.Data, _ = sale.SaleData()
		resp.Concat(sale.Data)

	} /*else {
		sale.Delete()

		spawnData, err := seller.SpawnCharacter()
		if err == nil {
			p := nats.CastPacket{CastNear: true, CharacterID: seller.ID, Type: nats.PLAYER_SPAWN, Data: spawnData}
			p.Cast()
			resp.Concat(spawnData)
		}

		visitors := FindSaleVisitors(sale.ID)
		for _, v := range visitors {
			v.Socket.Write(CLOSE_SALE)
			v.VisitedSaleID = 0
		}

		//resp.Concat(CLOSE_SALE)
		sellerResp.Concat(CLOSE_SALE)
	}*/

	return resp, nil
}

func (c *Character) UpdatePartyStatus() {

	user := c.Socket.User
	stat := c.Socket.Stats

	party := FindParty(c)
	if party == nil {
		return
	}

	coordinate := ConvertPointToLocation(c.Coordinate)

	resp := PARTY_STATUS
	resp.Insert(utils.IntToBytes(uint64(c.ID), 4, true), 6)             // character id
	resp.Insert(utils.IntToBytes(uint64(stat.HP), 4, true), 10)         // character hp
	resp.Insert(utils.IntToBytes(uint64(stat.MaxHP), 4, true), 14)      // character max hp
	resp.Insert(utils.FloatToBytes(float64(coordinate.X), 4, true), 19) // coordinate-x
	resp.Insert(utils.FloatToBytes(float64(coordinate.Y), 4, true), 23) // coordinate-y
	resp.Insert(utils.IntToBytes(uint64(stat.CHI), 4, true), 27)        // character chi
	resp.Insert(utils.IntToBytes(uint64(stat.MaxCHI), 4, true), 31)     // character max chi
	resp.Insert(utils.IntToBytes(uint64(c.Level), 4, true), 35)         // character level
	resp[39] = byte(c.Type)                                             // character type
	resp[41] = byte(user.ConnectedServer - 1)                           // connected server id

	members := party.GetMembers()
	members = funk.Filter(members, func(m *PartyMember) bool {
		return m.Accepted
	}).([]*PartyMember)

	party.Leader.Socket.Write(resp)
	for _, m := range members {
		m.Socket.Write(resp)
	}
}

func (c *Character) LeaveParty() {

	party := FindParty(c)
	if party == nil {
		return
	}

	c.PartyID = ""

	members := party.GetMembers()
	members = funk.Filter(members, func(m *PartyMember) bool {
		return m.Accepted
	}).([]*PartyMember)

	resp := utils.Packet{}
	if c.ID == party.Leader.ID { // disband party
		resp = PARTY_DISBANDED
		party.Leader.Socket.Write(resp)

		for _, member := range members {
			member.PartyID = ""
			member.Socket.Write(resp)
		}

		party.Delete()

	} else { // leave party
		member := party.GetMember(c.ID)
		party.RemoveMember(member)

		resp = LEFT_PARTY
		resp.Insert(utils.IntToBytes(uint64(c.ID), 4, true), 8) // character id

		leader := party.Leader
		if len(party.GetMembers()) == 0 {
			leader.PartyID = ""
			resp.Concat(PARTY_DISBANDED)
			party.Delete()

		}

		leader.Socket.Write(resp)
		for _, m := range members {
			m.Socket.Write(resp)
		}

	}
}

func (c *Character) GetGuildData() ([]byte, error) {

	if c.GuildID > 0 {
		guild, err := FindGuildByID(c.GuildID)
		if err != nil {
			return nil, err
		} else if guild == nil {
			return nil, nil
		}

		return guild.GetData(c)
	}

	return nil, nil
}

func (c *Character) JobPassives(stat *Stat) error {

	//stat := c.Socket.Stats
	skills, err := FindSkillsByID(c.ID)
	if err != nil {
		return err
	}

	skillSlots, err := skills.GetSkills()
	if err != nil {
		return err
	}

	if passive := skillSlots.Slots[5]; passive.BookID > 0 {
		info := JobPassives[int8(c.Class)]
		if info != nil {
			plus := passive.Skills[0].Plus
			stat.MaxHP += info.MaxHp * plus
			stat.MaxCHI += info.MaxChi * plus
			stat.MinATK += info.ATK * plus
			stat.MaxATK += info.ATK * plus
			stat.MinArtsATK += info.ArtsATK * plus
			stat.MaxArtsATK += info.ArtsATK * plus
			stat.DEF += info.DEF * plus
			stat.ArtsDEF += info.ArtsDef * plus
			stat.Accuracy += info.Accuracy * plus
			stat.Dodge += info.Dodge * plus
			stat.ConfusionDEF += info.ConfusionDEF * plus
			stat.PoisonDEF += info.PoisonDEF * plus
			stat.ParalysisDEF += info.ParalysisDEF * plus
			stat.HPRecoveryRate += info.HPRecoveryRate * plus
		}
	}

	slots := funk.Filter(skillSlots.Slots, func(slot *SkillSet) bool { // get 2nd job passive book
		return slot.BookID == 16100200 || slot.BookID == 16100300 || slot.BookID == 100030021 || slot.BookID == 100030023 || slot.BookID == 100030025 ||
			slot.BookID == 30000040 || slot.BookID == 30000041 || slot.BookID == 30000042 || slot.BookID == 30000043 || slot.BookID == 30000044 || slot.BookID == 30000045
	}).([]*SkillSet)

	for _, slot := range slots {
		for _, skill := range slot.Skills {
			info := SkillInfos[skill.SkillID]
			if info == nil {
				continue
			}

			amount := info.BasePassive + info.AdditionalPassive*skill.Plus
			switch info.PassiveType {
			case 1: // passive hp
				stat.MaxHP += amount
			case 2: // passive chi
				stat.MaxCHI += amount
			case 3: // passive arts defense
				stat.ArtsDEF += amount
			case 4: // passive defense
				stat.DEF += amount
			case 5: // passive accuracy
				stat.Accuracy += amount
			case 6: // passive dodge
				stat.Dodge += amount
			case 7: // passive arts atk
				stat.MinArtsATK += amount
				stat.MaxArtsATK += amount
			case 8: // passive atk
				stat.MinATK += amount
				stat.MaxATK += amount
			case 9: //HP AND CHI
				stat.MaxHP += amount
				stat.MaxCHI += amount
			case 11: //Dodge RAte AND ACCURACY
				stat.Accuracy += amount
				stat.Dodge += amount
			case 12: //EXTERNAL ATK AND INTERNAL ATK
				stat.MinArtsATK += amount
				stat.MaxArtsATK += amount
				stat.MinATK += amount
				stat.MaxATK += amount
			case 13: //INTERNAL ATTACK AND INTERNAL DEF
				stat.MinATK += amount
				stat.MaxATK += amount
				stat.DEF += amount
			case 14: //EXTERNAL ATK MINUS AND HP +
				stat.MaxHP += amount
				stat.MinArtsATK -= amount
				stat.MaxArtsATK -= amount
			case 15: //DAMAGE + HP
				stat.MaxHP += amount
				stat.MinATK += amount
				stat.MaxATK += amount
			case 16: //MINUS HP AND PLUS DEFENSE
				stat.MaxHP -= 15 //
				stat.DEF += amount
			}
		}
	}

	return nil
}

func (c *Character) BuffEffects(stat *Stat) error {

	buffs, err := FindBuffsByCharacterID(c.ID)
	if err != nil {
		return err
	}

	//stat := c.Socket.Stats

	for _, buff := range buffs {
		if buff.StartedAt+buff.Duration > c.Epoch || !buff.CanExpire {
			stat.MinATK += buff.ATK
			stat.MaxATK += buff.ATK
			stat.ATKRate += buff.ATKRate
			stat.Accuracy += buff.Accuracy
			stat.MinArtsATK += buff.ArtsATK
			stat.MaxArtsATK += buff.ArtsATK
			stat.ArtsATKRate += buff.ArtsATKRate
			stat.ArtsDEF += buff.ArtsDEF
			stat.ArtsDEFRate += buff.ArtsDEFRate
			stat.CHIRecoveryRate += buff.CHIRecoveryRate
			stat.ConfusionDEF += buff.ConfusionDEF
			stat.DEF += buff.DEF
			stat.DefRate += buff.DEFRate
			stat.DEXBuff += buff.DEX
			stat.Dodge += buff.Dodge
			stat.HPRecoveryRate += buff.HPRecoveryRate
			stat.INTBuff += buff.INT
			stat.MaxCHI += buff.MaxCHI
			stat.MaxHP += buff.MaxHP
			stat.ParalysisDEF += buff.ParalysisDEF
			stat.PoisonDEF += buff.PoisonDEF
			stat.STRBuff += buff.STR
		}
	}

	return nil
}

func (c *Character) GetLevelText() string {
	if c.Level < 10 {
		return fmt.Sprintf("%dKyu", c.Level)
	} else if c.Level <= 100 {
		return fmt.Sprintf("%dDan %dKyu", c.Level/10, c.Level%10)
	} else if c.Level < 110 {
		return fmt.Sprintf("Divine %dKyu", c.Level%100)
	} else if c.Level <= 200 {
		return fmt.Sprintf("Divine %dDan %dKyu", (c.Level-100)/10, c.Level%100)
	}

	return ""
}

func (c *Character) RelicDrop(itemID int64) []byte {

	itemName := Items[itemID].Name
	msg := fmt.Sprintf("%s has acquired [%s].", c.Name, itemName)
	length := int16(len(msg) + 3)

	now := time.Now().UTC()
	relic := &RelicLog{CharID: c.ID, ItemID: itemID, DropTime: null.TimeFrom(now)}
	RelicsLog[len(RelicsLog)] = relic
	err := relic.Create()
	if err != nil {
		fmt.Println("Error with load: ", err)
	}
	resp := RELIC_DROP
	resp.SetLength(length)
	resp[6] = byte(len(msg))
	resp.Insert([]byte(msg), 7)

	return resp
}

func (c *Character) AidStatus() []byte {

	resp := utils.Packet{}
	if c.AidMode {
		resp = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0xFA, 0x01, 0x55, 0xAA}
		resp.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 5) // pseudo id
		r2 := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x43, 0x01, 0x55, 0xAA}
		r2.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 5) // pseudo id

		resp.Concat(r2)

	} else {
		resp = utils.Packet{0xAA, 0x55, 0x04, 0x00, 0xFA, 0x00, 0x55, 0xAA}
		resp.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 5) // pseudo id
		r2 := utils.Packet{0xAA, 0x55, 0x04, 0x00, 0x43, 0x00, 0x55, 0xAA}
		r2.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 5) // pseudo id

		resp.Concat(r2)
	}

	return resp
}

func (c *Character) PickaxeActivated() bool {

	slots, err := c.InventorySlots()
	if err != nil {
		return false
	}

	pickaxeIDs := []int64{17200219, 17300005, 17501009, 17502536, 17502537, 17502538}

	return len(funk.Filter(slots, func(slot *InventorySlot) bool {
		return slot.Activated && funk.Contains(pickaxeIDs, slot.ItemID)
	}).([]*InventorySlot)) > 0
}

func (c *Character) TogglePet() []byte {
	slots, err := c.InventorySlots()
	if err != nil {
		return nil
	}

	petSlot := slots[0x0A]
	pet := petSlot.Pet
	if pet == nil {
		return nil
	}
	petInfo, _ := Pets[petSlot.ItemID]
	if petInfo.Combat {
		location := ConvertPointToLocation(c.Coordinate)
		pet.Coordinate = utils.Location{X: location.X, Y: location.Y}
		pet.IsOnline = !pet.IsOnline

		spawnData, _ := c.SpawnCharacter()
		if pet.IsOnline {
			GeneratePetID(c, pet)
			pet.PetCombatMode = 0
			c.PetHandlerCB = c.PetHandler
			go c.PetHandlerCB()

			resp := utils.Packet{
				0xAA, 0x55, 0x0B, 0x00, 0x75, 0x00, 0x01, 0x00, 0x80, 0xA1, 0x43, 0x00, 0x00, 0x3D, 0x43, 0x55, 0xAA,
				0xAA, 0x55, 0x05, 0x00, 0x51, 0x01, 0x0A, 0x00, 0x00, 0x55, 0xAA,
				0xAA, 0x55, 0x06, 0x00, 0x51, 0x05, 0x0A, 0x00, 0x00, 0x00, 0x55, 0xAA,
				0xAA, 0x55, 0x05, 0x00, 0x51, 0x06, 0x0A, 0x00, 0x00, 0x55, 0xAA,
				0xAA, 0x55, 0x05, 0x00, 0x51, 0x07, 0x0A, 0x00, 0x00, 0x55, 0xAA,
			}

			resp.Concat(spawnData)
			return resp
		}
	} else {
		location := ConvertPointToLocation(c.Coordinate)
		pet.Coordinate = utils.Location{X: location.X, Y: location.Y}
		pet.IsOnline = !pet.IsOnline
		spawnData, _ := c.SpawnCharacter()
		if pet.IsOnline {
			resp := utils.Packet{
				0xAA, 0x55, 0x05, 0x00, 0x51, 0x01, 0x0A, 0x00, 0x00, 0x55, 0xAA,
				0xAA, 0x55, 0x06, 0x00, 0x51, 0x05, 0x0A, 0x00, 0x00, 0x00, 0x55, 0xAA,
				0xAA, 0x55, 0x05, 0x00, 0x51, 0x07, 0x0A, 0x00, 0x00, 0x55, 0xAA,
			}
			resp.Concat(spawnData)
			return resp
		}
	}

	pet.Target = 0
	pet.Casting = false
	pet.IsMoving = false
	c.PetHandlerCB = nil
	RemovePetFromRegister(c)
	return DISMISS_PET
}

func (c *Character) ToggleMountPet() []byte {
	slots, err := c.InventorySlots()
	if err != nil {
		return nil
	}

	petSlot := slots[0x0A]
	pet := petSlot.Pet
	if pet == nil {
		return nil
	}
	petInfo, _ := Pets[petSlot.ItemID]
	if petInfo.Combat {
		location := ConvertPointToLocation(c.Coordinate)
		pet.Coordinate = utils.Location{X: location.X, Y: location.Y}
		pet.IsOnline = !pet.IsOnline

		spawnData, _ := c.SpawnCharacter()
		if pet.IsOnline {
			GeneratePetID(c, pet)
			pet.PetCombatMode = 0
			c.PetHandlerCB = c.PetHandler
			go c.PetHandlerCB()

			resp := utils.Packet{
				0xAA, 0x55, 0x0B, 0x00, 0x75, 0x00, 0x01, 0x00, 0x80, 0xA1, 0x43, 0x00, 0x00, 0x3D, 0x43, 0x55, 0xAA,
				0xAA, 0x55, 0x05, 0x00, 0x51, 0x01, 0x0A, 0x00, 0x00, 0x55, 0xAA,
				0xAA, 0x55, 0x06, 0x00, 0x51, 0x05, 0x0A, 0x00, 0x00, 0x00, 0x55, 0xAA,
				0xAA, 0x55, 0x05, 0x00, 0x51, 0x06, 0x0A, 0x00, 0x00, 0x55, 0xAA,
				0xAA, 0x55, 0x05, 0x00, 0x51, 0x07, 0x0A, 0x00, 0x00, 0x55, 0xAA,
			}

			resp.Concat(spawnData)
			return resp
		}
	} else {
		location := ConvertPointToLocation(c.Coordinate)
		pet.Coordinate = utils.Location{X: location.X, Y: location.Y}
		pet.IsOnline = !pet.IsOnline
		spawnData, _ := c.SpawnCharacter()
		if pet.IsOnline {
			resp := utils.Packet{
				0xAA, 0x55, 0x05, 0x00, 0x51, 0x01, 0x0A, 0x00, 0x00, 0x55, 0xAA,
				0xAA, 0x55, 0x06, 0x00, 0x51, 0x05, 0x0A, 0x00, 0x00, 0x00, 0x55, 0xAA,
				0xAA, 0x55, 0x05, 0x00, 0x51, 0x07, 0x0A, 0x00, 0x00, 0x55, 0xAA,
			}
			resp.Concat(spawnData)
			return resp
		}
	}

	pet.Target = 0
	pet.Casting = false
	pet.IsMoving = false
	c.PetHandlerCB = nil
	RemovePetFromRegister(c)
	return DISMISS_PET
}
func RemoveIndex(s []*AI, index int) []*AI {
	return append(s[:index], s[index+1:]...)
}
func (c *Character) DealDamageToPlayer(char *Character, dmg int) {
	if c == nil {
		log.Println("character is nil")
		return
	} else if char.Socket.Stats.HP <= 0 {
		return
	}
	if dmg > char.Socket.Stats.HP {
		dmg = char.Socket.Stats.HP
	}

	char.Socket.Stats.HP -= dmg
	if char.Socket.Stats.HP <= 0 {
		char.Socket.Stats.HP = 0
	}
	r := DEAL_DAMAGE
	r.Insert(utils.IntToBytes(uint64(char.PseudoID), 2, true), 5)          // ai pseudo id
	r.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 7)             // ai pseudo id
	r.Insert(utils.IntToBytes(uint64(char.Socket.Stats.HP), 4, true), 9)   // ai current hp
	r.Insert(utils.IntToBytes(uint64(char.Socket.Stats.CHI), 4, true), 13) // ai current chi
	char.Socket.Write(r)
	c.Socket.Write(r)
}
func (c *Character) DealDamage(ai *AI, dmg int) {

	if c == nil {
		log.Println("character is nil")
		return
	} else if ai.HP <= 0 {
		return
	}
	npcPos := NPCPos[ai.PosID]
	if npcPos == nil {
		log.Println("npc pos is nil")
		return
	}

	npc := NPCs[npcPos.NPCID]
	if npc == nil {
		log.Println("npc is nil")
		return
	}

	for _, factionNPC := range ZhuangFactionMobs {
		if factionNPC == npc.ID && c.Faction == 1 {
			return
		}
	}
	for _, factionNPC := range ShaoFactionMobs {
		if factionNPC == npc.ID && c.Faction == 2 {
			return
		}
	}

	s := c.Socket

	if dmg > ai.HP {
		dmg = ai.HP
	}

	ai.HP -= dmg
	if ai.HP <= 0 {
		ai.HP = 0
	}

	d := ai.DamageDealers.Get(c.ID)
	if d == nil {
		ai.DamageDealers.Add(c.ID, &Damage{Damage: dmg, DealerID: c.ID})
	} else {
		d.(*Damage).Damage += dmg
		ai.DamageDealers.Add(c.ID, d)
	}

	if c.Invisible {
		buff, _ := FindBuffByID(241, c.ID)
		if buff != nil {
			buff.Duration = 0
			go buff.Update()
		}

		buff, _ = FindBuffByID(244, c.ID)
		if buff != nil {
			buff.Duration = 0
			go buff.Update()
		}

		if c.DuelID > 0 {
			opponent, _ := FindCharacterByID(c.DuelID)
			spawnData, _ := c.SpawnCharacter()

			r := utils.Packet{}
			r.Concat(spawnData)
			r.Overwrite(utils.IntToBytes(500, 2, true), 13) // duel state

			sock := GetSocket(opponent.UserID)
			if sock != nil {
				sock.Write(r)
			}
		}
	}

	r := DEAL_DAMAGE
	r.Insert(utils.IntToBytes(uint64(ai.PseudoID), 2, true), 5) // ai pseudo id
	r.Insert(utils.IntToBytes(uint64(c.PseudoID), 2, true), 7)  // ai pseudo id
	r.Insert(utils.IntToBytes(uint64(ai.HP), 4, true), 9)       // ai current hp
	r.Insert(utils.IntToBytes(uint64(0), 4, true), 13)          // ai current chi
	p := &nats.CastPacket{CastNear: true, MobID: ai.ID, Data: r, Type: nats.MOB_ATTACK}
	if err := p.Cast(); err != nil {
		log.Println("deal damage broadcast error:", err)
		return
	}

	if !npcPos.Attackable {
		go ai.DropHandler(c)
	}
	if c.Map == 255 && IsFactionWarStarted() { //faction war
		if c.Faction == 1 {
			if npcPos.NPCID == 425506 {
				AddPointsToFactionWarFaction(5, 1)
			}
			if npcPos.NPCID == 425505 {
				AddPointsToFactionWarFaction(50, 1)
			}
			if npcPos.NPCID == 425507 {
				AddPointsToFactionWarFaction(7, 1)
			}
			if npcPos.NPCID == 425508 {
				AddPointsToFactionWarFaction(500, 1)
			}
		}
		if c.Faction == 2 {
			if npcPos.NPCID == 425501 {
				AddPointsToFactionWarFaction(5, 2)
			}
			if npcPos.NPCID == 425502 {
				AddPointsToFactionWarFaction(50, 2)
			}
			if npcPos.NPCID == 425503 {
				AddPointsToFactionWarFaction(7, 2)
			}
			if npcPos.NPCID == 425504 {
				AddPointsToFactionWarFaction(500, 2)
			}
		}
	}

	if ai.HP <= 0 { // ai died
		DeleteBuffsByAiPseudoID(ai.PseudoID)
		if ai.Once {
			ai.Handler = nil
			if funk.Contains(DungeonZones, ai.Map) {
				//	for i, action := range DungeonsByMap[ai.Server][ai.Map] {
				//if action.ID == ai.ID {
				//s2 := RemoveIndex(DungeonsByMap[ai.Server][ai.Map], i)
				//DungeonsByMap[ai.Server][ai.Map] = s2
				fmt.Println("Mobs: ", DungeonsByMap[c.Socket.User.ConnectedServer][c.Map])
				DungeonsByMap[c.Socket.User.ConnectedServer][c.Map]--
				/*mobsCount := DungeonsByMap[c.Socket.User.ConnectedServer][c.Map]
				if DungeonsByMap[c.Socket.User.ConnectedServer][c.Map] <= 10 {
					s.Conn.Write(messaging.InfoMessage(fmt.Sprintf("%d mobs remaining", mobsCount)))
				}*/
				//}
				//}
			}
		} else {
			time.AfterFunc(time.Duration(npcPos.RespawnTime)*time.Second/2, func() { // respawn mob n secs later
				curCoordinate := ConvertPointToLocation(ai.Coordinate)
				minCoordinate := ConvertPointToLocation(npcPos.MinLocation)
				maxCoordinate := ConvertPointToLocation(npcPos.MaxLocation)

				X := utils.RandFloat(minCoordinate.X, maxCoordinate.X)
				Y := utils.RandFloat(minCoordinate.Y, maxCoordinate.Y)

				X = (X / 3) + 2*curCoordinate.X/3
				Y = (Y / 3) + 2*curCoordinate.Y/3

				coordinate := &utils.Location{X: X, Y: Y}
				ai.TargetLocation = *coordinate
				ai.SetCoordinate(coordinate)

				ai.HP = npc.MaxHp
				ai.IsDead = false
			})
		}
		if npc.ID == 424201 && WarStarted {
			OrderPoints -= 200
		} else if npc.ID == 424202 && WarStarted {
			ShaoPoints -= 200
		}
		exp := int64(0)
		if c.Level <= 100 {
			exp = npc.Exp
		} else if c.Level <= 200 {
			exp = npc.DivineExp
		} else {
			exp = npc.DarknessExp
		}

		// EXP gained
		r, levelUp := c.AddExp(exp)
		if levelUp {
			statData, err := c.GetStats()
			if err == nil {
				s.Conn.Write(statData)
			}
		}
		s.Conn.Write(r)

		// EXP gain for party members
		party := FindParty(c)
		if party != nil {
			members := funk.Filter(party.GetMembers(), func(m *PartyMember) bool {
				return m.Accepted || m.ID == c.ID
			}).([]*PartyMember)
			members = append(members, &PartyMember{Character: party.Leader, Accepted: true})

			coordinate := ConvertPointToLocation(c.Coordinate)
			for _, m := range members {
				user, err := FindUserByID(m.UserID)
				if err != nil || user == nil || (c.Level-m.Level) > 20 {
					continue
				}

				memberCoordinate := ConvertPointToLocation(m.Coordinate)

				if m.ID == c.ID && !m.Accepted {
					break
				}

				if m.ID == c.ID || m.Map != c.Map || s.User.ConnectedServer != user.ConnectedServer ||
					utils.CalculateDistance(coordinate, memberCoordinate) > 100 || m.Socket.Stats.HP <= 0 {
					continue
				}

				exp := int64(0)
				if m.Level <= 100 {
					exp = npc.Exp
				} else if m.Level <= 200 {
					exp = npc.DivineExp
				} else {
					exp = npc.DarknessExp
				}

				exp /= int64(len(members))

				r, levelUp := m.AddExp(exp)
				if levelUp {
					statData, err := m.GetStats()
					if err == nil {
						m.Socket.Write(statData)
					}
				}
				m.Socket.Write(r)
			}
		}
		//GIVE 200+ ataraxia ITEM
		if npc.ID == 43401 {
			if c.Exp >= 544951059310 && c.Level == 200 {
				resp := utils.Packet{}
				slots, err := c.InventorySlots()
				if err != nil {
				}
				reward := NewSlot()
				reward.ItemID = int64(90000304)
				reward.Quantity = 1
				_, slot, _ := c.AddItem(reward, -1, true)
				resp.Concat(slots[slot].GetData(slot))
				s.Conn.Write(resp)
				s.Conn.Write(messaging.InfoMessage(fmt.Sprintf("You kill the Wyrm, now you can make the transformation.")))
			}
		}
		//60006	Underground(EXP)
		//60011	Waterfall(EXP)
		//60016	Forest(EXP)
		//60021	Sky Garden(EXP)
		//FIVE ELEMENT CASTLE
		claimer, err := ai.FindClaimer()
		if err == nil || claimer != nil {
			//exptime := time.Now().UTC().Add(time.Hour * 2)
			if npc.ID == 423308 { //HWARANG GUARDIAN STATUE //SOUTHERN WOOD TEMPLE
				if claimer.GuildID > 0 {
					/*					FiveClans[1].ClanID = claimer.GuildID
										FiveClans[1].ExpiresAt = null.TimeFrom(exptime)
										FiveClans[1].Update()*/
					CaptureFive(1, c)
				}
			} else if npc.ID == 423310 { //SUGUN GUARDIAN STATUE //LIGHTNING HILL TEMPLE
				if c.GuildID > 0 {
					/*FiveClans[2].ClanID = claimer.GuildID
					FiveClans[2].ExpiresAt = null.TimeFrom(exptime)
					FiveClans[2].Update()*/
					CaptureFive(2, c)
				}
			} else if npc.ID == 423312 { //CHUNKYUNG GUARDIAN STATUE //OCEAN ARMY TEMPLE
				if claimer.GuildID > 0 {
					/*FiveClans[3].ClanID = claimer.GuildID
					FiveClans[3].ExpiresAt = null.TimeFrom(exptime)
					FiveClans[3].Update()*/
					CaptureFive(3, c)
				}
			} else if npc.ID == 423314 { //MOKNAM GUARDIAN STATUE //FLAME WOLF TEMPLE
				if c.GuildID > 0 {
					/*FiveClans[4].ClanID = claimer.GuildID
					FiveClans[4].ExpiresAt = null.TimeFrom(exptime)
					FiveClans[4].Update()*/
					CaptureFive(4, c)
				}
			} else if npc.ID == 423316 { //JISU GUARDIAN STATUE //WESTERN LAND TEMPLE
				/*FiveClans[5].ClanID = claimer.GuildID
				FiveClans[5].ExpiresAt = null.TimeFrom(exptime)
				FiveClans[5].Update()*/
				CaptureFive(5, c)
			}
		}

		// PTS gained LOOT
		c.PTS++
		if c.PTS%50 == 0 {
			r = c.GetPTS()
			c.HasLot = true
			s.Conn.Write(r)
		}

		// Gold dropped
		goldDrop := int64(npc.GoldDrop)
		if goldDrop > 0 {
			amount := uint64(utils.RandInt(goldDrop/2, goldDrop))
			r = c.LootGold(amount)
			s.Conn.Write(r)
		}

		//Item dropped
		go func() {
			claimer, err := ai.FindClaimer()
			if err != nil || claimer == nil {
				return
			}

			dropMaxLevel := int(npc.Level + 250)
			if c.Level <= dropMaxLevel {
				ai.DropHandler(claimer)
			}
			time.AfterFunc(time.Second, func() {
				ai.DamageDealers.Clear()
			})
		}()

		time.AfterFunc(time.Second, func() { // disappear mob 1 sec later
			ai.TargetPlayerID = 0
			ai.TargetPetID = 0
			ai.IsDead = true
		})
	} else if ai.TargetPlayerID == 0 {
		ai.IsMoving = false
		ai.MovementToken = 0
		ai.TargetPlayerID = c.ID
	} else {
		ai.IsMoving = false
		ai.MovementToken = 0
	}

}

func (c *Character) GetPetStats() []byte {

	slots, err := c.InventorySlots()
	if err != nil {
		return nil
	}

	petSlot := slots[0x0A]
	pet := petSlot.Pet
	if pet == nil {
		return nil
	}

	resp := utils.Packet{}
	resp = petSlot.GetPetStats(c)
	resp.Concat(petSlot.GetData(0x0A))
	return resp
}

func (c *Character) StartPvP(timeLeft int) {

	info, resp := "", utils.Packet{}
	if timeLeft > 0 {
		info = fmt.Sprintf("Duel will start %d seconds later.", timeLeft)
		time.AfterFunc(time.Second, func() {
			c.StartPvP(timeLeft - 1)
		})

	} else if c.DuelID > 0 {
		info = "Duel has started."
		resp.Concat(c.OnDuelStarted())
	}

	resp.Concat(messaging.InfoMessage(info))
	c.Socket.Write(resp)
}

func (c *Character) CanAttack(enemy *Character) bool {
	servers, _ := GetServers()
	return (c.DuelID == enemy.ID && c.DuelStarted) || funk.Contains(PvPZones, c.Map) || servers[int16(c.Socket.User.ConnectedServer)-1].IsPVPServer
}

func (c *Character) OnDuelStarted() []byte {

	c.DuelStarted = true
	statData, _ := c.GetStats()

	opponent, err := FindCharacterByID(c.DuelID)
	if err != nil || opponent == nil {
		return nil
	}

	opData, err := opponent.SpawnCharacter()
	if err != nil || opData == nil || len(opData) < 13 {
		return nil
	}

	r := utils.Packet{}
	r.Concat(opData)
	r.Overwrite(utils.IntToBytes(500, 2, true), 13) // duel state

	resp := utils.Packet{}
	resp.Concat(opponent.GetHPandChi())
	resp.Concat(r)
	resp.Concat(statData)
	resp.Concat([]byte{0xAA, 0x55, 0x02, 0x00, 0x2A, 0x04, 0x55, 0xAA})
	return resp

}

func (c *Character) HasAidBuff() bool {
	slots, err := c.InventorySlots()
	if err != nil {
		return false
	}

	return len(funk.Filter(slots, func(s *InventorySlot) bool {
		return (s.ItemID == 13000170 || s.ItemID == 13000171 || s.ItemID == 13000173 || s.ItemID == 23000141) && s.Activated && s.InUse
	}).([]*InventorySlot)) > 0
}
