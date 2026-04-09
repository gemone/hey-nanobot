<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">{{ t('channel.title') }}</h2>
      <v-chip size="small" variant="tonal" color="primary">{{ enabledCount }} / {{ channelList.length }}</v-chip>
    </div>

    <v-row>
      <v-col v-for="ch in channelList" :key="ch.name" cols="12" sm="6" md="4">
        <v-card rounded="lg" class="pa-4" :style="channelCardStyle(ch)" @mouseenter="ch._hover = true" @mouseleave="ch._hover = false">
          <div class="d-flex align-center ga-2 mb-3">
            <v-icon size="18" color="primary">mdi-{{ channelIcon(ch.name) }}</v-icon>
            <span class="text-body-2 font-weight-semibold">{{ channelDisplayName(ch.name) }}</span>
            <v-switch
              :model-value="ch.enabled"
              @update:model-value="toggleChannel(ch.name, $event)"
              density="compact"
              hide-details
              color="primary"
              class="ml-auto"
            />
          </div>
          <template v-if="ch.enabled">
            <v-text-field
              v-for="field in getChannelFields(ch.name)"
              :key="field.key"
              :model-value="ch.data[field.key] || ''"
              @update:model-value="updateField(ch.name, field.key, $event)"
              :label="field.label"
              variant="outlined"
              density="compact"
              hide-details
              class="mb-2"
              :type="field.secret ? 'password' : 'text'"
              :prepend-inner-icon="field.icon"
            />
            <!-- Streaming toggle for channels that support it -->
            <div v-if="ch.data.streaming !== undefined || ch.enabled" class="d-flex align-center mt-1">
              <span class="text-caption text-medium-emphasis mr-2">{{ t('channel.streaming') }}</span>
              <v-switch
                :model-value="ch.data.streaming || false"
                @update:model-value="updateField(ch.name, 'streaming', JSON.stringify($event))"
                density="compact"
                hide-details
                color="primary"
                size="small"
              />
            </div>
          </template>
          <div v-else class="text-caption text-medium-emphasis pa-2">
            {{ t('channel.disabled') }}
          </div>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetChannels, SetChannelField } from '../../wailsjs/go/main/App'

const { t } = useI18n()

const channels = ref<Record<string, any>>({})

interface ChannelInfo {
  name: string
  enabled: boolean
  data: Record<string, any>
  _hover: boolean
}

// Channel field definitions — aligned with nanobot source code
const channelDefs: Record<string, { fields: { key: string; labelKey: string; secret?: boolean; icon: string }[] }> = {
  telegram: {
    fields: [
      { key: 'token', labelKey: 'channel.fields.token', secret: true, icon: 'mdi-key-variant' },
    ],
  },
  discord: {
    fields: [
      { key: 'token', labelKey: 'channel.fields.token', secret: true, icon: 'mdi-key-variant' },
    ],
  },
  qq: {
    fields: [
      { key: 'app_id', labelKey: 'channel.fields.appId', icon: 'mdi-identifier' },
      { key: 'secret', labelKey: 'channel.fields.appSecret', secret: true, icon: 'mdi-key-variant' },
    ],
  },
  slack: {
    fields: [
      { key: 'bot_token', labelKey: 'channel.fields.botToken', secret: true, icon: 'mdi-key-variant' },
      { key: 'app_token', labelKey: 'channel.fields.appToken', secret: true, icon: 'mdi-key-variant' },
    ],
  },
  feishu: {
    fields: [
      { key: 'app_id', labelKey: 'channel.fields.appId', icon: 'mdi-identifier' },
      { key: 'app_secret', labelKey: 'channel.fields.appSecret', secret: true, icon: 'mdi-key-variant' },
    ],
  },
  dingtalk: {
    fields: [
      { key: 'client_id', labelKey: 'channel.fields.clientId', icon: 'mdi-identifier' },
      { key: 'client_secret', labelKey: 'channel.fields.clientSecret', secret: true, icon: 'mdi-key-variant' },
    ],
  },
  wecom: {
    fields: [
      { key: 'corp_id', labelKey: 'channel.fields.corpId', icon: 'mdi-identifier' },
      { key: 'agent_id', labelKey: 'channel.fields.agentId', icon: 'mdi-identifier' },
      { key: 'secret', labelKey: 'channel.fields.secret', secret: true, icon: 'mdi-key-variant' },
    ],
  },
  whatsapp: {
    fields: [
      { key: 'token', labelKey: 'channel.fields.token', secret: true, icon: 'mdi-key-variant' },
      { key: 'phone_number_id', labelKey: 'channel.fields.phoneNumberId', icon: 'mdi-phone' },
    ],
  },
}

const allChannelNames = Object.keys(channelDefs)

const channelList = computed<ChannelInfo[]>(() => {
  return allChannelNames.map(name => {
    const data = channels.value[name] || {}
    return {
      name,
      enabled: !!data.enabled,
      data,
      _hover: false,
    }
  })
})

const enabledCount = computed(() => channelList.value.filter(c => c.enabled).length)

function channelIcon(name: string): string {
  const icons: Record<string, string> = {
    telegram: 'send', discord: 'chat', qq: 'chat',
    slack: 'slack', feishu: 'bird', dingtalk: 'bell',
    wecom: 'briefcase', whatsapp: 'phone',
  }
  return icons[name] || 'message-text'
}

function channelDisplayName(name: string): string {
  const names: Record<string, string> = {
    telegram: 'Telegram', discord: 'Discord', qq: 'QQ',
    slack: 'Slack', feishu: '飞书', dingtalk: '钉钉',
    wecom: '企业微信', whatsapp: 'WhatsApp',
  }
  return names[name] || name
}

function getChannelFields(name: string) {
  const def = channelDefs[name]
  if (!def) return []
  return def.fields.map(f => ({
    ...f,
    label: t(f.labelKey),
  }))
}

function channelCardStyle(ch: ChannelInfo) {
  return {
    border: ch.enabled
      ? (ch._hover ? '1px solid #6c5ce7' : '1px solid rgba(108,92,231,0.3)')
      : '1px solid #2a2a45',
    transition: 'border-color 0.2s',
  }
}

async function toggleChannel(name: string, enabled: boolean) {
  try {
    await SetChannelField(name, 'enabled', JSON.stringify(enabled))
    await loadData()
  } catch (e) {
    alert(t('common.error') + ': ' + e)
  }
}

async function updateField(channel: string, field: string, value: string) {
  try {
    await SetChannelField(channel, field, value)
    await loadData()
  } catch (e) {
    alert(t('common.error') + ': ' + e)
  }
}

async function loadData() {
  try { channels.value = await GetChannels() } catch {}
}

onMounted(loadData)
</script>
