import { flushPromises, mount } from '@vue/test-utils'
import { defineComponent, nextTick } from 'vue'
import { describe, expect, it, vi } from 'vitest'
import ReAuthAccountModal from '../ReAuthAccountModal.vue'

const {
  applyOAuthCredentialsMock,
  updateAccountMock,
  clearErrorMock,
  exchangeAuthCodeMock,
  buildCredentialsMock,
  buildExtraInfoMock
} = vi.hoisted(() => ({
  applyOAuthCredentialsMock: vi.fn(),
  updateAccountMock: vi.fn(),
  clearErrorMock: vi.fn(),
  exchangeAuthCodeMock: vi.fn(),
  buildCredentialsMock: vi.fn(),
  buildExtraInfoMock: vi.fn()
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    accounts: {
      applyOAuthCredentials: applyOAuthCredentialsMock,
      update: updateAccountMock,
      clearError: clearErrorMock,
      exchangeCode: vi.fn()
    }
  }
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({
    showError: vi.fn(),
    showSuccess: vi.fn()
  })
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({
      t: (key: string) => key
    })
  }
})

vi.mock('@/composables/useOpenAIOAuth', () => ({
  useOpenAIOAuth: () => ({
    sessionId: { value: 'session-1' },
    oauthState: { value: 'state-from-client' },
    authUrl: { value: '' },
    loading: { value: false },
    error: { value: '' },
    generateAuthUrl: vi.fn(),
    exchangeAuthCode: exchangeAuthCodeMock,
    buildCredentials: buildCredentialsMock,
    buildExtraInfo: buildExtraInfoMock,
    resetState: vi.fn()
  })
}))

vi.mock('@/composables/useAccountOAuth', () => ({
  useAccountOAuth: () => ({
    sessionId: { value: '' },
    authUrl: { value: '' },
    loading: { value: false },
    error: { value: '' },
    generateAuthUrl: vi.fn(),
    resetState: vi.fn(),
    buildExtraInfo: vi.fn()
  })
}))

vi.mock('@/composables/useGeminiOAuth', () => ({
  useGeminiOAuth: () => ({
    sessionId: { value: '' },
    state: { value: '' },
    authUrl: { value: '' },
    loading: { value: false },
    error: { value: '' },
    generateAuthUrl: vi.fn(),
    exchangeAuthCode: vi.fn(),
    buildCredentials: vi.fn(),
    resetState: vi.fn()
  })
}))

vi.mock('@/composables/useAntigravityOAuth', () => ({
  useAntigravityOAuth: () => ({
    sessionId: { value: '' },
    state: { value: '' },
    authUrl: { value: '' },
    loading: { value: false },
    error: { value: '' },
    generateAuthUrl: vi.fn(),
    exchangeAuthCode: vi.fn(),
    buildCredentials: vi.fn(),
    resetState: vi.fn()
  })
}))

const BaseDialogStub = defineComponent({
  name: 'BaseDialog',
  props: {
    show: {
      type: Boolean,
      default: false
    }
  },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>'
})

const OAuthAuthorizationFlowStub = defineComponent({
  name: 'OAuthAuthorizationFlow',
  setup(_, { expose }) {
    expose({
      authCode: 'code-1',
      oauthState: 'state-from-input',
      projectId: '',
      sessionKey: '',
      inputMethod: 'manual',
      reset: vi.fn()
    })
    return () => null
  }
})

function buildOpenAIAccount() {
  return {
    id: 42,
    name: 'OpenAI OAuth',
    platform: 'openai',
    type: 'oauth',
    credentials: {
      refresh_token: 'old-refresh-token'
    },
    extra: {
      enable_tls_fingerprint: true,
      tls_fingerprint_profile_id: -1,
      openai_compact_mode: 'force_on',
      openai_oauth_responses_websockets_v2_mode: 'auto'
    },
    proxy_id: null,
    status: 'active'
  } as any
}

function mountModal() {
  return mount(ReAuthAccountModal, {
    props: {
      show: true,
      account: buildOpenAIAccount()
    },
    global: {
      stubs: {
        BaseDialog: BaseDialogStub,
        OAuthAuthorizationFlow: OAuthAuthorizationFlowStub,
        Icon: true
      }
    }
  })
}

describe('admin ReAuthAccountModal', () => {
  it('uses applyOAuthCredentials for OpenAI OAuth reauth so persistent extra settings are merged server-side', async () => {
    const tokenInfo = {
      access_token: 'access-token',
      refresh_token: 'refresh-token',
      email: 'new@example.com'
    }
    const credentials = {
      access_token: 'access-token',
      refresh_token: 'refresh-token'
    }
    const extra = {
      email: 'new@example.com',
      privacy_mode: 'training_disabled'
    }
    const updatedAccount = buildOpenAIAccount()

    exchangeAuthCodeMock.mockResolvedValue(tokenInfo)
    buildCredentialsMock.mockReturnValue(credentials)
    buildExtraInfoMock.mockReturnValue(extra)
    applyOAuthCredentialsMock.mockResolvedValue(updatedAccount)
    exchangeAuthCodeMock.mockClear()
    buildCredentialsMock.mockClear()
    buildExtraInfoMock.mockClear()
    updateAccountMock.mockReset()
    clearErrorMock.mockReset()

    const wrapper = mountModal()
    await nextTick()
    await nextTick()

    const primaryButton = wrapper
      .findAll('button')
      .find((button) => button.text().includes('admin.accounts.oauth.completeAuth'))

    expect(primaryButton).toBeTruthy()
    await primaryButton!.trigger('click')
    await flushPromises()

    expect(exchangeAuthCodeMock).toHaveBeenCalledWith('code-1', 'session-1', 'state-from-input', null)
    expect(applyOAuthCredentialsMock).toHaveBeenCalledTimes(1)
    expect(applyOAuthCredentialsMock).toHaveBeenCalledWith(42, {
      type: 'oauth',
      credentials,
      extra
    })
    expect(updateAccountMock).not.toHaveBeenCalled()
    expect(clearErrorMock).not.toHaveBeenCalled()
  })
})
