package accounts

import (
	"context"
	"fmt"
	"reflect"

	"github.com/btcsuite/btcutil"
	"github.com/golang/protobuf/proto"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/channeldb"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/zpay32"
	"gopkg.in/macaroon.v2"
)

func (s *Service) intercept(_ context.Context,
	req *lnrpc.RPCMiddlewareRequest) *lnrpc.RPCMiddlewareResponse {

	mac := &macaroon.Macaroon{}
	err := mac.UnmarshalBinary(req.RawMacaroon)
	if err != nil {
		return rpcErrString(req, "error parsing macaroon: %v", err)
	}

	acctID, err := accountFromMacaroon(mac)
	if err != nil {
		return rpcErrString(
			req, "error parsing account from macaroon: %v", err,
		)
	}

	// No account lock in the macaroon, not our concern!
	if acctID == nil {
		return rpcOk(req)
	}

	acct, err := s.GetAccount(*acctID)
	if err != nil {
		return rpcErrString(
			req, "error getting account %x: %v", acctID[:], err,
		)
	}

	switch r := req.InterceptType.(type) {
	case *lnrpc.RPCMiddlewareRequest_StreamAuth:
		return rpcErr(req, s.checkCustomMacaroon(acct))

	case *lnrpc.RPCMiddlewareRequest_Request:
		msg, err := parseProto(r.Request.TypeName, r.Request.Serialized)
		if err != nil {
			return rpcErrString(req, "error parsing proto: %v", err)
		}

		return rpcErr(req, s.checkIncomingRequest(
			r.Request.MethodFullUri, msg, acct,
		))

	case *lnrpc.RPCMiddlewareRequest_Response:
		msg, err := parseProto(
			r.Response.TypeName, r.Response.Serialized,
		)
		if err != nil {
			return rpcErrString(req, "error parsing proto: %v", err)
		}

		replacement, err := s.replaceOutgoingResponse(msg, acct)
		if err != nil {
			return rpcErr(req, err)
		}

		// No error occurred but the response should be replaced with
		// the given custom response. Wrap it in the correct RPC
		// response of the interceptor now.
		if replacement != nil {
			return rpcReplacement(req, replacement)
		}

		// No error and no replacement, just return an empty response of
		// the correct type.
		return rpcOk(req)

	default:
		return rpcErrString(req, "invalid intercept type: %v", r)
	}
}

func (s *Service) checkCustomMacaroon(acct *OffChainBalanceAccount) error {
	log.Debugf("Account auth intercepted, ID=%x, balance_sat=%d, "+
		"expired=%v", acct.ID[:], acct.CurrentBalance.ToSatoshis(),
		acct.HasExpired())

	if acct.HasExpired() {
		return fmt.Errorf("account %x has expired", acct.ID[:])
	}

	// All good!
	return nil
}

func (s *Service) checkIncomingRequest(fullURI string, req proto.Message,
	acct *OffChainBalanceAccount) error {

	if err := s.checkCustomMacaroon(acct); err != nil {
		return err
	}

	switch t := req.(type) {
	case *lnrpc.SendRequest:
		return s.checkSend(t.Amt, t.AmtMsat, t.PaymentRequest, acct)

	case *routerrpc.SendPaymentRequest:
		return s.checkSend(t.Amt, t.AmtMsat, t.PaymentRequest, acct)

	case *lnrpc.SendToRouteRequest:
		return s.checkSendToRoute(t.Route, acct)

	case *routerrpc.SendToRouteRequest:
		if fullURI == "/routerrpc.Router.SendToRoute" {
			return fmt.Errorf("send to route v1 is unsupported")
		}

		return s.checkSendToRoute(t.Route, acct)

	case *lnrpc.SendCoinsRequest:
		return fmt.Errorf("on-chain send not allowed")

	case *lnrpc.SendManyRequest:
		return fmt.Errorf("on-chain send not allowed")
	}

	return nil
}

func (s *Service) replaceOutgoingResponse(resp proto.Message,
	acct *OffChainBalanceAccount) (proto.Message, error) {

	switch t := resp.(type) {
	case *lnrpc.AddInvoiceResponse:
		hash, err := lntypes.MakeHash(t.RHash)
		if err != nil {
			return nil, fmt.Errorf("error parsing invoice hash: %v",
				err)
		}

		return nil, s.Store.associateInvoice(acct.ID, hash)

	case *lnrpc.Payment:
		if t.Status != lnrpc.Payment_SUCCEEDED ||
			t.FailureReason != lnrpc.PaymentFailureReason_FAILURE_REASON_NONE {

			return nil, nil
		}

		hash, err := lntypes.MakeHashFromStr(t.PaymentHash)
		if err != nil {
			return nil, fmt.Errorf("error parsing payment hash: %v",
				err)
		}

		return nil, s.Store.chargeAccount(
			acct.ID, hash, lnwire.MilliSatoshi(t.ValueMsat),
		)

	case *lnrpc.HTLCAttempt:
		if t.Status != lnrpc.HTLCAttempt_SUCCEEDED || t.Failure != nil {
			return nil, nil
		}

		return nil, s.chargeHtlc(acct, t.Route, t.Preimage)

	case *routerrpc.PaymentStatus:
		if t.State != routerrpc.PaymentState_SUCCEEDED {
			return nil, nil
		}

		for _, htlc := range t.Htlcs {
			if htlc.Status != lnrpc.HTLCAttempt_SUCCEEDED {
				continue
			}

			err := s.chargeHtlc(acct, htlc.Route, htlc.Preimage)
			if err != nil {
				return nil, err
			}
		}

		return nil, nil

	case *lnrpc.ChannelBalanceResponse:
		balanceSat := acct.CurrentBalance.ToSatoshis()
		t.Balance = int64(balanceSat)
		t.LocalBalance.Msat = uint64(acct.CurrentBalance)
		t.LocalBalance.Sat = uint64(balanceSat)
		t.PendingOpenLocalBalance.Sat = 0
		t.PendingOpenLocalBalance.Msat = 0
		t.RemoteBalance.Sat = 0
		t.RemoteBalance.Msat = 0
		t.PendingOpenRemoteBalance.Sat = 0
		t.PendingOpenRemoteBalance.Msat = 0
		t.UnsettledLocalBalance.Sat = 0
		t.UnsettledLocalBalance.Msat = 0
		t.UnsettledRemoteBalance.Sat = 0
		t.UnsettledRemoteBalance.Msat = 0

		return t, nil

	case *lnrpc.ListPaymentsResponse:
		filteredPayments := make([]*lnrpc.Payment, 0, len(t.Payments))
		for _, payment := range t.Payments {
			hash, err := lntypes.MakeHashFromStr(
				payment.PaymentHash,
			)
			if err != nil {
				return nil, err
			}

			if _, ok := acct.Payments[hash]; ok {
				filteredPayments = append(
					filteredPayments, payment,
				)
			}
		}

		t.Payments = filteredPayments
		return t, nil

	case *lnrpc.ListInvoiceResponse:
		filteredInvoices := make([]*lnrpc.Invoice, 0, len(t.Invoices))
		for _, invoice := range t.Invoices {
			hash, err := lntypes.MakeHash(invoice.RHash)
			if err != nil {
				return nil, err
			}

			if _, ok := acct.Invoices[hash]; ok {
				filteredInvoices = append(
					filteredInvoices, invoice,
				)
			}
		}

		t.Invoices = filteredInvoices
		return t, nil

	case *lnrpc.ListChannelsResponse:
		t.Channels = []*lnrpc.Channel{}

		return t, nil

	case *lnrpc.WalletBalanceResponse:
		t.ConfirmedBalance = 0
		t.TotalBalance = 0
		t.UnconfirmedBalance = 0

		return t, nil

	case *lnrpc.TransactionDetails:
		t.Transactions = make([]*lnrpc.Transaction, 0)

		return t, nil
	}

	return nil, nil
}

func (s *Service) invoiceUpdate(invoice *lndclient.Invoice) error {
	if invoice == nil || invoice.State != channeldb.ContractSettled {
		return nil
	}

	return s.CreditAccount(invoice.Hash, invoice.AmountPaid)
}

func (s *Service) checkSend(amt, amtMsat int64, invoice string,
	acct *OffChainBalanceAccount) error {

	sendAmt := lnwire.NewMSatFromSatoshis(btcutil.Amount(amt))
	if lnwire.MilliSatoshi(amtMsat) > sendAmt {
		sendAmt = lnwire.MilliSatoshi(amtMsat)
	}

	payReq, err := zpay32.Decode(invoice, s.lnd.ChainParams)
	if err != nil {
		return fmt.Errorf("error decoding pay req: %v", err)
	}

	if payReq.MilliSat != nil && *payReq.MilliSat > sendAmt {
		sendAmt = *payReq.MilliSat
	}

	err = s.Store.checkBalance(acct.ID, sendAmt)
	if err != nil {
		return fmt.Errorf("error validating account balance: %v", err)
	}

	return nil
}

func (s *Service) checkSendToRoute(route *lnrpc.Route,
	acct *OffChainBalanceAccount) error {

	if route == nil {
		return fmt.Errorf("invalid route")
	}

	sendAmt := lnwire.NewMSatFromSatoshis(btcutil.Amount(route.TotalAmt))
	if lnwire.MilliSatoshi(route.TotalAmtMsat) > sendAmt {
		sendAmt = lnwire.MilliSatoshi(route.TotalAmtMsat)
	}

	err := s.Store.checkBalance(acct.ID, sendAmt)
	if err != nil {
		return fmt.Errorf("error validating account balance: %v", err)
	}

	return nil
}

func (s *Service) chargeHtlc(acct *OffChainBalanceAccount, route *lnrpc.Route,
	preimageBytes []byte) error {

	if route == nil {
		return fmt.Errorf("invalid route")
	}

	preimage, err := lntypes.MakePreimage(preimageBytes)
	if err != nil {
		return fmt.Errorf("error parsing preimage: %v", err)
	}

	hash := preimage.Hash()
	return s.Store.chargeAccount(
		acct.ID, hash, lnwire.MilliSatoshi(route.TotalAmtMsat),
	)
}

func rpcOk(req *lnrpc.RPCMiddlewareRequest) *lnrpc.RPCMiddlewareResponse {
	return rpcErrString(req, "")
}

func rpcErr(req *lnrpc.RPCMiddlewareRequest,
	err error) *lnrpc.RPCMiddlewareResponse {

	if err != nil {
		return rpcErrString(req, err.Error())
	}

	return rpcErrString(req, "")
}

func rpcErrString(req *lnrpc.RPCMiddlewareRequest,
	format string, args ...interface{}) *lnrpc.RPCMiddlewareResponse {

	feedback := &lnrpc.InterceptFeedback{}
	resp := &lnrpc.RPCMiddlewareResponse{
		RequestId: req.RequestId,
		MiddlewareMessage: &lnrpc.RPCMiddlewareResponse_Feedback{
			Feedback: feedback,
		},
	}

	if format != "" {
		feedback.Error = fmt.Sprintf(format, args...)
	}

	return resp
}

func rpcReplacement(req *lnrpc.RPCMiddlewareRequest,
	replacementResponse proto.Message) *lnrpc.RPCMiddlewareResponse {

	rawResponse, err := proto.Marshal(replacementResponse)
	if err != nil {
		return rpcErr(
			req, fmt.Errorf("cannot marshal proto msg: %v", err),
		)
	}

	feedback := &lnrpc.InterceptFeedback{
		ReplaceResponse:       true,
		ReplacementSerialized: rawResponse,
	}

	return &lnrpc.RPCMiddlewareResponse{
		RequestId: req.RequestId,
		MiddlewareMessage: &lnrpc.RPCMiddlewareResponse_Feedback{
			Feedback: feedback,
		},
	}
}

func parseProto(typeName string, serialized []byte) (proto.Message, error) {
	reflectType := proto.MessageType(typeName)
	msgValue := reflect.New(reflectType.Elem())
	msg := msgValue.Interface().(proto.Message)

	err := proto.Unmarshal(serialized, msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}
