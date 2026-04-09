<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">{{ t('sessions.title') }}</h2>
      <v-text-field
        v-model="search"
        :placeholder="t('sessions.search')"
        variant="outlined"
        density="compact"
        hide-details
        prepend-inner-icon="mdi-magnify"
        clearable
        style="max-width: 260px;"
      />
    </div>
    <div v-if="filteredSessions.length === 0" class="text-center py-12">
      <div style="font-size: 48px;" class="mb-3">📂</div>
      <p class="text-body-2 text-medium-emphasis">{{ search ? t('sessions.noResults') : t('sessions.noSessions') }}</p>
    </div>
    <v-list v-else bg-color="transparent" density="compact" rounded="lg" style="border: 1px solid #2a2a45;">
      <v-list-item
        v-for="s in filteredSessions"
        :key="s.key"
        @click="openInFinder(s.path)"
      >
        <template v-slot:prepend>
          <v-icon size="16" color="primary">mdi-file-document-outline</v-icon>
        </template>
        <v-list-item-title class="text-body-2">{{ s.key }}</v-list-item-title>
        <v-list-item-subtitle class="text-caption text-medium-emphasis">
          {{ s.updated_at || s.created_at }}
        </v-list-item-subtitle>
        <template v-slot:append>
          <v-btn icon size="x-small" variant="text" @click.stop="confirmDelete(s)">
            <v-icon size="14" color="error">mdi-delete-outline</v-icon>
          </v-btn>
          <v-btn icon size="x-small" variant="text">
            <v-icon size="14">mdi-open-in-new</v-icon>
          </v-btn>
        </template>
      </v-list-item>
    </v-list>

    <!-- Delete Dialog -->
    <v-dialog v-model="showDeleteModal" max-width="400">
      <v-card rounded="xl" class="pa-6" color="#161625">
        <h3 class="text-h6 mb-3">{{ t('sessions.deleteTitle') }}</h3>
        <p class="text-body-2 mb-4">{{ t('sessions.deleteConfirm', { name: deleteTarget?.key }) }}</p>
        <div class="d-flex justify-end ga-2">
          <v-btn variant="text" @click="showDeleteModal = false">{{ t('bot.cancel') }}</v-btn>
          <v-btn color="error" @click="deleteSession">{{ t('bot.delete') }}</v-btn>
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

interface SessionInfo {
  key: string; path: string; created_at: string; updated_at: string
}

const sessions = ref<SessionInfo[]>([])
const search = ref('')
const showDeleteModal = ref(false)
const deleteTarget = ref<SessionInfo | null>(null)

const filteredSessions = computed(() => {
  if (!search.value) return sessions.value
  const q = search.value.toLowerCase()
  return sessions.value.filter(s => s.key.toLowerCase().includes(q))
})

async function loadSessions() {
  try { sessions.value = (await GetSessions()) as any[] } catch {}
}

function openInFinder(path: string) {
  OpenInFinder(path)
}

function confirmDelete(s: SessionInfo) {
  deleteTarget.value = s
  showDeleteModal.value = true
}

async function deleteSession() {
  if (!deleteTarget.value) return
  try {
    await DeleteSession(deleteTarget.value.path)
    showDeleteModal.value = false
    deleteTarget.value = null
    await loadSessions()
  } catch (e) {
    alert(t('common.error') + ': ' + e)
  }
}

onMounted(loadSessions)
</script>
