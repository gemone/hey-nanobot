<template>
  <v-overlay :model-value="true" persistent class="d-flex align-center justify-center" style="z-index: 9999;">
    <v-card width="640" rounded="xl" class="setup-card">
      <!-- Stepper -->
      <v-stepper v-model="step" :items="steps" hide-actions bg-color="transparent" flat>
        <!-- Step 1: Welcome -->
        <template v-slot:item.1>
          <v-card-text class="text-center py-8">
            <div style="font-size: 64px;" class="mb-4">🐈</div>
            <h2 class="text-h5 font-weight-bold mb-2">Welcome to Hey Nanobot</h2>
            <p class="text-body-2 text-medium-emphasis mb-6" style="max-width: 420px; margin: 0 auto;">
              你的个人 AI 助手桌面客户端。内置 nanobot 引擎，只需配置 AI 服务即可开始。
            </p>
            <!-- Nanobot status (informational) -->
            <v-alert
              :type="nanobotInfo.available ? 'success' : 'warning'"
              variant="tonal"
              class="mb-4 text-left"
              density="compact"
            >
              <template v-slot:prepend>
                <v-icon :icon="nanobotInfo.available ? 'mdi-check-circle' : 'mdi-information'" />
              </template>
              <span v-if="nanobotInfo.available">
                ✅ nanobot 引擎就绪
                <span class="text-caption text-medium-emphasis ml-2">({{ nanobotInfo.source }})</span>
              </span>
              <span v-else>
                ⚠️ 未检测到 nanobot 引擎，可先完成配置，后续在设置中指定路径
              </span>
            </v-alert>
            <v-btn color="primary" size="large" rounded="lg" @click="step = 2">
              开始配置
              <v-icon end>mdi-arrow-right</v-icon>
            </v-btn>
          </v-card-text>
        </template>

        <!-- Step 2: Provider -->
        <template v-slot:item.2>
          <v-card-text class="py-6">
            <h3 class="text-h6 mb-2">配置 AI Provider</h3>
            <p class="text-body-2 text-medium-emphasis mb-4">选择一个 AI 服务商并填入 API Key（必填）。</p>

            <v-select
              v-model="selectedProvider"
              :items="providerList"
              item-title="label"
              item-value="key"
              label="选择 Provider"
              variant="outlined"
              density="comfortable"
              class="mb-4"
              prepend-inner-icon="mdi-cloud-outline"
            />

            <v-text-field
              v-model="providerApiKey"
              :label="selectedProvider ? providerList.find(p => p.key === selectedProvider)?.keyLabel : 'API Key'"
              :type="showKey ? 'text' : 'password'"
              variant="outlined"
              density="comfortable"
              :append-inner-icon="showKey ? 'mdi-eye-off' : 'mdi-eye'"
              @click:append-inner="showKey = !showKey"
              prepend-inner-icon="mdi-key"
            />

            <div class="d-flex justify-space-between">
              <v-btn variant="text" @click="step = 1">上一步</v-btn>
              <v-btn
                color="primary"
                @click="step = 3"
                :disabled="!providerApiKey.trim()"
              >
                下一步
                <v-icon end>mdi-arrow-right</v-icon>
              </v-btn>
            </div>
          </v-card-text>
        </template>

        <!-- Step 3: Channel (optional) -->
        <template v-slot:item.3>
          <v-card-text class="py-6">
            <h3 class="text-h6 mb-2">配置消息渠道</h3>
            <p class="text-body-2 text-medium-emphasis mb-4">选择一个消息平台，填入 Bot Token。（可选，可跳过）</p>

            <v-select
              v-model="selectedChannel"
              :items="channelList"
              item-title="label"
              item-value="key"
              label="选择 Channel"
              variant="outlined"
              density="comfortable"
              clearable
              class="mb-4"
              prepend-inner-icon="mdi-message-outline"
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
              <v-btn variant="text" @click="step = 2">上一步</v-btn>
              <v-btn color="primary" @click="saveAndFinish" :loading="saving">
                完成配置
                <v-icon end>mdi-check</v-icon>
              </v-btn>
            </div>
          </v-card-text>
        </template>
      </v-stepper>
    </v-card>
  </v-overlay>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import {
  GetNanobotInfo,
  SetProviderAPIKey,
  SetChannelField,
} from '../../wailsjs/go/main/App'

const emit = defineEmits<{ (e: 'done'): void }>()

const step = ref(1)
const steps = ['欢迎', 'AI Provider', '消息渠道']

// Nanobot info (informational)
const nanobotInfo = ref<{ available: boolean; source: string; path: string }>({
  available: false, source: 'none', path: ''
})

async function loadNanobotInfo() {
  try {
    const info = await GetNanobotInfo() as any
    nanobotInfo.value = info
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
  { key: 'ollama', label: 'Ollama (本地)', keyLabel: 'Ollama Base URL' },
]

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
  { key: 'feishu', label: '飞书 (Lark)', fields: [
    { key: 'app_id', label: 'App ID', icon: 'mdi-identifier' },
    { key: 'app_secret', label: 'App Secret', icon: 'mdi-key-variant' },
  ]},
  { key: 'dingtalk', label: '钉钉', fields: [
    { key: 'client_id', label: 'Client ID', icon: 'mdi-identifier' },
    { key: 'client_secret', label: 'Client Secret', icon: 'mdi-key-variant' },
  ]},
  { key: 'wecom', label: '企业微信', fields: [
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
    // Save provider (required)
    if (selectedProvider.value && providerApiKey.value.trim()) {
      await SetProviderAPIKey(selectedProvider.value, providerApiKey.value.trim())
    }

    // Save channel (optional)
    if (selectedChannel.value) {
      const chFields = channelFields.value
      for (const f of chFields) {
        const val = channelData.value[f.key] || ''
        if (val.trim()) {
          await SetChannelField(selectedChannel.value, f.key, val.trim())
        }
      }
      // Enable channel
      await SetChannelField(selectedChannel.value, 'enabled', 'true')
    }

    emit('done')
  } catch (e) {
    alert('配置保存失败: ' + e)
  }
  saving.value = false
}

onMounted(loadNanobotInfo)
</script>

<style scoped>
.setup-card {
  background: #161625 !important;
  border: 1px solid #2a2a45;
}
</style>
