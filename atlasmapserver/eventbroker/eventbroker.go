package eventbroker

import (
	"context"
	"errors"
	"sync"

	"github.com/antihax/AtlasMap/internal/atlasdb"
	"github.com/rs/zerolog/log"
)

type EventBroker struct {
	users     sync.Map
	usersMut  sync.Mutex
	tribes    sync.Map
	tribesMut sync.Mutex
	db        *atlasdb.AtlasDB
}

func NewEventBroker(db *atlasdb.AtlasDB) *EventBroker {
	return &EventBroker{db: db}
}

func (s *EventBroker) AddUser(steamID string, tribeID int64) chan string {
	channel := make(chan string, 20)
	s.usersMut.Lock()
	usersInterface, loaded := s.users.LoadOrStore(steamID, []chan string{channel})
	if loaded {
		users := usersInterface.([]chan string)
		s.users.Store(steamID, append(users, channel))
	}
	s.usersMut.Unlock()

	s.tribesMut.Lock()
	tribesInterface, loaded := s.tribes.LoadOrStore(tribeID, []chan string{channel})
	tribes := tribesInterface.([]chan string)
	if loaded {
		s.tribes.Store(tribeID, append(tribes, channel))
	}

	if !loaded || len(tribes) == 1 {
		s.subTribe(tribeID)
	}

	s.tribesMut.Unlock()
	return channel
}

func (s *EventBroker) RemoveChannel(channel chan string) {
	s.usersMut.Lock()
	s.tribesMut.Lock()

	s.users.Range(func(k, v interface{}) bool {
		users := v.([]chan string)
		changed := false
		for i := len(users) - 1; i >= 0; i-- {
			if users[i] == channel {
				users = append(users[:i], users[i+1:]...)
				changed = true
			}
		}
		if changed {
			s.users.Store(k, users)
		}
		return true
	})

	s.tribes.Range(func(k, v interface{}) bool {
		tribes := v.([]chan string)
		changed := false
		for i := len(tribes) - 1; i >= 0; i-- {
			if tribes[i] == channel {
				tribes = append(tribes[:i], tribes[i+1:]...)
				changed = true
			}
		}
		if changed {
			s.tribes.Store(k, tribes)
		}
		return true
	})

	s.usersMut.Unlock()
	s.tribesMut.Unlock()

	close(channel)
}

func (s *EventBroker) SendUser(steamID string, value string) error {
	v, ok := s.users.Load(steamID)
	if !ok {
		return errors.New("steamID not found")
	}
	for _, c := range v.([]chan string) {
		c <- value
	}

	return nil
}

func (s *EventBroker) SendTribe(tribeID int64, value string) error {
	v, ok := s.tribes.Load(tribeID)
	if !ok {
		return errors.New("tribeID not found")
	}
	for _, c := range v.([]chan string) {
		c <- value
	}

	return nil
}

func (s *EventBroker) subTribe(tribeID int64) {
	ctx, cancel := context.WithCancel(context.Background())
	c := s.db.SubTribe(ctx, tribeID)
	go func() {
		for {
			if err := s.SendTribe(tribeID, <-c); err != nil {
				// exit out and close the channel
				log.Err(err).Msg("broker.sendtribe")
				cancel()
				return
			}
		}
	}()
}
