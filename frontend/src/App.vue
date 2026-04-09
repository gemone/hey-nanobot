<template>
  <v-app style="background: #0f0f1a">
    <!-- Setup Wizard Overlay -->
    <SetupWizard v-if="showSetup" @done="onSetupDone" />

    <!-- Main layout (hidden during setup) -->
    <template v-if="!showSetup">
      <!-- Sidebar / Navigation Drawer -->
      <v-navigation-drawer
        v-model="drawer"
        :width="220"
        permanent
        :color="'#161625'"
        :border="true"
        class="app-sidebar"
        style="padding-top: 38px; -webkit-app-region: drag;"
      >
        <!-- Brand + Language Toggle -->
        <div class="d-flex align-center ga-2 pa-4 pb-3">
          <span style="font-size: 22px;">🐈</span>
          <span class="text-body-1 font-weight-bold">{{ t('app.title') }}</span>
          <v-spacer />
          <v-btn icon size="x-small" variant="text" @click="onToggleLocale" class="locale-btn">
            <span class="text-caption">{{ currentLocale === 'zh' ? 'EN' : '中' }}</span>
          </v-btn>
          <span class="text-caption text-medium-emphasis">v{{ appVersion }}</span>
        </div>

        <!-- Bot Switcher -->
        <div
          class="mx-2 mb-3 pa-2 d-flex align-center ga-2 rounded-lg cursor-pointer bot-switcher"
          style="background: #1e1e35; border: 1px solid #2a2a45; -webkit-app-region: no-drag;"
          @click="currentPage = 'bots'"
        >
          <v-avatar size="36" :color="'#1a1a30'" rounded="lg">
            <span style="font-size: 18px;">{{ activeBot.avatar || '🐱' }}</span>
          </v-avatar>
          <div class="flex-grow-1" style="min-width: 0;">
            <div class="text-body-2 font-weight-semibold text-truncate">{{ activeBot.name || t('common.loading') }}</div>
            <div class="d-flex align-center ga-1 text-caption text-medium-emphasis">
              <span class="status-dot" :class="{ running: gatewayRunning }"></span>
              {{ gatewayRunning ? t('common.online') : t('common.offline') }}
            </div>
          </div>
          <v-icon size="16" color="grey">mdi-chevron-right</v-icon>
        </div>

        <!-- Nav Items -->
        <v-list density="compact" nav class="px-2" style="-webkit-app-region: no-drag;">
          <v-list-item
            v-for="item in navItems"
            :key="item.id"
            :active="currentPage === item.id"
            @click="currentPage = item.id"
            :prepend-icon="item.mdiIcon"
            :title="t(item.labelKey)"
            rounded="lg"
            active-color="primary"
            :slim="true"
            class="mb-1"
          >
            <template v-slot:append v-if="item.badge">
              <v-chip size="x-small" :color="'error'" text-color="white" class="px-1">{{ item.badge }}</v-chip>
            </template>
          </v-list-item>
        </v-list>

        <template v-slot:append>
          <v-divider />
          <div class="pa-3 text-center text-caption text-medium-emphasis">
            Hey Nanobot v{{ appVersion }} · {{ t('app.multiBot') }}
          </div>
        </template>
      </v-navigation-drawer>

      <!-- Main Content -->
      <v-main style="padding-top: 38px !important;">
        <BotsPage v-if="currentPage === 'bots'" />
        <ChatPage v-else-if="currentPage === 'chat'" />
        <FeedPage v-else-if="currentPage === 'feed'" :gateway-running="gatewayRunning" />
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
        <SystemPage v-else-if="currentPage === 'system'" :info="systemInfo" :nanobot-info="nanobotInfo" />
      </v-main>
    </template>
  </v-app>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  GetConfig, SaveConfig, GetChannels, SetChannelField,
  GetProviders, SetProviderAPIKey, GetGatewayStatus,
  StartGateway, StopGateway, RestartGateway,
  GetSessions, GetSystemInfo,
  OpenInFinder,
  GetGatewayLogs, ClearGatewayLogs,
  GetActiveBot, ListBots, GetSetupState, GetNanobotInfo,
} from '../wailsjs/go/main/App'
import { EventsOn } from '../wailsjs/runtime/runtime'

import BotsPage from './components/BotsPage.vue'
import ChatPage from './components/ChatPage.vue'
import FeedPage from './components/FeedPage.vue'
import ChannelsPage from './components/ChannelsPage.vue'
import SessionsPage from './components/SessionsPage.vue'
import ProvidersPage from './components/ProvidersPage.vue'
import ConfigPage from './components/ConfigPage.vue'
import GatewayPage from './components/GatewayPage.vue'
import SystemPage from './components/SystemPage.vue'
import SetupWizard from './components/SetupWizard.vue'
import { toggleLocale, getLocale } from './i18n'

const { t } = useI18n()
const currentLocale = ref(getLocale())

const appVersion = '1.2.1'
const drawer = ref(true)
const currentPage = ref('chat')
const channelMsgCount = ref(0)
const showSetup = ref(false)

// ====== Bot State ======
const activeBot = ref<{ name: string; avatar: string; id: string }>({ name: '', avatar: '🐱', id: '' })

async function loadActiveBot() {
  try {
    const bot = await GetActiveBot()
    if (bot) activeBot.value = bot as any
  } catch {}
}

const navItems = computed(() => [
  { id: 'bots', labelKey: 'nav.bots', mdiIcon: 'mdi-robot-outline', badge: undefined },
  { id: 'chat', labelKey: 'nav.chat', mdiIcon: 'mdi-chat-outline', badge: undefined },
  { id: 'feed', labelKey: 'nav.feed', mdiIcon: 'mdi-broadcast', badge: channelMsgCount.value || undefined },
  { id: 'channels', labelKey: 'nav.channels', mdiIcon: 'mdi-link-variant', badge: undefined },
  { id: 'sessions', labelKey: 'nav.sessions', mdiIcon: 'mdi-folder-outline', badge: sessions.value.length || undefined },
  { id: 'providers', labelKey: 'nav.providers', mdiIcon: 'mdi-key-outline', badge: undefined },
  { id: 'config', labelKey: 'nav.config', mdiIcon: 'mdi-cog-outline', badge: undefined },
  { id: 'gateway', labelKey: 'nav.gateway', mdiIcon: 'mdi-web', badge: undefined },
  { id: 'system', labelKey: 'nav.system', mdiIcon: 'mdi-information-outline', badge: undefined },
])

// ====== Locale ======
function onToggleLocale() {
  const next = toggleLocale()
  currentLocale.value = next
}

// ====== Config ======
const configJson = ref('{}')
async function loadConfig() { try { configJson.value = await GetConfig() } catch {} }
async function saveConfig(json: string) {
  try { await SaveConfig(json); configJson.value = json }
  catch (e) { alert(t('common.saveFailed') + e) }
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
  catch (e) { alert(t('common.operationFailed') + e) }
}

// ====== Gateway ======
const gatewayStatus = ref<any>({ running: false })
const gatewayRunning = computed(() => gatewayStatus.value?.running || false)
const gatewayLogs = ref('')
async function loadGatewayStatus() { try { gatewayStatus.value = await GetGatewayStatus() } catch {} }
async function startGateway() { try { await StartGateway(); await loadGatewayStatus() } catch (e) { alert(t('common.error') + ': ' + e) } }
async function stopGateway() { try { await StopGateway(); await loadGatewayStatus() } catch (e) { alert(t('common.error') + ': ' + e) } }
async function restartGateway() { try { await RestartGateway(); await loadGatewayStatus() } catch (e) { alert(t('common.error') + ': ' + e) } }
async function clearLogs() { try { await ClearGatewayLogs(); gatewayLogs.value = '' } catch {} }

// ====== Sessions ======
const sessions = ref<any[]>([])

// ====== System ======
const systemInfo = ref<Record<string, string>>({})
const nanobotInfo = ref({ path: '', source: 'none', version: '', available: false })

async function loadNanobotInfo() {
  try {
    const info = await GetNanobotInfo() as any
    nanobotInfo.value = info
  } catch {}
}

function openInFinder(path: string) { OpenInFinder(path) }

function onSetupDone() {
  showSetup.value = false
  loadActiveBot()
  loadConfig()
  loadChannels()
  loadProviders()
  loadGatewayStatus()
}

// ====== Lifecycle ======
onMounted(async () => {
  // Check if setup is needed
  try {
    const state = await GetSetupState()
    if (state.needsSetup) {
      showSetup.value = true
      return
    }
  } catch {}

  await Promise.all([loadActiveBot(), loadConfig(), loadChannels(), loadProviders(), loadGatewayStatus(), loadNanobotInfo()])
  try { systemInfo.value = await GetSystemInfo() } catch {}

  // Event listeners
  EventsOn('navigate', (page: string) => { currentPage.value = page })
  EventsOn('gateway:status', (status: any) => { gatewayStatus.value = status })
  EventsOn('gateway:stdout', () => { refreshLogs() })
  EventsOn('gateway:stderr', () => { refreshLogs() })
  EventsOn('config:saved', () => { loadConfig(); loadChannels(); loadProviders() })
  EventsOn('sessions:updated', (s: any[]) => { sessions.value = s })
  EventsOn('channel:message', () => { channelMsgCount.value++ })
  EventsOn('channel:messages:cleared', () => { channelMsgCount.value = 0 })
  EventsOn('bot:switched', () => {
    loadActiveBot()
    loadConfig()
    loadChannels()
    loadProviders()
    loadGatewayStatus()
  })
  EventsOn('bots:updated', () => { loadActiveBot() })
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

<style>
/* Global overrides */
html, body {
  overflow: hidden;
  -webkit-user-select: none;
  user-select: none;
}

/* Scrollbar */
::-webkit-scrollbar { width: 5px; height: 5px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: #5a5a78; border-radius: 3px; }
::-webkit-scrollbar-thumb:hover { background: #9898b0; }

/* Status dot animation */
.status-dot {
  width: 7px; height: 7px; border-radius: 50%;
  background: #ff6b6b; flex-shrink: 0; display: inline-block;
}
.status-dot.running {
  background: #00cec9;
  box-shadow: 0 0 6px #00cec9;
  animation: pulse 2s ease-in-out infinite;
}
@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: .4; } }

/* Bot switcher hover */
.bot-switcher:hover {
  border-color: #6c5ce7 !important;
  background: #2a2a48 !important;
}

/* Locale toggle button */
.locale-btn {
  -webkit-app-region: no-drag;
}

/* Fix Vuetify nav drawer border */
.app-sidebar .v-navigation-drawer__border {
  background-color: #2a2a45 !important;
}

/* Page header style */
.page-header {
  padding: 12px 20px;
  border-bottom: 1px solid #2a2a45;
  display: flex;
  align-items: center;
  justify-content: space-between;
  background: #0f0f1a;
  -webkit-app-region: drag;
  gap: 12px;
}

/* Page body */
.page-body {
  flex: 1;
  overflow-y: auto;
  padding: 20px;
}
</style>
