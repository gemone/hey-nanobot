<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-4">
      <div class="d-flex align-center ga-2">
        <v-icon size="20" color="primary">mdi-folder-outline</v-icon>
        <span class="text-body-1 font-weight-bold">{{ t('sessions.title') }}</span>
        <v-chip v-if="sessions.length" size="x-small" variant="tonal" color="primary">{{ sessions.length }}</v-chip>
      </div>
      <v-text-field
        v-model="search"
        :placeholder="t('sessions.search')"
        variant="outlined"
        density="compact"
        hide-details
        prepend-inner-icon="mdi-magnify"
        clearable
        style="max-width: 220px;"
      />
    </div>

    <!-- Empty -->
    <div v-if="filteredSessions.length === 0" class="text-center py-16">
      <div style="font-size: 48px;" class="mb-3">📂</div>
      <p class="text-body-2" style="color: #5a5a78;">{{ search ? t('sessions.noResults') : t('sessions.noSessions') }}</p>
    </div>

    <!-- List -->
    <div v-else class="session-list">
      <div v-for="s in filteredSessions" :key="s.key" class="session-item card-base" @click="openInFinder(s.path)">
        <div class="d-flex align-center ga-3 flex-grow-1" style="min-width: 0;">
          <v-icon size="16" color="primary">mdi-file-document-outline</v-icon>
          <div style="min-width: 0;">
            <div class="text-body-2 text-truncate">{{ s.key }}</div>
            <div class="text-caption" style="color: #5a5a78;">{{ formatDate(s.updated_at || s.created_at) }}</div>
          </div>
        </div>
        <div class="d-flex ga-0">
          <v-btn icon size="x-small" variant="text" @click.stop="confirmDelete(s)" :title="t('sessions.deleteTitle')">
            <v-icon size="14" color="#ff6b6b">mdi-delete-outline</v-icon>
          </v-btn>
          <v-btn icon size="x-small" variant="text" @click.stop="openInFinder(s.path)" :title="t('sessions.showInFinder')">
            <v-icon size="14" color="#9898b0">mdi-open-in-new</v-icon>
          </v-btn>
        </div>
      </div>
    </div>

    <!-- Delete Dialog -->
    <v-dialog v-model="showDeleteModal" max-width="380">
      <v-card rounded="xl" class="pa-5" color="#161628" style="border: 1px solid #222240;">
        <h3 class="text-subtitle-1 font-weight-bold mb-2">{{ t('sessions.deleteTitle') }}</h3>
        <p class="text-body-2 mb-4">{{ t('sessions.deleteConfirm', { name: deleteTarget?.key }) }}</p>
        <div class="d-flex justify-end ga-2">
          <v-btn variant="text" size="small" @click="showDeleteModal = false">{{ t('bot.cancel') }}</v-btn>
          <v-btn color="error" size="small" @click="deleteSession">{{ t('bot.delete') }}</v-btn>
        </div>
      </v-card>
    </v-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { GetSessions, OpenInFinder, DeleteSession } from '../../wailsjs/go/main/App'

const { t } = useI18n()
interface SessionInfo { key: string; path: string; created_at: string; updated_at: string }

const sessions = ref<SessionInfo[]>([])
const search = ref('')
const showDeleteModal = ref(false)
const deleteTarget = ref<SessionInfo | null>(null)

const filteredSessions = computed(() => {
  if (!search.value) return sessions.value
  const q = search.value.toLowerCase()
  return sessions.value.filter(s => s.key.toLowerCase().includes(q))
})

function formatDate(d: string): string {
  if (!d) return ''
  try { return new Date(d).toLocaleString() } catch { return d }
}

async function loadSessions() { try { sessions.value = (await GetSessions()) as any[] } catch {} }
function openInFinder(path: string) { OpenInFinder(path) }
function confirmDelete(s: SessionInfo) { deleteTarget.value = s; showDeleteModal.value = true }

async function deleteSession() {
  if (!deleteTarget.value) return
  try { await DeleteSession(deleteTarget.value.path); showDeleteModal.value = false; deleteTarget.value = null; await loadSessions(); window.__notify?.(t('bot.delete') + ' ✓', 'info', 'mdi-delete') }
  catch (e) { window.__notify?.(String(e), 'error', 'mdi-alert-circle') }
}

onMounted(loadSessions)
</script>

<style scoped>
.session-list { display: flex; flex-direction: column; gap: 6px; }
.session-item { padding: 10px 14px; cursor: pointer; display: flex; align-items: center; }
.session-item:hover { background: #181832 !important; }
</style>
