package cmds

import (
	"fmt"
	"os"

	"github.com/pachyderm/pachyderm/src/client"
	"github.com/pachyderm/pachyderm/src/client/pkg/config"
	"github.com/pachyderm/pachyderm/src/client/transaction"
)

// getActiveTransaction will read the active transaction from the config file
// (if it exists) and return it.  If the config file is uninitialized or the
// active transaction is unset, `nil` will be returned.
func getActiveTransaction() (*transaction.Transaction, error) {
	cfg, err := config.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading Pachyderm config: %v", err)
	}
	if cfg.V1 == nil || cfg.V1.ActiveTransaction == "" {
		return nil, nil
	}
	return &transaction.Transaction{ID: cfg.V1.ActiveTransaction}, nil
}

func requireActiveTransaction() (*transaction.Transaction, error) {
	txn, err := getActiveTransaction()
	if err != nil {
		return nil, err
	} else if txn == nil {
		return nil, fmt.Errorf("no active transaction")
	}
	return txn, nil
}

func setActiveTransaction(txn *transaction.Transaction) error {
	cfg, err := config.Read()
	if err != nil {
		return fmt.Errorf("error reading Pachyderm config: %v", err)
	}
	if cfg.V1 == nil {
		cfg.V1 = &config.ConfigV1{}
	}
	if txn == nil {
		cfg.V1.ActiveTransaction = ""
	} else {
		cfg.V1.ActiveTransaction = txn.ID
	}
	if err := cfg.Write(); err != nil {
		return fmt.Errorf("error writing Pachyderm config: %v", err)
	}
	return nil
}

func WithActiveTransaction(c *client.APIClient, callback func(*client.APIClient) error) error {
	txn, err := getActiveTransaction()
	if err != nil {
		return err
	}
	if txn != nil {
		c = c.WithTransaction(txn)
	}
	err = callback(c)
	if err == nil && txn != nil {
		fmt.Fprintf(os.Stderr, "Added to transaction: %s\n", txn.ID)
	}
	return err
}
