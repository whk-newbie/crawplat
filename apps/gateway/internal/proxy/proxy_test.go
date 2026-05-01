package proxy

import "testing"

func TestResolveServiceURL_DefaultMappings(t *testing.T) {
	tests := map[string]string{
		"iam-service":        "http://iam-service:8081",
		"project-service":    "http://project-service:8082",
		"spider-service":     "http://spider-service:8083",
		"execution-service":  "http://execution-service:8085",
		"node-service":       "http://node-service:8084",
		"datasource-service": "http://datasource-service:8086",
		"scheduler-service":  "http://scheduler-service:8087",
		"monitor-service":    "http://monitor-service:8088",
	}

	for serviceName, want := range tests {
		if got := ResolveServiceURL(serviceName); got != want {
			t.Fatalf("ResolveServiceURL(%q) = %q, want %q", serviceName, got, want)
		}
	}
}

func TestResolveServiceURL_PrefersNonEmptyEnvOverride(t *testing.T) {
	t.Setenv("GATEWAY_UPSTREAM_IAM_SERVICE", "http://iam-override.internal:18081")

	got := ResolveServiceURL("iam-service")
	want := "http://iam-override.internal:18081"
	if got != want {
		t.Fatalf("ResolveServiceURL(%q) = %q, want %q", "iam-service", got, want)
	}
}

func TestResolveServiceURL_EmptyEnvFallsBackToDefault(t *testing.T) {
	t.Setenv("GATEWAY_UPSTREAM_IAM_SERVICE", "")

	got := ResolveServiceURL("iam-service")
	want := "http://iam-service:8081"
	if got != want {
		t.Fatalf("ResolveServiceURL(%q) = %q, want %q", "iam-service", got, want)
	}
}
