package verify

import "github.com/stellar/go/clients/horizon"
import hProtocol "github.com/stellar/go/protocols/horizon"

func LoadTransaction(t string) (transaction hProtocol.Transaction, err error) {
	transaction, err = horizon.DefaultTestNetClient.LoadTransaction(t)
	if err != nil {
		return _, err
	}

	return transaction, nil
}
