package dht

import (
	"testing"
)

func TestCreateParentChord(t *testing.T) {
	c, err := NewChord(":5000", 1099)
	if err != nil {
		t.Fatalf("Failed to create parent chord: %v\n", err)
	}
	c.Set("Hello", "World")
}

func TestConnectToParent1(t *testing.T) {
	c, err := JoinChord(":5001", ":5000", 1099)
	if err != nil {
		t.Fatalf("Failed to connect to parent chord: %v\n", err)
	}
	//val, err := c.Get("Hello")
	c.Set("Fire", "and flame")
	if err != nil {
		t.Fatalf("Failed to get value from parent chord: %v\n", err)
	}
}

func TestConnectToParent2(t *testing.T) {
	c, err := JoinChord(":5002", ":5000", 1099)
	if err != nil {
		t.Fatalf("Failed to connect to parent chord: %v\n", err)
	}
	c.GetAllKeys()
	c.Set("Super", "Mario")
}
