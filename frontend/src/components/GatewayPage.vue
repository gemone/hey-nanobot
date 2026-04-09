<template>
  <div class="d-flex flex-column" style="height: 100%;">
    <div class="page-header">
      <div class="d-flex align-center ga-2">
        <v-icon size="20" color="primary">mdi-web</v-icon>
        <span class="text-body-1 font-weight-bold">{{ t('gateway.title') }}</span>
        <span class="status-dot" :class="{ running: status.running }" style="margin-left: 4px;"></span>
      </div>
      <div class="actions d-flex ga-2">
        <v-btn v-if="!status.running" color="success" size="small" variant="tonal" prepend-icon="mdi-play" @click="$emit('start')">{{ t('gateway.start') }}</v-btn>
        <v-btn v-if="status.running" color="error" size="small" variant="tonal" prepend-icon="mdi-stop" @click="$emit('stop')">{{ t('gateway.stop') }}</v-btn>
        <v-btn v-if="status.running" size="small" variant="outlined" prepend-icon="mdi-refresh" @click="$emit('restart')">{{ t('gateway.restart') }}</v-btn>
      </div>
    </div>

    <!-- Stats -->
    <div class="pa-4 pb-2">
      <div class="stats-grid">
        <div class="stat-card">
          <div class="stat-label">{{ t('gateway.pid') }}</div>
          <div class="stat-value">{{ status.pid || '—' }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">{{ t('gateway.port') }}</div>
          <div class="stat-value">{{ status.port || '—' }}</div>
        </div>
        <div class="stat-card">
          <div class="stat-label">{{ t('gateway.uptime') }}</div>
          <div class="stat-value">{{ status.uptime || '—' }}</div>
        </div>
        <div class="stat-card" :class="{ 'stat-active': status.running }">
          <div class="stat-label">{{ t('gateway.status') }}</div>
          <div class="stat-value" :style="{ color: status.running ? '#00cec9' : '#ff6b6b' }">
            {{ status.running ? t('common.online') : t('common.offline') }}
          </div>
        </div>
      </div>
    </div>

    <!-- Logs -->
    <div class="px-4 pb-2 d-flex align-center justify-space-between">
      <span class="text-caption font-weight-semibold" style="text-transform: uppercase; letter-spacing: 0.5px; color: #5a5a78;">{{ t('gateway.logs') }}</span>
      <v-btn variant="text" size="x-small" prepend-icon="mdi-delete-sweep" @click="$emit('clear-logs')" style="color: #5a5a78;">{{ t('gateway.clearLogs') }}</v-btn>
    </div>
    <div class="log-view flex-grow-1 mx-4 mb-4 pa-3 rounded-lg">
      <pre>{{ logs || t('gateway.noLogs') }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()
defineProps<{ status: any; logs: string }>()
defineEmits(['start', 'stop', 'restart', 'clear-logs'])
</script>

<style scoped>
.stats-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
}
.stat-card {
  background: rgb(var(--v-theme-surface));
  border: 1px solid rgba(128,128,128,0.12);
  border-radius: 8px;
  padding: 12px;
  text-align: center;
}
.stat-card.stat-active {
  border-color: rgb(var(--v-theme-success));
  background: rgba(var(--v-theme-success), 0.06);
}
.stat-label { font-size: 10px; color: #5a5a78; text-transform: uppercase; letter-spacing: 0.5px; margin-bottom: 4px; }
.stat-value { font-size: 16px; font-weight: 700; }
.log-view {
  background: rgb(var(--v-theme-background));
  border: 1px solid rgba(128,128,128,0.1);
  overflow-y: auto;
  font-family: 'SF Mono', 'Menlo', monospace;
  font-size: 11px;
  line-height: 1.5;
  opacity: 0.7;
}
.log-view pre { white-space: pre-wrap; word-break: break-all; margin: 0; }
</style>
