<template>
  <div class="page-body">
    <!-- Page Title -->
    <div class="d-flex align-center justify-space-between mb-4">
      <div class="d-flex align-center ga-2">
        <v-icon size="20" color="primary">mdi-key-outline</v-icon>
        <span class="text-body-1 font-weight-bold">{{ t('provider.title') }}</span>
      </div>
    </div>

    <!-- Agent Defaults -->
    <div class="card-base pa-4 mb-4 agent-defaults-card">
      <div class="d-flex align-center ga-2 mb-3">
        <v-icon size="16" color="primary">mdi-robot-outline</v-icon>
        <span class="text-caption font-weight-semibold" style="text-transform: uppercase; letter-spacing: 0.5px;">{{ t('provider.agentDefaults') }}</span>
      </div>
      <v-row dense>
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
            @update:model-value="debouncedAgentField('model', $event)"
            :label="t('provider.defaultModel')"
            variant="outlined"
            density="compact"
            hide-details
            placeholder="anthropic/claude-opus-4-5"
            prepend-inner-icon="mdi-brain"
          />
        </v-col>
      </v-row>
    </div>

    <!-- Configured Providers -->
    <div v-if="configuredList.length" class="mb-3">
      <div class="text-caption font-weight-semibold mb-2" style="text-transform: uppercase; letter-spacing: 0.5px; color: #5a5a78;">
        {{ t('provider.configured') }}
      </div>
      <div class="provider-grid">
        <div v-for="pv in configuredList" :key="pv.name" class="card-base pa-4">
          <div class="d-flex align-center ga-2 mb-3">
            <v-icon size="18" color="success">mdi-{{ providerIcon(pv.name) }}</v-icon>
            <span class="text-body-2 font-weight-semibold">{{ providerDisplayName(pv.name) }}</span>
            <v-chip size="x-small" color="success" variant="tonal" class="ml-auto">✓</v-chip>
          </div>
          <v-text-field
            :model-value="pv.apiKey"
            @update:model-value="debouncedProviderField(pv.name, 'apiKey', $event)"
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
            @update:model-value="debouncedProviderField(pv.name, 'apiBase', $event)"
            :label="t('provider.apiBase')"
            variant="outlined"
            density="compact"
            hide-details
            :placeholder="defaultApiBase(pv.name)"
            prepend-inner-icon="mdi-web"
          />
        </div>
      </div>
    </div>

    <!-- Other Providers (collapsible) -->
    <div v-if="unconfiguredList.length">
      <div class="d-flex align-center cursor-pointer py-2" @click="showAll = !showAll">
        <v-icon size="14" class="mr-1" :class="{ 'rotate-90': showAll }">mdi-chevron-right</v-icon>
        <span class="text-caption font-weight-semibold" style="text-transform: uppercase; letter-spacing: 0.5px; color: #5a5a78;">
          {{ t('provider.allProviders') }} ({{ unconfiguredList.length }})
        </span>
      </div>
      <template v-if="showAll">
        <div class="provider-grid">
          <div v-for="pv in unconfiguredList" :key="pv.name" class="card-base pa-4" style="opacity: 0.85;">
            <div class="d-flex align-center ga-2 mb-3">
              <v-icon size="18" color="grey">mdi-{{ providerIcon(pv.name) }}</v-icon>
              <span class="text-body-2 font-weight-semibold" style="color: #9898b0;">{{ providerDisplayName(pv.name) }}</span>
            </div>
            <v-text-field
              :model-value="pv.apiKey"
              @update:model-value="debouncedProviderField(pv.name, 'apiKey', $event)"
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
              @update:model-value="debouncedProviderField(pv.name, 'apiBase', $event)"
              :label="t('provider.apiBase')"
              variant="outlined"
              density="compact"
              hide-details
              :placeholder="defaultApiBase(pv.name)"
              prepend-inner-icon="mdi-web"
            />
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, reactive, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetProviders, SetProviderField, GetAgentDefaults, SetAgentDefaults } from '../../wailsjs/go/main/App'

const { t } = useI18n()

interface ProviderInfo {
  name: string; apiKey: string; apiBase: string; hasKey: boolean
}

const providers = ref<Record<string, any>>({})
const agentDefaults = ref<Record<string, any>>({})
const showKeys = reactive<Record<string, boolean>>({})
const showAll = ref(false)

const allProviders = [
  'openai', 'anthropic', 'google', 'deepseek', 'openrouter',
  'ollama', 'groq', 'zhipu', 'dashscope', 'vllm', 'gemini',
  'moonshot', 'minimax', 'mistral', 'stepfun', 'aihubmix',
  'azure_openai', 'custom',
]

const configuredList = computed<ProviderInfo[]>(() =>
  allProviders
    .filter(name => { const pv = providers.value[name] || {}; return !!(pv.apiKey || pv.api_key || pv.apiBase || pv.api_base) })
    .map(name => { const pv = providers.value[name] || {}; return { name, apiKey: pv.apiKey || pv.api_key || '', apiBase: pv.apiBase || pv.api_base || '', hasKey: !!(pv.apiKey || pv.api_key) } })
)

const unconfiguredList = computed<ProviderInfo[]>(() =>
  allProviders
    .filter(name => { const pv = providers.value[name] || {}; return !(pv.apiKey || pv.api_key || pv.apiBase || pv.api_base) })
    .map(name => ({ name, apiKey: '', apiBase: '', hasKey: false }))
)

const providerSelectItems = computed(() => [
  { title: t('provider.autoDetect'), value: 'auto' },
  ...allProviders.filter(n => !['custom', 'azure_openai'].includes(n)).map(n => ({ title: providerDisplayName(n), value: n })),
])

function providerDisplayName(name: string): string {
  const m: Record<string, string> = { openai: 'OpenAI', anthropic: 'Anthropic', google: 'Google', deepseek: 'DeepSeek', openrouter: 'OpenRouter', ollama: 'Ollama', groq: 'Groq', zhipu: '智谱 AI', dashscope: '通义千问', vllm: 'vLLM', gemini: 'Gemini', moonshot: 'Moonshot', minimax: 'MiniMax', mistral: 'Mistral', stepfun: '阶跃星辰', aihubmix: 'AiHubMix', azure_openai: 'Azure OpenAI', custom: 'Custom' }
  return m[name] || name
}
function providerIcon(name: string): string {
  const m: Record<string, string> = { openai: 'alpha-o-circle', anthropic: 'alpha-a-circle', google: 'google', deepseek: 'brain', openrouter: 'transfer', ollama: 'llama' }
  return m[name] || 'cloud-outline'
}
function apiKeyPlaceholder(name: string): string {
  const m: Record<string, string> = { openai: 'sk-...', anthropic: 'sk-ant-...', google: 'AI...', deepseek: 'sk-...', openrouter: 'sk-or-...', ollama: '(not required)' }
  return m[name] || ''
}
function defaultApiBase(name: string): string {
  const m: Record<string, string> = { openai: 'https://api.openai.com/v1', anthropic: 'https://api.anthropic.com', deepseek: 'https://api.deepseek.com', openrouter: 'https://openrouter.ai/api/v1', ollama: 'http://localhost:11434' }
  return m[name] || ''
}

// ====== Debounced save ======
const _timers: Record<string, any> = {}
function debouncedProviderField(provider: string, field: string, value: string) {
  const key = `pv.${provider}.${field}`
  clearTimeout(_timers[key])
  _timers[key] = setTimeout(async () => {
    try { await SetProviderField(provider, field, value); await loadData() } catch (e) { window.__notify?.(String(e), 'error', 'mdi-alert-circle') }
  }, 600)
}
function debouncedAgentField(field: string, value: string) {
  const key = `agent.${field}`
  clearTimeout(_timers[key])
  _timers[key] = setTimeout(async () => {
    try { const updated = { ...agentDefaults.value, [field]: value }; await SetAgentDefaults(JSON.stringify(updated)); await loadData() } catch (e) { window.__notify?.(String(e), 'error', 'mdi-alert-circle') }
  }, 600)
}
async function updateAgentField(field: string, value: string) {
  try { const updated = { ...agentDefaults.value, [field]: value }; await SetAgentDefaults(JSON.stringify(updated)); await loadData() } catch (e) { window.__notify?.(String(e), 'error', 'mdi-alert-circle') }
}

async function loadData() {
  try { providers.value = await GetProviders() } catch {}
  try { agentDefaults.value = await GetAgentDefaults() || {} } catch {}
}
onMounted(loadData)
</script>

<style scoped>
.provider-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 10px;
}
.rotate-90 { transform: rotate(90deg); }
.agent-defaults-card { border-color: rgba(108,92,231,0.35); }
</style>
