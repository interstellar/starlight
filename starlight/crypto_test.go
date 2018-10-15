package starlight

import (
	"bytes"
	"testing"
)

func TestSealOpen(t *testing.T) {
	plaintext := []byte("Hello world!")
	password := []byte("very secret password")

	box := sealBox(plaintext, password)
	result := openBox(box, password)
	if !bytes.Equal(result, plaintext) {
		t.Errorf("Error opening box: expected %s, got %s", plaintext, result)
	}
	result = openBox(box, []byte("wrong secret"))
	if result != nil {
		t.Errorf("Box accepted incorrect password: expected nil, got %s", result)
	}
}
