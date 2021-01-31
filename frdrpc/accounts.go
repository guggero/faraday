package frdrpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"path"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/lightninglabs/faraday/accounts"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnwire"
)

func (s *RPCServer) startAccountService(mainCtx context.Context,
	lnd lndclient.LndServices) error {

	acctService, err := accounts.NewService(
		path.Dir(s.cfg.MacaroonPath), lnd,
	)
	if err != nil {
		return fmt.Errorf("error starting account service: %v", err)
	}
	s.accountService = acctService

	return s.accountService.Start(mainCtx)
}

func (s *RPCServer) stopAccountService() error {
	return s.accountService.Stop()
}

// CreateAccount adds an entry to the account database. This entry represents
// an amount of satoshis (account balance) that can be spent using off-chain
// transactions (e.g. paying invoices).
//
// Macaroons can be created to be locked to an account. This makes sure that
// the bearer of the macaroon can only spend at most that amount of satoshis
// through the daemon that has issued the macaroon.
//
// Accounts only assert a maximum amount spendable. Having a certain account
// balance does not guarantee that the node has the channel liquidity to
// actually spend that amount.
func (s *RPCServer) CreateAccount(_ context.Context,
	req *CreateAccountRequest) (*CreateAccountResponse, error) {

	log.Debugf("[createaccount]")

	var (
		balanceMsat    lnwire.MilliSatoshi
		expirationDate time.Time
	)

	// If the expiration date was set, parse it as an unix time stamp.
	// Otherwise we leave it nil to indicate the account has no expiration
	// date.
	if req.ExpirationDate >= 0 {
		expirationDate = time.Unix(req.ExpirationDate, 0)
	}

	// Convert from satoshis to millisatoshis for storage.
	balance := btcutil.Amount(req.AccountBalance)
	balanceMsat = lnwire.NewMSatFromSatoshis(balance)

	// Create the actual account in the macaroon account store.
	account, err := s.accountService.NewAccount(
		balanceMsat, expirationDate,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create account: %v", err)
	}

	// Map the response into the proper response type and return it.
	rpcAccount := &Account{
		Id:             hex.EncodeToString(account.ID[:]),
		InitialBalance: uint64(account.InitialBalance.ToSatoshis()),
		CurrentBalance: uint64(account.CurrentBalance.ToSatoshis()),
		LastUpdate:     account.LastUpdate.Unix(),
		ExpirationDate: int64(0),
	}
	if !account.ExpirationDate.IsZero() {
		rpcAccount.ExpirationDate = account.ExpirationDate.Unix()
	}

	return &CreateAccountResponse{
		Account: rpcAccount,
	}, nil
}

// ListAccounts returns all accounts that are currently stored in the account
// database.
func (s *RPCServer) ListAccounts(_ context.Context,
	_ *ListAccountsRequest) (*ListAccountsResponse,
	error) {

	log.Debugf("[listaccounts]")

	// Retrieve all accounts from the macaroon account store.
	accts, err := s.accountService.GetAccounts()
	if err != nil {
		return nil, fmt.Errorf("unable to list accounts: %v", err)
	}

	// Map the response into the proper response type and return it.
	rpcAccounts := make([]*Account, len(accts))
	for i, account := range accts {
		rpcAccounts[i] = &Account{
			Id: hex.EncodeToString(account.ID[:]),
			InitialBalance: uint64(
				account.InitialBalance.ToSatoshis(),
			),
			CurrentBalance: uint64(
				account.CurrentBalance.ToSatoshis(),
			),
			LastUpdate:     account.LastUpdate.Unix(),
			ExpirationDate: int64(0),
		}
		if !account.ExpirationDate.IsZero() {
			rpcAccounts[i].ExpirationDate =
				account.ExpirationDate.Unix()
		}
	}

	return &ListAccountsResponse{
		Accounts: rpcAccounts,
	}, nil
}

// RemoveAccount removes the given account from the account database.
func (s *RPCServer) RemoveAccount(_ context.Context,
	req *RemoveAccountRequest) (*RemoveAccountResponse, error) {

	// Account ID is always a hex string, convert it to byte array.
	var accountID = accounts.AccountID{}
	decoded, err := hex.DecodeString(req.Id)
	if err != nil {
		return nil, err
	}
	copy(accountID[:], decoded)

	// Now remove the account.
	err = s.accountService.RemoveAccount(accountID)
	if err != nil {
		return nil, err
	}

	return &RemoveAccountResponse{}, nil
}
