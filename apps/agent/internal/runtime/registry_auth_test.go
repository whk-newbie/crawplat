package runtime

import "testing"

func TestParseRegistryCredentials(t *testing.T) {
	creds, err := ParseRegistryCredentials(`{"ghcr.io":{"username":"user","password":"pass"}}`)
	if err != nil {
		t.Fatalf("ParseRegistryCredentials returned error: %v", err)
	}
	cred, ok := creds["ghcr.io"]
	if !ok {
		t.Fatalf("expected ghcr.io credentials, got %+v", creds)
	}
	if cred.Server != "ghcr.io" || cred.Username != "user" || cred.Password != "pass" {
		t.Fatalf("unexpected credential: %+v", cred)
	}
}

func TestRegistryHostFromImage(t *testing.T) {
	cases := []struct {
		image string
		host  string
	}{
		{image: "ghcr.io/acme/crawler:latest", host: "ghcr.io"},
		{image: "registry.example.com:5000/acme/crawler:latest", host: "registry.example.com:5000"},
		{image: "library/nginx:latest", host: "docker.io"},
		{image: "nginx:latest", host: "docker.io"},
	}
	for _, tc := range cases {
		if got := registryHostFromImage(tc.image); got != tc.host {
			t.Fatalf("registryHostFromImage(%q)=%q, want %q", tc.image, got, tc.host)
		}
	}
}
