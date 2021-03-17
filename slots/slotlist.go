package slots

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidOperationType = errors.New("invalid operation type")
	ErrOutOfRange           = errors.New("out of range")
	ErrOldEvent             = errors.New("old event")
)

type Mode int

const (
	ServerMode Mode = iota
	ClientMode
)

type Slot struct {
	ID        string
	Value     int32
	UpdatedAt time.Time
}

type SlotList struct {
	mu          sync.RWMutex
	slots       []Slot
	uuidToIndex map[string]int
	maxIndex    int
	mode        Mode
}

func NewSlotList(size int) *SlotList {
	return &SlotList{
		slots:       make([]Slot, 0, size * 2),
		maxIndex:    -1,
		uuidToIndex: make(map[string]int, size * 2),
		mode:        ServerMode,
	}
}

func NewGeneratedSlotList(size int) *SlotList {
	slotlist := NewSlotList(size)
	now := time.Now()
	for i := 0; i < size; i++ {
		_, _, _ = slotlist.insert(uuid.New().String(), i, rand.Int31(), now)
	}
	return slotlist
}

func (sl *SlotList) Get(pos int) (*Slot, bool) {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	if pos < 0 || pos > sl.maxIndex {
		return nil, false
	}
	s := &sl.slots[pos]
	return s, true
}

func (sl *SlotList) Copy() *SlotList {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	newSlotList := &SlotList{
		slots:       make([]Slot, len(sl.slots), len(sl.slots) * 2),
		maxIndex:    sl.maxIndex,
		uuidToIndex: make(map[string]int, len(sl.slots) * 2),
	}
	for i := range sl.slots {
		newSlotList.slots[i] = sl.slots[i]
	}
	for k, v := range sl.uuidToIndex {
		newSlotList.uuidToIndex[k] = v
	}
	return newSlotList
}

func (sl *SlotList) Slots() []Slot {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.slots
}

func (sl *SlotList) MaxIndex() int {
	sl.mu.RLock()
	defer sl.mu.RUnlock()
	return sl.maxIndex
}

func (sl *SlotList) SetMode(mode Mode) {
	sl.mu.Lock()
	sl.mode = mode
	sl.mu.Unlock()
}

func (sl *SlotList) DoOperation(op *SlotOperation) (time.Time, int, error) {
	switch op.OperationType {
	case Insert:
		return sl.insert(op.ID, op.Position, op.Value, op.UpdatedAt)
	case Delete:
		return sl.delete(op.ID, op.Position, op.UpdatedAt)
	case Update:
		return sl.update(op.ID, op.Position, op.Value, op.UpdatedAt)
	}
	return time.Time{}, 0, ErrInvalidOperationType
}

func (sl *SlotList) insert(id string, pos int, value int32, updatedAt time.Time) (time.Time, int, error) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	// check if the position is accurate
	if pos < 0 || sl.maxIndex != -1 && pos > sl.maxIndex+1 {
		return time.Time{}, pos, ErrOutOfRange
	}
	// if the ID and the pos is not the same
	// it means that client had old data
	// trying to figure out where whould be the actual position
	if sl.mode == ServerMode && pos != sl.maxIndex+1 && sl.slots[pos].ID != id {
		if actualPos, ok := sl.uuidToIndex[id]; ok {
			// found a correct position for the operation
			pos = actualPos
		}
		// didn't find a correct position for the operation
		// proceeding inplace
	}
	// we do not check the oldness of the operation as any insert should be done
	now := time.Now()
	sl.slots = append(sl.slots, Slot{})
	sl.maxIndex++
	if pos != sl.maxIndex+1 {
		for i := sl.maxIndex; i > pos; i-- {
			sl.slots[i].UpdatedAt = now
			sl.slots[i] = sl.slots[i-1]
			sl.uuidToIndex[sl.slots[i].ID] = i
		}
	}
	sl.slots[pos] = Slot{
		UpdatedAt: now,
		ID:        uuid.New().String(),
		Value:     value,
	}
	sl.uuidToIndex[sl.slots[pos].ID] = pos
	return now, pos, nil
}

func (sl *SlotList) delete(id string, pos int, updatedAt time.Time) (time.Time, int, error) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	// check if the position is accurate
	if pos < 0 || pos > sl.maxIndex {
		return time.Time{}, pos, ErrOutOfRange
	}
	// if the ID and the pos is not the same
	// it means that client had old data
	// trying to figure out where whould be the actual position
	if sl.mode == ServerMode && sl.slots[pos].ID != id {
		if actualPos, ok := sl.uuidToIndex[id]; ok {
			// found a correct position for the operation
			pos = actualPos
		} else {
			// didn't find a correct position for the operation
			// slot is already deleted
			return time.Time{}, pos, ErrOldEvent
		}
	}
	// the event was updated by someone else, we still delete it
	now := time.Now()
	delete(sl.uuidToIndex, sl.slots[sl.maxIndex].ID)
	for i := pos; i < sl.maxIndex; i++ {
		sl.slots[i] = sl.slots[i+1]
		sl.slots[i].UpdatedAt = now
		sl.uuidToIndex[sl.slots[i].ID] = i
	}
	sl.maxIndex--
	sl.slots = sl.slots[:sl.maxIndex+1]
	return now, pos, nil
}

func (sl *SlotList) update(id string, pos int, value int32, updatedAt time.Time) (time.Time, int, error) {
	sl.mu.Lock()
	defer sl.mu.Unlock()
	// check if the position is accurate
	if pos < 0 || pos > sl.maxIndex {
		return time.Time{}, pos, ErrOutOfRange
	}
	// if the ID and the pos is not the same
	// it means that client had old data
	// trying to figure out where whould be the actual position
	if sl.mode == ServerMode && sl.slots[pos].ID != id {
		if actualPos, ok := sl.uuidToIndex[id]; ok {
			// found a correct position for the operation
			pos = actualPos
		} else {
			// didn't find a correct position for the operation
			// slot is already deleted
			return time.Time{}, pos, ErrOldEvent
		}
	}
	// if someone updated the position with more relative data
	// then skip the update
	if sl.mode == ServerMode && sl.maxIndex != -1 && updatedAt.Before(sl.slots[pos].UpdatedAt) {
		return time.Time{}, pos, ErrOldEvent
	}

	now := time.Now()
	sl.slots[pos].UpdatedAt = now
	sl.slots[pos].Value = value
	return now, pos, nil
}
