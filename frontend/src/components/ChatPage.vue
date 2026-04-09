<template>
  <div class="chat-container">
    <div class="page-header">
      <h2>💬 Chat</h2>
      <div class="actions">
        <button class="btn btn-ghost btn-sm" @click="clearChat">🗑️ Clear</button>
      </div>
    </div>

    <div class="chat-messages" ref="messagesEl">
      <div v-if="messages.length === 0" class="empty-state">
        <span class="emoji">🐈</span>
        <span class="text">Send a message to nanobot</span>
      </div>
      <div v-for="(msg) in messages" :key="msg.id" class="chat-message" :class="msg.role">
        <div class="avatar">{{ msg.role === 'user' ? '👤' : '🐈' }}</div>
        <div class="bubble">
          <div v-if="msg.role === 'assistant'" class="md-content" v-html="renderMd(msg.content)"></div>
          <template v-else>{{ msg.content }}</template>
          <span v-if="msg.streaming" class="streaming-cursor"></span>
        </div>
      </div>
      <div v-if="loading && !streamingId" class="chat-message assistant">
        <div class="avatar">🐈</div>
        <div class="bubble" style="color: var(--text-muted); font-style: italic;">Thinking...</div>
      </div>
    </div>

    <div class="chat-input-area">
      <input
        v-model="input"
        placeholder="Message nanobot... (Enter to send)"
        @keydown.enter="send"
        :disabled="loading"
      />
      <button class="btn btn-primary" @click="send" :disabled="loading || !input.trim()">
        {{ loading ? '⏳' : '➤' }} Send
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, nextTick, onMounted } from 'vue'
import { SendMessage, GetMessages, ClearMessages } from '../../wailsjs/go/main/App'
import { EventsOn } from '../../wailsjs/runtime/runtime'
import { marked } from 'marked'

// Configure marked for safe rendering
marked.setOptions({
  breaks: true,
  gfm: true,
})

function renderMd(text: string): string {
  if (!text) return ''
  try {
    return marked.parse(text) as string
  } catch {
    return text
  }
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
    messages.value.push({ id: 'err', role: 'assistant', content: `Error: ${e}`, streaming: false })
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
