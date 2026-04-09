<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">{{ t('provider.title') }}</h2>
    </div>

    <!-- Agent Defaults (model selection) -->
    <v-card rounded="lg" class="pa-4 mb-4" style="border: 1px solid #6c5ce7; background: rgba(108,92,231,0.05);">
      <div class="d-flex align-center ga-2 mb-3">
        <v-icon size="18" color="primary">mdi-robot-outline</v-icon>
        <span class="text-body-2 font-weight-semibold">{{ t('provider.agentDefaults') }}</span>
      </div>
      <v-row>
        <v-col cols="6">
          <v-select
            :model-value="agentDefaults.provider || 'auto'"
            @update:model-value="updateAgentField('provider', $event)"
            :items="providerSelectItems"
            :label="t('provider.defaultProvider')"
            variant="outlined"
            density="compact"
            hide-details
            prepend-inner-icon="mdi-cloud-outline"
          />
        </v-col>
        <v-col cols="6">
          <v-text-field
            :model-value="agentDefaults.model || ''"
            @update:model-value="updateAgentField('model', $event)"
            :label="t('provider.defaultModel')"
            variant="outlined"
            density="compact"
            hide-details
            placeholder="anthropic/claude-opus-4-5"
            prepend-inner-icon="mdi-brain"
          />
        </v-col>
      </v-row>
    </v-card>

    <!-- Provider Cards -->
    <v-row>
      <v-col v-for="pv in providerList" :key="pv.name" cols="12" sm="6" md="4" lg="3">
        <v-card rounded="lg" class="pa-4" :style="providerCardStyle(pv)" @mouseenter="pv._hover = true" @mouseleave="pv._hover = false">
          <div class="d-flex align-center ga-2 mb-3">
            <v-icon size="18" :color="pv.hasKey ? 'success' : 'grey'">mdi-{{ providerIcon(pv.name) }}</v-icon>
            <span class="text-body-2 font-weight-semibold">{{ providerDisplayName(pv.name) }}</span>
            <v-chip v-if="pv.hasKey" size="x-small" color="success" variant="tonal" class="ml-auto">✓</v-chip>
          </div>
          <v-text-field
            :model-value="pv.apiKey"
            @update:model-value="updateProviderField(pv.name, 'apiKey', $event)"
            :label="t('provider.apiKey')"
            variant="outlined"
            density="compact"
            hide-details
            :type="showKeys[pv.name] ? 'text' : 'password'"
            :placeholder="apiKeyPlaceholder(pv.name)"
            :append-inner-icon="showKeys[pv.name] ? 'mdi-eye-off' : 'mdi-eye'"
            @click:append-inner="showKeys[pv.name] = !showKeys[pv.name]"
            class="mb-2"
          />
          <v-text-field
            :model-value="pv.apiBase"
            @update:model-value="updateProviderField(pv.name, 'apiBase', $event)"
            :label="t('provider.apiBase')"
            variant="outlined"
            density="compact"
            hide-details
            :placeholder="defaultApiBase(pv.name)"
            prepend-inner-icon="mdi-web"
          />
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetProviders, SetProviderField, GetAgentDefaults, SetAgentDefaults } from '../../wailsjs/go/main/App'

const { t } = useI18n()

interface ProviderInfo {
  name: string
  apiKey: string
  apiBase: string
  hasKey: boolean
  _hover: boolean
}

const providers = ref<Record<string, any>>({})
const agentDefaults = ref<Record<string, any>>({})
const showKeys = reactive<Record<string, boolean>>({})

// All known provider names
const allProviders = [
  'openai', 'anthropic', 'google', 'deepseek', 'openrouter',
  'ollama', 'groq', 'zhipu', 'dashscope', 'vllm', 'gemini',
  'moonshot', 'minimax', 'mistral', 'stepfun', 'aihubmix',
  'azure_openai', 'custom',
]

const providerList = computed<ProviderInfo[]>(() => {
  const configured: ProviderInfo[] = []
  const unconfigured: ProviderInfo[] = []
  for (const name of allProviders) {
    const pv = providers.value[name] || {}
    const hasKey = !!(pv.apiKey || pv.api_key)
    const info: ProviderInfo = {
      name,
      apiKey: pv.apiKey || pv.api_key || '',
      apiBase: pv.apiBase || pv.api_base || '',
      hasKey,
      _hover: false,
    }
    if (hasKey || pv.apiBase || pv.api_base) {
      configured.push(info)
    } else {
      unconfigured.push(info)
    }
  }
  return [...configured, ...unconfigured]
})

const providerSelectItems = computed(() => [
  { title: t('provider.autoDetect'), value: 'auto' },
  ...allProviders.filter(n => !['custom', 'azure_openai'].includes(n)).map(n => ({
    title: providerDisplayName(n), value: n,
  })),
])

function providerDisplayName(name: string): string {
  const names: Record<string, string> = {
    openai: 'OpenAI', anthropic: 'Anthropic', google: 'Google',
    deepseek: 'DeepSeek', openrouter: 'OpenRouter', ollama: 'Ollama',
    groq: 'Groq', zhipu: '智谱 AI', dashscope: '通义千问',
    vllm: 'vLLM', gemini: 'Gemini', moonshot: 'Moonshot',
    minimax: 'MiniMax', mistral: 'Mistral', stepfun: '阶跃星辰',
    aihubmix: 'AiHubMix', azure_openai: 'Azure OpenAI', custom: 'Custom',
  }
  return names[name] || name
}

function providerIcon(name: string): string {
  const icons: Record<string, string> = {
    openai: 'alpha-o-circle', anthropic: 'alpha-a-circle', google: 'google',
    deepseek: 'brain', openrouter: 'transfer', ollama: 'llama',
  }
  return icons[name] || 'cloud-outline'
}

function apiKeyPlaceholder(name: string): string {
  const placeholders: Record<string, string> = {
    openai: 'sk-...', anthropic: 'sk-ant-...', google: 'AI...', deepseek: 'sk-...',
    openrouter: 'sk-or-...', ollama: '(not required)',
  }
  return placeholders[name] || ''
}

function defaultApiBase(name: string): string {
  const bases: Record<string, string> = {
    openai: 'https://api.openai.com/v1',
    anthropic: 'https://api.anthropic.com',
    deepseek: 'https://api.deepseek.com',
    openrouter: 'https://openrouter.ai/api/v1',
    ollama: 'http://localhost:11434',
  }
  return bases[name] || ''
}

function providerCardStyle(pv: ProviderInfo) {
  return {
    border: pv._hover ? '1px solid #6c5ce7' : '1px solid #2a2a45',
    transition: 'border-color 0.2s',
  }
}

async function updateProviderField(provider: string, field: string, value: string) {
  try {
    // nanobot config uses camelCase (apiKey, apiBase)
    const configField = field === 'apiKey' ? 'apiKey' : 'apiBase'
    await SetProviderField(provider, configField, value)
    await loadData()
  } catch (e) {
    alert(t('common.error') + ': ' + e)
  }
}

async function updateAgentField(field: string, value: string) {
  try {
    const updated = { ...agentDefaults.value, [field]: value }
    await SetAgentDefaults(JSON.stringify(updated))
    await loadData()
  } catch (e) {
    alert(t('common.error') + ': ' + e)
  }
}

async function loadData() {
  try { providers.value = await GetProviders() } catch {}
  try { agentDefaults.value = await GetAgentDefaults() || {} } catch {}
}

onMounted(loadData)
</script>
