package database

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"hero-emulator/gold"

	"github.com/thoas/go-funk"
	gorp "gopkg.in/gorp.v1"
	null "gopkg.in/guregu/null.v3"
)

var (
	orders     = map[int]string{1: "item_name", 2: "quantity", 3: "expires_at", 4: "price"}
	categories = map[int][]int{
		1:  {70, 71, 99, 100, 101, 102, 103, 104, 105, 107, 108},
		2:  {121, 122, 123, 124, 175},
		3:  {131, 132, 133, 134, 90},
		4:  {64},
		5:  {135, 136, 137, 221, 222, 223},
		6:  {161},
		7:  {147, 148, 149, 150, 151, 152, 153, 154, 156},
		8:  {80, 190, 191, 192, 194, 195},
		9:  {164, 166, 167, 187, 189, 219},
		10: {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, 27, 28, 29, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 65, 66, 67, 68, 69, 72, 73, 74, 75, 76, 77, 78, 79, 81, 82, 83, 84, 85, 86, 87, 88, 89, 91, 92, 93, 94, 95, 96, 97, 98, 106, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 125, 126, 127, 128, 129, 130, 138, 139, 140, 141, 142, 143, 144, 145, 146, 155, 157, 158, 159, 160, 162, 163, 165, 168, 169, 170, 171, 172, 173, 174, 176, 177, 178, 179, 180, 181, 182, 183, 184, 185, 186, 188, 193, 196, 197, 198, 199, 200, 201, 202, 203, 204, 205, 206, 207, 208, 209, 210, 211, 212, 213, 214, 215, 216, 217, 218, 220, 224, 225, 226, 227, 228, 229, 230, 231, 232, 233, 234, 235, 236, 237, 238, 239, 240, 241, 242, 243, 244, 245, 246, 247, 248, 249, 250, 251, 252, 253, 254, 255},
		11: {102, 103}, 12: {105}, 13: {108}, 14: {104}, 15: {107}, 16: {101}, 17: {99, 100}, 18: {70, 71}, 19: {},
		20: {121}, 21: {122}, 22: {123}, 23: {124}, 24: {-121}, 25: {-122}, 26: {-123}, 27: {-124}, 28: {175},
		29: {131}, 30: {132}, 31: {133}, 32: {134}, 33: {312}, 34: {313}, 35: {314}, 36: {315},
		37: {397}, 38: {398}, 39: {399}, 40: {400}, 41: {401},
		42: {221}, 43: {135}, 44: {136}, 45: {137}, 46: {222, 223},
		47: {161}, // skill book
		48: {147, 148, 149, 150, 151, 152, 153, 154, 156},
		49: {194}, 50: {195}, 51: {192}, 52: {191}, 53: {80}, 54: {190},
		55: {164, 166, 167, 219}, 56: {187, 189},
	}
)

type ConsignmentItem struct {
	ID        int       `db:"id" json:"id"`
	SellerID  int       `db:"seller_id" json:"seller_id"`
	ItemName  string    `db:"item_name" json:"item_name"`
	Quantity  int       `db:"quantity" json:"quantity"`
	Price     uint64    `db:"price" json:"price"`
	IsSold    bool      `db:"is_sold" json:"is_sold"`
	ExpiresAt null.Time `db:"expires_at" json:"expires_at"`
}

func (e *ConsignmentItem) PreInsert(s gorp.SqlExecutor) error {
	exp := time.Now().UTC().Add(time.Hour * 24 * 3)
	e.ExpiresAt = null.TimeFrom(exp)

	return nil
}

func (e *ConsignmentItem) Create() error {
	return db.Insert(e)
}

func (e *ConsignmentItem) Delete() error {
	_, err := db.Delete(e)
	return err
}

func (e *ConsignmentItem) Update() error {
	_, err := db.Update(e)
	return err
}

func GetConsignmentItems(page, category, minUpgLevel, maxUpgLevel, orderBy int, minPrice, maxPrice uint64, itemName string) ([]*ConsignmentItem, int64, error) {

	if maxPrice == 150*gold.B {
		maxPrice = math.MaxInt64
	}

	order := int(math.Abs(float64(int8(orderBy))))

	direction := "asc"
	if int8(orderBy) < 0 {
		direction = "desc"
	}

	quotedName := fmt.Sprintf("%%%s%%", itemName)

	query := ``
	cats, ok := categories[category]
	if !ok {
		query = `select c.* from hops.consignment c
		inner join hops.items_characters ic on c.id = ic.id
		inner join data.items i on i.id = ic.item_id
		where is_sold = false and c.price >= $1 and c.price <= $2 and 
		lower(c.item_name) like lower($3) and ic.plus >= $4 and ic.plus <= $5`

	} else if cats[0] < 0 {
		query = `select c.* from hops.consignment c
		inner join hops.items_characters ic on c.id = ic.id
		inner join data.items i on i.id = ic.item_id
		where is_sold = false and c.price >= $1 and c.price <= $2 and 
		lower(c.item_name) like lower($3) and ic.plus >= $4 and ic.plus <= $5 and 
		-i."type" in (%s) and i.ht_type > 0 `

		query = fmt.Sprintf(query, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(cats)), ","), "[]"))

	} else if cats[0] > 255 {
		query = `select c.* from hops.consignment c
		inner join hops.items_characters ic on c.id = ic.id
		inner join data.items i on i.id = ic.item_id
		where is_sold = false and c.price >= $1 and c.price <= $2 and 
		lower(c.item_name) like lower($3) and ic.plus >= $4 and ic.plus <= $5 and i.slot in (%s)`

		query = fmt.Sprintf(query, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(cats)), ","), "[]"))

	} else {
		query = `select c.* from hops.consignment c
		inner join hops.items_characters ic on c.id = ic.id
		inner join data.items i on i.id = ic.item_id
		where is_sold = false and c.price >= $1 and c.price <= $2 and 
		lower(c.item_name) like lower($3) and ic.plus >= $4 and ic.plus <= $5 and i.type in (%s)`

		query = fmt.Sprintf(query, strings.Trim(strings.Join(strings.Fields(fmt.Sprint(cats)), ","), "[]"))
	}

	countQuery := strings.Replace(query, "c.*", "count(*)", 1)
	query = fmt.Sprintf("%s order by %s %s offset $6 limit 20", query, orders[order], direction)

	var err error
	items := []*ConsignmentItem{}
	if _, err = db.Select(&items, query, minPrice, maxPrice, quotedName, minUpgLevel, maxUpgLevel, (page-1)*20); err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("GetConsignmentItems: %s", err.Error())
	}

	count := int64(0)
	if count, err = db.SelectInt(countQuery, minPrice, maxPrice, quotedName, minUpgLevel, maxUpgLevel); err != nil {
		if err == sql.ErrNoRows {
			return nil, 0, nil
		}
		return nil, 0, fmt.Errorf("GetConsignmentItemCount: %s", err.Error())
	}

	return items, count, nil
}

func CountConsignmentItems(category, minUpgLevel, maxUpgLevel int, minPrice, maxPrice uint64, itemName string) (int64, error) {

	var (
		err error
	)

	if maxPrice == 50*gold.B {
		maxPrice = math.MaxInt64
	}

	query := `select * from hops.consignment c where is_sold = false and price >= $1 and price <= $2`

	var items []*ConsignmentItem
	if _, err = db.Select(&items, query, minPrice, maxPrice); err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, fmt.Errorf("GetConsignmentItems: %s", err.Error())
	}

	cats, ok := categories[category]
	if !ok {
		return int64(len(items)), nil
	}

	items = funk.Filter(items, func(item *ConsignmentItem) bool {
		slot, err := FindInventorySlotByID(item.ID)
		if err != nil {
			return false
		}

		info := Items[slot.ItemID]
		if info == nil || !strings.Contains(strings.ToLower(info.Name), strings.ToLower(itemName)) ||
			slot.Plus < uint8(minUpgLevel) || slot.Plus > uint8(maxUpgLevel) || len(cats) == 0 {

			return false
		}

		if cats[0] < 0 {
			return funk.ContainsInt(cats, -int(info.Type)) && info.HtType > 0

		} else if cats[0] > 255 {
			return funk.ContainsInt(cats, info.Slot)
		}

		return funk.ContainsInt(cats, int(info.Type))
	}).([]*ConsignmentItem)

	return int64(len(items)), nil
}

func FindConsignmentItemsBySellerID(sellerID int) ([]*ConsignmentItem, error) {

	query := `select * from hops.consignment where seller_id = $1 order by expires_at desc`

	items := []*ConsignmentItem{}
	if _, err := db.Select(&items, query, sellerID); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindConsignmentItemsBySellerID: %s", err.Error())
	}

	return items, nil
}

func FindConsignmentItemByID(id int) (*ConsignmentItem, error) {

	query := `select * from hops.consignment where id = $1`

	item := &ConsignmentItem{}
	if err := db.SelectOne(&item, query, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("FindConsignmentItemByID: %s", err.Error())
	}

	return item, nil
}
