package database

// TransactionRepository is a repository fro handling transactions
type TransactionRepository interface {
	RemoveOddTransactions() error
	SaveTransaction(t *Transaction) error
}
