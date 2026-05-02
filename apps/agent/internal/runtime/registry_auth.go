package runtime

import (
	"encoding/json"
	"fmt"
	"strings"
)

func ParseRegistryCredentials(raw string) (map[string]RegistryCredential, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}

	parsed := map[string]RegistryCredential{}
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, err
	}

	out := make(map[string]RegistryCredential, len(parsed))
	for host, cred := range parsed {
		key := strings.TrimSpace(strings.ToLower(host))
		if key == "" {
			return nil, fmt.Errorf("registry host is empty")
		}
		cred.Username = strings.TrimSpace(cred.Username)
		cred.Password = strings.TrimSpace(cred.Password)
		cred.Server = strings.TrimSpace(cred.Server)
		if cred.Username == "" || cred.Password == "" {
			return nil, fmt.Errorf("registry credentials incomplete for host %s", key)
		}
		if cred.Server == "" {
			cred.Server = key
		}
		out[key] = cred
	}

	return out, nil
}
