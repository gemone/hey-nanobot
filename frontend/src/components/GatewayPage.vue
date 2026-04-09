<template>
  <div>
    <div class="page-header">
      <h2>🌐 Gateway</h2>
      <div class="actions">
        <button v-if="!status?.running" class="btn btn-success btn-sm" @click="$emit('start')">▶ Start</button>
        <button v-else class="btn btn-danger btn-sm" @click="$emit('stop')">⏹ Stop</button>
        <button class="btn btn-secondary btn-sm" @click="$emit('restart')" :disabled="!status?.running">🔄 Restart</button>
      </div>
    </div>
    <div class="page-body">
      <!-- Stats -->
      <div class="stats-row">
        <div class="stat-card">
          <div class="stat-value" :style="{ color: status?.running ? 'var(--green)' : 'var(--red)' }">
            {{ status?.running ? '● Online' : '● Offline' }}
          </div>
          <div class="stat-label">Status</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ status?.pid || '—' }}</div>
          <div class="stat-label">PID</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ status?.port || '—' }}</div>
          <div class="stat-label">Port</div>
        </div>
        <div class="stat-card">
          <div class="stat-value">{{ status?.uptime || '—' }}</div>
          <div class="stat-label">Uptime</div>
        </div>
      </div>

      <!-- Logs -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">📋 Logs</span>
          <button class="btn btn-ghost btn-xs" @click="$emit('clear-logs')">Clear</button>
        </div>
        <div class="log-view" ref="logEl">{{ logs || 'No logs yet...' }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
defineProps<{ status: any; logs: string }>()
defineEmits(['start', 'stop', 'restart', 'clear-logs'])
const logEl = ref<HTMLDivElement>()
watch(() => arguments, () => {
  nextTick(() => { if (logEl.value) logEl.value.scrollTop = logEl.value.scrollHeight })
})
</script>
