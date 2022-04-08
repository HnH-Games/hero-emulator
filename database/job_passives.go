package database

import (
	"database/sql"
	"fmt"

	gorp "gopkg.in/gorp.v1"
)

type JobPassive struct {
	ID             int8 `db:"id"`
	MaxHp          int  `db:"max_hp"`
	MaxChi         int  `db:"max_chi"`
	ATK            int  `db:"atk"`
	ArtsATK        int  `db:"arts_atk"`
	DEF            int  `db:"def"`
	ArtsDef        int  `db:"arts_def"`
	Accuracy       int  `db:"accuracy"`
	Dodge          int  `db:"dodge"`
	ConfusionDEF   int  `db:"confusion_def"`
	PoisonDEF      int  `db:"poison_def"`
	ParalysisDEF   int  `db:"paralysis_def"`
	HPRecoveryRate int  `db:"hp_recovery_rate"`
}

var (
	JobPassives = make(map[int8]*JobPassive)
)

func (p *JobPassive) Create() error {
	return db.Insert(p)
}

func (p *JobPassive) CreateWithTransaction(tr *gorp.Transaction) error {
	return tr.Insert(p)
}

func (p *JobPassive) Update() error {
	_, err := db.Update(p)
	return err
}

func (p *JobPassive) Delete() error {
	_, err := db.Delete(p)
	return err
}

func getJobPassives() error {
	var passives []*JobPassive
	query := `select * from data.job_passives`

	if _, err := db.Select(&passives, query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("getJobPassives: %s", err.Error())
	}

	for _, p := range passives {
		JobPassives[p.ID] = p
	}

	return nil
}
