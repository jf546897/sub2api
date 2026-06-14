package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAccountFromService_ShowsTLSFingerprintForOpenAIOAuth(t *testing.T) {
	t.Parallel()

	account := &service.Account{
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeOAuth,
		Extra: map[string]any{
			"enable_tls_fingerprint":     true,
			"tls_fingerprint_profile_id": -1,
		},
	}

	got := AccountFromServiceShallow(account)
	require.NotNil(t, got)
	require.NotNil(t, got.EnableTLSFingerprint)
	require.True(t, *got.EnableTLSFingerprint)
	require.NotNil(t, got.TLSFingerprintProfileID)
	require.Equal(t, int64(-1), *got.TLSFingerprintProfileID)
}

func TestAccountFromService_HidesTLSFingerprintForOpenAIApiKey(t *testing.T) {
	t.Parallel()

	account := &service.Account{
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeAPIKey,
		Extra: map[string]any{
			"enable_tls_fingerprint":     true,
			"tls_fingerprint_profile_id": -1,
		},
	}

	got := AccountFromServiceShallow(account)
	require.NotNil(t, got)
	require.Nil(t, got.EnableTLSFingerprint)
	require.Nil(t, got.TLSFingerprintProfileID)
}
