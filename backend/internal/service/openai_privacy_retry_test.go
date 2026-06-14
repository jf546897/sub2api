//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/imroc/req/v3"
	"github.com/stretchr/testify/require"
)

func TestShouldSkipOpenAIPrivacyEnsure_BlankAndInvalidModeDoNotSkip(t *testing.T) {
	require.False(t, shouldSkipOpenAIPrivacyEnsure(nil))
	require.False(t, shouldSkipOpenAIPrivacyEnsure(map[string]any{"privacy_mode": ""}))
	require.False(t, shouldSkipOpenAIPrivacyEnsure(map[string]any{"privacy_mode": "  "}))
	require.False(t, shouldSkipOpenAIPrivacyEnsure(map[string]any{"privacy_mode": 123}))
	require.False(t, shouldSkipOpenAIPrivacyEnsure(map[string]any{"privacy_mode": PrivacyModeFailed}))
	require.True(t, shouldSkipOpenAIPrivacyEnsure(map[string]any{"privacy_mode": PrivacyModeTrainingOff}))
}

func TestAdminService_EnsureOpenAIPrivacy_RetriesNonSuccessModes(t *testing.T) {
	t.Parallel()

	for _, mode := range []string{PrivacyModeFailed, PrivacyModeCFBlocked} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()

			privacyCalls := 0
			svc := &adminServiceImpl{
				accountRepo: &mockAccountRepoForGemini{},
				privacyClientFactory: func(proxyURL string) (*req.Client, error) {
					privacyCalls++
					return nil, errors.New("factory failed")
				},
			}

			account := &Account{
				ID:       101,
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
				Credentials: map[string]any{
					"access_token": "token-1",
				},
				Extra: map[string]any{
					openAIExtraAutoPrivacyEnsureKey: true,
					"privacy_mode":                  mode,
				},
			}

			got := svc.EnsureOpenAIPrivacy(context.Background(), account)

			require.Equal(t, PrivacyModeFailed, got)
			require.Equal(t, 1, privacyCalls)
		})
	}
}

func TestAdminService_EnsureOpenAIPrivacy_SkipsWhenProxyMissing(t *testing.T) {
	t.Parallel()

	privacyCalls := 0
	svc := &adminServiceImpl{
		accountRepo: &mockAccountRepoForGemini{},
		privacyClientFactory: func(proxyURL string) (*req.Client, error) {
			privacyCalls++
			return nil, errors.New("factory should not be called")
		},
	}
	proxyID := int64(303)
	account := &Account{
		ID:       303,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		ProxyID:  &proxyID,
		Credentials: map[string]any{
			"access_token": "token-3",
		},
		Extra: map[string]any{
			openAIExtraAutoPrivacyEnsureKey: true,
		},
	}

	got := svc.EnsureOpenAIPrivacy(context.Background(), account)

	require.Empty(t, got)
	require.Zero(t, privacyCalls)
}

func TestAdminService_EnsureOpenAIPrivacy_RequiresOptIn(t *testing.T) {
	t.Parallel()

	privacyCalls := 0
	svc := &adminServiceImpl{
		accountRepo: &mockAccountRepoForGemini{},
		privacyClientFactory: func(proxyURL string) (*req.Client, error) {
			privacyCalls++
			return nil, errors.New("factory should not be called")
		},
	}
	account := &Account{
		ID:       304,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "token-4",
		},
	}

	got := svc.EnsureOpenAIPrivacy(context.Background(), account)

	require.Empty(t, got)
	require.Zero(t, privacyCalls)
}

func TestAdminService_ForceOpenAIPrivacy_ManualContextForcesWithoutOptIn(t *testing.T) {
	t.Parallel()

	privacyCalls := 0
	svc := &adminServiceImpl{
		accountRepo: &mockAccountRepoForGemini{},
		privacyClientFactory: func(proxyURL string) (*req.Client, error) {
			privacyCalls++
			return nil, errors.New("factory failed")
		},
	}
	account := &Account{
		ID:       304,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "token-4",
		},
	}

	got := svc.ForceOpenAIPrivacy(context.Background(), account)

	require.Equal(t, PrivacyModeFailed, got)
	require.Equal(t, 1, privacyCalls)
}

func TestAdminService_ForceOpenAIPrivacy_RequestContextStillForces(t *testing.T) {
	t.Parallel()

	privacyCalls := 0
	svc := &adminServiceImpl{
		accountRepo: &mockAccountRepoForGemini{},
		privacyClientFactory: func(proxyURL string) (*req.Client, error) {
			privacyCalls++
			return nil, errors.New("factory failed")
		},
	}
	account := &Account{
		ID:       305,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "token-5",
		},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	got := svc.ForceOpenAIPrivacy(ctx, account)

	require.Equal(t, PrivacyModeFailed, got)
	require.Equal(t, 1, privacyCalls)
}

func TestTokenRefreshService_ensureOpenAIPrivacy_RetriesNonSuccessModes(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		TokenRefresh: config.TokenRefreshConfig{
			MaxRetries:          1,
			RetryBackoffSeconds: 0,
		},
	}

	for _, mode := range []string{PrivacyModeFailed, PrivacyModeCFBlocked} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()

			service := NewTokenRefreshService(&tokenRefreshAccountRepo{}, nil, nil, nil, nil, nil, nil, cfg, nil)
			privacyCalls := 0
			service.SetPrivacyDeps(func(proxyURL string) (*req.Client, error) {
				privacyCalls++
				return nil, errors.New("factory failed")
			}, nil)

			account := &Account{
				ID:       202,
				Platform: PlatformOpenAI,
				Type:     AccountTypeOAuth,
				Credentials: map[string]any{
					"access_token": "token-2",
				},
				Extra: map[string]any{
					openAIExtraAutoPrivacyEnsureKey: true,
					"privacy_mode":                  mode,
				},
			}

			service.ensureOpenAIPrivacy(context.Background(), account)

			require.Equal(t, 1, privacyCalls)
		})
	}
}

func TestTokenRefreshService_ensureOpenAIPrivacy_SkipsWhenProxyMissing(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		TokenRefresh: config.TokenRefreshConfig{
			MaxRetries:          1,
			RetryBackoffSeconds: 0,
		},
	}
	service := NewTokenRefreshService(&tokenRefreshAccountRepo{}, nil, nil, nil, nil, nil, nil, cfg, nil)
	privacyCalls := 0
	service.SetPrivacyDeps(func(proxyURL string) (*req.Client, error) {
		privacyCalls++
		return nil, errors.New("factory should not be called")
	}, nil)

	proxyID := int64(404)
	account := &Account{
		ID:       404,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		ProxyID:  &proxyID,
		Credentials: map[string]any{
			"access_token": "token-4",
		},
		Extra: map[string]any{
			openAIExtraAutoPrivacyEnsureKey: true,
		},
	}

	service.ensureOpenAIPrivacy(context.Background(), account)

	require.Zero(t, privacyCalls)
}

func TestTokenRefreshService_ensureOpenAIPrivacy_RequiresOptIn(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		TokenRefresh: config.TokenRefreshConfig{
			MaxRetries:          1,
			RetryBackoffSeconds: 0,
		},
	}
	service := NewTokenRefreshService(&tokenRefreshAccountRepo{}, nil, nil, nil, nil, nil, nil, cfg, nil)
	privacyCalls := 0
	service.SetPrivacyDeps(func(proxyURL string) (*req.Client, error) {
		privacyCalls++
		return nil, errors.New("factory should not be called")
	}, nil)

	account := &Account{
		ID:       405,
		Platform: PlatformOpenAI,
		Type:     AccountTypeOAuth,
		Credentials: map[string]any{
			"access_token": "token-5",
		},
	}

	service.ensureOpenAIPrivacy(context.Background(), account)

	require.Zero(t, privacyCalls)
}
