package starlight

import (
	"context"
	"reflect"
	"testing"

	bolt "github.com/coreos/bbolt"
	"github.com/stellar/go/keypair"

	"github.com/interstellar/starlight/starlight/db"
	"github.com/interstellar/starlight/starlight/fsm"
)

func testMsg() (*fsm.Message, error) {
	kp := keypair.MustParse("SC4OY3XA7VOFPREKUNDW5T7ZO45LKVEFIDTQXK5MT4GJRLVB3H25JLM6")
	sig, err := kp.SignDecorated([]byte("some random input"))
	if err != nil {
		return nil, err
	}
	return &fsm.Message{
		ChannelID: kp.Address(),
		PaymentCompleteMsg: &fsm.PaymentCompleteMsg{
			RoundNumber:      5,
			SenderRatchetSig: sig,
		},
	}, nil
}

func TestEncodeMsg(t *testing.T) {
	g := startTestAgent(t)
	defer g.CloseWait()
	codec := tbCodec{g: g}
	msg, err := testMsg()
	if err != nil {
		t.Fatal(err)
	}
	m := &TbMsg{
		g:         g,
		RemoteURL: "https://starlight.com",
		Msg:       *msg,
	}
	bytes, err := codec.Encode(m)
	if err != nil {
		t.Fatal(err)
	}
	got, err := codec.Decode(bytes)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, m) {
		t.Fatalf("decoded message doesn't match: want %#v, got %#v", m, got)
	}
}

func TestRunMsg(t *testing.T) {
	g := startTestAgent(t)
	defer g.CloseWait()
	err := g.ConfigInit(&Config{
		Username:   "alice",
		Password:   "passw0rd",
		HorizonURL: testHorizonURL,
	})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := testMsg()
	if err != nil {
		t.Fatal(err)
	}
	m := &TbMsg{
		g:         g,
		RemoteURL: "https://starlight.com",
		Msg:       *msg,
	}
	db.Update(m.g.db, func(root *db.Root) error {
		root.Agent().Channels().Put([]byte(msg.ChannelID), &fsm.Channel{ID: msg.ChannelID})
		return nil
	})
	err = m.Run(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	m.RemoteURL = "https://does-not-exist.com"
	err = m.Run(context.TODO())
	if err == nil {
		t.Fatalf("expected http status 404 not found, got %s", err)
	}
}

func TestAddMsgs(t *testing.T) {
	g := startTestAgent(t)
	defer g.CloseWait()
	err := g.ConfigInit(&Config{
		Username:   "alice",
		Password:   "passw0rd",
		HorizonURL: testHorizonURL,
	})
	if err != nil {
		t.Fatal(err)
	}
	msg, err := testMsg()
	if err != nil {
		t.Fatal(err)
	}
	m := &TbMsg{
		g:         g,
		RemoteURL: "https://starlight.com",
		Msg:       *msg,
	}
	db.Update(m.g.db, func(root *db.Root) error {
		root.Agent().Channels().Put([]byte(msg.ChannelID), &fsm.Channel{ID: msg.ChannelID})
		return nil
	})
	err = g.db.Update(func(tx *bolt.Tx) error {
		return g.addMsgTask(tx, m.RemoteURL, msg)
	})
	if err != nil {
		t.Fatal(err)
	}
}
