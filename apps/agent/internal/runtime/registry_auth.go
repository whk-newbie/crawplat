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
		cred.Server = strings.TrimSpace(strings.ToLower(cred.Server))
		if cred.Username == "" || cred.Password == "" {
			return nil, fmt.Errorf("registry credentials incomplete for host %s", key)
		}
		if cred.Server == "" {
			if !isLikelyRegistryHost(key) {
				return nil, fmt.Errorf("registry credential %s requires server when key is not a registry host", key)
			}
			cred.Server = key
		}
		out[key] = cred
	}

	return out, nil
}

func registryHostFromImage(image string) string {
	image = strings.TrimSpace(image)
	if image == "" {
		return "docker.io"
	}

	first := image
	hasSlash := false
	if idx := strings.IndexByte(image, '/'); idx >= 0 {
		hasSlash = true
		first = image[:idx]
	}

	if strings.Contains(first, ".") || first == "localhost" {
		return strings.ToLower(first)
	}
	if hasSlash && strings.Contains(first, ":") {
		return strings.ToLower(first)
	}
	return "docker.io"
}

func isLikelyRegistryHost(value string) bool {
	if value == "localhost" {
		return true
	}
	return strings.Contains(value, ".") || strings.Contains(value, ":")
}
