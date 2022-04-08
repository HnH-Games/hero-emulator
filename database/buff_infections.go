package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

var (
	BuffInfections = make(map[int]*BuffInfection)
)

type BuffInfection struct {
	ID                                int     `db:"id"`
	Name                              string  `db:"name"`
	PoisonDef                         int     `db:"poison_def"`
	AdditionalPoisonDef               int     `db:"additional_poison_def"`
	ParalysisDef                      int     `db:"paralysis_def"`
	AdditionalParalysisDef            int     `db:"additional_para_def"`
	ConfusionDef                      int     `db:"confusion_def"`
	AdditionalConfusionDef            int     `db:"additional_confusion_def"`
	BaseDef                           int     `db:"base_def"`
	AdditionalDEF                     int     `db:"additional_def"`
	ArtsDEF                           int     `db:"arts_def"`
	AdditionalArtsDEF                 int     `db:"additional_arts_def"`
	MaxHP                             int     `db:"max_hp"`
	HPRecoveryRate                    int     `db:"hp_recovery_rate"`
	STR                               int     `db:"str"`
	AdditionalSTR                     int     `db:"additional_str"`
	DEX                               int     `db:"dex"`
	AdditionalDEX                     int     `db:"additional_dex"`
	INT                               int     `db:"int"`
	AdditionalINT                     int     `db:"additional_int"`
	Wind                              int     `db:"wind"`
	AdditionalWind                    int     `db:"additional_wind"`
	Water                             int     `db:"water"`
	AdditionalWater                   int     `db:"additional_water"`
	Fire                              int     `db:"fire"`
	AdditionalFire                    int     `db:"additional_fire"`
	AdditionalHP                      int     `db:"additional_hp"`
	BaseATK                           int     `db:"base_atk"`
	AdditionalATK                     int     `db:"additional_atk"`
	BaseArtsATK                       int     `db:"base_arts_atk"`
	AdditionalArtsATK                 int     `db:"additional_arts_atk"`
	Accuracy                          int     `db:"accuracy"`
	AdditionalAccuracy                int     `db:"additional_accuracy"`
	DodgeRate                         int     `db:"dodge_rate"`
	AdditionalDodgeRate               int     `db:"additional_dodge_rate"`
	MovSpeed                          float64 `db:"movement_speed"`
	AdditionalMovSpeed                float64 `db:"additional_movement_speed"`
	ExpRate                           int     `db:"exp_rate"`
	HyeolgongCost                     int     `db:"hyeolgong_cost"`
	NPCSellingCost                    int     `db:"npc_selling"`
	NPCBuyingCost                     int     `db:"npc_buying"`
	IsPercent                         bool    `db:"ispercent"`
	LightningRadius                   int     `db:"lightning_radius"`
	AttackSpeed                       int     `db:"attack_speed"`
	AdditionalHPRecovery              int     `db:"additional_hp_recovery"`
	MaxChi                            int     `db:"max_chi"`
	DamageReflection                  int     `db:"damage_reflection"`
	DropItem                          int     `db:"drop_item"`
	EnchancedProb                     int     `db:"enchanced_prob"`
	SyntheticComposite                int     `db:"synthetic_composite"`
	AdvancedComposite                 int     `db:"advanced_composite"`
	PetMaxHP                          int     `db:"pet_base_hp"`
	PetAdditionalHP                   int     `db:"pet_additional_hp"`
	PetBaseDEF                        int     `db:"pet_base_def"`
	PetAdditionalDEF                  int     `db:"pet_additional_def"`
	PetArtsDEF                        int     `db:"pet_base_arts_def"`
	PetAdditinalArtsDEF               int     `db:"pet_additional_arts_def"`
	AdditionalAttackSpeed             int     `db:"additional_attack_speed"`
	MakeSize                          float64 `db:"makesize"`
	Critical_Strike                   int     `db:"critical_strike"`
	AdditionalCriticalStrike          int     `db:"additional_critical_strike"`
	CashAcquired                      int     `db:"cash_acquired"`
	TakingEffectProbability           int     `db:"taking_effect_probability"`
	AdditionalTakingEffectProbability int     `db:"additional_taking_effect_probability"`
	AdditionalDamageReflection        int     `db:"additional_damage_reflection"`
}

func (e *BuffInfection) Create() error {
	return db.Insert(e)
}

func (e *BuffInfection) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(e)
}

func (e *BuffInfection) Delete() error {
	_, err := db.Delete(e)
	return err
}

func (e *BuffInfection) Update() error {
	_, err := db.Update(e)
	return err
}

func getBuffInfections() error {
	var buffInfections []*BuffInfection
	query := `select * from data.buff_infections`

	if _, err := db.Select(&buffInfections, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getBuffInfections: %s", err.Error())
	}

	for _, b := range buffInfections {
		BuffInfections[b.ID] = b
	}

	return nil
}

func RefreshBuffInfections() error {
	var buffInfections []*BuffInfection
	query := `select * from data.buff_infections`

	if _, err := db.Select(&buffInfections, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getBuffInfections: %s", err.Error())
	}

	for _, b := range buffInfections {
		BuffInfections[b.ID] = b
	}

	return nil
}
