<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-4">
      <div class="d-flex align-center ga-2">
        <v-icon size="20" color="primary">mdi-robot-outline</v-icon>
        <span class="text-body-1 font-weight-bold">{{ t('bot.manager') }}</span>
      </div>
      <v-btn color="primary" size="small" variant="tonal" prepend-icon="mdi-plus" @click="openCreate">{{ t('bot.newBot') }}</v-btn>
    </div>

    <!-- Bot Grid -->
    <div v-if="bots.length" class="bots-grid">
      <div v-for="bot in bots" :key="bot.id" class="card-base pa-4" :class="{ 'bot-active': bot.isActive }">
        <div class="d-flex align-center ga-3">
          <v-avatar size="40" rounded="lg" :color="bot.isActive ? '#1a1a40' : '#1a1a30'">
            <span style="font-size: 22px;">{{ bot.avatar }}</span>
          </v-avatar>
          <div class="flex-grow-1" style="min-width: 0;">
            <div class="d-flex align-center ga-2">
              <span class="text-body-2 font-weight-semibold">{{ bot.name }}</span>
              <v-chip v-if="bot.isActive" size="x-small" color="primary" variant="tonal">{{ t('bot.active') }}</v-chip>
            </div>
            <div class="d-flex align-center ga-2 mt-1" style="font-size: 10px; color: #5a5a78;">
              <span>{{ t('gateway.port') }} {{ bot.port }}</span>
              <span>·</span>
              <span>{{ bot.id.slice(0, 8) }}</span>
              <span v-if="bot.running">
                <span>·</span>
                <span style="color: #00cec9;">{{ t('bot.running') }}</span>
              </span>
            </div>
          </div>
          <div class="d-flex ga-0">
            <v-btn icon size="x-small" variant="text" @click.stop="openEdit(bot)" :title="t('bot.editTitle')">
              <v-icon size="15" color="#9898b0">mdi-pencil-outline</v-icon>
            </v-btn>
            <v-btn v-if="!bot.isActive" icon size="x-small" variant="text" @click.stop="switchBot(bot.id)" :title="t('bot.switch')">
              <v-icon size="15" color="#9898b0">mdi-swap-horizontal</v-icon>
            </v-btn>
            <v-btn v-if="!bot.isActive" icon size="x-small" variant="text" @click.stop="confirmDelete(bot)" :title="t('bot.delete')">
              <v-icon size="15" color="#ff6b6b">mdi-close</v-icon>
            </v-btn>
          </div>
        </div>
      </div>
    </div>

    <!-- Empty -->
    <div v-else class="text-center py-16">
      <div style="font-size: 48px;" class="mb-3">🤖</div>
      <p class="text-body-2" style="color: #5a5a78;">{{ t('bot.noBots') }}</p>
    </div>

    <!-- Create Dialog -->
    <v-dialog v-model="showCreateModal" max-width="380">
      <v-card rounded="xl" class="pa-5" color="#161628" style="border: 1px solid #222240;">
        <h3 class="text-subtitle-1 font-weight-bold mb-4">{{ t('bot.createTitle') }}</h3>
        <v-text-field v-model="formName" :label="t('bot.botName')" variant="outlined" density="compact" @keyup.enter="createBot" />
        <div class="mb-3">
          <label class="text-caption" style="color: #5a5a78;">{{ t('bot.avatar') }}</label>
          <div class="d-flex flex-wrap ga-2 mt-2">
            <v-avatar v-for="av in avatars" :key="av" size="34" rounded="lg"
              :color="formAvatar === av ? '#6c5ce7' : '#1a1a30'" class="cursor-pointer"
              @click="formAvatar = av"
            ><span style="font-size: 18px;">{{ av }}</span></v-avatar>
          </div>
        </div>
        <div class="d-flex justify-end ga-2">
          <v-btn variant="text" size="small" @click="showCreateModal = false">{{ t('bot.cancel') }}</v-btn>
          <v-btn color="primary" size="small" @click="createBot" :disabled="!formName.trim()">{{ t('bot.create') }}</v-btn>
        </div>
      </v-card>
    </v-dialog>

    <!-- Edit Dialog -->
    <v-dialog v-model="showEditModal" max-width="380">
      <v-card rounded="xl" class="pa-5" color="#161628" style="border: 1px solid #222240;">
        <h3 class="text-subtitle-1 font-weight-bold mb-4">{{ t('bot.editTitle') }}</h3>
        <v-text-field v-model="formName" :label="t('bot.botName')" variant="outlined" density="compact" />
        <div class="mb-3">
          <label class="text-caption" style="color: #5a5a78;">{{ t('bot.avatar') }}</label>
          <div class="d-flex flex-wrap ga-2 mt-2">
            <v-avatar v-for="av in avatars" :key="av" size="34" rounded="lg"
              :color="formAvatar === av ? '#6c5ce7' : '#1a1a30'" class="cursor-pointer"
              @click="formAvatar = av"
            ><span style="font-size: 18px;">{{ av }}</span></v-avatar>
          </div>
        </div>
        <div class="d-flex justify-end ga-2">
          <v-btn variant="text" size="small" @click="showEditModal = false">{{ t('bot.cancel') }}</v-btn>
          <v-btn color="primary" size="small" @click="saveEdit" :disabled="!formName.trim()">{{ t('bot.save') }}</v-btn>
        </div>
      </v-card>
    </v-dialog>

    <!-- Delete Dialog -->
    <v-dialog v-model="showDeleteModal" max-width="380">
      <v-card rounded="xl" class="pa-5" color="#161628" style="border: 1px solid #222240;">
        <h3 class="text-subtitle-1 font-weight-bold mb-2">{{ t('bot.deleteTitle') }}</h3>
        <p class="text-body-2 mb-1">{{ t('bot.deleteConfirm', { name: deleteTarget?.name }) }}</p>
        <p class="text-caption" style="color: #5a5a78;">{{ t('bot.deleteWarning') }}</p>
        <div class="d-flex justify-end ga-2 mt-4">
          <v-btn variant="text" size="small" @click="showDeleteModal = false">{{ t('bot.cancel') }}</v-btn>
          <v-btn color="error" size="small" @click="deleteBot">{{ t('bot.delete') }}</v-btn>
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

interface BotInfo { id: string; name: string; avatar: string; port: number; isActive: boolean; running: boolean; pid?: number }

const bots = ref<BotInfo[]>([])
const showCreateModal = ref(false)
const showEditModal = ref(false)
const showDeleteModal = ref(false)
const formName = ref('')
const formAvatar = ref('🐱')
const editTarget = ref<BotInfo | null>(null)
const deleteTarget = ref<BotInfo | null>(null)
const avatars = ['🐱', '🐶', '🦊', '🐻', '🐼', '🦄', '🐙', '🤖', '👾', '🧠', '⚡', '🔥']

async function loadBots() { try { bots.value = (await ListBots()) as any[] } catch {} }
function openCreate() { formName.value = ''; formAvatar.value = '🐱'; showCreateModal.value = true }
function openEdit(bot: BotInfo) { editTarget.value = bot; formName.value = bot.name; formAvatar.value = bot.avatar; showEditModal.value = true }

async function createBot() {
  if (!formName.value.trim()) return
  try { await CreateBot(formName.value.trim(), formAvatar.value); showCreateModal.value = false; await loadBots(); window.__notify?.(t('bot.create') + ' ✓') }
  catch (e) { window.__notify?.(String(e), 'error', 'mdi-alert-circle') }
}
async function saveEdit() {
  if (!editTarget.value || !formName.value.trim()) return
  try { await UpdateBot(editTarget.value.id, formName.value.trim(), formAvatar.value); showEditModal.value = false; editTarget.value = null; await loadBots(); window.__notify?.(t('bot.save') + ' ✓') }
  catch (e) { window.__notify?.(String(e), 'error', 'mdi-alert-circle') }
}
function confirmDelete(bot: BotInfo) { deleteTarget.value = bot; showDeleteModal.value = true }
async function deleteBot() {
  if (!deleteTarget.value) return
  try { await DeleteBot(deleteTarget.value.id); showDeleteModal.value = false; deleteTarget.value = null; await loadBots(); window.__notify?.(t('bot.delete') + ' ✓', 'info', 'mdi-delete') }
  catch (e) { window.__notify?.(String(e), 'error', 'mdi-alert-circle') }
}
async function switchBot(id: string) {
  try { await SwitchBot(id); await loadBots(); window.__notify?.(t('bot.switch') + ' ✓') }
  catch (e) { window.__notify?.(String(e), 'error', 'mdi-alert-circle') }
}

onMounted(loadBots)
</script>

<style scoped>
.bots-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 10px; }
.bot-active { border-color: rgba(108,92,231,0.4) !important; background: #161630 !important; }
</style>
