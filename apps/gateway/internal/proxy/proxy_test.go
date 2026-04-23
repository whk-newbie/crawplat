package proxy

import "testing"

func TestResolveServiceURL(t *testing.T) {
	tests := map[string]string{
		"iam-service":        "http://iam-service:8081",
		"project-service":    "http://project-service:8082",
		"spider-service":     "http://spider-service:8083",
		"execution-service":  "http://execution-service:8085",
		"node-service":       "http://node-service:8084",
		"datasource-service": "http://datasource-service:8086",
		"scheduler-service":  "http://scheduler-service:8087",
	}

	for serviceName, want := range tests {
		if got := ResolveServiceURL(serviceName); got != want {
			t.Fatalf("ResolveServiceURL(%q) = %q, want %q", serviceName, got, want)
		}
	}
}
