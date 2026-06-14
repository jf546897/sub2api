package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTLSFingerprintProfileService_ResolveTLSProfileNilServiceReturnsBuiltInDefault(t *testing.T) {
	t.Parallel()

	account := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Extra: map[string]any{
			"enable_tls_fingerprint": true,
		},
	}

	profile := (*TLSFingerprintProfileService)(nil).ResolveTLSProfile(account)
	require.NotNil(t, profile)
	require.Equal(t, "Built-in Default (Node.js 24.x)", profile.Name)
}

func TestAccount_TLSFingerprintSupportMatrix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		account   *Account
		supported bool
		enabled   bool
	}{
		{
			name: "OpenAI OAuth 支持 TLS 指纹",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
				Extra: map[string]any{
					"enable_tls_fingerprint": true,
				},
			},
			supported: true,
			enabled:   true,
		},
		{
			name: "OpenAI API Key 不启用客户端 TLS 指纹",
			account: &Account{
				Platform: PlatformOpenAI,
				Type:     AccountTypeAPIKey,
				Extra: map[string]any{
					"enable_tls_fingerprint": true,
				},
			},
			supported: false,
			enabled:   false,
		},
		{
			name: "Anthropic OAuth 继续支持 TLS 指纹",
			account: &Account{
				Platform: PlatformAnthropic,
				Type:     AccountTypeOAuth,
				Extra: map[string]any{
					"enable_tls_fingerprint": true,
				},
			},
			supported: true,
			enabled:   true,
		},
		{
			name:      "空账号不支持",
			account:   nil,
			supported: false,
			enabled:   false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.supported, tt.account.SupportsTLSFingerprint())
			require.Equal(t, tt.enabled, tt.account.IsTLSFingerprintEnabled())
		})
	}
}
