package accounting

import (
	"fmt"
	"strings"

	"github.com/btcsuite/btcutil"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
)

// feeReference returns a special unique reference for the fee paid on a
// transaction. We use the reference of the original entry with :-1 to denote
// that this entry is associated with the original entry.
func feeReference(reference string) string {
	return fmt.Sprintf("%v:-1", reference)
}

// channelOpenNote creates a note for a channel open entry type.
func channelOpenNote(initiator bool, remotePubkey string,
	capacity btcutil.Amount) string {

	if !initiator {
		return fmt.Sprintf("remote peer %v initated channel open "+
			"with capacity: %v sat", remotePubkey,
			capacity)
	}

	return fmt.Sprintf("initiated channel with remote peer: %v "+
		"capacity: %v sats", remotePubkey, capacity)
}

// channelOpenFeeNote creates a note for channel open types.
func channelOpenFeeNote(channelID uint64) string {
	return fmt.Sprintf("fees to open channel: %v", channelID)
}

// channelOpenEntries produces the relevant set of entries for a currently open
// channel.
func channelOpenEntries(channel lndclient.ChannelInfo, tx lndclient.Transaction,
	convert msatToFiat) ([]*HarmonyEntry, error) {

	var (
		amtMsat   = satsToMsat(tx.Amount)
		entryType = EntryTypeLocalChannelOpen
	)

	if !channel.Initiator {
		amtMsat = 0
		entryType = EntryTypeRemoteChannelOpen
	}

	return openEntries(
		tx, convert, amtMsat, channel.Capacity, entryType,
		channel.PubKeyBytes.String(), channel.ChannelID,
		channel.Initiator,
	)
}

// openChannelFromCloseSummary returns entries for a channel that was opened
// and closed within the period we are reporting on. Since the channel has
// already been closed, we need to produce channel opening records from the
// close summary.
func openChannelFromCloseSummary(channel lndclient.ClosedChannel,
	tx lndclient.Transaction, convert msatToFiat) ([]*HarmonyEntry, error) {

	// If the transaction has a negative amount, we can infer that this
	// transaction was a local channel open, because a remote party opening
	// a channel to us does not affect our balance.
	amtMsat := satsToMsat(tx.Amount)
	initiator := amtMsat < 0

	entryType := EntryTypeLocalChannelOpen
	if !initiator {
		entryType = EntryTypeRemoteChannelOpen
	}

	return openEntries(
		tx, convert, amtMsat, channel.Capacity, entryType,
		channel.PubKeyBytes.String(), channel.ChannelID, initiator,
	)
}

// openEntries creates channel open entries from a set of rpc-indifferent
// fields. This is required because we create channel open entries from already
// open channels using lnrpc.Channel and from closed channels using
// lnrpc.ChannelCloseSummary.
func openEntries(tx lndclient.Transaction, convert msatToFiat, amtMsat int64,
	capacity btcutil.Amount, entryType EntryType, remote string,
	channelID uint64, initiator bool) ([]*HarmonyEntry, error) {

	ref := fmt.Sprintf("%v", channelID)
	note := channelOpenNote(initiator, remote, capacity)

	openEntry, err := newHarmonyEntry(
		tx.Timestamp, amtMsat, entryType, tx.TxHash, ref, note,
		true, convert,
	)
	if err != nil {
		return nil, err
	}

	// If we did not initiate opening the channel, we can just return the
	// channel open entry and do not need a fee entry.
	if !initiator {
		return []*HarmonyEntry{openEntry}, nil
	}

	// We also need an entry for the fees we paid for the on chain tx.
	// Transactions record fees in absolute amounts in sats, so we need
	// to convert fees to msat and filp it to a negative value so it
	// records as a debit.
	feeMsat := invertedSatsToMsats(tx.Fee)

	note = channelOpenFeeNote(channelID)
	feeEntry, err := newHarmonyEntry(
		tx.Timestamp, feeMsat, EntryTypeChannelOpenFee,
		tx.TxHash, feeReference(tx.TxHash), note, true, convert,
	)
	if err != nil {
		return nil, err
	}

	return []*HarmonyEntry{openEntry, feeEntry}, nil
}

// channelCloseNote creates a close note for a channel close entry type.
func channelCloseNote(channelID uint64, closeType, initiated string) string {
	return fmt.Sprintf("close channel: %v close type: %v closed by: %v",
		channelID, closeType, initiated)
}

// closedChannelEntries produces the entries associated with a channel close.
// Note that this entry only reflects the balance we were directly paid out
// in the close transaction. It *does not* include any on chain resolutions, so
// it is excluding htlcs that are resolved on chain, and will not reflect our
// balance when we force close (because it is behind a timelock).
func closedChannelEntries(channel lndclient.ClosedChannel,
	tx lndclient.Transaction, convert msatToFiat) ([]*HarmonyEntry, error) {

	amtMsat := satsToMsat(tx.Amount)
	note := channelCloseNote(
		channel.ChannelID, channel.CloseType.String(),
		channel.CloseInitiator.String(),
	)

	closeEntry, err := newHarmonyEntry(
		tx.Timestamp, amtMsat, EntryTypeChannelClose,
		channel.ClosingTxHash, channel.ClosingTxHash, note, true,
		convert,
	)
	if err != nil {
		return nil, err
	}

	// TODO(carla): add channel close fee entry.

	return []*HarmonyEntry{closeEntry}, nil
}

// onChainEntries produces relevant entries for an on chain transaction.
func onChainEntries(tx lndclient.Transaction, isSweep bool,
	convert msatToFiat) ([]*HarmonyEntry, error) {

	var (
		amtMsat   = satsToMsat(tx.Amount)
		entryType EntryType
		feeType   = EntryTypeFee
	)

	// Determine the type of entry we are creating. If this is a sweep, we
	// set our fee as well, otherwise we set type based on the amount of the
	// transaction.
	switch {
	case isSweep:
		entryType = EntryTypeSweep
		feeType = EntryTypeSweepFee

	case amtMsat < 0:
		entryType = EntryTypePayment

	case amtMsat >= 0:
		entryType = EntryTypeReceipt
	}

	txEntry, err := newHarmonyEntry(
		tx.Timestamp, amtMsat, entryType, tx.TxHash, tx.TxHash,
		tx.Label, true, convert,
	)
	if err != nil {
		return nil, err
	}

	// If we did not pay any fees, we can just return a single entry.
	if tx.Fee == 0 {
		return []*HarmonyEntry{txEntry}, nil
	}

	// Total fees are expressed as a positive value in sats, we convert to
	// msat here and make the value negative so that it reflects as a
	// debit.
	feeAmt := invertedSatsToMsats(tx.Fee)

	feeEntry, err := newHarmonyEntry(
		tx.Timestamp, feeAmt, feeType, tx.TxHash,
		feeReference(tx.TxHash), "", true, convert,
	)
	if err != nil {
		return nil, err
	}

	return []*HarmonyEntry{txEntry, feeEntry}, nil
}

// invoiceNote creates an optional note for an invoice if it had a memo, was
// overpaid, or both.
func invoiceNote(memo string, amt, amtPaid lnwire.MilliSatoshi,
	keysend bool) string {

	var notes []string

	if memo != "" {
		notes = append(notes, fmt.Sprintf("memo: %v", memo))
	}

	if amt != amtPaid {
		notes = append(notes, fmt.Sprintf("invoice overpaid "+
			"original amount: %v msat, paid: %v", amt, amtPaid))
	}

	if keysend {
		notes = append(notes, "keysend payment")
	}

	if len(notes) == 0 {
		return ""
	}

	return strings.Join(notes, "/")
}

// invoiceEntry creates an entry for an invoice.
func invoiceEntry(invoice lndclient.Invoice, circularReceipt bool,
	convert msatToFiat) (*HarmonyEntry, error) {

	eventType := EntryTypeReceipt
	if circularReceipt {
		eventType = EntryTypeCircularReceipt
	}

	note := invoiceNote(
		invoice.Memo, invoice.Amount, invoice.AmountPaid,
		invoice.IsKeysend,
	)

	return newHarmonyEntry(
		invoice.SettleDate, int64(invoice.AmountPaid), eventType,
		invoice.Hash.String(), invoice.Preimage.String(), note, false,
		convert,
	)
}

// paymentReference produces a unique reference for a payment. Since payment
// hash is not guaranteed to be unique, we use the payments unique sequence
// number and its hash.
func paymentReference(sequenceNumber uint64, paymentHash lntypes.Hash) string {
	return fmt.Sprintf("%v:%v", sequenceNumber, paymentHash)
}

// paymentNote creates a note for payments from our node.
func paymentNote(preimage lntypes.Preimage) string {
	return fmt.Sprintf("Preimage: %v", preimage)
}

// paymentFeeNote creates a note for a payment fee entry.
func paymentFeeNote(htlcs []*lnrpc.HTLCAttempt) string {
	return fmt.Sprintf("Settled with: %v htlc(s)", len(htlcs))
}

// paymentEntry creates an entry for an off chain payment, including fee entries
// where required.
func paymentEntry(payment settledPayment, paidToSelf bool,
	convert msatToFiat) ([]*HarmonyEntry, error) {

	// It is possible to make a payment to ourselves as part of a circular
	// rebalance which is operationally used to shift funds between
	// channels. For these payment types, we lose balance from fees, but do
	// not change our balance from the actual payment because it is paid
	// back to ourselves.
	var (
		paymentType = EntryTypePayment
		feeType     = EntryTypeFee
	)

	// If we made the payment to ourselves, we set special entry types,
	// since the payment amount did not actually affect our balance.
	if paidToSelf {
		paymentType = EntryTypeCircularPayment
		feeType = EntryTypeCircularPaymentFee
	}

	// Create a note for our payment. Since we have already checked that our
	// payment is settled, we will not have a nil preimage.
	note := paymentNote(*payment.Preimage)
	ref := paymentReference(payment.SequenceNumber, payment.Hash)

	// Payment values are expressed as positive values over rpc, but they
	// decrease our balance so we flip our value to a negative one.
	amt := invertMsat(int64(payment.Amount))

	paymentEntry, err := newHarmonyEntry(
		payment.settleTime, amt, paymentType,
		payment.Hash.String(), ref, note, false, convert,
	)
	if err != nil {
		return nil, err
	}

	// If we paid no fees (possible for payments to our direct peer), then
	// we just return the payment entry.
	if payment.Fee == 0 {
		return []*HarmonyEntry{paymentEntry}, nil
	}

	feeNote := paymentFeeNote(payment.Htlcs)
	feeRef := feeReference(ref)
	feeAmt := invertMsat(int64(payment.Fee))

	feeEntry, err := newHarmonyEntry(
		payment.settleTime, feeAmt, feeType,
		payment.Hash.String(), feeRef, feeNote, false, convert,
	)
	if err != nil {
		return nil, err
	}
	return []*HarmonyEntry{paymentEntry, feeEntry}, nil
}

// forwardTxid provides a best effort txid using incoming and outgoing channel
// ID paired with timestamp in an effort to make txid unique per htlc forwarded.
// This is not used as a reference because we could theoretically have duplicate
// timestamps.
func forwardTxid(forward lndclient.ForwardingEvent) string {
	return fmt.Sprintf("%v:%v:%v", forward.Timestamp, forward.ChannelIn,
		forward.ChannelOut)
}

// forwardNote creates a note that indicates the amuonts that were forwarded in
// and out of our node.
func forwardNote(amtIn, amtOut lnwire.MilliSatoshi) string {
	return fmt.Sprintf("incoming: %v msat outgoing: %v msat", amtIn, amtOut)
}

// forwardingEntry produces a forwarding entry with a zero amount which reflects
// shifting of funds in our channels, and fees entry which reflects the fees we
// earned form the forward.
func forwardingEntry(forward lndclient.ForwardingEvent,
	convert msatToFiat) ([]*HarmonyEntry, error) {

	txid := forwardTxid(forward)
	note := forwardNote(forward.AmountMsatIn, forward.AmountMsatOut)

	fwdEntry, err := newHarmonyEntry(
		forward.Timestamp, 0, EntryTypeForward, txid, "", note,
		false, convert,
	)
	if err != nil {
		return nil, err
	}

	// If we did not earn any fees, return the forwarding entry.
	if forward.FeeMsat == 0 {
		return []*HarmonyEntry{fwdEntry}, nil
	}

	feeEntry, err := newHarmonyEntry(
		forward.Timestamp, int64(forward.FeeMsat),
		EntryTypeForwardFee, txid, "", "", false, convert,
	)
	if err != nil {
		return nil, err
	}

	return []*HarmonyEntry{fwdEntry, feeEntry}, nil
}