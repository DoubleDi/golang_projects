package slots

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

type OperationType int

const (
	Insert OperationType = iota
	Update
	Delete
	AllOperationTypes
)

type SlotOperation struct {
	OperationType OperationType
	ID            string
	Position      int
	Value         int32
	UpdatedAt     time.Time

	UniqueID   string
	PreviousID string
}

type Server struct {
	slotlist      *SlotList
	clients       []chan<- *SlotOperation
	pipeline      chan *SlotOperation
	notifications chan *SlotOperation
}

func NewServer(size int) *Server {
	return &Server{
		slotlist:      NewSlotList(size),
		pipeline:      make(chan *SlotOperation, 100),
		notifications: make(chan *SlotOperation, 10000),
	}
}

func NewGeneratedServer(size int) *Server {
	server := NewServer(size)
	server.slotlist = NewGeneratedSlotList(size)
	return server
}

func (s *Server) Register(c *Client) chan *SlotOperation {
	ch := make(chan *SlotOperation, 100)
	s.clients = append(s.clients, ch)
	return ch
}

func (s *Server) CurrentState() (*SlotList, time.Time) {
	return s.slotlist.Copy(), time.Now()
}

func (s *Server) Insert(id string, pos int, value int32, updatedAt time.Time) error {
	return s.Do(Insert, id, pos, value, updatedAt)
}

func (s *Server) Update(id string, pos int, value int32, updatedAt time.Time) error {
	return s.Do(Update, id, pos, value, updatedAt)
}

func (s *Server) Delete(id string, pos int, updatedAt time.Time) error {
	return s.Do(Delete, id, pos, 0, updatedAt)
}

func (s *Server) Do(op OperationType, id string, pos int, value int32, updatedAt time.Time) error {
	s.pipeline <- &SlotOperation{
		OperationType: op,
		ID:            id,
		Position:      pos,
		UpdatedAt:     updatedAt,
		Value:         value,
		UniqueID:      uuid.New().String(),
	}
	return nil
}

func (s *Server) ProcessOperations() {
	var previous string
	for op := range s.pipeline {
		t, pos, err := s.slotlist.DoOperation(op)
		if err != nil {
			// log.Printf("Error on operation %#v: %v\n", op, err)
			continue
		}
		op.UpdatedAt = t
		op.Position = pos
		op.PreviousID = previous
		s.notifications <- op
		previous = op.UniqueID
	}
}

func (s *Server) ProcessNotifications() {
	for op := range s.notifications {
		wg := &sync.WaitGroup{}
		for _, c := range s.clients {
			wg.Add(1)
			func(c chan<- *SlotOperation) {
				defer wg.Done()
				select {
				case c <- op:
				case <-time.After(time.Millisecond):
				}
			}(c)
		}
		wg.Wait()
	}
}
