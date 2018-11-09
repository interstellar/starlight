package fsm

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/stellar/go/xdr"
)

func TestChannelJSON(t *testing.T) {
	ch := &Channel{}
	empty, err := json.Marshal(ch)
	if err != nil {
		t.Fatal(err)
	}
	var emptych Channel
	err = json.Unmarshal(empty, &emptych)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"ID":"GDNY5IMBRIESB4YP3LCRZF6Q7TFLVJDU2ZWGIM4Q4BHK7TOKXNDY35PU","Role":"","State":"","PrevState":"",` +
		`"CounterpartyAddress":"","RemoteURL":"","Passphrase":"Test SDF Network ; September 2015","Cursor":"","BaseSequenceNumber":0,` +
		`"RoundNumber":1,"CounterpartyMsgIndex":0,"MaxRoundDuration":60000000000,"FinalityDelay":1000000000,"ChannelFeerate":0,"HostFeerate":0,"FundingTime":"2018-09-24T11:02:00Z",` +
		`"FundingTimedOut":false,"FundingTxSeqnum":0,"HostAmount":20000000,"GuestAmount":20000000,"TopUpAmount":0,"PendingAmountSent":10000000,` +
		`"PendingAmountReceived":0,"PaymentTime":"0001-01-01T00:00:00Z","PendingPaymentTime":"2018-09-24T11:02:30Z",` +
		`"HostAcct":"GDVIAIZXN2UQ6ZIW5VDQR7XZPAXBTXEETMAV3R676SE2KWO5LSHOEPST","GuestAcct":"GBZQBS5FDR2F3CAIYGFWOGYIZC3QNXVL2HTSLPUVI43PCNYMBOWTIMY6",` +
		`"EscrowAcct":"GDNY5IMBRIESB4YP3LCRZF6Q7TFLVJDU2ZWGIM4Q4BHK7TOKXNDY35PU","HostRatchetAcct":"GAXLMHJO5YSIB6DHEI3G45IDGNF3D7YA63ZPWINTZ4X72UZLC2K3FEPP",` +
		`"GuestRatchetAcct":"GBKRPV3F4GGOFELFRABLPEJCVHVSNBOTEVNLZTE646PYVYWB3UFYCUWJ","KeyIndex":1,"HostRatchetAcctSeqNum":0,"GuestRatchetAcctSeqNum":0,` +
		`"CurrentRatchetTx":{"Tx":{"SourceAccount":{"Type":0,"Ed25519":null},"Fee":0,"SeqNum":0,"TimeBounds":null,"Memo":{"Type":0,"Text":null,"Id":null,` +
		`"Hash":null,"RetHash":null},"Operations":null,"Ext":{"V":0}},"Signatures":null},"CounterpartyLatestSettleWithGuestTx":null,` +
		`"CounterpartyLatestSettleWithHostTx":{"Tx":{"SourceAccount":{"Type":0,"Ed25519":null},"Fee":0,"SeqNum":0,"TimeBounds":null,"Memo":{"Type":0,` +
		`"Text":null,"Id":null,"Hash":null,"RetHash":null},"Operations":null,"Ext":{"V":0}},"Signatures":null},"CurrentSettleWithGuestTx":null,` +
		`"CurrentSettleWithHostTx":{"Tx":{"SourceAccount":{"Type":0,"Ed25519":null},"Fee":0,"SeqNum":0,"TimeBounds":null,"Memo":{"Type":0,"Text":null,"Id":null,` +
		`"Hash":null,"RetHash":null},"Operations":null,"Ext":{"V":0}},"Signatures":null},"CounterpartyCoopCloseSig":{"Hint":[0,0,0,0],"Signature":null}}`
	ch, err = createTestChannel()
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.Marshal(ch)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("channel JSON doesn't match: want %s, got %s", want, string(got))
	}
	var unmarshaledChannel Channel
	err = json.Unmarshal(got, &unmarshaledChannel)
	if err != nil {
		t.Fatal("error unmarshaling channel JSON", err)
	}
	if !reflect.DeepEqual(*ch, unmarshaledChannel) {
		t.Fatalf("unmarshaled channel doesn't match: want %#v, got %#v", *ch, unmarshaledChannel)
	}
	u := &Updater{
		C: ch,
		O: ono{},
	}
	err = u.transitionTo(Open)
	if err != nil {
		t.Fatal(err)
	}
	changedState, err := json.Marshal(ch)
	if err != nil {
		t.Fatal(err)
	}
	if string(changedState) == want {
		t.Fatalf("expected channel with Open state, got %s", changedState)
	}
}

// ono is a no-op Outputter
type ono struct{}

func (o ono) OutputMsg(*Message)               {}
func (o ono) OutputTx(xdr.TransactionEnvelope) {}
func (o ono) SetTimer(time.Time)               {}
