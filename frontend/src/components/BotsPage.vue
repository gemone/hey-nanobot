<template>
  <div class="page bots-page">
    <div class="page-header">
      <h2>🤖 Bot Manager</h2>
      <button class="btn btn-primary" @click="showCreateModal = true">+ New Bot</button>
    </div>

    <!-- Bot List -->
    <div class="bots-grid">
      <div
        v-for="bot in bots"
        :key="bot.id"
        class="bot-card"
        :class="{ active: bot.isActive, running: bot.running }"
        @click="selectBot(bot)"
      >
        <div class="bot-avatar">{{ bot.avatar }}</div>
        <div class="bot-info">
          <div class="bot-name">
            {{ bot.name }}
            <span v-if="bot.isActive" class="badge badge-primary">Active</span>
            <span v-if="bot.running" class="badge badge-success">Running</span>
          </div>
          <div class="bot-meta">
            Port {{ bot.port }} · {{ bot.id }}
          </div>
        </div>
        <div class="bot-actions">
          <button
            v-if="!bot.isActive"
            class="btn btn-sm btn-ghost"
            @click.stop="switchBot(bot.id)"
            title="Switch to this bot"
          >↗</button>
          <button
            v-if="!bot.isActive"
            class="btn btn-sm btn-ghost text-danger"
            @click.stop="confirmDelete(bot)"
            title="Delete bot"
          >✕</button>
        </div>
      </div>
    </div>

    <!-- Empty state -->
    <div v-if="bots.length === 0" class="empty-state">
      <div class="empty-icon">🤖</div>
      <p>No bots yet. Create one to get started.</p>
    </div>

    <!-- Create Bot Modal -->
    <div v-if="showCreateModal" class="modal-overlay" @click.self="showCreateModal = false">
      <div class="modal">
        <h3>Create New Bot</h3>
        <div class="form-group">
          <label>Bot Name</label>
          <input v-model="newBotName" type="text" placeholder="My Bot" @keyup.enter="createBot" />
        </div>
        <div class="form-group">
          <label>Avatar</label>
          <div class="avatar-picker">
            <span
              v-for="av in avatarOptions"
              :key="av"
              class="avatar-option"
              :class="{ selected: newBotAvatar === av }"
              @click="newBotAvatar = av"
            >{{ av }}</span>
          </div>
        </div>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="showCreateModal = false">Cancel</button>
          <button class="btn btn-primary" @click="createBot" :disabled="!newBotName.trim()">Create</button>
        </div>
      </div>
    </div>

    <!-- Delete Confirm Modal -->
    <div v-if="showDeleteModal" class="modal-overlay" @click.self="showDeleteModal = false">
      <div class="modal">
        <h3>Delete Bot</h3>
        <p>Are you sure you want to delete <strong>{{ deleteTarget?.name }}</strong>?</p>
        <p class="text-muted">This will remove the bot's config and workspace. This cannot be undone.</p>
        <div class="modal-actions">
          <button class="btn btn-ghost" @click="showDeleteModal = false">Cancel</button>
          <button class="btn btn-danger" @click="deleteBot">Delete</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import {
  ListBots,
  CreateBot,
  DeleteBot,
  SwitchBot,
} from '../../wailsjs/go/main/App'

interface BotInfo {
  id: string
  name: string
  avatar: string
  port: number
  isActive: boolean
  running: boolean
  pid?: number
  createdAt: string
}

const bots = ref<BotInfo[]>([])
const showCreateModal = ref(false)
const showDeleteModal = ref(false)
const newBotName = ref('')
const newBotAvatar = ref('🤖')
const deleteTarget = ref<BotInfo | null>(null)

const avatarOptions = ['🐱', '🤖', '🦊', '🐶', '🐸', '🦁', '🐼', '🦄', '🐲', '👻', '🧠', '⚡', '🌟', '🎯', '🚀', '💡']

async function refresh() {
  try {
    const list = await ListBots()
    bots.value = list || []
  } catch (e) {
    console.error('Failed to list bots:', e)
  }
}

async function createBot() {
  if (!newBotName.value.trim()) return
  try {
    await CreateBot(newBotName.value.trim(), newBotAvatar.value)
    newBotName.value = ''
    newBotAvatar.value = '🤖'
    showCreateModal.value = false
    await refresh()
  } catch (e) {
    console.error('Failed to create bot:', e)
  }
}

function confirmDelete(bot: BotInfo) {
  deleteTarget.value = bot
  showDeleteModal.value = true
}

async function deleteBot() {
  if (!deleteTarget.value) return
  try {
    await DeleteBot(deleteTarget.value.id)
    showDeleteModal.value = false
    deleteTarget.value = null
    await refresh()
  } catch (e) {
    console.error('Failed to delete bot:', e)
  }
}

async function switchBot(id: string) {
  try {
    await SwitchBot(id)
    await refresh()
  } catch (e) {
    console.error('Failed to switch bot:', e)
  }
}

function selectBot(bot: BotInfo) {
  // Clicking the card does nothing special for now
}

onMounted(refresh)
</script>

<style scoped>
.bots-page {
  padding: 20px;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.page-header h2 {
  margin: 0;
  font-size: 18px;
  color: var(--text);
}

.bots-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(300px, 1fr));
  gap: 12px;
}

.bot-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 16px;
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.15s ease;
}

.bot-card:hover {
  border-color: var(--accent);
  background: var(--card-hover);
}

.bot-card.active {
  border-color: var(--accent);
  box-shadow: 0 0 0 1px var(--accent);
}

.bot-card.running {
  border-left: 3px solid #22c55e;
}

.bot-avatar {
  font-size: 28px;
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg);
  border-radius: 10px;
  flex-shrink: 0;
}

.bot-info {
  flex: 1;
  min-width: 0;
}

.bot-name {
  font-weight: 600;
  font-size: 14px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.bot-meta {
  font-size: 12px;
  color: var(--text-muted);
  margin-top: 2px;
}

.badge {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 4px;
  font-weight: 600;
  text-transform: uppercase;
}

.badge-primary {
  background: rgba(139, 92, 246, 0.2);
  color: var(--accent);
}

.badge-success {
  background: rgba(34, 197, 94, 0.2);
  color: #22c55e;
}

.bot-actions {
  display: flex;
  gap: 4px;
}

.btn-sm {
  padding: 4px 8px;
  font-size: 14px;
}

.text-danger {
  color: #ef4444;
}

.text-danger:hover {
  background: rgba(239, 68, 68, 0.1);
}

/* Modal */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  backdrop-filter: blur(4px);
}

.modal {
  background: var(--card);
  border: 1px solid var(--border);
  border-radius: 12px;
  padding: 24px;
  width: 400px;
  max-width: 90vw;
}

.modal h3 {
  margin: 0 0 16px 0;
  font-size: 16px;
}

.modal p {
  margin: 8px 0;
  font-size: 14px;
}

.text-muted {
  color: var(--text-muted);
  font-size: 13px;
}

.form-group {
  margin-bottom: 16px;
}

.form-group label {
  display: block;
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 6px;
  color: var(--text-muted);
}

.form-group input {
  width: 100%;
  padding: 8px 12px;
  background: var(--bg);
  border: 1px solid var(--border);
  border-radius: 8px;
  color: var(--text);
  font-size: 14px;
  outline: none;
  box-sizing: border-box;
}

.form-group input:focus {
  border-color: var(--accent);
}

.avatar-picker {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.avatar-option {
  font-size: 24px;
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  cursor: pointer;
  border: 2px solid transparent;
  transition: all 0.15s;
}

.avatar-option:hover {
  background: var(--bg);
}

.avatar-option.selected {
  border-color: var(--accent);
  background: rgba(139, 92, 246, 0.1);
}

.modal-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 20px;
}

.empty-state {
  text-align: center;
  padding: 60px 20px;
  color: var(--text-muted);
}

.empty-icon {
  font-size: 48px;
  margin-bottom: 12px;
}

.btn-danger {
  background: #ef4444;
  color: white;
  border: none;
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-weight: 600;
}

.btn-danger:hover {
  background: #dc2626;
}
</style>
