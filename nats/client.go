package nats

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/gnatsd/server"
	"github.com/nats-io/nats.go"
)

const (
	HOUSTON_CH = "Houston"
)

var (
	conn *nats.Conn
)

type CastPacket struct {
	CastNear    bool `json:"cast_near"`
	CharacterID int  `json:"character_id"`
	MobID       int  `json:"mob_id"`
	PetID       int  `json:"pet_id"`
	DropID      int  `json:"loot_id"`
	Location    *struct {
		X float64
		Y float64
	} `json:"location"`
	MaxDistance float64 `json:"max_distance"`
	Data        []byte  `json:"data"`
	Type        int8    `json:"type"`
}

func ConnectSelf(opts *server.Options) (*nats.Conn, error) {
	var err error
	if opts == nil {
		opts = &DefaultOptions
	}

	url := fmt.Sprintf("nats://%s:%d", opts.Host, opts.Port)
	conn, err = nats.Connect(url, nats.Timeout(5*time.Second))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func Connection() *nats.Conn {
	return conn
}

func (p *CastPacket) Cast() error {

	data, err := json.Marshal(p)
	if err != nil {
		return err
	}

	return Connection().Publish(HOUSTON_CH, data)
}
