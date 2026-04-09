<template>
  <div class="setup-overlay">
    <div class="setup-card">
      <!-- Step Indicator -->
      <div class="d-flex align-center justify-center ga-1 pa-4 pb-0">
        <div v-for="i in 3" :key="i" class="step-dot" :class="{ active: step >= i }"></div>
        <span class="text-caption text-medium-emphasis ml-2">{{ step }} / 3</span>
      </div>

      <!-- Step 1: Welcome -->
      <div v-if="step === 1" class="pa-6 text-center">
        <div style="font-size: 64px;" class="mb-4">🐈</div>
        <h2 class="text-h5 font-weight-bold mb-2">{{ t('setup.welcome') }}</h2>
        <p class="text-body-2 text-medium-emphasis mb-6" style="max-width: 420px; margin: 0 auto;">
          {{ t('setup.welcomeDesc') }}
        </p>
        <v-alert
          :type="nanobotInfo.available ? 'success' : 'warning'"
          variant="tonal"
          class="mb-4 text-left"
          density="compact"
        >
          {{ nanobotInfo.available ? t('setup.engineReady') : t('setup.engineNotFound') }}
        </v-alert>
        <v-btn color="primary" size="large" rounded="lg" @click="step = 2">
          {{ t('setup.startSetup') }}
          <v-icon end>mdi-arrow-right</v-icon>
        </v-btn>
      </div>

      <!-- Step 2: Provider -->
      <div v-if="step === 2" class="pa-6">
        <h3 class="text-h6 mb-2">{{ t('setup.configureProvider') }}</h3>
        <p class="text-body-2 text-medium-emphasis mb-4">{{ t('setup.providerDesc') }}</p>

        <v-select
          v-model="selectedProvider"
          :items="providerList"
          item-title="label"
          item-value="key"
          :label="t('setup.selectProvider')"
          variant="outlined"
          density="comfortable"
          class="mb-4"
          prepend-inner-icon="mdi-cloud-outline"
          menu-icon="mdi-chevron-down"
        />

        <v-text-field
          v-model="providerApiKey"
          :label="currentProviderKeyLabel"
          :type="showKey ? 'text' : 'password'"
          variant="outlined"
          density="comfortable"
          :append-inner-icon="showKey ? 'mdi-eye-off' : 'mdi-eye'"
          @click:append-inner="showKey = !showKey"
          prepend-inner-icon="mdi-key"
          class="mb-4"
        />

        <div class="d-flex justify-space-between">
          <v-btn variant="text" @click="step = 1">{{ t('setup.previous') }}</v-btn>
          <v-btn color="primary" @click="step = 3" :disabled="!providerApiKey.trim()">
            {{ t('setup.next') }}
            <v-icon end>mdi-arrow-right</v-icon>
          </v-btn>
        </div>
      </div>

      <!-- Step 3: Channel (optional) -->
      <div v-if="step === 3" class="pa-6">
        <h3 class="text-h6 mb-2">{{ t('setup.configureChannel') }}</h3>
        <p class="text-body-2 text-medium-emphasis mb-4">{{ t('setup.channelDesc') }}</p>

        <v-select
          v-model="selectedChannel"
          :items="channelList"
          item-title="label"
          item-value="key"
          :label="t('setup.selectChannel')"
          variant="outlined"
          density="comfortable"
          clearable
          class="mb-4"
          prepend-inner-icon="mdi-message-outline"
          menu-icon="mdi-chevron-down"
        />

        <template v-if="selectedChannel">
          <v-text-field
            v-for="field in channelFields"
            :key="field.key"
            v-model="channelData[field.key]"
            :label="field.label"
            variant="outlined"
            density="comfortable"
            class="mb-3"
            :prepend-inner-icon="field.icon"
          />
        </template>

        <div class="d-flex justify-space-between">
          <v-btn variant="text" @click="step = 2">{{ t('setup.previous') }}</v-btn>
          <v-btn color="primary" @click="saveAndFinish" :loading="saving">
            {{ t('setup.finish') }}
            <v-icon end>mdi-check</v-icon>
          </v-btn>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  GetNanobotInfo,
  SetProviderAPIKey,
  SetChannelField,
} from '../../wailsjs/go/main/App'

const { t } = useI18n()

const emit = defineEmits<{ (e: 'done'): void }>()

const step = ref(1)

// Nanobot info
const nanobotInfo = ref<{ available: boolean; source: string; path: string }>({
  available: false, source: 'none', path: ''
})

async function loadNanobotInfo() {
  try {
    nanobotInfo.value = await GetNanobotInfo() as any
  } catch {}
}

// Step 2: Provider
const showKey = ref(false)
const selectedProvider = ref('openai')
const providerApiKey = ref('')

const providerList = [
  { key: 'openai', label: 'OpenAI', keyLabel: 'OpenAI API Key' },
  { key: 'anthropic', label: 'Anthropic (Claude)', keyLabel: 'Anthropic API Key' },
  { key: 'google', label: 'Google (Gemini)', keyLabel: 'Google AI API Key' },
  { key: 'deepseek', label: 'DeepSeek', keyLabel: 'DeepSeek API Key' },
  { key: 'openrouter', label: 'OpenRouter', keyLabel: 'OpenRouter API Key' },
  { key: 'ollama', label: 'Ollama (Local)', keyLabel: 'Ollama Base URL' },
]

const currentProviderKeyLabel = computed(() => {
  const p = providerList.find(p => p.key === selectedProvider.value)
  return p?.keyLabel || 'API Key'
})

// Step 3: Channel
const selectedChannel = ref<string | null>(null)
const channelData = ref<Record<string, string>>({})

const channelList = [
  { key: 'telegram', label: 'Telegram', fields: [
    { key: 'token', label: 'Bot Token', icon: 'mdi-key-variant' },
  ]},
  { key: 'discord', label: 'Discord', fields: [
    { key: 'token', label: 'Bot Token', icon: 'mdi-key-variant' },
  ]},
  { key: 'qq', label: 'QQ', fields: [
    { key: 'app_id', label: 'App ID', icon: 'mdi-identifier' },
    { key: 'app_secret', label: 'App Secret', icon: 'mdi-key-variant' },
  ]},
  { key: 'slack', label: 'Slack', fields: [
    { key: 'bot_token', label: 'Bot Token (xoxb-...)', icon: 'mdi-key-variant' },
    { key: 'app_token', label: 'App Token (xapp-...)', icon: 'mdi-key-variant' },
  ]},
  { key: 'feishu', label: 'Lark / Feishu', fields: [
    { key: 'app_id', label: 'App ID', icon: 'mdi-identifier' },
    { key: 'app_secret', label: 'App Secret', icon: 'mdi-key-variant' },
  ]},
  { key: 'dingtalk', label: 'DingTalk', fields: [
    { key: 'client_id', label: 'Client ID', icon: 'mdi-identifier' },
    { key: 'client_secret', label: 'Client Secret', icon: 'mdi-key-variant' },
  ]},
  { key: 'wecom', label: 'WeCom', fields: [
    { key: 'corp_id', label: 'Corp ID', icon: 'mdi-identifier' },
    { key: 'agent_id', label: 'Agent ID', icon: 'mdi-identifier' },
    { key: 'secret', label: 'Secret', icon: 'mdi-key-variant' },
  ]},
  { key: 'whatsapp', label: 'WhatsApp', fields: [
    { key: 'token', label: 'Access Token', icon: 'mdi-key-variant' },
    { key: 'phone_number_id', label: 'Phone Number ID', icon: 'mdi-phone' },
  ]},
]

const channelFields = computed(() => {
  const ch = channelList.find(c => c.key === selectedChannel.value)
  return ch?.fields || []
})

// Save
const saving = ref(false)

async function saveAndFinish() {
  saving.value = true
  try {
    if (selectedProvider.value && providerApiKey.value.trim()) {
      await SetProviderAPIKey(selectedProvider.value, providerApiKey.value.trim())
    }

    if (selectedChannel.value) {
      for (const f of channelFields.value) {
        const val = channelData.value[f.key] || ''
        if (val.trim()) {
          await SetChannelField(selectedChannel.value, f.key, val.trim())
        }
      }
      await SetChannelField(selectedChannel.value, 'enabled', 'true')
    }

    emit('done')
  } catch (e) {
    alert(t('setup.saveFailed') + e)
  }
  saving.value = false
}

onMounted(loadNanobotInfo)
</script>

<style scoped>
.setup-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(15, 15, 26, 0.92);
  backdrop-filter: blur(8px);
}
.setup-card {
  width: 520px;
  max-height: 90vh;
  overflow-y: auto;
  background: #161625;
  border: 1px solid #2a2a45;
  border-radius: 16px;
}
.step-dot {
  width: 8px; height: 8px; border-radius: 50%;
  background: #2a2a45; transition: all 0.2s;
}
.step-dot.active {
  background: #6c5ce7;
  box-shadow: 0 0 6px #6c5ce7;
}
</style>
