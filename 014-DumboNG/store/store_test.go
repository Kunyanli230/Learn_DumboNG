package store

import "testing"

func TestNustDB(t *testing.T) {
	db := NewDefaultNutsDB(t.TempDir())
	store := NewStore(db)
	key, val := []byte("dcz"), []byte("1234")
	if err := store.Write(key, val); err != nil {
		t.Fatal(err)
	}
	if rVal, err := store.Read(key); err != nil {
		t.Fatal(err)
	} else if string(rVal) != string(val) {
		t.Fatalf("value is not equal\n")
	}
}
