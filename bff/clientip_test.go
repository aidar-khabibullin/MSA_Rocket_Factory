package main

import (
	"net/http/httptest"
	"testing"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		forwarded  string
		want       string
	}{
		{"IPv4 с портом", "203.0.113.7:51234", "", "203.0.113.7"},
		{"IPv6 с портом", "[2001:db8::1]:51234", "", "2001:db8::1"},
		{"адрес без порта", "203.0.113.7", "", "203.0.113.7"},
		// Заголовок X-Forwarded-For подделывает любой клиент: доверять ему без доверенного прокси нельзя,
		// иначе ограничение попыток входа обходится подменой заголовка.
		{"X-Forwarded-For от клиента игнорируется", "203.0.113.7:51234", "198.51.100.1", "203.0.113.7"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/api/auth/login", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.forwarded != "" {
				req.Header.Set("X-Forwarded-For", tt.forwarded)
			}

			if got := clientIP(req); got != tt.want {
				t.Fatalf("clientIP = %q, want %q", got, tt.want)
			}
		})
	}
}
