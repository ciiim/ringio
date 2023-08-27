package fs

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"sync"

	"golang.org/x/exp/slices"
)

//事务模块

type TransactionID string

type Transaction struct {
	rollbackFn []func()
}

type TransactionSystem struct {
	mu      sync.RWMutex
	actions map[TransactionID]*Transaction
}

var TS = &TransactionSystem{
	actions: make(map[TransactionID]*Transaction),
}

func (t *TransactionSystem) NewTransaction(name string) TransactionID {
	t.mu.Lock()
	defer t.mu.Unlock()
	sum := md5.Sum([]byte(name))
	ID := TransactionID(hex.EncodeToString(sum[:]))
	t.actions[ID] = &Transaction{}
	return ID
}

func (t *TransactionSystem) Commit(ID TransactionID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.actions, ID)
}

func (t *TransactionSystem) AddRollbackAction(ID TransactionID, fn func()) {
	t.mu.RLock()
	defer t.mu.RUnlock()
	t.actions[ID].rollbackFn = append(t.actions[ID].rollbackFn, fn)
}

func (t *TransactionSystem) Rollback(ID TransactionID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	log.Printf("[Transaction] Rollback transaction <%s>", ID)
	slices.Reverse[[]func()](t.actions[ID].rollbackFn)
	for _, fn := range t.actions[ID].rollbackFn {
		fn()
	}
	delete(t.actions, ID)
}
