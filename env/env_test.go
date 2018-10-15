package env

import (
	"bytes"
	"encoding/base64"
	"os"
	"reflect"
	"testing"
	"time"
)

func TestInt(t *testing.T) {
	result := Int("nonexistent", 15)

	if result != 15 {
		t.Fatalf("expected result=15, got result=%d", result)
	}

	err := os.Setenv("int-key", "25")
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	result = Int("int-key", 15)

	if result != 25 {
		t.Fatalf("expected result=25, got result=%d", result)
	}
}

func TestBool(t *testing.T) {
	result := Bool("nonexistent", true)

	if result != true {
		t.Fatalf("expected result=true, got result=%t", result)
	}

	err := os.Setenv("bool-key", "true")
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	result = Bool("bool-key", false)

	if result != true {
		t.Fatalf("expected result=true, got result=%t", result)
	}
}

func TestBytes(t *testing.T) {
	fallback := []byte{0xc0, 0x01}
	got := Bytes("nonexistent", fallback)
	if !bytes.Equal(got, got) {
		t.Fatalf("Bytes(\"nonexistent\", %x) = %x, want %x", fallback, got, fallback)
	}

	want := []byte{0xca, 0xfe}
	err := os.Setenv("bytes-key", base64.StdEncoding.EncodeToString(want))
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	got = Bytes("bytes-key", fallback)
	if !bytes.Equal(got, want) {
		t.Fatalf("Bytes(\"bytes-key\", %x) = %x, want %x", fallback, got, want)
	}
}

func TestDuration(t *testing.T) {
	result := Duration("nonexistent", 15*time.Second)

	if result != 15*time.Second {
		t.Fatalf("expected result=15s, got result=%v", result)
	}

	err := os.Setenv("duration-key", "25s")
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	result = Duration("duration-key", 15*time.Second)

	if result != 25*time.Second {
		t.Fatalf("expected result=25s, got result=%v", result)
	}
}

func TestURL(t *testing.T) {
	example := "http://example.com"
	newExample := "http://something-new.com"
	result := URL("nonexistent", example)

	if result.String() != example {
		t.Fatalf("expected result=%s, got result=%v", example, result)
	}

	err := os.Setenv("url-key", newExample)
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	result = URL("url-key", example)

	if result.String() != newExample {
		t.Fatalf("expected result=%v, got result=%v", newExample, result)
	}
}

func TestString(t *testing.T) {
	result := String("nonexistent", "default")

	if result != "default" {
		t.Fatalf("expected result=default, got result=%s", result)
	}

	err := os.Setenv("string-key", "something-new")
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	result = String("string-key", "default")

	if result != "something-new" {
		t.Fatalf("expected result=something-new, got result=%s", result)
	}
}

func TestStringSlice(t *testing.T) {
	result := StringSlice("empty", "hi")

	exp := []string{"hi"}
	if !reflect.DeepEqual(exp, result) {
		t.Fatalf("expected %v, got %v", exp, result)
	}

	err := os.Setenv("string-slice-key", "hello,world")
	if err != nil {
		t.Fatal("unexpected error", err)
	}

	result = StringSlice("string-slice-key", "hi")

	exp = []string{"hello", "world"}
	if !reflect.DeepEqual(exp, result) {
		t.Fatalf("expected %v, got %v", exp, result)
	}
}
