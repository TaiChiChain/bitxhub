package common

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

	"github.com/axiomesh/axiom-ledger/internal/ledger"
)

type VMMap[K, V any] struct {
	contractAccount ledger.IAccount
	mapName         string
	keyToString     func(key K) string
}

func NewVMMap[K, V any](contractAccount ledger.IAccount, mapName string, keyToString func(key K) string) *VMMap[K, V] {
	return &VMMap[K, V]{
		contractAccount: contractAccount,
		mapName:         mapName,
		keyToString:     keyToString,
	}
}

func (m *VMMap[K, V]) stateKey(key K) []byte {
	return []byte(fmt.Sprintf("%s_%s", m.mapName, m.keyToString(key)))
}

func (m *VMMap[K, V]) Get(k K) (exist bool, v V, err error) {
	exist, data := m.contractAccount.GetState(m.stateKey(k))
	if !exist || len(data) == 0 || data[0] == 0 {
		return false, v, nil
	}
	if err := json.Unmarshal(data[1:], &v); err != nil {
		return false, v, err
	}
	return true, v, nil
}

func (m *VMMap[K, V]) MustGet(k K) (v V, err error) {
	exist, data := m.contractAccount.GetState(m.stateKey(k))
	if !exist || len(data) == 0 || data[0] == 0 {
		return v, errors.Errorf("system contract[%s] map[%s] key[%s] not exist", m.contractAccount.GetAddress(), m.mapName, m.keyToString(k))
	}
	if err := json.Unmarshal(data[1:], &v); err != nil {
		return v, err
	}
	return v, nil
}

func (m *VMMap[K, V]) Has(k K) bool {
	exist, data := m.contractAccount.GetState(m.stateKey(k))
	return !(!exist || len(data) == 0 || data[0] == 0)
}

func (m *VMMap[K, V]) Put(k K, v V) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	m.contractAccount.SetState(m.stateKey(k), append([]byte{1}, data...))
	return nil
}

func (m *VMMap[K, V]) Delete(k K) error {
	m.contractAccount.SetState(m.stateKey(k), []byte{0})
	return nil
}

type VMSlot[V any] struct {
	contractAccount ledger.IAccount
	slotName        string
}

func NewVMSlot[V any](contractAccount ledger.IAccount, slotName string) *VMSlot[V] {
	return &VMSlot[V]{
		contractAccount: contractAccount,
		slotName:        slotName,
	}
}

func (s *VMSlot[V]) stateKey() []byte {
	return []byte(s.slotName)
}

func (s *VMSlot[V]) Get() (exist bool, v V, err error) {
	exist, data := s.contractAccount.GetState(s.stateKey())
	if !exist || len(data) == 0 || data[0] == 0 {
		return false, v, nil
	}
	if err := json.Unmarshal(data[1:], &v); err != nil {
		return false, v, err
	}
	return true, v, nil
}

func (s *VMSlot[V]) MustGet() (v V, err error) {
	exist, data := s.contractAccount.GetState(s.stateKey())
	if !exist || len(data) == 0 || data[0] == 0 {
		return v, errors.Errorf("system contract[%s] slot[%s] not exist", s.contractAccount.GetAddress(), s.slotName)
	}
	if err := json.Unmarshal(data[1:], &v); err != nil {
		return v, err
	}
	return v, nil
}

func (s *VMSlot[V]) Has() bool {
	exist, data := s.contractAccount.GetState(s.stateKey())
	return !(!exist || len(data) == 0 || data[0] == 0)
}

func (s *VMSlot[V]) Put(v V) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	s.contractAccount.SetState(s.stateKey(), append([]byte{1}, data...))
	return nil
}

func (s *VMSlot[V]) Delete() error {
	s.contractAccount.SetState(s.stateKey(), []byte{0})
	return nil
}
