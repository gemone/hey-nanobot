<template>
  <div class="d-flex flex-column" style="height: 100%;">
    <div class="page-header">
      <div class="d-flex align-center ga-2">
        <v-icon size="20" color="primary">mdi-broadcast</v-icon>
        <span class="text-body-1 font-weight-bold">{{ t('feed.title') }}</span>
      </div>
      <v-btn variant="text" size="small" prepend-icon="mdi-notification-clear-all" @click="$emit('clear')" style="color: #5a5a78;">{{ t('feed.clear') }}</v-btn>
    </div>
    <div class="flex-grow-1 overflow-y-auto pa-3">
      <div v-if="!gatewayRunning" class="d-flex flex-column align-center justify-center" style="height: 100%; opacity: 0.35;">
        <v-icon size="44" color="#5a5a78">mdi-web-off</v-icon>
        <span class="text-body-2 mt-3" style="color: #5a5a78;">{{ t('feed.noMessages') }}</span>
      </div>
      <div v-else-if="!messages.length" class="d-flex flex-column align-center justify-center" style="height: 100%; opacity: 0.35;">
        <v-icon size="44" color="#5a5a78">mdi-broadcast</v-icon>
        <span class="text-body-2 mt-3" style="color: #5a5a78;">{{ t('feed.noMessages') }}</span>
      </div>
      <template v-else>
        <div v-for="(msg, i) in messages" :key="i" class="feed-item">
          <div class="feed-avatar" :style="{ background: channelColor(msg.channel) + '20' }">
            <span style="font-size: 14px;">{{ channelEmoji(msg.channel) }}</span>
          </div>
          <div class="flex-grow-1" style="min-width: 0;">
            <div class="d-flex align-center ga-2 mb-1">
              <span class="feed-channel-tag" :style="{ background: channelColor(msg.channel) + '20', color: channelColor(msg.channel) }">{{ msg.channel }}</span>
              <span class="text-caption font-weight-semibold">{{ msg.sender }}</span>
              <span class="text-caption ml-auto" style="color: #3a3a58;">{{ msg.time }}</span>
            </div>
            <div class="text-body-2 text-truncate" style="color: #7a7a98;">{{ msg.text }}</div>
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

function channelEmoji(ch: string) { return { telegram: '✈️', discord: '🎮', slack: '💬', qq: '🐧', wecom: '💼', feishu: '🐦', dingtalk: '🔔', whatsapp: '📱', email: '📧' }[ch] || '📡' }
function channelColor(ch: string) { return { telegram: '#0088cc', discord: '#5865f2', slack: '#e01e5a', qq: '#12b7f5', wecom: '#07c160', feishu: '#3370ff', dingtalk: '#0089ff', whatsapp: '#25d366', email: '#ea4335' }[ch] || '#6c5ce7' }
</script>

<style scoped>
.feed-item {
  display: flex; gap: 10px; padding: 8px 12px;
  border-radius: 8px; margin-bottom: 4px;
  transition: background 0.15s;
  cursor: pointer;
}
.feed-item:hover { background: rgb(var(--v-theme-surface)); }
.feed-avatar {
  width: 28px; height: 28px; border-radius: 8px;
  display: flex; align-items: center; justify-content: center;
  flex-shrink: 0;
}
.feed-channel-tag {
  font-size: 10px; padding: 1px 6px; border-radius: 4px;
  font-weight: 600; text-transform: uppercase; letter-spacing: 0.3px;
}
</style>
