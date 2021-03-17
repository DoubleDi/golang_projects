package slots

import (
	"errors"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func randomOperation() OperationType {
	return OperationType(rand.Intn(int(AllOperationTypes)))
}


// this test creates a server and 20 clients
// each client downloads the current state and starts to 
// update/insert/delete slots
// at some point they finish and we sleep for one second
// and check the consistency between the server and all the clients
// some of the clients miss parts of data so they get resynced with the server
func TestServer(t *testing.T) {
	server := NewGeneratedServer(10000)
	go server.ProcessOperations()
	go server.ProcessNotifications()

	wg := &sync.WaitGroup{}
	clients := []*Client{}
	for i := 0; i < 20; i++ {
		client := NewClient(server)
		assert.NoError(t, client.UpdateCurrentState())
		assert.NoError(t, client.Connect())
		clients = append(clients, client)
		wg.Add(1)
		go func(client *Client) {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				op := randomOperation()
				max := client.MaxIndex() + 1
				if max == 0 {
					max += 1
				}
				index := rand.Intn(max)
				value := rand.Int31()
				err := client.Do(op, index, value)
				if !errors.Is(err, ErrClientOutOfRange) {
					assert.NoError(t, err)
				}
				time.Sleep(time.Millisecond * 10)
			}
		}(client)
		time.Sleep(time.Millisecond * 100)
	}
	wg.Wait()
	time.Sleep(time.Second)

	serverState, _ := server.CurrentState()
	serverSlots := serverState.Slots()
	m := 0
	for j, client := range clients {
		clientSlots := client.CurrentState().Slots()
		assert.EqualValues(t, len(serverSlots), len(clientSlots), "%d", j)
		for i := range serverSlots {
			if serverSlots[i].Value != clientSlots[i].Value {
				m++
			}
			assert.EqualValues(t, serverSlots[i].Value, clientSlots[i].Value, "%d", j)
		}
	}
	assert.EqualValues(t, 0, m)
}
