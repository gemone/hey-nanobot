<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-4">
      <div class="d-flex align-center ga-2">
        <v-icon size="20" color="primary">mdi-link-variant</v-icon>
        <span class="text-body-1 font-weight-bold">{{ t('channel.title') }}</span>
        <v-chip v-if="enabledCount" size="x-small" color="primary" variant="tonal">{{ enabledCount }} {{ t('channel.enabled') }}</v-chip>
      </div>
    </div>

    <div class="channel-grid">
      <div v-for="ch in channelList" :key="ch.name" class="card-base pa-4" :class="{ 'channel-active': ch.enabled }">
        <!-- Header -->
        <div class="d-flex align-center ga-2 mb-2">
          <div class="ch-icon-wrap" :style="{ background: channelColor(ch.name) + '18' }">
            <v-icon size="16" :color="channelColor(ch.name)">mdi-{{ channelIcon(ch.name) }}</v-icon>
          </div>
          <span class="text-body-2 font-weight-semibold">{{ channelDisplayName(ch.name) }}</span>
          <v-switch
            :model-value="ch.enabled"
            @update:model-value="toggleChannel(ch.name, $event)"
            density="compact"
            hide-details
            color="primary"
            class="ml-auto"
            style="margin-top: 0;"
          />
        </div>

        <!-- Fields (when enabled) -->
        <template v-if="ch.enabled">
          <v-text-field
            v-for="field in getChannelFields(ch.name)"
            :key="field.key"
            :model-value="ch.data[field.key] || ''"
            @update:model-value="debouncedField(ch.name, field.key, $event)"
            :label="field.label"
            variant="outlined"
            density="compact"
            hide-details
            class="mb-2"
            :type="field.secret ? 'password' : 'text'"
          />
          <div class="d-flex align-center mt-1">
            <span class="text-caption" style="color: #5a5a78;">{{ t('channel.streaming') }}</span>
            <v-switch
              :model-value="ch.data.streaming || false"
              @update:model-value="setField(ch.name, 'streaming', JSON.stringify($event))"
              density="compact"
              hide-details
              color="primary"
              size="small"
              class="ml-2"
              style="margin-top: 0;"
            />
          </div>
        </template>
        <div v-else class="text-caption" style="color: #3a3a58; padding: 4px 0;">
          {{ t('channel.disabled') }}
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetChannels, SetChannelField } from '../../wailsjs/go/main/App'

const { t } = useI18n()
const channels = ref<Record<string, any>>({})

interface ChannelInfo { name: string; enabled: boolean; data: Record<string, any> }

const channelDefs: Record<string, { fields: { key: string; labelKey: string; secret?: boolean }[] }> = {
  telegram: { fields: [{ key: 'token', labelKey: 'channel.fields.token', secret: true }] },
  discord: { fields: [{ key: 'token', labelKey: 'channel.fields.token', secret: true }] },
  qq: { fields: [{ key: 'app_id', labelKey: 'channel.fields.appId' }, { key: 'secret', labelKey: 'channel.fields.appSecret', secret: true }] },
  slack: { fields: [{ key: 'bot_token', labelKey: 'channel.fields.botToken', secret: true }, { key: 'app_token', labelKey: 'channel.fields.appToken', secret: true }] },
  feishu: { fields: [{ key: 'app_id', labelKey: 'channel.fields.appId' }, { key: 'app_secret', labelKey: 'channel.fields.appSecret', secret: true }] },
  dingtalk: { fields: [{ key: 'client_id', labelKey: 'channel.fields.clientId' }, { key: 'client_secret', labelKey: 'channel.fields.clientSecret', secret: true }] },
  wecom: { fields: [{ key: 'corp_id', labelKey: 'channel.fields.corpId' }, { key: 'agent_id', labelKey: 'channel.fields.agentId' }, { key: 'secret', labelKey: 'channel.fields.secret', secret: true }] },
  whatsapp: { fields: [{ key: 'token', labelKey: 'channel.fields.token', secret: true }, { key: 'phone_number_id', labelKey: 'channel.fields.phoneNumberId' }] },
}

const allChannelNames = Object.keys(channelDefs)
const channelList = computed<ChannelInfo[]>(() => allChannelNames.map(name => {
  const data = channels.value[name] || {}
  return { name, enabled: !!data.enabled, data }
}))
const enabledCount = computed(() => channelList.value.filter(c => c.enabled).length)

function channelIcon(name: string): string { return { telegram: 'send-variant', discord: 'chat', qq: 'chat', slack: 'slack', feishu: 'bird', dingtalk: 'bell', wecom: 'briefcase-variant', whatsapp: 'phone' }[name] || 'message-text' }
function channelDisplayName(name: string): string { return { telegram: 'Telegram', discord: 'Discord', qq: 'QQ', slack: 'Slack', feishu: '飞书', dingtalk: '钉钉', wecom: '企业微信', whatsapp: 'WhatsApp' }[name] || name }
function channelColor(name: string): string { return { telegram: '#0088cc', discord: '#5865f2', qq: '#12b7f5', slack: '#e01e5a', feishu: '#3370ff', dingtalk: '#0089ff', wecom: '#07c160', whatsapp: '#25d366' }[name] || '#6c5ce7' }

function getChannelFields(name: string) {
  const def = channelDefs[name]; if (!def) return []
  return def.fields.map(f => ({ ...f, label: t(f.labelKey) }))
}

async function toggleChannel(name: string, enabled: boolean) {
  try { await SetChannelField(name, 'enabled', JSON.stringify(enabled)); await loadData() } catch (e) { window.__notify?.(String(e), 'error', 'mdi-alert-circle') }
}
async function setField(channel: string, field: string, value: string) {
  try { await SetChannelField(channel, field, value); await loadData() } catch (e) { window.__notify?.(String(e), 'error', 'mdi-alert-circle') }
}

// Debounced field save
const _timers: Record<string, any> = {}
function debouncedField(channel: string, field: string, value: string) {
  const key = `ch.${channel}.${field}`
  clearTimeout(_timers[key])
  _timers[key] = setTimeout(async () => { await setField(channel, field, value) }, 600)
}

async function loadData() { try { channels.value = await GetChannels() } catch {} }
onMounted(loadData)
</script>

<style scoped>
.channel-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 10px;
}
.ch-icon-wrap {
  width: 28px; height: 28px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.channel-active { border-color: rgba(108,92,231,0.3) !important; }
</style>
