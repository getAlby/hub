package api

import "testing"

func TestMintUrlSanitization(t *testing.T) {
	tests := []struct {
		name          string
		raw           string
		public        string
		authenticated string
	}{
		{
			name:          "plain origin",
			raw:           "https://cashu.mutinynet.com",
			public:        "https://cashu.mutinynet.com",
			authenticated: "https://cashu.mutinynet.com",
		},
		{
			name:          "trailing slash on origin",
			raw:           "https://cashu.mutinynet.com/",
			public:        "https://cashu.mutinynet.com",
			authenticated: "https://cashu.mutinynet.com/",
		},
		{
			name:          "userinfo and secret path",
			raw:           "https://user:secret@mint.example/private-token",
			public:        "https://mint.example",
			authenticated: "https://mint.example/private-token",
		},
		{
			name:          "secret path only",
			raw:           "https://mint.example/private-token",
			public:        "https://mint.example",
			authenticated: "https://mint.example/private-token",
		},
		{
			name:          "path and query",
			raw:           "https://mint.example/v1?api_key=s",
			public:        "https://mint.example",
			authenticated: "https://mint.example/v1?api_key=s",
		},
		{
			name:          "host and port",
			raw:           "http://127.0.0.1:3338/secret",
			public:        "http://127.0.0.1:3338",
			authenticated: "http://127.0.0.1:3338/secret",
		},
		{
			name:          "empty",
			raw:           "",
			public:        "",
			authenticated: "",
		},
		{
			name:          "missing scheme",
			raw:           "mint.example/secret",
			public:        "",
			authenticated: "",
		},
		{
			name:          "invalid",
			raw:           "://",
			public:        "",
			authenticated: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := publicMintUrl(tt.raw); got != tt.public {
				t.Fatalf("publicMintUrl(%q) = %q, want %q", tt.raw, got, tt.public)
			}
			if got := authenticatedMintUrl(tt.raw); got != tt.authenticated {
				t.Fatalf("authenticatedMintUrl(%q) = %q, want %q", tt.raw, got, tt.authenticated)
			}
		})
	}
}

func TestIncludeAuthenticatedCashuMintUrl(t *testing.T) {
	info := &InfoResponse{
		CashuMintUrl:        publicMintUrl("https://mint.example/private-token"),
		privateCashuMintUrl: authenticatedMintUrl("https://mint.example/private-token"),
	}
	if info.CashuMintUrl != "https://mint.example" {
		t.Fatalf("public URL = %q", info.CashuMintUrl)
	}
	info.IncludeAuthenticatedCashuMintUrl()
	if info.CashuMintUrl != "https://mint.example/private-token" {
		t.Fatalf("authenticated URL = %q", info.CashuMintUrl)
	}
}
