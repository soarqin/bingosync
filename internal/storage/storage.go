package storage

import (
	"bingosync/internal/game"
	"encoding/json"
	"log"

	"github.com/dgraph-io/badger/v4"
)

// RoomData represents the persistable room state
type RoomData struct {
	ID       string     `json:"id"`
	Name     string     `json:"name"`
	Password string     `json:"password"`
	Game     *game.Game `json:"game"`
}

// Storage handles persistence using Badger
type Storage struct {
	db *badger.DB
}

// New creates a new Storage instance
func New(dataDir string) (*Storage, error) {
	opts := badger.DefaultOptions(dataDir)
	opts.Logger = nil // Disable badger's default logger

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &Storage{db: db}, nil
}

// Close closes the storage
func (s *Storage) Close() error {
	return s.db.Close()
}

// SaveRoom saves a room to storage
func (s *Storage) SaveRoom(data *RoomData) error {
	return s.db.Update(func(txn *badger.Txn) error {
		value, err := json.Marshal(data)
		if err != nil {
			return err
		}
		return txn.Set([]byte("room:"+data.ID), value)
	})
}

// DeleteRoom removes a room from storage
func (s *Storage) DeleteRoom(id string) error {
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte("room:" + id))
	})
}

// LoadAllRooms loads all rooms from storage
func (s *Storage) LoadAllRooms() ([]*RoomData, error) {
	var rooms []*RoomData

	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		prefix := []byte("room:")
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			var data RoomData
			err := item.Value(func(val []byte) error {
				return json.Unmarshal(val, &data)
			})
			if err != nil {
				log.Printf("Error unmarshaling room data: %v", err)
				continue
			}
			rooms = append(rooms, &data)
		}
		return nil
	})

	return rooms, err
}
