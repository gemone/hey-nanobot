<template>
  <div class="d-flex flex-column" style="height: 100%;">
    <div class="page-header">
      <div class="d-flex align-center ga-2">
        <v-icon size="20" color="primary">mdi-chat-outline</v-icon>
        <span class="text-body-1 font-weight-bold">{{ t('chat.title') }}</span>
      </div>
      <v-btn variant="text" size="small" prepend-icon="mdi-delete-outline" @click="clearChat" style="color: #5a5a78;">{{ t('chat.clear') }}</v-btn>
    </div>

    <!-- Messages -->
    <div class="flex-grow-1 overflow-y-auto pa-4" ref="messagesEl" style="scroll-behavior: smooth;">
      <div v-if="messages.length === 0" class="d-flex flex-column align-center justify-center" style="height: 100%; opacity: 0.35;">
        <span style="font-size: 56px;">🐈</span>
        <span class="text-body-2 mt-3" style="color: #5a5a78;">{{ t('chat.welcome') }}</span>
      </div>

      <div v-for="msg in messages" :key="msg.id" class="d-flex mb-3" :class="msg.role === 'user' ? 'justify-end' : 'justify-start'">
        <div class="d-flex ga-2" :style="{ flexDirection: msg.role === 'user' ? 'row-reverse' : 'row' }">
          <v-avatar size="26" :color="msg.role === 'user' ? '#6c5ce7' : '#00cec9'" class="mt-1 flex-shrink-0">
            <span style="font-size: 13px;">{{ msg.role === 'user' ? '👤' : '🐈' }}</span>
          </v-avatar>
          <div
            class="msg-bubble"
            :class="msg.role"
          >
            <div v-if="msg.role === 'assistant'" class="md-content" v-html="renderMd(msg.content)" style="user-select: text; -webkit-user-select: text;"></div>
            <template v-else>{{ msg.content }}</template>
            <span v-if="msg.streaming" class="streaming-cursor"></span>
          </div>
        </div>
      </div>

      <div v-if="loading && !streamingId" class="d-flex ga-2">
        <v-avatar size="26" color="#00cec9" class="mt-1">
          <span style="font-size: 13px;">🐈</span>
        </v-avatar>
        <div class="msg-bubble assistant" style="font-style: italic; color: #5a5a78;">
          {{ t('common.loading') }}
        </div>
      </div>
    </div>

    <!-- Input -->
    <div class="chat-input-area">
      <v-text-field
        v-model="input"
        :placeholder="t('chat.placeholder')"
        variant="outlined"
        density="compact"
        hide-details
        :disabled="loading"
        @keydown.enter="send"
      />
      <v-btn color="primary" @click="send" :disabled="loading || !input.trim()" icon size="small">
        <v-icon size="18">{{ loading ? 'mdi-hourglass-empty' : 'mdi-send' }}</v-icon>
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
  try { return marked.parse(text) as string } catch { return text }
}

const input = ref('')
const messages = ref<any[]>([])
const loading = ref(false)
const streamingId = ref('')
const messagesEl = ref<HTMLDivElement>()

function scrollToBottom() {
  nextTick(() => { if (messagesEl.value) messagesEl.value.scrollTop = messagesEl.value.scrollHeight })
}

async function send() {
  const msg = input.value.trim()
  if (!msg || loading.value) return
  input.value = ''
  loading.value = true
  scrollToBottom()
  try { await SendMessage(msg) }
  catch (e) { messages.value.push({ id: 'err', role: 'assistant', content: t('chat.error', { error: String(e) }), streaming: false }) }
  loading.value = false
  scrollToBottom()
}

async function clearChat() { await ClearMessages(); messages.value = [] }

onMounted(async () => {
  messages.value = await GetMessages()
  scrollToBottom()

  EventsOn('chat:message', () => {
    GetMessages().then(msgs => { messages.value = msgs; loading.value = false; scrollToBottom() })
  })
  EventsOn('chat:stream', (data: any) => {
    streamingId.value = data?.id || ''
    const idx = messages.value.findIndex((m: any) => m.id === data?.id)
    if (idx >= 0) messages.value[idx].content = data.content
    scrollToBottom()
  })
  EventsOn('chat:stream:done', (data: any) => {
    if (streamingId.value === data?.id) streamingId.value = ''
    GetMessages().then(msgs => { messages.value = msgs; scrollToBottom() })
  })
})
</script>

<style scoped>
.msg-bubble {
  max-width: 75%;
  padding: 8px 12px;
  border-radius: 10px;
  font-size: 13px;
  line-height: 1.6;
  word-break: break-word;
}
.msg-bubble.user {
  background: rgba(var(--v-theme-primary), 0.1);
  border: 1px solid rgba(var(--v-theme-primary), 0.2);
}
.msg-bubble.assistant {
  background: rgb(var(--v-theme-surface));
  border: 1px solid rgba(128,128,128,0.1);
}

.streaming-cursor::after { content: '▊'; animation: blink 1s step-end infinite; color: #6c5ce7; }
@keyframes blink { 0%,100% { opacity: 1; } 50% { opacity: 0; } }

.md-content :deep(p) { margin: 0 0 6px; }
.md-content :deep(p:last-child) { margin-bottom: 0; }
.md-content :deep(code) { background: #252540; padding: 1px 4px; border-radius: 3px; font-size: 12px; font-family: 'SF Mono', monospace; }
.md-content :deep(pre) { background: #0f0f1a; padding: 8px; border-radius: 6px; overflow-x: auto; margin: 6px 0; }
.md-content :deep(pre code) { background: none; padding: 0; }
.md-content :deep(ul), .md-content :deep(ol) { padding-left: 20px; margin: 4px 0; }
.md-content :deep(blockquote) { border-left: 3px solid #6c5ce7; padding-left: 10px; margin: 6px 0; color: #9898b0; }

.chat-input-area {
  padding: 8px 16px;
  border-top: 1px solid rgba(128,128,128,0.1);
  display: flex;
  gap: 8px;
  align-items: center;
  -webkit-app-region: no-drag;
}
</style>
