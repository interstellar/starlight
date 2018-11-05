package fsm

import (
	"bytes"
	"crypto/rand"
	"reflect"
	"testing"
	"time"

	b "github.com/stellar/go/build"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/network"
	"github.com/stellar/go/xdr"

	"github.com/interstellar/starlight/starlight/key"
	"github.com/interstellar/starlight/worizon"
	"github.com/interstellar/starlight/worizon/xlm"
)

var seed = "SA26PHIKZM6CXDGR472SSGUQQRYXM6S437ZNHZGRM6QA4FOPLLLFRGDX"

// Corresponding public key: GDYOVXQVXTUNAWYJBRQTKTWOBEMEYR2YBZYWI4CTGV7ICIQDPS7LUI5B
var hostSeed = "SBHRLPZCQARHBNYMYQDRI6VWRSE7V6HEYMPVWISYI76BQ3I552FAOA4C"

// Corresponding public key: GAGGENNSLRD7XEAUV2UAI7GFPZ7IBS3UWFNLBICTYBSPV65YJB2N5VR7
var guestSeed = "SAZLNS7Z6LHHJODTFCYMWKLNGTZGZYU336OXJIUFI2GM4R5IRTAP53AF"

func createTestChannel() (*Channel, error) {
	var hostAcct AccountID
	hostAcctKeyPair := key.DeriveAccountPrimary([]byte(hostSeed))
	err := hostAcct.SetAddress(hostAcctKeyPair.Address())
	if err != nil {
		return nil, err
	}
	var guestAcct AccountID
	guestAcctKeyPair := key.DeriveAccountPrimary([]byte(guestSeed))
	err = guestAcct.SetAddress(guestAcctKeyPair.Address())
	if err != nil {
		return nil, err
	}
	var escrowAcct AccountID
	var channelKeyIndex uint32 = 1 // 0 used by DeriveAccountPrimary
	escrowAcctKeyPair := key.DeriveAccount([]byte(hostSeed), channelKeyIndex)
	err = escrowAcct.SetAddress(escrowAcctKeyPair.Address())
	if err != nil {
		return nil, err
	}
	var hostRatchetAcct AccountID
	hostRatchetAcctKeyPair := key.DeriveAccount([]byte(hostSeed), channelKeyIndex+1)
	err = hostRatchetAcct.SetAddress(hostRatchetAcctKeyPair.Address())
	if err != nil {
		return nil, err
	}
	var guestRatchetAcct AccountID
	guestRatchetAcctKeyPair := key.DeriveAccount([]byte(hostSeed), channelKeyIndex+2)
	err = guestRatchetAcct.SetAddress(guestRatchetAcctKeyPair.Address())
	if err != nil {
		return nil, err
	}
	fundingTime := time.Date(2018, 9, 24, 11, 02, 00, 0, time.UTC)
	paymentTime := time.Date(2018, 9, 24, 11, 02, 30, 0, time.UTC)
	return &Channel{
		ID:                 escrowAcct.Address(),
		EscrowAcct:         escrowAcct,
		HostAcct:           hostAcct,
		GuestAcct:          guestAcct,
		GuestRatchetAcct:   guestRatchetAcct,
		HostRatchetAcct:    hostRatchetAcct,
		HostAmount:         2 * xlm.Lumen,
		GuestAmount:        2 * xlm.Lumen,
		PendingAmountSent:  1 * xlm.Lumen,
		BaseSequenceNumber: 0,
		RoundNumber:        1,
		FundingTime:        fundingTime,
		FinalityDelay:      1 * time.Second,
		MaxRoundDuration:   1 * time.Minute,
		Passphrase:         network.TestNetworkPassphrase,
		PendingPaymentTime: paymentTime,
		KeyIndex:           channelKeyIndex,
	}, nil
}

func createTestHost() *WalletAcct {
	return &WalletAcct{
		Balance: 10 * xlm.Lumen,
		Seqnum:  xdr.SequenceNumber(5),
	}
}

func TestBuildAndSignSettlementTxes(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := buildSettleOnlyWithHostTx(ch, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
	tx, err = buildSettleWithGuestTx(ch, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
	tx, err = buildSettleWithHostTx(ch, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildAndSignFundingTx(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	h := createTestHost()
	tx, err := buildFundingTx(ch, h)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBuildAndSignRatchetTxes(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := buildRatchetTx(ch, time.Now(), ch.HostRatchetAcct, ch.HostRatchetAcctSeqNum)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
	tx, err = buildRatchetTx(ch, time.Now(), ch.GuestRatchetAcct, ch.GuestRatchetAcctSeqNum)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHandleRatchetTx(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	builder, err := buildRatchetTx(ch, now, ch.HostRatchetAcct, ch.HostRatchetAcctSeqNum)
	if err != nil {
		t.Fatal(err)
	}
	txenv, err := builder.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
	tx := &worizon.Tx{
		Env: txenv.E,
	}
	u := &Updater{
		C:          ch,
		O:          ono{},
		LedgerTime: now,
		Seed:       []byte(seed),
	}
	ok, err := handleRatchetTx(u, tx, true)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("handleRatchetTx returned not-ok status")
	}
	u.C.Role = Host
	_, err = handleRatchetTx(u, tx, false)
	if err != errRatchetTxFailed {
		t.Fatal("handleRatchetTx succeeded with success==false")
	}
}

func TestHandleFundingTx(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	u := &Updater{
		C: ch,
		O: ono{},
	}
	err = u.transitionTo(AwaitingFunding)
	if err != nil {
		t.Fatal(err)
	}
	h := createTestHost()
	builder, err := buildFundingTx(ch, h)
	if err != nil {
		t.Fatal(err)
	}
	txenv, err := builder.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
	tx := &worizon.Tx{
		Env: txenv.E,
	}
	ok, err := handleFundingTx(u, tx, true)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("handleFundingTx returned not-ok status")
	}
}

func TestHandleSettleWithGuestTx(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	u := &Updater{
		C: ch,
		O: ono{},
	}
	err = u.transitionTo(AwaitingSettlement)
	if err != nil {
		t.Fatal(err)
	}
	builder, err := buildSettleWithGuestTx(ch, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	txenv, err := builder.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
	tx := &worizon.Tx{
		Env: txenv.E,
	}
	ok, err := handleSettleWithGuestTx(u, tx, true)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("handleSettleWithGuestTx returned not-ok status")
	}
}

func TestHandleSettleWithHostTx(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	builder, err := buildSettleWithHostTx(ch, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	txenv, err := builder.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
	tx := &worizon.Tx{
		Env: txenv.E,
	}
	u := &Updater{
		C: ch,
		O: ono{},
	}
	ok, err := handleSettleWithHostTx(u, tx, true)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("handleSettleWithHostTx returned not-ok status")
	}
}

func TestUpdateFailedSetupAccountTx(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	h := createTestHost()
	u := &Updater{
		C:    ch,
		O:    ono{},
		H:    h,
		Seed: []byte(seed),
	}
	err = u.transitionTo(SettingUp)
	if err != nil {
		t.Fatal(err)
	}
	builder, err := buildSetupAccountTx(ch, ch.EscrowAcct, xdr.SequenceNumber(0))
	if err != nil {
		t.Fatal(err)
	}
	txenv, err := builder.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
	tx := &worizon.Tx{
		Env: txenv.E,
		Result: &xdr.TransactionResult{
			Result: xdr.TransactionResultResult{
				Code: xdr.TransactionResultCodeTxInsufficientFee,
			},
		},
	}
	err = u.Tx(tx)
	if err != nil {
		t.Errorf("error updating from failed tx: got %s, want nil", err)
	}
	if h.Balance != 11*xlm.Lumen {
		t.Errorf("error unreserving host balance: got %s, want %s", h.Balance, 11*xlm.Lumen)
	}
}

func TestUpdateFailedFundingTx(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	h := createTestHost()
	u := &Updater{
		C:    ch,
		O:    ono{},
		H:    h,
		Seed: []byte(seed),
	}
	err = u.transitionTo(AwaitingFunding)
	if err != nil {
		t.Fatal(err)
	}
	ch.Role = Host
	original := *h
	builder, err := buildFundingTx(ch, h)
	if err != nil {
		t.Fatal(err)
	}
	txenv, err := builder.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
	tx := &worizon.Tx{
		Env: txenv.E,
		Result: &xdr.TransactionResult{
			Result: xdr.TransactionResultResult{
				Code: xdr.TransactionResultCodeTxTooLate,
			},
		},
	}
	err = u.Tx(tx)
	if err != nil {
		t.Errorf("error updating from failed tx: got %s, want nil", err)
	}
	if h.Balance != original.Balance+u.C.totalFundingTxAmount() {
		t.Errorf("error unreserving host balance: got %s, want %s", h.Balance, original.Balance+u.C.totalFundingTxAmount())
	}
	if h.Seqnum != original.Seqnum+1 {
		t.Errorf("error incrementing host seqnum: got %d, want %d", h.Seqnum, original.Seqnum+1)
	}
	if ch.State != AwaitingCleanup {
		t.Errorf("unexpected state: got %s, want %s", ch.State, AwaitingCleanup)
	}
}

func TestUpdateFailedRatchetTx(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	h := createTestHost()
	u := &Updater{
		C:    ch,
		O:    ono{},
		H:    h,
		Seed: []byte(seed),
	}
	err = u.transitionTo(AwaitingFunding)
	if err != nil {
		t.Fatal(err)
	}
	ch.Role = Host
	builder, err := buildRatchetTx(ch, ch.PendingPaymentTime, ch.GuestRatchetAcct, ch.GuestRatchetAcctSeqNum)
	if err != nil {
		t.Fatal(err)
	}
	txenv, err := builder.Sign(seed)
	if err != nil {
		t.Fatal(err)
	}
	tx := &worizon.Tx{
		Env: txenv.E,
		Result: &xdr.TransactionResult{
			Result: xdr.TransactionResultResult{
				Code: xdr.TransactionResultCodeTxBadSeq,
			},
		},
	}
	err = u.Tx(tx)
	if err != nil {
		t.Fatal(err)
	}
	if ch.State != AwaitingRatchet {
		t.Errorf("unexpected state: got %s, want %s", ch.State, AwaitingRatchet)
	}
}

func TestVerifySig(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	builder, err := buildSettleWithHostTx(ch, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	skp, err := keypair.Parse(seed)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := builder.Hash()
	if err != nil {
		t.Fatal(err)
	}
	signature, err := skp.Sign(hash[:])
	if err != nil {
		t.Fatal(err)
	}
	ds := xdr.DecoratedSignature{
		Hint:      skp.Hint(),
		Signature: xdr.Signature(signature[:]),
	}
	err = verifySig(builder, skp, ds)
	if err != nil {
		t.Fatal(err)
	}
	ds = xdr.DecoratedSignature{
		Hint:      skp.Hint(),
		Signature: xdr.Signature(signature[1:]),
	}
	err = verifySig(builder, skp, ds)
	if err == nil {
		t.Fatal("verifySig returned ok on an incorrect signature")
	}
}

func TestDetachedSig(t *testing.T) {
	const (
		builderB64 = "AAAAAQAAAADbjqGBigkg8w/axRyX0PzKuqR01mxkM5DgTq/NyrtHjQAAASwAAAAAAAAABwAAAAEVUziHHyI44AAAAAAAAAAAAAAAAAAAAAMAAAABAAAAANuOoYGKCSDzD9rFHJfQ/Mq6pHTWbGQzkOBOr83Ku0eNAAAACAAAAADqgCM3bqkPZRbtRwj++XguGdyEmwFdx9/0iaVZ3VyO4gAAAAEAAAAAVRfXZeGM4pFliAK3kSKp6yaF0yVavMye55+K4sHdC4EAAAAIAAAAAOqAIzduqQ9lFu1HCP75eC4Z3ISbAV3H3/SJpVndXI7iAAAAAQAAAAAuth0u7iSA+GciNm51AzNLsf8A9vL7IbPPL/1TKxaVsgAAAAgAAAAA6oAjN26pD2UW7UcI/vl4LhnchJsBXcff9ImlWd1cjuIAAAAAAAAAIVRlc3QgU0RGIE5ldHdvcmsgOyBTZXB0ZW1iZXIgMjAxNQAAAAAAAAAAAABk"
		sigB64     = "MAnpiwAAAECzMIpB53pmJOsnqdKMTCza+TjaJDvoUTtaR3cvvc5IWMsg7LhCIJjtzYUVITsjx/6ueX+ZKJK0z2PmjIAb8CgF"
	)
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	var builder b.TransactionBuilder
	xdr.SafeUnmarshalBase64(builderB64, &builder)
	var wantDecSig xdr.DecoratedSignature
	xdr.SafeUnmarshalBase64(sigB64, &wantDecSig)
	gotDecSig, err := detachedSig(builder.TX, []byte(seed), ch.Passphrase, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(wantDecSig.Signature, gotDecSig.Signature) {
		t.Fatal("detachedSig returned incorrect signature")
	}
	txhash, _ := network.HashTransaction(builder.TX, ch.Passphrase)
	kp := key.DeriveAccount([]byte(seed), 0)
	err = kp.Verify(txhash[:], gotDecSig.Signature)
	if err != nil {
		t.Fatal(err)
	}
	_, err = detachedSig(builder.TX, nil, ch.Passphrase, 0)
	if err == nil {
		t.Fatal("detachedSig returned ok on nil seed")
	}
	gotDecSig, err = detachedSig(builder.TX, []byte(seed), ch.Passphrase, 1)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Equal(wantDecSig.Signature, gotDecSig.Signature) {
		t.Fatal("detachedSig returned ok on unequal signatures")
	}
}

func TestTxSig(t *testing.T) {
	const (
		builderB64 = "AAAAAQAAAACjrS92l+b/6lzOv4m7yUQwNKtKXLU4NOr2P3w8GwGnuwAAASwAAAAAAAAABwAAAAEVUenResIImAAAAAAAAAAAAAAAAAAAAAMAAAABAAAAAKOtL3aX5v/qXM6/ibvJRDA0q0pctTg06vY/fDwbAae7AAAACAAAAABSkt8ldfJhMCsmqyFiHRrwnXFSxjouc7JDizYnij7JsgAAAAEAAAAAv5YiUXJGh5wr6NLnnkNGnDVddFG5r2lq3yEgz/vM6e8AAAAIAAAAAFKS3yV18mEwKyarIWIdGvCdcVLGOi5zskOLNieKPsmyAAAAAQAAAAAnRuxwVnBBO91MI+EUweVBccRgg2leZWpYHxXRkxwR3wAAAAgAAAAAUpLfJXXyYTArJqshYh0a8J1xUsY6LnOyQ4s2J4o+ybIAAAAAAAAAIVRlc3QgU0RGIE5ldHdvcmsgOyBTZXB0ZW1iZXIgMjAxNQAAAAAAAAAAAABk"
		sigB64     = "MAnpiwAAAECCGu9rDK1SqhSiUJqxTHt75HjDGcdXf/1wspM+2d7NEVC7KtiqkrhnP1jFL4T+haOfX8bGCE6B9RW/5bSMh1cG"
	)
	var builder b.TransactionBuilder
	xdr.SafeUnmarshalBase64(builderB64, &builder)
	_, err := txSig(&builder, nil, 0)
	if err == nil {
		t.Fatal("txSig returned ok on nil seed")
	}
	env, err := txSig(&builder, []byte(seed), 0)
	if err != nil {
		t.Fatal(err)
	}
	gotSig := env.E.Signatures[0]
	var wantSig xdr.DecoratedSignature
	xdr.SafeUnmarshalBase64(sigB64, &wantSig)
	if !reflect.DeepEqual(gotSig, wantSig) {
		t.Errorf("got %v, want %v", gotSig, wantSig)
	}
}

func TestVerifyTxSig(t *testing.T) {
	ch, err := createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	builder, err := buildSettleOnlyWithHostTx(ch, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	testSeed := make([]byte, 32)
	_, err = rand.Read(testSeed)
	if err != nil {
		t.Fatal(err)
	}
	eb, err := txSig(builder, testSeed, 0)
	if err != nil {
		t.Fatal(err)
	}
	skp := key.DeriveAccount(testSeed, 0)
	err = verifySig(builder, skp, eb.E.Signatures[0])
	if err != nil {
		t.Fatal(err)
	}
	wrongkp := key.DeriveAccount(testSeed, 1)
	err = verifySig(builder, wrongkp, eb.E.Signatures[0])
	if err != keypair.ErrInvalidSignature {
		t.Errorf("expected %s, got %s", keypair.ErrInvalidSignature, err)
	}
	wrongkp = key.DeriveAccount([]byte(seed), 0)
	err = verifySig(builder, wrongkp, eb.E.Signatures[0])
	if err != keypair.ErrInvalidSignature {
		t.Errorf("expected %s, got %s", keypair.ErrInvalidSignature, err)
	}
}
