package starlight

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"

	bolt "github.com/coreos/bbolt"
	"github.com/stellar/go/clients/horizon"
	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/starlight/db"
	"github.com/interstellar/starlight/starlight/fsm"
	"github.com/interstellar/starlight/starlight/internal/update"
	"github.com/interstellar/starlight/starlight/taskbasket"
	"github.com/interstellar/starlight/worizon"
	"github.com/interstellar/starlight/worizon/xlm"
)

type tbCodec struct {
	g *Agent
}

type encodedTask struct {
	*TbTx  `json:",omitempty"`
	*TbMsg `json:",omitempty"`
}

// Encode implements taskbasket.Codec.Encode.
func (c tbCodec) Encode(t taskbasket.Task) ([]byte, error) {
	var et encodedTask

	switch t := t.(type) {
	case *TbTx:
		et.TbTx = t
	case *TbMsg:
		et.TbMsg = t
	default:
		return nil, fmt.Errorf("unknown task type %T", t)
	}

	return json.Marshal(et)
}

// Decode implements taskbasket.Codec.Decode.
func (c tbCodec) Decode(b []byte) (taskbasket.Task, error) {
	var et encodedTask
	err := json.Unmarshal(b, &et)
	if err != nil {
		return nil, err
	}
	switch {
	case et.TbTx != nil:
		et.TbTx.g = c.g
		return et.TbTx, nil
	case et.TbMsg != nil:
		et.TbMsg.g = c.g
		return et.TbMsg, nil
	}

	return nil, errors.New("empty task")
}

const walletBucket = "wallet"

// TbTx is a taskbasket transaction-submitting task.
type TbTx struct {
	g      *Agent
	ChanID string // Starlight channel ID, or "wallet" for wallet txs
	E      xdr.TransactionEnvelope
}

// Run implements taskbasket.Task.Run.
func (t *TbTx) Run(ctx context.Context) error {
	isWalletTx := t.ChanID == walletBucket

	if !isWalletTx {
		exists, err := channelExists(t.g.db, t.ChanID)
		if err != nil {
			return err
		}
		if !exists {
			return nil
		}
	}

	txstr, err := xdr.MarshalBase64(t.E)
	if err != nil {
		return err
	}

	succ, submitErr := t.g.wclient.SubmitTx(txstr)
	if submitErr != nil {
		log.Printf("SubmitTx error (channel %s): %s\ntx: %s", string(t.ChanID), submitErr, txstr)

		var (
			resultStr string
			err       error
			tr        xdr.TransactionResult
		)

		if herr, ok := submitErr.(*horizon.Error); ok {
			resultStr, err = herr.ResultString()
			if err != nil {
				log.Printf("extracting result string from horizon.Error: %s", err)
				resultStr = ""
			}
		}
		if resultStr == "" {
			resultStr = succ.Result
			if resultStr == "" {
				log.Print("cannot locate result string from failed SubmitTx call")
				return err // will retry
			}
		}

		log.Printf("result string: %s", resultStr)

		err = xdr.SafeUnmarshalBase64(resultStr, &tr)
		if err != nil {
			log.Printf("unmarshaling TransactionResult: %s", err)
			return err // will retry
		}

		if !isRetriableSubmitErr(t.g.wclient, &t.E.Tx, &tr, submitErr) {
			if isWalletTx {
				err = db.Update(t.g.db, func(root *db.Root) error {
					walletAddr := root.Agent().PrimaryAcct().Address()
					isWalletSrcTx := t.E.Tx.SourceAccount.Address() == walletAddr

					var delta int64
					for _, op := range t.E.Tx.Operations {
						if op.SourceAccount == nil && !isWalletSrcTx {
							continue
						}
						if op.SourceAccount != nil && op.SourceAccount.Address() != walletAddr {
							continue
						}
						if op.Body.Type != xdr.OperationTypePayment {
							continue
						}
						if op.Body.PaymentOp.Asset.Type != xdr.AssetTypeAssetTypeNative {
							continue
						}
						delta += int64(op.Body.PaymentOp.Amount)
					}

					if delta != 0 {
						w := root.Agent().Wallet()
						w.Balance += xlm.Amount(delta)
						root.Agent().PutWallet(w)
					}

					t.g.putUpdate(root, &Update{
						Type: update.TxFailureType,
						InputTx: &fsm.Tx{
							Env:    &t.E,
							Result: &tr,
							SeqNum: strconv.FormatUint(uint64(t.E.Tx.SeqNum), 10),
						},
					})
					return nil
				})
				if err != nil {
					log.Printf("unreserving wallet funds after unretriable tx failure: %s", err)
					t.g.mustDeauthenticate()
				}
				return nil // will not retry
			}

			// Failed channel tx (not wallet tx).
			// Hand this transaction to the channel for UpdateFromTx handling
			// (which is failed-transaction-aware).
			ftx := &fsm.Tx{
				Env:    &t.E,
				Result: &tr,
			}
			return t.g.updateChannel(t.ChanID, updateFromTxCaller(ftx))
			// TODO(bobg): add a tx_failure Update for the UI to consume.
		}
	}
	return submitErr
}

// TbMsg is a taskbasket message-sending task.
type TbMsg struct {
	g         *Agent
	RemoteURL string
	Msg       fsm.Message
}

// Run implements taskbasket.Task.
func (m *TbMsg) Run(ctx context.Context) error {
	j, err := json.Marshal(m.Msg)
	if err != nil {
		// If m cannot be marshaled, the implementation is broken.
		return err
	}
	// Check if channel has closed
	exists, err := channelExists(m.g.db, m.Msg.ChannelID)
	if err != nil {
		log.Printf("channelExists error: %s", err)
		return err
	}
	if !exists {
		log.Printf("channel doesn't exist")
		return nil
	}
	url := strings.TrimRight(m.RemoteURL, "/") + "/starlight/message"
	err = post(&m.g.httpclient, url, bytes.NewReader(j))
	if err != nil {
		log.Printf("error %s sending message to %s", err, url)
	}
	return err
}

func post(client *http.Client, url string, body io.Reader) error {
	resp, err := client.Post(url, "application/json", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	r, ok := parse(resp.Body)
	if ok && !r.Retriable {
		return nil
	}

	if resp.StatusCode/100 != 2 {
		return errors.New("bad status " + resp.Status)
	}

	return nil
}

func channelExists(boltDB *bolt.DB, chanID string) (bool, error) {
	var exists bool

	err := db.View(boltDB, func(root *db.Root) error {
		chans := root.Agent().Channels()
		c := chans.Get([]byte(chanID))
		exists = len(c.ID) > 0
		return nil
	})
	return exists, err
}

func isRetriableSubmitErr(wclient *worizon.Client, tx *xdr.Transaction, tr *xdr.TransactionResult, err error) bool {
	if _, ok := err.(*horizon.Error); !ok {
		// TODO(bobg): are there any non-horizon errors that are
		// non-retriable?
		log.Printf("error %s not a horizon error", err)
		return true
	}

	log.Printf("tx failed with result code %s", tr.Result.Code)

	switch tr.Result.Code {
	case xdr.TransactionResultCodeTxTooLate:
		return false

	case xdr.TransactionResultCodeTxInsufficientBalance, xdr.TransactionResultCodeTxNoAccount:
		// Technically retriable if someone went in and fixed the
		// underlying issue (added funds, created the missing account).
		return false

	case xdr.TransactionResultCodeTxInsufficientFee:
		return false

	case xdr.TransactionResultCodeTxMissingOperation:
		return false

	case xdr.TransactionResultCodeTxBadSeq:
		// Don't know if seqnum is too low or too high. One case is
		// retriable, the other's not.
		acctSeqNum, err := wclient.SequenceForAccount(tx.SourceAccount.Address())
		if err != nil {
			return true // whatevs
		}
		return acctSeqNum < tx.SeqNum

	case xdr.TransactionResultCodeTxFailed:
		if tr.Result.Results == nil {
			return false
		}

		for _, opRes := range *tr.Result.Results {
			if !isRetriableOpResult(opRes) {
				return false
			}
		}

		return true
	}

	return true
}

// TODO(bobg): Which operations are really retriable and which are not?
// Retriable under what assumptions?
// (E.g. if a payment fails because the destination account doesn't exist,
// is that retriable because someone might create the account,
// or unretriable because it would require creating the account?)
// Right now it's all just an educated guess.
// Note: https://www.stellar.org/developers/horizon/reference/errors/transaction-failed.html
// contains this language:
//     In almost every case, this error indicates that the transaction
//     submitted in the initial request will never succeed. There is
//     one exception: a transaction that fails with the tx_bad_seq
//     result code (as expressed in the result_code field of the
//     error) may become valid in the future if the sequence number it
//     used was too high.
func isRetriableOpResult(opRes xdr.OperationResult) bool {
	switch opRes.Code {
	case xdr.OperationResultCodeOpInner:
	case xdr.OperationResultCodeOpBadAuth:
		return false
	case xdr.OperationResultCodeOpNoAccount:
		return false
	case xdr.OperationResultCodeOpNotSupported:
		return false
	}

	if opRes.Tr == nil {
		return true
	}

	switch opRes.Tr.Type {
	case xdr.OperationTypeCreateAccount:
		if res := opRes.Tr.CreateAccountResult; res != nil {
			switch res.Code {
			case xdr.CreateAccountResultCodeCreateAccountSuccess:
			case xdr.CreateAccountResultCodeCreateAccountMalformed:
				return false
			case xdr.CreateAccountResultCodeCreateAccountUnderfunded:
			case xdr.CreateAccountResultCodeCreateAccountLowReserve:
				return false
			case xdr.CreateAccountResultCodeCreateAccountAlreadyExist:
				return false
			}
		}

	case xdr.OperationTypePayment:
		if res := opRes.Tr.PaymentResult; res != nil {
			switch res.Code {
			case xdr.PaymentResultCodePaymentSuccess:
			case xdr.PaymentResultCodePaymentMalformed:
				return false
			case xdr.PaymentResultCodePaymentUnderfunded:
			case xdr.PaymentResultCodePaymentSrcNoTrust:
			case xdr.PaymentResultCodePaymentSrcNotAuthorized:
			case xdr.PaymentResultCodePaymentNoDestination:
			case xdr.PaymentResultCodePaymentNoTrust:
			case xdr.PaymentResultCodePaymentNotAuthorized:
			case xdr.PaymentResultCodePaymentLineFull:
			case xdr.PaymentResultCodePaymentNoIssuer:
			}
		}

	case xdr.OperationTypePathPayment:
		if res := opRes.Tr.PathPaymentResult; res != nil {
			switch res.Code {
			case xdr.PathPaymentResultCodePathPaymentSuccess:
			case xdr.PathPaymentResultCodePathPaymentMalformed:
				return false
			case xdr.PathPaymentResultCodePathPaymentUnderfunded:
			case xdr.PathPaymentResultCodePathPaymentSrcNoTrust:
				return false
			case xdr.PathPaymentResultCodePathPaymentSrcNotAuthorized:
			case xdr.PathPaymentResultCodePathPaymentNoDestination:
			case xdr.PathPaymentResultCodePathPaymentNoTrust:
			case xdr.PathPaymentResultCodePathPaymentNotAuthorized:
			case xdr.PathPaymentResultCodePathPaymentLineFull:
			case xdr.PathPaymentResultCodePathPaymentNoIssuer:
				return false
			case xdr.PathPaymentResultCodePathPaymentTooFewOffers:
			case xdr.PathPaymentResultCodePathPaymentOfferCrossSelf:
			case xdr.PathPaymentResultCodePathPaymentOverSendmax:
			}
		}

	case xdr.OperationTypeManageOffer:
		if res := opRes.Tr.ManageOfferResult; res != nil {
			switch res.Code {
			case xdr.ManageOfferResultCodeManageOfferSuccess:
			case xdr.ManageOfferResultCodeManageOfferMalformed:
				return false
			case xdr.ManageOfferResultCodeManageOfferSellNoTrust:
				return false
			case xdr.ManageOfferResultCodeManageOfferBuyNoTrust:
				return false
			case xdr.ManageOfferResultCodeManageOfferSellNotAuthorized:
			case xdr.ManageOfferResultCodeManageOfferBuyNotAuthorized:
			case xdr.ManageOfferResultCodeManageOfferLineFull:
			case xdr.ManageOfferResultCodeManageOfferUnderfunded:
			case xdr.ManageOfferResultCodeManageOfferCrossSelf:
			case xdr.ManageOfferResultCodeManageOfferSellNoIssuer:
				return false
			case xdr.ManageOfferResultCodeManageOfferBuyNoIssuer:
				return false
			case xdr.ManageOfferResultCodeManageOfferNotFound:
				return false
			case xdr.ManageOfferResultCodeManageOfferLowReserve:
			}
		}

	case xdr.OperationTypeCreatePassiveOffer:
		if res := opRes.Tr.CreatePassiveOfferResult; res != nil {
			switch res.Code {
			case xdr.ManageOfferResultCodeManageOfferSuccess:
			case xdr.ManageOfferResultCodeManageOfferMalformed:
				return false
			case xdr.ManageOfferResultCodeManageOfferSellNoTrust:
				return false
			case xdr.ManageOfferResultCodeManageOfferBuyNoTrust:
				return false
			case xdr.ManageOfferResultCodeManageOfferSellNotAuthorized:
			case xdr.ManageOfferResultCodeManageOfferBuyNotAuthorized:
			case xdr.ManageOfferResultCodeManageOfferLineFull:
			case xdr.ManageOfferResultCodeManageOfferUnderfunded:
			case xdr.ManageOfferResultCodeManageOfferCrossSelf:
			case xdr.ManageOfferResultCodeManageOfferSellNoIssuer:
				return false
			case xdr.ManageOfferResultCodeManageOfferBuyNoIssuer:
				return false
			case xdr.ManageOfferResultCodeManageOfferNotFound:
				return false
			case xdr.ManageOfferResultCodeManageOfferLowReserve:
			}
		}

	case xdr.OperationTypeSetOptions:
		if res := opRes.Tr.SetOptionsResult; res != nil {
			switch res.Code {
			case xdr.SetOptionsResultCodeSetOptionsSuccess:
			case xdr.SetOptionsResultCodeSetOptionsLowReserve:
			case xdr.SetOptionsResultCodeSetOptionsTooManySigners:
				return false
			case xdr.SetOptionsResultCodeSetOptionsBadFlags:
				return false
			case xdr.SetOptionsResultCodeSetOptionsInvalidInflation:
			case xdr.SetOptionsResultCodeSetOptionsCantChange:
				return false
			case xdr.SetOptionsResultCodeSetOptionsUnknownFlag:
				return false
			case xdr.SetOptionsResultCodeSetOptionsThresholdOutOfRange:
				return false
			case xdr.SetOptionsResultCodeSetOptionsBadSigner:
				return false
			case xdr.SetOptionsResultCodeSetOptionsInvalidHomeDomain:
				return false
			}
		}

	case xdr.OperationTypeChangeTrust:
		if res := opRes.Tr.ChangeTrustResult; res != nil {
			switch res.Code {
			case xdr.ChangeTrustResultCodeChangeTrustSuccess:
			case xdr.ChangeTrustResultCodeChangeTrustMalformed:
				return false
			case xdr.ChangeTrustResultCodeChangeTrustNoIssuer:
				return false
			case xdr.ChangeTrustResultCodeChangeTrustInvalidLimit:
				return false
			case xdr.ChangeTrustResultCodeChangeTrustLowReserve:
			case xdr.ChangeTrustResultCodeChangeTrustSelfNotAllowed:
				return false
			}
		}

	case xdr.OperationTypeAllowTrust:
		if res := opRes.Tr.AllowTrustResult; res != nil {
			switch res.Code {
			case xdr.AllowTrustResultCodeAllowTrustSuccess:
			case xdr.AllowTrustResultCodeAllowTrustMalformed:
				return false
			case xdr.AllowTrustResultCodeAllowTrustNoTrustLine:
				return false
			case xdr.AllowTrustResultCodeAllowTrustTrustNotRequired:
				return false
			case xdr.AllowTrustResultCodeAllowTrustCantRevoke:
				return false
			case xdr.AllowTrustResultCodeAllowTrustSelfNotAllowed:
				return false
			}
		}

	case xdr.OperationTypeAccountMerge:
		if res := opRes.Tr.AccountMergeResult; res != nil {
			switch res.Code {
			case xdr.AccountMergeResultCodeAccountMergeSuccess:
			case xdr.AccountMergeResultCodeAccountMergeMalformed:
				return false
			case xdr.AccountMergeResultCodeAccountMergeNoAccount:
				return false
			case xdr.AccountMergeResultCodeAccountMergeImmutableSet:
				return false
			case xdr.AccountMergeResultCodeAccountMergeHasSubEntries:
				return false
			case xdr.AccountMergeResultCodeAccountMergeSeqnumTooFar:
				return false
			case xdr.AccountMergeResultCodeAccountMergeDestFull:
			}
		}

	case xdr.OperationTypeInflation:
		if res := opRes.Tr.InflationResult; res != nil {
			switch res.Code {
			case xdr.InflationResultCodeInflationSuccess:
			case xdr.InflationResultCodeInflationNotTime:
			}
		}

	case xdr.OperationTypeManageData:
		if res := opRes.Tr.ManageDataResult; res != nil {
			switch res.Code {
			case xdr.ManageDataResultCodeManageDataSuccess:
			case xdr.ManageDataResultCodeManageDataNotSupportedYet:
			case xdr.ManageDataResultCodeManageDataNameNotFound:
				return false
			case xdr.ManageDataResultCodeManageDataLowReserve:
			case xdr.ManageDataResultCodeManageDataInvalidName:
				return false
			}
		}

	case xdr.OperationTypeBumpSequence:
		if res := opRes.Tr.BumpSeqResult; res != nil {
			switch res.Code {
			case xdr.BumpSequenceResultCodeBumpSequenceSuccess:
			case xdr.BumpSequenceResultCodeBumpSequenceBadSeq:
				return false
			}
		}
	}

	return true
}
