package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"hero-emulator/config"
	"hero-emulator/logging"
	"hero-emulator/utils"

	_ "github.com/lib/pq"
	gorp "gopkg.in/gorp.v1"
)

var (
	DROP_LIFETIME     = time.Duration(30) * time.Second
	FREEDROP_LIFETIME = time.Duration(3) * time.Second
	DROP_RATE         = utils.ParseFloat("10.0")
	DEFAULT_DROP_RATE = utils.ParseFloat("10.0")
	EXP_RATE          = utils.ParseFloat("500.0")
	DEFAULT_EXP_RATE  = utils.ParseFloat("500.0")
)

var (
	db                      *gorp.DbMap
	Init                    = make(chan bool, 1)
	GetFromRegister         func(int, int16, uint16) interface{}
	RemoveFromRegister      func(*Character)
	RemovePetFromRegister   func(c *Character)
	FindCharacterByPseudoID func(server int, ID uint16) *Character

	AccUpgrades    []byte
	ArmorUpgrades  []byte
	WeaponUpgrades []byte
	plusRates      = []int{800, 900, 950, 980, 990, 996, 999}
	logger         = logging.Logger
)

func InitDB() error {

	var (
		cfg = config.Default
		//drv         = cfg.Database.Driver
		ip          = cfg.Database.IP
		port        = cfg.Database.Port
		user        = cfg.Database.User
		pass        = cfg.Database.Password
		name        = cfg.Database.Name
		maxIdle     = cfg.Database.ConnMaxIdle
		maxOpen     = cfg.Database.ConnMaxOpen
		maxLifetime = cfg.Database.ConnMaxLifetime
		debug       = cfg.Database.Debug
		sslMode     = cfg.Database.SSLMode
		err         error
		conn        *sql.DB
	)

	conn, err = sql.Open("postgres", fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s", ip, port, user, pass, name, sslMode))
	if err != nil {
		return fmt.Errorf("Database connection error: %s", err.Error())
	}

	conn.SetMaxIdleConns(maxIdle)
	conn.SetMaxOpenConns(maxOpen)
	conn.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)

	if err = conn.Ping(); err != nil {
		return fmt.Errorf("Database connection error: %s", err.Error())
	}

	db = &gorp.DbMap{Db: conn, Dialect: gorp.PostgresDialect{}}
	db.AddTableWithNameAndSchema(PetExpInfo{}, "data", "pet_exp_table")
	db.AddTableWithNameAndSchema(ExpInfo{}, "data", "exp_table")
	db.AddTableWithNameAndSchema(NpcPosition{}, "data", "npc_pos_table").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(Item{}, "data", "items").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(SkillInfo{}, "data", "skills").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(Production{}, "data", "productions")
	db.AddTableWithNameAndSchema(CraftItem{}, "data", "craft_items")
	db.AddTableWithNameAndSchema(Stackable{}, "data", "stackables")
	db.AddTableWithNameAndSchema(Gambling{}, "data", "gambling")
	db.AddTableWithNameAndSchema(JobPassive{}, "data", "job_passives")
	db.AddTableWithNameAndSchema(SavePoint{}, "data", "save_points")
	db.AddTableWithNameAndSchema(HaxCode{}, "data", "hax_codes")
	db.AddTableWithNameAndSchema(ItemMelting{}, "data", "item_meltings")
	db.AddTableWithNameAndSchema(Gate{}, "data", "gates")
	db.AddTableWithNameAndSchema(DropInfo{}, "data", "drops").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(HtItem{}, "data", "ht_shop").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(NPCScript{}, "data", "npc_scripts")
	db.AddTableWithNameAndSchema(Fusion{}, "data", "advanced_fusion")
	db.AddTableWithNameAndSchema(Pet{}, "data", "pets").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(NPC{}, "data", "npc_table").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(BuffIcon{}, "data", "buff_icons")
	db.AddTableWithNameAndSchema(BuffInfection{}, "data", "buff_infections").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(Shop{}, "data", "shop_table").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(ShopItem{}, "data", "shop_items").SetKeys(false, "type")
	db.AddTableWithNameAndSchema(RelicLog{}, "data", "relic_log")
	db.AddTableWithNameAndSchema(RelicLog{}, "data", "item_set")
	db.AddTableWithNameAndSchema(Rank{}, "data", "reborn_system").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(ItemJudgement{}, "data", "item_judgement")
	db.AddTableWithNameAndSchema(FiveClan{}, "data", "fiveclan_war").SetKeys(false, "id")

	db.AddTableWithNameAndSchema(AI{}, "hops", "ai").SetKeys(true, "id")
	db.AddTableWithNameAndSchema(AiBuff{}, "hops", "ai_buffs").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(Character{}, "hops", "characters").SetKeys(true, "id")
	db.AddTableWithNameAndSchema(Buff{}, "hops", "characters_buffs").SetKeys(false, "id", "character_id")
	db.AddTableWithNameAndSchema(ConsignmentItem{}, "hops", "consignment").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(Guild{}, "hops", "guilds").SetKeys(true, "id")
	db.AddTableWithNameAndSchema(InventorySlot{}, "hops", "items_characters").SetKeys(true, "id")
	db.AddTableWithNameAndSchema(Relic{}, "hops", "relics")
	db.AddTableWithNameAndSchema(Server{}, "hops", "servers").SetKeys(true, "id")
	db.AddTableWithNameAndSchema(Skills{}, "hops", "skills").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(Stat{}, "hops", "stats").SetKeys(false, "id")
	db.AddTableWithNameAndSchema(User{}, "hops", "users").SetKeys(true, "id")

	if debug {
		db.TraceOn("[gorp]", log.New(os.Stdout, "myapp:", log.Lmicroseconds))
	}

	if err = resetDB(); err != nil {
		return err
	}

	if err = getAll(); err != nil {
		return err
	}

	Init <- err == nil
	return nil
}

func resetDB() error {

	query := `update hops.characters set is_active = false, is_online = false`
	if _, err := db.Exec(query); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("Reset DB error: %s", err.Error())
	}

	query = `update hops.users set ip = $1, server = 0`
	if _, err := db.Exec(query, ""); err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return fmt.Errorf("Reset DB error: %s", err.Error())
	}

	return nil
}

func getAll() error {

	callBacks := []func() error{getAllDrops, getScripts, getHaxCodes, getHTItems, getProductions, getCraftItem, getAdvancedFusions, getItemMeltings, getGates,
		getStackables, getAllItems, getSkillInfos, getGamblingItems, getJobPassives, getItemJudgements, getItemSet, getBuffIcons, getBuffInfections, getExps, getAllSavePoints,
		getRelics, getRelicLog, GetAllPetExps, GetAllPets, getAllShops, getAllShopItems}

	for _, cb := range callBacks {
		if err := cb(); err != nil {
			return err
		}
	}

	return nil
}
