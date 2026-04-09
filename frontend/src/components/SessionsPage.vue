<template>
  <div>
    <div class="page-header">
      <h2>📂 Sessions</h2>
    </div>
    <div class="page-body">
      <div v-if="sessions.length === 0" class="empty-state">
        <span class="emoji">📂</span>
        <span class="text">No sessions found</span>
      </div>
      <div class="session-list">
        <div v-for="session in sessions" :key="session.key" class="session-item"
          @click="$emit('open-in-finder', session.path)">
          <div>
            <div class="s-key">{{ session.key }}</div>
            <div class="s-time">{{ formatTime(session.updated_at) }}</div>
          </div>
          <span style="color:var(--text-muted)">📁</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { GetSessions } from '../../wailsjs/go/main/App'
defineProps<{ sessions: any[] }>()
defineEmits(['open-in-finder'])
function refresh() { GetSessions() }
function formatTime(iso: string): string {
  if (!iso) return ''
  try { return new Date(iso).toLocaleString() } catch { return iso }
}
</script>
