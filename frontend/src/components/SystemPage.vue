<template>
  <div>
    <div class="page-header">
      <h2>ℹ️ System</h2>
    </div>
    <div class="page-body">
      <div class="card">
        <div class="card-header"><span class="card-title">🖥️ Environment</span></div>
        <div style="display:flex; flex-direction:column; gap:6px">
          <div v-for="(value, key) in info" :key="key" style="display:flex; gap:12px">
            <span style="color:var(--text-secondary); min-width:120px; font-size:12px">{{ key }}</span>
            <span style="font-size:12px; font-family:monospace; user-select:text; -webkit-user-select:text">{{ value || '—' }}</span>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header"><span class="card-title">📁 Quick Actions</span></div>
        <div style="display:flex; gap:6px; flex-wrap:wrap">
          <button class="btn btn-secondary btn-sm" @click="openConfig">📄 Config File</button>
          <button class="btn btn-secondary btn-sm" @click="openWorkspace">📂 Workspace</button>
          <button class="btn btn-secondary btn-sm" @click="openSessions">📂 Sessions</button>
          <button class="btn btn-secondary btn-sm" @click="openLogs">📋 Log Files</button>
        </div>
      </div>

      <div class="card">
        <div class="card-header"><span class="card-title">🐈 About</span></div>
        <div style="font-size:13px; color:var(--text-secondary); line-height:1.6">
          Hey Nanobot — Personal AI Assistant<br/>
          Built with Go + Wails + Vue 3<br/>
          <span style="color:var(--text-muted); font-size:11px">v1.1.0 · {{ info.goVersion }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { OpenInFinder } from '../../wailsjs/go/main/App'
const props = defineProps<{ info: Record<string, string> }>()
const configPath = computed(() => props.info?.configPath || '')
function openConfig() { if (configPath.value) OpenInFinder(configPath.value) }
function openWorkspace() { if (configPath.value) OpenInFinder(configPath.value.replace('/config.json', '/workspace')) }
function openSessions() { if (configPath.value) OpenInFinder(configPath.value.replace('/config.json', '/workspace/sessions')) }
function openLogs() { if (configPath.value) OpenInFinder('/tmp/realtime_monitor_continuous.log') }
</script>
