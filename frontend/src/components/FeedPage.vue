<template>
  <div class="d-flex flex-column" style="height: 100%;">
    <div class="page-header">
      <h2 class="text-body-1 font-weight-bold">{{ t('feed.title') }}</h2>
      <v-btn variant="text" size="small" prepend-icon="mdi-notification-clear-all" @click="$emit('clear')">{{ t('feed.clear') }}</v-btn>
    </div>
    <div class="flex-grow-1 overflow-y-auto pa-3">
      <div v-if="!gatewayRunning" class="d-flex flex-column align-center justify-center" style="height: 100%; opacity: 0.4;">
        <v-icon size="48" color="grey">mdi-web-off</v-icon>
        <span class="text-body-2 text-medium-emphasis mt-2">{{ t('feed.noMessages') }}</span>
      </div>
      <div v-else-if="!messages.length" class="d-flex flex-column align-center justify-center" style="height: 100%; opacity: 0.4;">
        <v-icon size="48" color="grey">mdi-broadcast</v-icon>
        <span class="text-body-2 text-medium-emphasis mt-2">{{ t('feed.noMessages') }}</span>
      </div>
      <template v-else>
        <div v-for="(msg, i) in messages" :key="i" class="d-flex ga-2 pa-2 rounded-lg mb-1 cursor-pointer" style="border: 1px solid transparent;" @mouseenter="$event.currentTarget.style.borderColor='#2a2a45'" @mouseleave="$event.currentTarget.style.borderColor='transparent'">
          <v-avatar size="28" rounded="circle" :color="channelColor(msg.channel)">
            <span style="font-size: 14px;">{{ channelEmoji(msg.channel) }}</span>
          </v-avatar>
          <div class="flex-grow-1" style="min-width: 0;">
            <div class="d-flex align-center ga-2 mb-1">
              <v-chip size="x-small" :color="channelColor(msg.channel)" variant="tonal" class="text-uppercase">{{ msg.channel }}</v-chip>
              <span class="text-body-2 font-weight-semibold">{{ msg.sender }}</span>
              <span class="text-caption text-medium-emphasis ml-auto">{{ msg.time }}</span>
            </div>
            <div class="text-body-2 text-medium-emphasis text-truncate">{{ msg.text }}</div>
          </div>
        </div>
      </template>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()

defineProps<{ gatewayRunning: boolean }>()
defineEmits(['clear'])

const messages = defineModel<any[]>('messages', { default: [] })

function channelEmoji(ch: string) {
  const m: Record<string, string> = { telegram: '✈️', discord: '🎮', slack: '💬', qq: '🐧', wecom: '💼', feishu: '🐦', dingtalk: '🔔', whatsapp: '📱', email: '📧' }
  return m[ch] || '📡'
}
function channelColor(ch: string) {
  const m: Record<string, string> = { telegram: '#0088cc', discord: '#5865f2', slack: '#e01e5a', qq: '#12b7f5', wecom: '#07c160', feishu: '#3370ff', dingtalk: '#0089ff', whatsapp: '#25d366', email: '#ea4335' }
  return m[ch] || '#6c5ce7'
}
</script>
