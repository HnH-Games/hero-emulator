package api

import (
	"encoding/json"
	"log"
	"net"
	"time"

	"hero-emulator/database"

	"github.com/thoas/go-funk"
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
	"gopkg.in/guregu/null.v3"
)

type ApiService struct{}

const (
	port = ":9000"
)

func InitGRPC() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	RegisterApiServer(s, &ApiService{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func (s *ApiService) GetUserByName(ctx context.Context, req *GetUserRequest) (*User, error) {

	user, err := database.FindUserByName(req.Username)
	if err != nil || user == nil {
		return nil, err
	}

	return &User{
		Id:         user.ID,
		Cash:       int64(user.NCash),
		CreatedAt:  user.CreatedAt.Time.String(),
		DisabledAt: user.DisabledUntil.Time.String(),
		Ip:         user.ConnectedIP,
		Mail:       user.Mail,
		Password:   user.Password,
		Server:     int32(user.ConnectedServer),
		Username:   user.Username,
		Usertype:   int32(user.UserType),
	}, nil
}

func (s *ApiService) GetUserByID(ctx context.Context, req *GetUserRequest) (*User, error) {

	user, err := database.FindUserByID(req.Id)
	if err != nil || user == nil {
		return nil, err
	}

	return &User{
		Id:         user.ID,
		Cash:       int64(user.NCash),
		CreatedAt:  user.CreatedAt.Time.String(),
		DisabledAt: user.DisabledUntil.Time.String(),
		Ip:         user.ConnectedIP,
		Mail:       user.Mail,
		Password:   user.Password,
		Server:     int32(user.ConnectedServer),
		Username:   user.Username,
		Usertype:   int32(user.UserType),
	}, nil
}

func (s *ApiService) Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error) {

	user, err := database.FindUserByMail(req.Mail)
	if err != nil || user != nil {
		return &RegisterResponse{Ok: false}, err
	}

	user, err = database.FindUserByName(req.Username)
	if err != nil || user != nil {
		return &RegisterResponse{Ok: false}, err
	}

	user = &database.User{
		BankGold:  0,
		CreatedAt: null.NewTime(time.Now(), true),
		Mail:      req.Mail,
		NCash:     0,
		Password:  req.Password,
		UserType:  1,
		Username:  req.Username,
	}

	err = user.Create()
	if err != nil {
		return &RegisterResponse{Ok: false}, err
	}

	return &RegisterResponse{Ok: true, UserID: user.ID}, nil
}

func (s *ApiService) GetServers(ctx context.Context, req *Empty) (*GetServerResponse, error) {

	resp := &GetServerResponse{Servers: []*Server{}}

	servers, err := database.GetServers()
	if err != nil {
		return resp, err
	}

	for _, s := range servers {
		item := &Server{
			Name:         s.Name,
			Maxplayers:   int32(s.MaxUsers),
			Totalplayers: int32(s.ConnectedUsers),
		}

		resp.Servers = append(resp.Servers, item)
	}

	return resp, nil
}

func (s *ApiService) GetTavern(ctx context.Context, req *Empty) (*GetTavernResponse, error) {

	items := funk.Values(database.HTItems).([]*database.HtItem)
	items = funk.Filter(items, func(i *database.HtItem) bool {
		return i.IsActive
	}).([]*database.HtItem)

	titles := []string{"Medicine", "Book", "Pet", "Costume", "Premium", "Talisman", "Etc."}
	tavernMap := make(map[string][]interface{})

	for _, i := range items {
		title := titles[i.HTID/1000]
		info := database.Items[int64(i.ID)]

		quantity := int16(1)
		if info.Timer > 0 {
			quantity = int16(info.Timer)
		}

		item := struct {
			Name      string `json:"name"`
			NCash     int    `json:"ncash"`
			Quantity  int16  `json:"quantity"`
			IsNew     bool   `json:"is_new"`
			IsPopular bool   `json:"is_popular"`
		}{info.Name, i.Cash, quantity, i.IsNew, i.IsPopular}

		tavernMap[title] = append(tavernMap[title], item)
	}

	data, err := json.Marshal(tavernMap)
	if err != nil {
		return nil, err
	}

	resp := &GetTavernResponse{Items: data}
	return resp, nil
}
