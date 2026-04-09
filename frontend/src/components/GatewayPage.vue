<template>
  <div class="d-flex flex-column" style="height: 100%;">
    <div class="page-header">
      <h2 class="text-body-1 font-weight-bold">🌐 Gateway</h2>
      <div class="d-flex ga-2">
        <v-btn v-if="!status.running" color="success" size="small" prepend-icon="mdi-play" @click="$emit('start')">Start</v-btn>
        <v-btn v-if="status.running" color="error" size="small" prepend-icon="mdi-stop" @click="$emit('stop')">Stop</v-btn>
        <v-btn v-if="status.running" size="small" variant="outlined" prepend-icon="mdi-refresh" @click="$emit('restart')">Restart</v-btn>
      </div>
    </div>

    <!-- Stats -->
    <div class="pa-4 pb-2">
      <v-row>
        <v-col cols="3">
          <v-card rounded="lg" class="pa-3 text-center" color="#1a1a30" style="border: 1px solid #2a2a45;">
            <div class="text-caption text-medium-emphasis mb-1">PID</div>
            <div class="text-h6 font-weight-bold">{{ status.pid || '—' }}</div>
          </v-card>
        </v-col>
        <v-col cols="3">
          <v-card rounded="lg" class="pa-3 text-center" color="#1a1a30" style="border: 1px solid #2a2a45;">
            <div class="text-caption text-medium-emphasis mb-1">Port</div>
            <div class="text-h6 font-weight-bold">{{ status.port || '—' }}</div>
          </v-card>
        </v-col>
        <v-col cols="3">
          <v-card rounded="lg" class="pa-3 text-center" color="#1a1a30" style="border: 1px solid #2a2a45;">
            <div class="text-caption text-medium-emphasis mb-1">Uptime</div>
            <div class="text-h6 font-weight-bold">{{ status.uptime || '—' }}</div>
          </v-card>
        </v-col>
        <v-col cols="3">
          <v-card rounded="lg" class="pa-3 text-center" :color="status.running ? 'rgba(0,206,201,0.1)' : '#1a1a30'" style="border: 1px solid #2a2a45;">
            <div class="text-caption text-medium-emphasis mb-1">Status</div>
            <div class="text-h6 font-weight-bold" :color="status.running ? 'success' : 'error'">
              {{ status.running ? 'Online' : 'Offline' }}
            </div>
          </v-card>
        </v-col>
      </v-row>
    </div>

    <!-- Logs -->
    <div class="px-4 pb-1 d-flex align-center justify-space-between">
      <span class="text-caption text-medium-emphasis">Logs</span>
      <v-btn variant="text" size="x-small" prepend-icon="mdi-delete-sweep" @click="$emit('clear-logs')">Clear</v-btn>
    </div>
    <div class="flex-grow-1 mx-4 mb-4 pa-3 rounded-lg overflow-y-auto" style="background: #0f0f1a; border: 1px solid #2a2a45; font-family: 'SF Mono', monospace; font-size: 11px; line-height: 1.5;">
      <pre style="white-space: pre-wrap; word-break: break-all; color: #9898b0;">{{ logs || 'No logs yet.' }}</pre>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ status: any; logs: string }>()
defineEmits(['start', 'stop', 'restart', 'clear-logs'])
</script>
