package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

type Rank struct {
	ID       int `db:"id" json:"id"`
	HonorID  int `db:"honor_id" json:"honor_id"`
	PlusSTR  int `db:"str" json:"str"`
	PlusDEX  int `db:"dex" json:"dex"`
	PlusINT  int `db:"int" json:"int"`
	PlusStat int `db:"plus_stat" json:"plus_stat"`
	PlusSP   int `db:"plus_skillpoint" json:"plus_skillpoint"`
}

func (b *Rank) Create() error {
	return db.Insert(b)
}

func (b *Rank) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(b)
}

func (b *Rank) Delete() error {
	_, err := db.Delete(b)
	return err
}
func (b *Rank) Update() error {
	_, err := db.Update(b)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error: %s", err.Error()))
	}
	return err
}

func FindRankByHonorID(rankID int64) (*Rank, error) {

	var g Rank
	query := `select * from data.reborn_system where honor_id = $1`

	if err := db.SelectOne(&g, query, rankID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindRevornByHonorID: %s", err.Error())
	}

	return &g, nil
}

func (c *Character) RebornEffects(st *Stat) error {
	rank, err := FindRankByHonorID(c.HonorRank)
	if err != nil {
		return err
	}
	if rank == nil {
		return nil
	}
	st.STRBuff += rank.PlusSTR
	st.DEXBuff += rank.PlusDEX
	st.INTBuff += rank.PlusINT
	c.ExpMultiplier -= float64(0.05) * float64(rank.ID)
	c.DropMultiplier += float64(0.1) * float64(rank.ID)
	c.Update()
	/*st.WindBuff += rank.Wind
	st.WaterBuff += rank.Water
	st.FireBuff += rank.Fire

	st.DEF += rank.Def + ((rank.BaseDef1 + rank.BaseDef2 + rank.BaseDef3) / 3)
	st.DefRate += rank.DefRate

	st.ArtsDEF += rank.ArtsDef
	st.ArtsDEFRate += rank.ArtsDefRate

	st.MaxHP += rank.MaxHp
	st.MaxCHI += rank.MaxChi

	st.Accuracy += rank.Accuracy
	st.Dodge += rank.Dodge

	st.MinATK += rank.BaseMinAtk + rank.MinAtk
	st.MaxATK += rank.BaseMaxAtk + rank.MaxAtk
	st.ATKRate += rank.AtkRate

	st.PoisonATK += rank.PoisonATK
	st.PoisonDEF += rank.PoisonDEF
	st.ParalysisATK += rank.ParaATK
	st.ParalysisDEF += rank.ParaDEF
	st.ConfusionATK += rank.ConfusionATK
	st.ConfusionDEF += rank.ConfusionDEF

	st.MinArtsATK += rank.MinArtsAtk
	st.MaxArtsATK += rank.MaxArtsAtk
	st.ArtsATKRate += rank.ArtsAtkRate
	additionalExpMultiplier += rank.ExpRate / 100
	additionalDropMultiplier += rank.DropRate / 100
	additionalRunningSpeed += rank.RunningSpeed
	st.HPRecoveryRate += rank.HPRecoveryRate
	st.CHIRecoveryRate += rank.CHIRecoveryRate
	*/
	return nil
}
