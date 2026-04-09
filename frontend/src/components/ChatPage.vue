<template>
  <div class="d-flex flex-column" style="height: 100%;">
    <div class="page-header">
      <h2 class="text-body-1 font-weight-bold">{{ t('chat.title') }}</h2>
      <v-btn variant="text" size="small" prepend-icon="mdi-delete-outline" @click="clearChat">{{ t('chat.clear') }}</v-btn>
    </div>

    <div class="flex-grow-1 overflow-y-auto pa-4" ref="messagesEl">
      <!-- Empty state -->
      <div v-if="messages.length === 0" class="d-flex flex-column align-center justify-center" style="height: 100%; opacity: 0.4;">
        <span style="font-size: 48px;">🐈</span>
        <span class="text-body-2 text-medium-emphasis mt-2">{{ t('chat.welcome') }}</span>
      </div>

      <!-- Messages -->
      <div v-for="msg in messages" :key="msg.id" class="d-flex mb-3" :class="msg.role === 'user' ? 'justify-end' : 'justify-start'">
        <v-avatar size="28" :color="msg.role === 'user' ? '#74b9ff' : '#00cec9'" class="mr-2 mt-1 flex-shrink-0">
          <span style="font-size: 14px;">{{ msg.role === 'user' ? '👤' : '🐈' }}</span>
        </v-avatar>
        <v-card
          :color="msg.role === 'user' ? 'rgba(108,92,231,0.15)' : '#1a1a30'"
          :style="msg.role === 'user' ? 'border: 1px solid #6c5ce7;' : 'border: 1px solid #2a2a45;'"
          rounded="lg"
          max-width="80%"
          class="pa-3 text-body-2"
          style="line-height: 1.6; word-break: break-word;"
        >
          <div v-if="msg.role === 'assistant'" class="md-content" v-html="renderMd(msg.content)" style="user-select: text; -webkit-user-select: text;"></div>
          <template v-else>{{ msg.content }}</template>
          <span v-if="msg.streaming" class="streaming-cursor"></span>
        </v-card>
      </div>

      <!-- Thinking -->
      <div v-if="loading && !streamingId" class="d-flex">
        <v-avatar size="28" color="#00cec9" class="mr-2 mt-1">
          <span style="font-size: 14px;">🐈</span>
        </v-avatar>
        <v-card color="#1a1a30" rounded="lg" class="pa-3 text-body-2 text-medium-emphasis" style="font-style: italic; border: 1px solid #2a2a45;">
          {{ t('common.loading') }}
        </v-card>
      </div>
    </div>

    <!-- Input -->
    <div class="d-flex ga-2 pa-3" style="border-top: 1px solid #2a2a45;">
      <v-text-field
        v-model="input"
        :placeholder="t('chat.placeholder')"
        variant="outlined"
        density="compact"
        hide-details
        :disabled="loading"
        @keydown.enter="send"
      />
      <v-btn color="primary" @click="send" :disabled="loading || !input.trim()" icon>
        <v-icon>{{ loading ? 'mdi-hourglass-empty' : 'mdi-send' }}</v-icon>
      </v-btn>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { SendMessage, GetMessages, ClearMessages } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { marked } from 'marked'

const { t } = useI18n()

marked.setOptions({ breaks: true, gfm: true })

function renderMd(text: string): string {
  if (!text) return ''
  try { return marked.parse(text) as string }
  catch { return text }
}

const input = ref('')
const messages = ref<any[]>([])
const loading = ref(false)
const streamingId = ref('')
const messagesEl = ref<HTMLDivElement>()

function scrollToBottom() {
  nextTick(() => {
    if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight
  })
}

async function send() {
  const msg = input.value.trim()
  if (!msg || loading.value) return
  input.value = ''
  loading.value = true
  scrollToBottom()
  try { await SendMessage(msg) }
  catch (e) {
    messages.value.push({ id: 'err', role: 'assistant', content: t('chat.error', { error: String(e) }), streaming: false })
  }
  loading.value = false
  scrollToBottom()
}

async function clearChat() {
  await ClearMessages()
  messages.value = []
}

onMounted(async () => {
  messages.value = await GetMessages()
  scrollToBottom()

  EventsOn('chat:message', () => {
    GetMessages().then(msgs => {
      messages.value = msgs
      loading.value = false
      scrollToBottom()
    })
  })

  EventsOn('chat:stream', (data: any) => {
    streamingId.value = data?.id || ''
    const idx = messages.value.findIndex((m: any) => m.id === data?.id)
    if (idx >= 0) {
      messages.value[idx].content = data.content
    }
    scrollToBottom()
  })

  EventsOn('chat:stream:done', (data: any) => {
    if (streamingId.value === data?.id) streamingId.value = ''
    GetMessages().then(msgs => { messages.value = msgs; scrollToBottom() })
  })
})
</script>

<style scoped>
.streaming-cursor::after {
  content: '▊'; animation: blink 1s step-end infinite; color: #6c5ce7;
}
@keyframes blink { 0%,100% { opacity: 1; } 50% { opacity: 0; } }

.md-content :deep(p) { margin: 0 0 6px; }
.md-content :deep(p:last-child) { margin-bottom: 0; }
.md-content :deep(code) {
  background: #252540; padding: 1px 4px; border-radius: 3px;
  font-size: 12px; font-family: 'SF Mono', monospace;
}
.md-content :deep(pre) {
  background: #0f0f1a; padding: 8px; border-radius: 6px;
  overflow-x: auto; margin: 6px 0;
}
.md-content :deep(pre code) { background: none; padding: 0; }
.md-content :deep(ul), .md-content :deep(ol) { padding-left: 20px; margin: 4px 0; }
.md-content :deep(blockquote) {
  border-left: 3px solid #6c5ce7; padding-left: 10px; margin: 6px 0;
  color: #9898b0;
}
</style>
