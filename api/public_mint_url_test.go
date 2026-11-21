package api

import "testing"

func TestPublicMintUrlStripsUserinfo(t *testing.T) {
	got := publicMintUrl("https://user:secret@mint.example/path")
	if got != "https://mint.example/path" {
		t.Fatalf("got %s", got)
	}
	plain := publicMintUrl("https://cashu.mutinynet.com")
	if plain != "https://cashu.mutinynet.com" {
		t.Fatalf("got %s", plain)
	}
}
