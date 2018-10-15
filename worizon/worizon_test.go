package worizon

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/stellar/go/build"
	"github.com/stellar/go/keypair"
	"github.com/stellar/go/xdr"
)

func TestNow(t *testing.T) {
	wor := NewClient(nil, &FakeHorizonClient{}, nil)
	got := wor.Now()
	if got.IsZero() {
		t.Error("got zero time")
	}
}

type FakeTXSigner struct {
	enclave map[string]string
}

func (fst *FakeTXSigner) SignTransaction(ctx context.Context, tx *build.TransactionBuilder, pubKeys ...string) (build.TransactionEnvelopeBuilder, error) {
	privateKeys := make([]string, len(pubKeys))
	for i, pubKey := range pubKeys {
		priv, ok := fst.enclave[pubKey]
		if !ok {
			return build.TransactionEnvelopeBuilder{}, errors.New(fmt.Sprintf("Could not find corresponding private key for %s", pubKey))
		}
		privateKeys[i] = priv
	}
	return tx.Sign(privateKeys...)
}

// PersistSeed will encrypt and call p, a persistence function
func (fst *FakeTXSigner) PersistSeed(ctx context.Context, seed string) (string, error) {
	kp := keypair.MustParse(seed).(*keypair.Full)
	return kp.Address(), nil
}

func TestCreateAccount(t *testing.T) {
	sourceSeed := "SAV76USXIJOBMEQXPANUOQM6F5LIOTLPDIDVRJBFFE2MDJXG24TAPUU7"
	kpA, err := keypair.Parse(sourceSeed)
	if err != nil {
		t.Fatalf("Could not parse sourceSeed as keypair: %s", err)
	}
	kp := kpA.(*keypair.Full)
	sourceAddr := kp.Address()
	hc := &FakeHorizonClient{}
	fts := &FakeTXSigner{enclave: map[string]string{
		sourceAddr: kp.Seed(),
	}}
	wor := NewClient(nil, hc, fts)
	_, _, err = wor.CreateAccounts(context.Background(), sourceAddr, "", "1", false, 1)
	if err != nil {
		t.Fatalf("CreateAccount error: %+v", err)
	}
	if len(hc.transactionEnvelopes) != 1 {
		t.Fatal("No transactions were submitted.")
	}
	txe := &xdr.TransactionEnvelope{}
	err = xdr.SafeUnmarshalBase64(hc.transactionEnvelopes[0], txe)
	if err != nil {
		t.Errorf("Could not decode transaction envelope from base64 %v", err)
	}
	if len(txe.Signatures) != 1 {
		t.Errorf("Found %d signatures, expected exactly 1", len(txe.Signatures))
	}
}

func TestCreditAccount(t *testing.T) {
	kp, err := keypair.Random()
	if err != nil {
		t.Fatalf("could not make the source keypair: %v", err)
	}
	fts := &FakeTXSigner{enclave: map[string]string{}}
	fts.enclave[kp.Address()] = kp.Seed()

	sourceID := kp.Address()
	kp, err = keypair.Random()
	if err != nil {
		t.Fatalf("could not make the destination keypair: %v", err)
	}
	fts.enclave[kp.Address()] = kp.Seed()
	destPubKey := kp.Address()
	hc := &FakeHorizonClient{}
	wor := NewClient(nil, hc, fts)
	wor.hclient = hc
	_, err = wor.CreditAccount(context.Background(), sourceID, "", false, Payment{Credit: Credit{Amount: "10000"}, DestAddr: destPubKey})
	if err != nil {
		t.Fatalf("CreditAccount error: %+v", err)
	}
	if len(hc.transactionEnvelopes) != 1 {
		t.Fatal("No transactions were submitted.")
	}

	txe := &xdr.TransactionEnvelope{}
	err = xdr.SafeUnmarshalBase64(hc.transactionEnvelopes[0], txe)
	if err != nil {
		t.Errorf("Could not decode transaction envelope from base64 %v", err)
	}
	if len(txe.Signatures) != 1 {
		t.Errorf("Found %d signatures, expected exactly 1", len(txe.Signatures))
	}
}
