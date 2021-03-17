package slots

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestSlotList(t *testing.T) {
	sl := NewSlotList(100)
	// insert slots in empty list
	for i := 0; i < 3; i++ {
		_, _, err := sl.insert(uuid.New().String(), i, int32(i), time.Now())
		assert.NoError(t, err)
	}
	for i := 0; i < 3; i++ {
		assert.EqualValues(t, i, sl.slots[i].Value)
	}
	testUUIDIndex(t, sl)

	// insert slots in the middle
	for i := 0; i < 3; i++ {
		_, _, err := sl.insert(uuid.New().String(), 2*i, int32(i), time.Now())
		assert.NoError(t, err)
	}
	for i := 0; i < 3; i++ {
		assert.EqualValues(t, i, sl.slots[2*i].Value)
		assert.EqualValues(t, i, sl.slots[2*i+1].Value)
	}
	testUUIDIndex(t, sl)

	// insert invalid
	_, _, err := sl.insert(uuid.New().String(), -1, 10, time.Now())
	assert.EqualValues(t, ErrOutOfRange, err)
	_, _, err = sl.insert(uuid.New().String(), 8, 10, time.Now())
	assert.EqualValues(t, ErrOutOfRange, err)

	// update slots
	for i := 0; i < 6; i++ {
		_, _, err := sl.update(sl.slots[i].ID, i, int32(-i), time.Now())
		assert.NoError(t, err)
	}
	for i := 0; i < 6; i++ {
		assert.EqualValues(t, i, -sl.slots[i].Value)
	}
	testUUIDIndex(t, sl)

	// update deleted slot
	// uuid is not stated in storage
	_, _, err = sl.update(uuid.New().String(), 0, 0, time.Now())
	assert.EqualValues(t, ErrOldEvent, err)

	// old slot
	// not updating
	_, _, err = sl.update(sl.slots[0].ID, 0, 0, time.Now().Add(-time.Second))
	assert.EqualValues(t, ErrOldEvent, err)

	// delete slots
	for i := 2; i >= 0; i-- {
		_, _, err := sl.delete(sl.slots[2*i].ID, 2*i, time.Now())
		assert.NoError(t, err)
	}
	for i := 0; i < 3; i++ {
		assert.EqualValues(t, 2*i+1, -sl.slots[i].Value)
	}
	testUUIDIndex(t, sl)

	// delete slot
	// uuid is not stated in storage
	_, _, err = sl.delete(uuid.New().String(), 0, time.Now())
	assert.EqualValues(t, ErrOldEvent, err)

	// insert slot in wrong position but insert correct
	_, _, err = sl.insert(sl.slots[0].ID, 2, 10, time.Now())
	assert.NoError(t, err)
	assert.EqualValues(t, 10, sl.slots[0].Value)

	// update slot in wrong position but update correct
	_, _, err = sl.update(sl.slots[0].ID, 2, 11, time.Now())
	assert.NoError(t, err)
	assert.EqualValues(t, 11, sl.slots[0].Value)

	// delete slot in wrong position but delete correct
	v := sl.slots[1].Value
	_, _, err = sl.delete(sl.slots[0].ID, 2, time.Now())
	assert.NoError(t, err)
	assert.EqualValues(t, v, sl.slots[0].Value)
}

func testUUIDIndex(t *testing.T, sl *SlotList) {
	for i := 0; i < len(sl.slots); i++ {
		assert.EqualValues(t, i, sl.uuidToIndex[sl.slots[i].ID])
	}
}
