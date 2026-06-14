package service

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCRSSyncService_RefreshOAuthToken_SkipsOpenAIByDefault(t *testing.T) {
	client := &openaiOAuthClientRefreshStub{}
	svc := &CRSSyncService{
		openaiOAuthService: NewOpenAIOAuthService(nil, client),
	}
	account := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"refresh_token": "refresh-token",
		},
	}

	creds := svc.refreshOAuthToken(context.Background(), account)

	require.Nil(t, creds)
	require.Zero(t, atomic.LoadInt32(&client.refreshCalls))
}

func TestCRSSyncService_RefreshOAuthToken_OpenAIOptInCallsRefresh(t *testing.T) {
	client := &openaiOAuthClientRefreshStub{}
	svc := &CRSSyncService{
		openaiOAuthService: NewOpenAIOAuthService(nil, client),
	}
	account := &Account{
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"refresh_token": "refresh-token",
		},
		Extra: map[string]any{
			openAIExtraRefreshOnCRSSyncKey: true,
		},
	}

	creds := svc.refreshOAuthToken(context.Background(), account)

	require.Nil(t, creds)
	require.Equal(t, int32(1), atomic.LoadInt32(&client.refreshCalls))
}

func TestCRSSyncService_OpenAIOAuthRequiresRefreshToken(t *testing.T) {
	require.False(t, hasCRSOpenAIRefreshToken(nil))
	require.False(t, hasCRSOpenAIRefreshToken(map[string]any{"access_token": "access-token"}))
	require.False(t, hasCRSOpenAIRefreshToken(map[string]any{"refresh_token": "  "}))
	require.True(t, hasCRSOpenAIRefreshToken(map[string]any{"refresh_token": "refresh-token"}))
}

func TestCRSSyncService_MapOrCreateProxyRejectsInvalidProxy(t *testing.T) {
	svc := &CRSSyncService{}
	cached := []Proxy{}

	id, err := svc.mapOrCreateProxy(context.Background(), true, &cached, &crsProxy{
		Protocol: "http",
		Port:     8080,
	}, "bad")

	require.Error(t, err)
	require.Contains(t, err.Error(), "proxy host is required")
	require.Nil(t, id)
}
