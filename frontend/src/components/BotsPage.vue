<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">{{ t('bot.manager') }}</h2>
      <v-btn color="primary" size="small" prepend-icon="mdi-plus" @click="openCreate">{{ t('bot.newBot') }}</v-btn>
    </div>

    <!-- Bot Grid -->
    <v-row v-if="bots.length">
      <v-col v-for="bot in bots" :key="bot.id" cols="12" sm="6" md="4">
        <v-card
          rounded="lg"
          class="pa-4"
          :style="botCardStyle(bot)"
        >
          <div class="d-flex align-center ga-3">
            <v-avatar size="44" rounded="lg" color="#1a1a30">
              <span style="font-size: 24px;">{{ bot.avatar }}</span>
            </v-avatar>
            <div class="flex-grow-1" style="min-width: 0;">
              <div class="text-body-2 font-weight-semibold d-flex align-center ga-2">
                {{ bot.name }}
                <v-chip v-if="bot.isActive" size="x-small" color="primary" variant="tonal">{{ t('bot.active') }}</v-chip>
                <v-chip v-if="bot.running" size="x-small" color="success" variant="tonal">{{ t('bot.running') }}</v-chip>
              </div>
              <div class="text-caption text-medium-emphasis">{{ t('gateway.port') }} {{ bot.port }} · {{ bot.id.slice(0, 8) }}</div>
            </div>
            <div class="d-flex ga-1">
              <v-btn icon size="x-small" variant="text" @click.stop="openEdit(bot)">
                <v-icon size="16">mdi-pencil</v-icon>
              </v-btn>
              <v-btn v-if="!bot.isActive" icon size="x-small" variant="text" @click.stop="switchBot(bot.id)">
                <v-icon size="16">mdi-swap-horizontal</v-icon>
              </v-btn>
              <v-btn v-if="!bot.isActive" icon size="x-small" variant="text" color="error" @click.stop="confirmDelete(bot)">
                <v-icon size="16">mdi-close</v-icon>
              </v-btn>
            </div>
          </div>
        </v-card>
      </v-col>
    </v-row>

    <!-- Empty -->
    <div v-else class="text-center py-12">
      <div style="font-size: 48px;" class="mb-3">🤖</div>
      <p class="text-body-2 text-medium-emphasis">{{ t('bot.noBots') }}</p>
    </div>

    <!-- Create Dialog -->
    <v-dialog v-model="showCreateModal" max-width="400">
      <v-card rounded="xl" class="pa-6" color="#161625">
        <h3 class="text-h6 mb-4">{{ t('bot.createTitle') }}</h3>
        <v-text-field v-model="formName" :label="t('bot.botName')" variant="outlined" density="comfortable" @keyup.enter="createBot" />
        <div class="mb-4">
          <label class="text-caption text-medium-emphasis mb-2 d-block">{{ t('bot.avatar') }}</label>
          <div class="d-flex flex-wrap ga-2">
            <v-avatar
              v-for="av in avatars"
              :key="av"
              size="40"
              rounded="lg"
              :color="formAvatar === av ? '#6c5ce7' : '#1a1a30'"
              class="cursor-pointer"
              @click="formAvatar = av"
            >
              <span style="font-size: 20px;">{{ av }}</span>
            </v-avatar>
          </div>
        </div>
        <div class="d-flex justify-end ga-2">
          <v-btn variant="text" @click="showCreateModal = false">{{ t('bot.cancel') }}</v-btn>
          <v-btn color="primary" @click="createBot" :disabled="!formName.trim()">{{ t('bot.create') }}</v-btn>
        </div>
      </v-card>
    </v-dialog>

    <!-- Edit Dialog -->
    <v-dialog v-model="showEditModal" max-width="400">
      <v-card rounded="xl" class="pa-6" color="#161625">
        <h3 class="text-h6 mb-4">{{ t('bot.editTitle') }}</h3>
        <v-text-field v-model="formName" :label="t('bot.botName')" variant="outlined" density="comfortable" />
        <div class="mb-4">
          <label class="text-caption text-medium-emphasis mb-2 d-block">{{ t('bot.avatar') }}</label>
          <div class="d-flex flex-wrap ga-2">
            <v-avatar
              v-for="av in avatars"
              :key="av"
              size="40"
              rounded="lg"
              :color="formAvatar === av ? '#6c5ce7' : '#1a1a30'"
              class="cursor-pointer"
              @click="formAvatar = av"
            >
              <span style="font-size: 20px;">{{ av }}</span>
            </v-avatar>
          </div>
        </div>
        <div class="d-flex justify-end ga-2">
          <v-btn variant="text" @click="showEditModal = false">{{ t('bot.cancel') }}</v-btn>
          <v-btn color="primary" @click="saveEdit" :disabled="!formName.trim()">{{ t('bot.save') }}</v-btn>
        </div>
      </v-card>
    </v-dialog>

    <!-- Delete Dialog -->
    <v-dialog v-model="showDeleteModal" max-width="400">
      <v-card rounded="xl" class="pa-6" color="#161625">
        <h3 class="text-h6 mb-3">{{ t('bot.deleteTitle') }}</h3>
        <p class="text-body-2 mb-2">{{ t('bot.deleteConfirm', { name: deleteTarget?.name }) }}</p>
        <p class="text-caption text-medium-emphasis">{{ t('bot.deleteWarning') }}</p>
        <div class="d-flex justify-end ga-2 mt-4">
          <v-btn variant="text" @click="showDeleteModal = false">{{ t('bot.cancel') }}</v-btn>
          <v-btn color="error" @click="deleteBot">{{ t('bot.delete') }}</v-btn>
        </div>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ListBots, CreateBot, DeleteBot, SwitchBot, UpdateBot } from '../../wailsjs/go/main/App'

const { t } = useI18n()

interface BotInfo {
  id: string; name: string; avatar: string; port: number
  isActive: boolean; running: boolean; pid?: number
}

const bots = ref<BotInfo[]>([])
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const formName = ref('')
const formAvatar = ref('🐱')
const editTarget = ref<BotInfo | null>(null)
const deleteTarget = ref<BotInfo | null>(null)

const avatars = ['🐱', '🐶', '🦊', '🐻', '🐼', '🦄', '🐙', '🤖', '👾', '🧠', '⚡', '🔥']

async function loadBots() {
  try { bots.value = (await ListBots()) as any[] } catch {}
}

function openCreate() {
  formName.value = ''
  formAvatar.value = '🐱'
  showCreateModal.value = true
}

function openEdit(bot: BotInfo) {
  editTarget.value = bot
  formName.value = bot.name
  formAvatar.value = bot.avatar
  showEditModal.value = true
}

async function createBot() {
  if (!formName.value.trim()) return
  try {
    await CreateBot(formName.value.trim(), formAvatar.value)
    showCreateModal.value = false
    await loadBots()
  } catch (e) { alert(t('common.error') + ': ' + e) }
}

async function saveEdit() {
  if (!editTarget.value || !formName.value.trim()) return
  try {
    await UpdateBot(editTarget.value.id, formName.value.trim(), formAvatar.value)
    showEditModal.value = false
    editTarget.value = null
    await loadBots()
  } catch (e) { alert(t('common.error') + ': ' + e) }
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
    await loadBots()
  } catch (e) { alert(String(e)) }
}

async function switchBot(id: string) {
  try { await SwitchBot(id); await loadBots() } catch (e) { alert(String(e)) }
}

function botCardStyle(bot: BotInfo) {
  return {
    border: bot.isActive ? '1px solid #6c5ce7' : '1px solid #2a2a45',
    background: bot.isActive ? '#1a1a35' : '#161625',
  }
}

onMounted(loadBots)
</script>
