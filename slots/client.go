package slots

import (
	"errors"
	"log"
	"time"
	"sync"
)

var (
	ErrClientOutOfRange = errors.New("client out of range")
)

type Client struct {
	mu sync.RWMutex
	slotlist *SlotList
	server   *Server
	lastSync time.Time
}

func NewClient(s *Server) *Client {
	return &Client{
		server: s,
	}
}

func (c *Client) UpdateCurrentState() error {
	slotlist, lastSync := c.server.CurrentState()
	slotlist.SetMode(ClientMode)
	c.mu.Lock()
	c.slotlist = slotlist
	c.lastSync = lastSync
	c.mu.Unlock()
	return nil
}

func (c *Client) CurrentState() *SlotList {
	c.mu.RLock()
	slotlist := c.slotlist
	c.mu.RUnlock()
	return slotlist.Copy()
}

func (c *Client) Connect() error {
	updateCh := c.server.Register(c)
	go func(c *Client, updateCh <-chan *SlotOperation) {
		var previous string
		for op := range updateCh {
			if c.lastSync.After(op.UpdatedAt) || c.lastSync.Equal(op.UpdatedAt) {
				previous = op.UniqueID
				continue
			}
			if op.PreviousID != previous {
				if err := c.UpdateCurrentState(); err != nil {
					log.Printf("resync err: %v", err)
				}
				log.Println("current state resync")
				previous = op.UniqueID
				continue
			}
			if _, _, err := c.slotlist.DoOperation(op); err != nil {
				log.Printf("Operation %#v: %v", op, err)
			}
			previous = op.UniqueID
		}
	}(c, updateCh)
	return nil
}

func (c *Client) GetLocal(pos int) (*Slot, bool) {
	c.mu.RLock()
	slotlist := c.slotlist
	c.mu.RUnlock()
	return slotlist.Get(pos)
}

func (c *Client) MaxIndex() int {
	c.mu.RLock()
	slotlist := c.slotlist
	c.mu.RUnlock()
	return slotlist.MaxIndex()
}

func (c *Client) Insert(pos int, value int32) error {
	return c.Do(Insert, pos, value)
}

func (c *Client) Update(pos int, value int32) error {
	return c.Do(Update, pos, value)
}

func (c *Client) Delete(pos int) error {
	return c.Do(Delete, pos, 0)
}

func (c *Client) Do(op OperationType, pos int, value int32) error {
	s, ok := c.GetLocal(pos)
	if !ok {
		return ErrClientOutOfRange
	}
	return c.server.Do(op, s.ID, pos, value, s.UpdatedAt)
}
