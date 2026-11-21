package keys

import "testing"

// NIP-06 test vector from nostr-protocol/nips/06.md.
// mnemonic: leader monkey parrot ring guide accident before fence cannon height naive bean
// private key: 7f7ff03d123792d6ac594bfa67bf6d0c0ab55b6b1fdb6249303fe861f1ccba9a
func TestDeriveNostrKeyFromMnemonicNIP06Vector(t *testing.T) {
	mnemonic := "leader monkey parrot ring guide accident before fence cannon height naive bean"
	want := "7f7ff03d123792d6ac594bfa67bf6d0c0ab55b6b1fdb6249303fe861f1ccba9a"
	got, err := deriveNostrKeyFromMnemonic(mnemonic)
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
