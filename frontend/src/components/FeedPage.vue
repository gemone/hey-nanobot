<template>
  <div class="feed-container">
    <div class="page-header">
      <h2>📡 Live Feed</h2>
      <div class="actions">
        <button class="btn btn-ghost btn-sm" @click="refresh">🔄</button>
        <button class="btn btn-ghost btn-sm" @click="clearAll">🗑️ Clear</button>
      </div>
    </div>

    <div class="feed-messages" ref="feedEl">
      <div v-if="channelMsgs.length === 0" class="empty-state">
        <span class="emoji">📡</span>
        <span class="text">{{ gatewayRunning ? 'Waiting for messages...' : 'Start gateway to see live messages' }}</span>
      </div>

      <div
        v-for="(msg, i) in channelMsgs"
        :key="i"
        class="feed-item"
      >
        <div class="feed-icon">{{ getChannelIcon(msg.channel) }}</div>
        <div class="feed-body">
          <div class="feed-meta">
            <span class="feed-channel" :class="getChannelClass(msg.channel)">
              {{ msg.channel }}
            </span>
            <span class="feed-sender">{{ msg.sender_id || msg.role }}</span>
            <span class="feed-time">{{ formatTime(msg.time) }}</span>
          </div>
          <div class="feed-text">{{ msg.content }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, nextTick, onMounted } from 'vue'
import { GetChannelMessages, ClearChannelMessages } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'

const props = defineProps<{ gatewayRunning?: boolean }>()

const channelMsgs = ref<any[]>([])
const feedEl = ref<HTMLDivElement>()

function scrollToBottom() {
  nextTick(() => {
    if (feedEl.value) feedEl.value.scrollTop = feedEl.value.scrollHeight
  })
}

async function refresh() {
  try {
    channelMsgs.value = await GetChannelMessages()
    scrollToBottom()
  } catch {}
}

async function clearAll() {
  try {
    await ClearChannelMessages()
    channelMsgs.value = []
  } catch {}
}

function getChannelIcon(ch: string): string {
  const icons: Record<string, string> = {
    telegram: '✈️', discord: '🎮', slack: '💼', qq: '🐧',
    wecom: '🏢', feishu: '🪽', dingtalk: '🔔', whatsapp: '📱',
    email: '📧', mochat: '🫧',
  }
  return icons[ch] || '📡'
}

function getChannelClass(ch: string): string {
  return `ch-${ch}`
}

function formatTime(t: string): string {
  if (!t) return ''
  try { return new Date(t).toLocaleTimeString() } catch { return t }
}

onMounted(async () => {
  await refresh()
  EventsOn('channel:message', () => {
    refresh()
  })
  EventsOn('channel:messages:cleared', () => { channelMsgs.value = [] })
})
</script>
