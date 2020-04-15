package transaction

import (
	"log"
	"time"

	"github.com/DoubleDi/golang_projects/transactions/pkg/config"
	"github.com/DoubleDi/golang_projects/transactions/pkg/database"
)

// Worker is used for periodical transaction reverting
type Worker struct {
	every time.Duration
	repo  database.TransactionRepository
	stop  func()
}

// NewWorker returns a new Worker instance
func NewWorker(cfg *config.Config, repo database.TransactionRepository) *Worker {
	return &Worker{
		every: cfg.CleanEvery,
		repo:  repo,
	}
}

// Run creates an infinite loop which runs at every specified time
func (w *Worker) Run() {
	log.Printf("Starting worker to clean every %v", w.every)
	t := time.NewTicker(w.every)
	w.stop = t.Stop
	for {
		select {
		case <-t.C:
			w.removeOddTransactions()
		}
	}
}

func (w *Worker) removeOddTransactions() {
	log.Println("Removing odd transactions")
	if err := w.repo.RemoveOddTransactions(); err != nil {
		log.Println(err)
	}
	log.Println("Done removing odd transactions")
}

// Close stops the worker
func (w *Worker) Close() {
	if w.stop != nil {
		w.stop()
	}
}
