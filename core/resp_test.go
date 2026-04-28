package core

import "testing"

func TestReadSimpleString(t *testing.T) {
	data := []byte("+OK\r\n")
	value, n, err := readSimpleString(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "OK" {
		t.Fatalf("expected 'OK', got '%s'", value)
	}
	if n != 5 {
		t.Fatalf("expected 5 bytes consumed, got %d", n)
	}
}

func TestReadError(t *testing.T) {
	data := []byte("-invalid command\r\n")
	value, n, err := readError(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if value != "invalid command" {
		t.Fatalf("expected 'invlaid command', got '%s'", value)
	}
	if n != 18 {
		t.Fatalf("expected 18 bytes consumed, got %d", n)
	}
}
