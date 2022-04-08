package database

import "sync"

var (
	Trades = make(map[string]*Trade)
	tMutex sync.RWMutex
)

type Trade struct {
	Sender     *Trader
	Receiver   *Trader
	Completing bool
}

type Trader struct {
	Character *Character
	Slots     map[int16]int16
	Gold      uint64
	Accepted  bool
}

func (t *Trade) New(sender, receiver *Character) {
	t.Sender = &Trader{Character: sender, Slots: make(map[int16]int16)}
	t.Receiver = &Trader{Character: receiver, Slots: make(map[int16]int16)}

	tMutex.Lock()
	defer tMutex.Unlock()

	Trades[sender.UserID] = t
	Trades[receiver.UserID] = t
}

func (t *Trade) Delete() {
	suID := t.Sender.Character.UserID
	ruID := t.Receiver.Character.UserID

	t.Sender.Character.TradeID = ""
	t.Receiver.Character.TradeID = ""

	tMutex.Lock()
	defer tMutex.Unlock()

	delete(Trades, suID)
	delete(Trades, ruID)
}

func FindTrade(c *Character) *Trade {
	tMutex.RLock()
	t := Trades[c.UserID]
	tMutex.RUnlock()

	return t
}
