<template>
  <div class="app-layout">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="sidebar-brand">
        <span class="emoji">🐈</span>
        <span>nanobot</span>
        <span class="version">v1.1</span>
      </div>

      <nav class="sidebar-nav">
        <div
          v-for="item in navItems"
          :key="item.id"
          class="nav-item"
          :class="{ active: currentPage === item.id }"
          @click="currentPage = item.id"
        >
          <span class="icon">{{ item.icon }}</span>
          <span>{{ item.label }}</span>
          <span v-if="item.badge" class="badge">{{ item.badge }}</span>
        </div>
      </nav>

      <div class="sidebar-footer">
        <div class="gateway-status" @click="currentPage = 'gateway'">
          <span class="status-dot" :class="{ running: gatewayRunning }"></span>
          <span>{{ gatewayRunning ? 'Online' : 'Offline' }}</span>
        </div>
      </div>
    </aside>

    <!-- Main -->
    <main class="main-content">
      <ChatPage v-if="currentPage === 'chat'" />
      <FeedPage v-else-if="currentPage === 'feed'" />
      <ChannelsPage
        v-else-if="currentPage === 'channels'"
        :channels="channels"
        @toggle-channel="toggleChannel"
        @update-field="updateChannelField"
      />
      <SessionsPage v-else-if="currentPage === 'sessions'" :sessions="sessions" @open-in-finder="openInFinder" />
      <ProvidersPage v-else-if="currentPage === 'providers'" :providers="providers" @set-key="setProviderKey" />
      <ConfigPage v-else-if="currentPage === 'config'" :config-json="configJson" @save="saveConfig" />
      <GatewayPage v-else-if="currentPage === 'gateway'" :status="gatewayStatus" :logs="gatewayLogs" @start="startGateway" @stop="stopGateway" @restart="restartGateway" @clear-logs="clearLogs" />
      <SystemPage v-else-if="currentPage === 'system'" :info="systemInfo" />
    </main>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import {
  GetConfig, SaveConfig, GetChannels, SetChannelField,
  GetProviders, SetProviderAPIKey, GetGatewayStatus,
  StartGateway, StopGateway, RestartGateway,
  GetSessions, GetSystemInfo, GetMessages,
  SendMessage, ClearMessages, OpenInFinder,
  GetGatewayLogs, ClearGatewayLogs, GetChannelMessages, ClearChannelMessages,
} from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

import ChatPage from './components/ChatPage.vue'
import FeedPage from './components/FeedPage.vue'
import ChannelsPage from './components/ChannelsPage.vue'
import SessionsPage from './components/SessionsPage.vue'
import ProvidersPage from './components/ProvidersPage.vue'
import ConfigPage from './components/ConfigPage.vue'
import GatewayPage from './components/GatewayPage.vue'
import SystemPage from './components/SystemPage.vue'

const currentPage = ref('chat')

const channelMsgCount = ref(0)
const navItems = computed(() => [
  { id: 'chat', icon: '💬', label: 'Chat' },
  { id: 'feed', icon: '📡', label: 'Live Feed', badge: channelMsgCount.value || undefined },
  { id: 'channels', icon: '🔗', label: 'Channels' },
  { id: 'sessions', icon: '📂', label: 'Sessions', badge: sessions.value.length || undefined },
  { id: 'providers', icon: '🔑', label: 'Providers' },
  { id: 'config', icon: '⚙️', label: 'Config' },
  { id: 'gateway', icon: '🌐', label: 'Gateway' },
  { id: 'system', icon: 'ℹ️', label: 'System' },
])

// ====== Config ======
const configJson = ref('{}')
async function loadConfig() { try { configJson.value = await GetConfig() } catch {} }
async function saveConfig(json: string) {
  try { await SaveConfig(json); configJson.value = json }
  catch (e) { alert('Save failed: ' + e) }
}

// ====== Channels ======
const channels = ref<Record<string, any>>({})
async function loadChannels() { try { channels.value = await GetChannels() } catch {} }
async function toggleChannel(name: string, enabled: boolean) {
  try { await SetChannelField(name, 'enabled', JSON.stringify(enabled)); await loadChannels() } catch {}
}
async function updateChannelField(channel: string, field: string, value: string) {
  try { await SetChannelField(channel, field, value); await loadChannels() } catch {}
}

// ====== Providers ======
const providers = ref<Record<string, any>>({})
async function loadProviders() { try { providers.value = await GetProviders() } catch {} }
async function setProviderKey(provider: string, key: string) {
  try { await SetProviderAPIKey(provider, key); await loadProviders() }
  catch (e) { alert('Failed: ' + e) }
}

// ====== Gateway ======
const gatewayStatus = ref<any>({ running: false })
const gatewayRunning = computed(() => gatewayStatus.value?.running || false)
const gatewayLogs = ref('')
async function loadGatewayStatus() { try { gatewayStatus.value = await GetGatewayStatus() } catch {} }
async function startGateway() { try { await StartGateway(); await loadGatewayStatus() } catch (e) { alert(e) } }
async function stopGateway() { try { await StopGateway(); await loadGatewayStatus() } catch (e) { alert(e) } }
async function restartGateway() { try { await RestartGateway(); await loadGatewayStatus() } catch (e) { alert(e) } }
async function clearLogs() { try { await ClearGatewayLogs(); gatewayLogs.value = '' } catch {} }

// ====== Sessions ======
const sessions = ref<any[]>([])

// ====== System ======
const systemInfo = ref<Record<string, string>>({})

function openInFinder(path: string) { OpenInFinder(path) }

// ====== Lifecycle ======
onMounted(async () => {
  await Promise.all([loadConfig(), loadChannels(), loadProviders(), loadGatewayStatus()])
  try { systemInfo.value = await GetSystemInfo() } catch {}

  // Event listeners
  EventsOn('navigate', (page: string) => { currentPage.value = page })
  EventsOn('gateway:status', (status: any) => { gatewayStatus.value = status })
  EventsOn('gateway:stdout', () => { refreshLogs() })
  EventsOn('gateway:stderr', () => { refreshLogs() })
  EventsOn('config:saved', () => { loadConfig(); loadChannels(); loadProviders() })
  EventsOn('sessions:updated', (s: any[]) => { sessions.value = s })
  EventsOn('channel:message', () => {
    channelMsgCount.value++
    if (currentPage.value !== 'feed') {
      // Could show desktop notification here
    }
  })
  EventsOn('channel:messages:cleared', () => { channelMsgCount.value = 0 })
})

let logThrottle: any = null
function refreshLogs() {
  if (logThrottle) return
  logThrottle = setTimeout(async () => {
    logThrottle = null
    try { gatewayLogs.value = await GetGatewayLogs() } catch {}
  }, 200)
}
</script>
