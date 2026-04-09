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
          :items="providerItems"
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
          :items="channelItems"
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
            v-for="f in channelFields"
            :key="f.key"
            v-model="channelData[f.key]"
            :label="f.label"
            variant="outlined"
            density="comfortable"
            class="mb-3"
            :prepend-inner-icon="f.icon"
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

// Step 2: Provider — labels from i18n
const showKey = ref(false)
const selectedProvider = ref('openai')
const providerApiKey = ref('')

const providerKeys = ['openai', 'anthropic', 'google', 'deepseek', 'openrouter', 'ollama']

const providerItems = computed(() =>
  providerKeys.map(key => ({
    key,
    label: t(`setup.providers.${key}.label`),
    keyLabel: t(`setup.providers.${key}.keyLabel`),
  }))
)

const currentProviderKeyLabel = computed(() => {
  const p = providerItems.value.find(p => p.key === selectedProvider.value)
  return p?.keyLabel || t('setup.defaultApiKeyLabel')
})

// Step 3: Channel — labels from i18n
const selectedChannel = ref<string | null>(null)
const channelData = ref<Record<string, string>>({})

const channelDefs: Record<string, { fields: { key: string; icon: string }[] }> = {
  telegram:   { fields: [{ key: 'token', icon: 'mdi-key-variant' }] },
  discord:    { fields: [{ key: 'token', icon: 'mdi-key-variant' }] },
  qq:         { fields: [{ key: 'app_id', icon: 'mdi-identifier' }, { key: 'app_secret', icon: 'mdi-key-variant' }] },
  slack:      { fields: [{ key: 'bot_token', icon: 'mdi-key-variant' }, { key: 'app_token', icon: 'mdi-key-variant' }] },
  feishu:     { fields: [{ key: 'app_id', icon: 'mdi-identifier' }, { key: 'app_secret', icon: 'mdi-key-variant' }] },
  dingtalk:   { fields: [{ key: 'client_id', icon: 'mdi-identifier' }, { key: 'client_secret', icon: 'mdi-key-variant' }] },
  wecom:      { fields: [{ key: 'corp_id', icon: 'mdi-identifier' }, { key: 'agent_id', icon: 'mdi-identifier' }, { key: 'secret', icon: 'mdi-key-variant' }] },
  whatsapp:   { fields: [{ key: 'token', icon: 'mdi-key-variant' }, { key: 'phone_number_id', icon: 'mdi-phone' }] },
}

const channelKeys = Object.keys(channelDefs)

const channelItems = computed(() =>
  channelKeys.map(key => ({
    key,
    label: t(`setup.channels.${key}.label`),
  }))
)

const channelFields = computed(() => {
  const def = channelDefs[selectedChannel.value || '']
  if (!def) return []
  return def.fields.map(f => ({
    ...f,
    label: t(`setup.channels.${selectedChannel.value}.fields.${f.key}`),
  }))
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
