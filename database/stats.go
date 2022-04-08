package database

import (
	"database/sql"
	"fmt"
	"sync"

	gorp "gopkg.in/gorp.v1"
)

var (
	stats         = make(map[int]*Stat)
	stMutex       sync.RWMutex
	startingStats = map[int]*Stat{
		50: {STR: 13, DEX: 12, INT: 8, HP: 72, MaxHP: 72, CHI: 48, MaxCHI: 48, HPRecoveryRate: 10, CHIRecoveryRate: 10}, //Beast Male
		51: {STR: 13, DEX: 12, INT: 8, HP: 72, MaxHP: 72, CHI: 48, MaxCHI: 48, HPRecoveryRate: 10, CHIRecoveryRate: 10}, //Best Female
		52: {STR: 13, DEX: 12, INT: 8, HP: 72, MaxHP: 72, CHI: 48, MaxCHI: 48, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Monk
		53: {STR: 12, DEX: 12, INT: 6, HP: 90, MaxHP: 90, CHI: 30, MaxCHI: 30, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Male Blade
		54: {STR: 11, DEX: 13, INT: 6, HP: 90, MaxHP: 90, CHI: 30, MaxCHI: 30, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Female Blade
		56: {STR: 15, DEX: 10, INT: 5, HP: 96, MaxHP: 96, CHI: 24, MaxCHI: 24, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Axe
		57: {STR: 14, DEX: 11, INT: 5, HP: 96, MaxHP: 96, CHI: 24, MaxCHI: 24, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Female Spear
		59: {STR: 10, DEX: 18, INT: 2, HP: 84, MaxHP: 84, CHI: 36, MaxCHI: 36, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Dual Sword

		60: {STR: 13, DEX: 12, INT: 8, HP: 72, MaxHP: 72, CHI: 48, MaxCHI: 48, HPRecoveryRate: 10, CHIRecoveryRate: 10}, //Divine Beast Male
		61: {STR: 13, DEX: 12, INT: 8, HP: 72, MaxHP: 72, CHI: 48, MaxCHI: 48, HPRecoveryRate: 10, CHIRecoveryRate: 10}, //Divine Beast Female
		62: {STR: 13, DEX: 12, INT: 8, HP: 72, MaxHP: 72, CHI: 48, MaxCHI: 48, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Divine Monk
		63: {STR: 12, DEX: 12, INT: 6, HP: 90, MaxHP: 90, CHI: 30, MaxCHI: 30, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Divine Male Blade
		64: {STR: 11, DEX: 13, INT: 6, HP: 90, MaxHP: 90, CHI: 30, MaxCHI: 30, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Divine Female Blade
		66: {STR: 15, DEX: 10, INT: 5, HP: 96, MaxHP: 96, CHI: 24, MaxCHI: 24, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Divine Axe
		67: {STR: 14, DEX: 11, INT: 5, HP: 96, MaxHP: 96, CHI: 24, MaxCHI: 24, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Divine Female Spear
		69: {STR: 10, DEX: 18, INT: 2, HP: 84, MaxHP: 84, CHI: 36, MaxCHI: 36, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Divine Dual Sword

		70: {STR: 13, DEX: 12, INT: 8, HP: 72, MaxHP: 72, CHI: 48, MaxCHI: 48, HPRecoveryRate: 10, CHIRecoveryRate: 10}, //Darknes Beast Male
		71: {STR: 13, DEX: 12, INT: 8, HP: 72, MaxHP: 72, CHI: 48, MaxCHI: 48, HPRecoveryRate: 10, CHIRecoveryRate: 10}, //Darknes Beast Female
		72: {STR: 13, DEX: 12, INT: 8, HP: 72, MaxHP: 72, CHI: 48, MaxCHI: 48, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Dark Monk
		73: {STR: 12, DEX: 12, INT: 6, HP: 90, MaxHP: 90, CHI: 30, MaxCHI: 30, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Dark Male Blade
		74: {STR: 11, DEX: 13, INT: 6, HP: 90, MaxHP: 90, CHI: 30, MaxCHI: 30, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Dark Female Blade
		76: {STR: 15, DEX: 10, INT: 5, HP: 96, MaxHP: 96, CHI: 24, MaxCHI: 24, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Dark Axe
		77: {STR: 14, DEX: 11, INT: 5, HP: 96, MaxHP: 96, CHI: 24, MaxCHI: 24, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Dark Female Spear
		79: {STR: 10, DEX: 18, INT: 2, HP: 84, MaxHP: 84, CHI: 36, MaxCHI: 36, HPRecoveryRate: 10, CHIRecoveryRate: 10}, // Dark Dual Sword
	}
)

type Stat struct {
	ID              int `db:"id"`
	HP              int `db:"hp"`
	MaxHP           int `db:"max_hp"`
	HPRecoveryRate  int `db:"hp_recovery_rate"`
	CHI             int `db:"chi"`
	MaxCHI          int `db:"max_chi"`
	CHIRecoveryRate int `db:"chi_recovery_rate"`
	STR             int `db:"str"`
	DEX             int `db:"dex"`
	INT             int `db:"int"`
	STRBuff         int `db:"str_buff"`
	DEXBuff         int `db:"dex_buff"`
	INTBuff         int `db:"int_buff"`
	StatPoints      int `db:"stat_points"`
	Honor           int `db:"honor"`
	MinATK          int `db:"min_atk"`
	MaxATK          int `db:"max_atk"`
	ATKRate         int `db:"atk_rate"`
	MinArtsATK      int `db:"min_arts_atk"`
	MaxArtsATK      int `db:"max_arts_atk"`
	ArtsATKRate     int `db:"arts_atk_rate"`
	DEF             int `db:"def"`
	DefRate         int `db:"def_rate"`
	ArtsDEF         int `db:"arts_def"`
	ArtsDEFRate     int `db:"arts_def_rate"`
	Accuracy        int `db:"accuracy"`
	Dodge           int `db:"dodge"`
	PoisonATK       int `db:"poison_atk"`
	ParalysisATK    int `db:"paralysis_atk"`
	ConfusionATK    int `db:"confusion_atk"`
	PoisonDEF       int `db:"poison_def"`
	ParalysisDEF    int `db:"paralysis_def"`
	ConfusionDEF    int `db:"confusion_def"`
	Wind            int `db:"wind"`
	WindBuff        int `db:"wind_buff"`
	Water           int `db:"water"`
	WaterBuff       int `db:"water_buff"`
	Fire            int `db:"fire"`
	FireBuff        int `db:"fire_buff"`
	NaturePoints    int `db:"nature_points"`
}

func (t *Stat) Create(c *Character) error {
	t = startingStats[c.Type]
	t.ID = c.ID
	t.StatPoints = 4
	t.NaturePoints = 0
	t.Honor = 10000
	return db.Insert(t)
}

func (t *Stat) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(t)
}

func (t *Stat) Update() error {
	_, err := db.Update(t)
	return err
}

func (t *Stat) Delete() error {
	stMutex.Lock()
	delete(stats, t.ID)
	stMutex.Unlock()

	_, err := db.Delete(t)
	return err
}

func (t *Stat) Calculate() error {

	c, err := FindCharacterByID(t.ID)
	if err != nil {
		return err
	} else if c == nil {
		return nil
	}

	temp := *t
	stStat := startingStats[c.Type]
	temp.MaxHP = stStat.MaxHP
	temp.MaxCHI = stStat.MaxCHI
	temp.STRBuff = 0
	temp.DEXBuff = 0
	temp.INTBuff = 0
	temp.WindBuff = 0
	temp.WaterBuff = 0
	temp.FireBuff = 0
	temp.MinATK = temp.STR
	temp.MaxATK = temp.STR
	temp.ATKRate = 0
	temp.MinArtsATK = temp.STR
	temp.MaxArtsATK = temp.STR
	temp.ArtsATKRate = 0
	temp.DEF = temp.DEX
	temp.DefRate = 0
	temp.ArtsDEF = 2*temp.INT + temp.DEX
	temp.ArtsDEFRate = 0
	temp.Accuracy = int(float32(temp.STR) * 0.925)
	temp.Dodge = temp.DEX

	c.BuffEffects(&temp)
	c.JobPassives(&temp)
	c.ExpMultiplier = 1
	c.DropMultiplier = 1
	c.AdditionalDropMultiplier = 0
	c.AdditionalExpMultiplier = 0
	c.AdditionalRunningSpeed = 0
	c.ItemEffects(&temp, 0, 9)         // NORMAL ITEMS
	c.ItemEffects(&temp, 307, 315)     // HT ITEMS
	c.ItemEffects(&temp, 0x0B, 0x43)   // INV BUFFS1
	c.ItemEffects(&temp, 0x155, 0x18D) // INV BUFFS2
	c.ItemEffects(&temp, 397, 399)     // MARBLES(1-3)
	c.RebornEffects(&temp)
	//totalDEX := temp.DEX + temp.DEXBuff
	totalWind := temp.Wind + temp.WindBuff
	totalWater := temp.Water + temp.WaterBuff
	totalFire := temp.Fire + temp.FireBuff

	temp.DEF += temp.DEXBuff + 2*totalWind + 1*totalWater + 1*totalFire
	temp.DEF += temp.DEF * temp.DefRate / 100
	temp.ArtsDEF += 2*temp.INTBuff + temp.DEXBuff + 1*totalWind + 2*totalWater + 1*totalFire
	temp.ArtsDEF += temp.ArtsDEF * temp.ArtsDEFRate / 100

	c.ItemEffects(&temp, 400, 401) // MARBLES(4-5)

	totalSTR := temp.STR + temp.STRBuff
	totalINT := temp.INT + temp.INTBuff

	temp.MaxHP += 10 * totalSTR
	temp.MaxCHI += 3 * totalINT

	temp.MinATK += temp.STRBuff + 1*totalWind + 1*totalWater + 2*totalFire
	temp.MinATK += temp.MinATK * temp.ATKRate / 100
	temp.MaxATK += temp.STRBuff + 1*totalWind + 1*totalWater + 2*totalFire
	temp.MaxATK += temp.MaxATK * temp.ATKRate / 100

	temp.MinArtsATK += temp.STRBuff + 2*totalINT + int(float32(totalINT*temp.MinATK)/200)
	temp.MinArtsATK += temp.MinArtsATK * temp.ArtsATKRate / 100
	temp.MaxArtsATK += temp.STRBuff + 2*totalINT + int(float32(totalINT*temp.MaxATK)/200)
	temp.MaxArtsATK += temp.MaxArtsATK * temp.ArtsATKRate / 100

	temp.Accuracy += int(float32(temp.STRBuff) * 0.925)
	temp.Dodge += temp.DEXBuff

	*t = temp
	go t.Update()
	return nil
}

func (t *Stat) Reset() error {

	c, err := FindCharacterByID(t.ID)
	if err != nil {
		return err
	} else if c == nil {
		return fmt.Errorf("CalculateTotalStatPoints: character not found")
	}

	tens := c.Level / 10
	statPts := ((((tens + 3) * (tens + 4) / 2) - 6) * 10) - 4
	statPts += (tens + 4) * ((c.Level + 1) % 10)

	stat := startingStats[c.Type]
	t.STR = stat.STR
	t.DEX = stat.DEX
	t.INT = stat.INT
	t.StatPoints = statPts

	return t.Update()
}

func FindStatByID(id int) (*Stat, error) {

	stMutex.RLock()
	s, ok := stats[id]
	stMutex.RUnlock()
	if ok {
		return s, nil
	}

	stat := &Stat{}
	query := `select * from hops.stats where id = $1`

	if err := db.SelectOne(&stat, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindStatByID: %s", err.Error())
	}

	stMutex.Lock()
	defer stMutex.Unlock()
	stats[stat.ID] = stat

	return stat, nil
}

func DeleteStatFromCache(id int) {
	delete(stats, id)
}
