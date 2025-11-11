package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProxyService_validateDomainName(t *testing.T) {
	tests := []struct {
		name    string
		domain  string
		wantErr bool
	}{
		{
			name:    "valid domain",
			domain:  "example.com",
			wantErr: false,
		},
		{
			name:    "valid subdomain",
			domain:  "media.example.com",
			wantErr: false,
		},
		{
			name:    "valid localhost",
			domain:  "localhost",
			wantErr: false,
		},
		{
			name:    "valid local domain",
			domain:  "mediacheky.local",
			wantErr: false,
		},
		{
			name:    "empty domain",
			domain:  "",
			wantErr: true,
		},
		{
			name:    "domain with spaces",
			domain:  "exam ple.com",
			wantErr: true,
		},
		{
			name:    "domain starting with hyphen",
			domain:  "-example.com",
			wantErr: true,
		},
		{
			name:    "domain ending with hyphen",
			domain:  "example.com-",
			wantErr: true,
		},
		{
			name:    "domain with double dots",
			domain:  "example..com",
			wantErr: true,
		},
	}

	s := &ProxyService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.validateDomainName(tt.domain)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProxyService_validateSubdomain(t *testing.T) {
	tests := []struct {
		name      string
		subdomain string
		wantErr   bool
	}{
		{
			name:      "valid subdomain",
			subdomain: "radarr",
			wantErr:   false,
		},
		{
			name:      "valid subdomain with number",
			subdomain: "radarr123",
			wantErr:   false,
		},
		{
			name:      "valid subdomain with hyphen",
			subdomain: "tv-shows",
			wantErr:   false,
		},
		{
			name:      "empty subdomain",
			subdomain: "",
			wantErr:   true,
		},
		{
			name:      "subdomain with spaces",
			subdomain: "rad arr",
			wantErr:   true,
		},
		{
			name:      "subdomain starting with hyphen",
			subdomain: "-radarr",
			wantErr:   true,
		},
		{
			name:      "subdomain ending with hyphen",
			subdomain: "radarr-",
			wantErr:   true,
		},
		{
			name:      "subdomain with special chars",
			subdomain: "radarr@123",
			wantErr:   true,
		},
		{
			name:      "subdomain with dots",
			subdomain: "radarr.tv",
			wantErr:   true,
		},
	}

	s := &ProxyService{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := s.validateSubdomain(tt.subdomain)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
