<template>
  <v-app>
    <!-- Setup Wizard Overlay -->
    <SetupWizard v-if="showSetup" @done="onSetupDone" />

    <!-- Main layout (hidden during setup) -->
    <template v-if="!showSetup">
      <!-- Sidebar / Navigation Drawer -->
      <v-navigation-drawer
        v-model="drawer"
        :width="210"
        permanent
        color="surface"
        :border="true"
        class="app-sidebar"
        style="padding-top: 40px; -webkit-app-region: drag;"
      >
        <!-- Brand -->
        <div class="d-flex align-center ga-2 px-4 pb-2 pt-1">
          <span style="font-size: 20px;">🐈</span>
          <span class="text-subtitle-2 font-weight-bold" style="letter-spacing: -0.3px;">Hey Nanobot</span>
          <v-spacer />
          <v-btn icon size="x-small" variant="text" @click="onToggleTheme" class="locale-btn" :title="isDark ? t('common.lightMode') : t('common.darkMode')">
            <v-icon size="16">{{ isDark ? 'mdi-weather-sunny' : 'mdi-weather-night' }}</v-icon>
          </v-btn>
          <v-btn icon size="x-small" variant="text" @click="onToggleLocale" class="locale-btn" :title="currentLocale === 'zh' ? 'Switch to English' : '切换到中文'">
            <span class="text-caption font-weight-medium">{{ currentLocale === 'zh' ? 'EN' : '中' }}</span>
          </v-btn>
        </div>

        <!-- Bot Switcher -->
        <div
          class="mx-3 my-2 pa-2 d-flex align-center ga-2 rounded-lg bot-switcher"
          @click="currentPage = 'bots'"
        >
          <v-avatar size="32" rounded="lg" color="surface-variant">
            <span style="font-size: 16px;">{{ activeBot.avatar || '🐱' }}</span>
          </v-avatar>
          <div class="flex-grow-1" style="min-width: 0;">
            <div class="text-caption font-weight-semibold text-truncate">{{ activeBot.name || t('common.loading') }}</div>
            <div class="d-flex align-center ga-1" style="font-size: 10px; opacity: 0.5;">
              <span class="status-dot" :class="{ running: gatewayRunning }"></span>
              {{ gatewayRunning ? t('common.online') : t('common.offline') }}
            </div>
          </div>
          <v-icon size="14" style="opacity: 0.4;">mdi-chevron-right</v-icon>
        </div>

        <!-- Nav Items -->
        <v-list density="compact" nav class="px-2 py-0" style="-webkit-app-region: no-drag;">
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
            class="nav-item-custom mb-0"
          >
            <template v-slot:append v-if="item.badge">
              <span class="nav-badge">{{ item.badge }}</span>
            </template>
          </v-list-item>
        </v-list>

        <template v-slot:append>
          <v-divider />
          <div class="pa-3 text-center" style="font-size: 10px; opacity: 0.3;">
            v{{ appVersion }}
          </div>
        </template>
      </v-navigation-drawer>

      <!-- Main Content -->
      <v-main style="padding-top: 40px !important;">
        <BotsPage v-if="currentPage === 'bots'" />
        <ChatPage v-else-if="currentPage === 'chat'" />
        <FeedPage v-else-if="currentPage === 'feed'" :gateway-running="gatewayRunning" />
        <ChannelsPage v-else-if="currentPage === 'channels'" />
        <SessionsPage v-else-if="currentPage === 'sessions'" />
        <ProvidersPage v-else-if="currentPage === 'providers'" />
        <ConfigPage v-else-if="currentPage === 'config'" :config-json="configJson" @save="saveConfig" />
        <GatewayPage v-else-if="currentPage === 'gateway'" :status="gatewayStatus" :logs="gatewayLogs" @start="startGateway" @stop="stopGateway" @restart="restartGateway" @clear-logs="clearLogs" />
        <SystemPage v-else-if="currentPage === 'system'" :info="systemInfo" :nanobot-info="nanobotInfo" />
      </v-main>
    </template>

    <!-- Global Snackbar -->
    <v-snackbar
      v-model="snackbar.show"
      :color="snackbar.color"
      :timeout="snackbar.timeout"
      location="top right"
      rounded="lg"
      density="compact"
    >
      <div class="d-flex align-center ga-2">
        <v-icon size="16">{{ snackbar.icon }}</v-icon>
        <span style="font-size: 13px;">{{ snackbar.text }}</span>
      </div>
    </v-snackbar>
  </v-app>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, reactive } from 'vue'
import { useI18n } from 'vue-i18n'
import { useTheme } from 'vuetify'
import {
  GetConfig, SaveConfig,
  GetGatewayStatus,
  StartGateway, StopGateway, RestartGateway,
  GetSystemInfo,
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
const theme = useTheme()
const currentLocale = ref(getLocale())

// ====== Theme Toggle ======
const isDark = computed(() => theme.global.current.value.dark)

function onToggleTheme() {
  const next = isDark.value ? 'light' : 'dark'
  theme.global.name.value = next
  localStorage.setItem('hey-nanobot-theme', next)
}

// Restore saved theme
const savedTheme = localStorage.getItem('hey-nanobot-theme')
if (savedTheme === 'light' || savedTheme === 'dark') {
  theme.global.name.value = savedTheme
}

const appVersion = '1.3.0'
const drawer = ref(true)
const currentPage = ref('chat')
const channelMsgCount = ref(0)
const showSetup = ref(false)

// ====== Global Snackbar ======
const snackbar = reactive({ show: false, text: '', color: 'success', icon: 'mdi-check-circle', timeout: 2500 })

function notify(text: string, color = 'success', icon = 'mdi-check-circle') {
  snackbar.text = text
  snackbar.color = color
  snackbar.icon = icon
  snackbar.show = true
}
window.__notify = notify

// ====== Bot State ======
const activeBot = ref<{ name: string; avatar: string; id: string }>({ name: '', avatar: '🐱', id: '' })

async function loadActiveBot() {
  try {
    const bot = await GetActiveBot()
    if (bot) activeBot.value = bot as any
  } catch {}
}

const navItems = computed(() => [
  { id: 'chat', labelKey: 'nav.chat', mdiIcon: 'mdi-chat-outline', badge: undefined },
  { id: 'feed', labelKey: 'nav.feed', mdiIcon: 'mdi-broadcast', badge: channelMsgCount.value || undefined },
  { id: 'bots', labelKey: 'nav.bots', mdiIcon: 'mdi-robot-outline', badge: undefined },
  { id: 'providers', labelKey: 'nav.providers', mdiIcon: 'mdi-key-outline', badge: undefined },
  { id: 'channels', labelKey: 'nav.channels', mdiIcon: 'mdi-link-variant', badge: undefined },
  { id: 'gateway', labelKey: 'nav.gateway', mdiIcon: 'mdi-web', badge: undefined },
  { id: 'sessions', labelKey: 'nav.sessions', mdiIcon: 'mdi-folder-outline', badge: sessions.value.length || undefined },
  { id: 'config', labelKey: 'nav.config', mdiIcon: 'mdi-cog-outline', badge: undefined },
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
  try { await SaveConfig(json); configJson.value = json; notify(t('common.success')) }
  catch (e) { notify(t('common.saveFailed') + e, 'error', 'mdi-alert-circle') }
}

// ====== Gateway ======
const gatewayStatus = ref<any>({ running: false })
const gatewayRunning = computed(() => gatewayStatus.value?.running || false)
const gatewayLogs = ref('')
async function loadGatewayStatus() { try { gatewayStatus.value = await GetGatewayStatus() } catch {} }
async function startGateway() { try { await StartGateway(); await loadGatewayStatus(); notify(t('gateway.start')) } catch (e) { notify(String(e), 'error', 'mdi-alert-circle') } }
async function stopGateway() { try { await StopGateway(); await loadGatewayStatus(); notify(t('gateway.stop'), 'info', 'mdi-stop') } catch (e) { notify(String(e), 'error', 'mdi-alert-circle') } }
async function restartGateway() { try { await RestartGateway(); await loadGatewayStatus(); notify(t('gateway.restart')) } catch (e) { notify(String(e), 'error', 'mdi-alert-circle') } }
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

function onSetupDone() {
  showSetup.value = false
  loadActiveBot()
  loadConfig()
  loadGatewayStatus()
}

// ====== Lifecycle ======
onMounted(async () => {
  try {
    const state = await GetSetupState()
    if (state.needsSetup) {
      showSetup.value = true
      return
    }
  } catch {}

  await Promise.all([loadActiveBot(), loadConfig(), loadGatewayStatus(), loadNanobotInfo()])
  try { systemInfo.value = await GetSystemInfo() } catch {}

  EventsOn('navigate', (page: string) => { currentPage.value = page })
  EventsOn('gateway:status', (status: any) => { gatewayStatus.value = status })
  EventsOn('gateway:stdout', () => { refreshLogs() })
  EventsOn('gateway:stderr', () => { refreshLogs() })
  EventsOn('config:saved', () => { loadConfig() })
  EventsOn('sessions:updated', (s: any[]) => { sessions.value = s })
  EventsOn('channel:message', () => { channelMsgCount.value++ })
  EventsOn('channel:messages:cleared', () => { channelMsgCount.value = 0 })
  EventsOn('bot:switched', () => {
    loadActiveBot()
    loadConfig()
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
/* ====== Global ====== */
html, body { overflow: hidden; -webkit-user-select: none; user-select: none; }

/* Scrollbar */
::-webkit-scrollbar { width: 4px; height: 4px; }
::-webkit-scrollbar-track { background: transparent; }
::-webkit-scrollbar-thumb { background: rgba(128,128,128,0.3); border-radius: 2px; }
::-webkit-scrollbar-thumb:hover { background: rgba(128,128,128,0.5); }

/* Status dot */
.status-dot {
  width: 6px; height: 6px; border-radius: 50%;
  background: rgb(var(--v-theme-error)); flex-shrink: 0; display: inline-block;
}
.status-dot.running {
  background: rgb(var(--v-theme-success));
  box-shadow: 0 0 5px rgb(var(--v-theme-success));
  animation: pulse 2s ease-in-out infinite;
}
@keyframes pulse { 0%,100% { opacity: 1; } 50% { opacity: .35; } }

/* Bot switcher */
.bot-switcher {
  background: rgb(var(--v-theme-surface-variant));
  border: 1px solid rgba(128,128,128,0.15);
  -webkit-app-region: no-drag;
  transition: all 0.15s ease;
}
.bot-switcher:hover { border-color: rgb(var(--v-theme-primary)) !important; }

/* Locale toggle */
.locale-btn { -webkit-app-region: no-drag; }

/* Nav item custom */
.nav-item-custom { min-height: 34px !important; margin-bottom: 1px !important; }
.nav-item-custom .v-list-item-title { font-size: 13px !important; }
.nav-item-custom .v-list-item__prepend { margin-inline-end: 10px !important; }

/* Nav badge */
.nav-badge {
  background: rgb(var(--v-theme-error)); color: white;
  padding: 0 5px; border-radius: 8px;
  font-size: 10px; font-weight: 600;
  min-width: 16px; text-align: center; line-height: 16px;
}

/* Page header */
.page-header {
  padding: 10px 20px;
  border-bottom: 1px solid rgba(128,128,128,0.12);
  display: flex; align-items: center;
  justify-content: space-between;
  background: rgb(var(--v-theme-background));
  -webkit-app-region: drag;
  gap: 12px;
  min-height: 48px;
}
.page-header h2 { font-size: 14px; font-weight: 600; white-space: nowrap; }
.page-header .actions { -webkit-app-region: no-drag; }

/* Page body */
.page-body { flex: 1; overflow-y: auto; padding: 20px; }

/* Card base */
.card-base {
  background: rgb(var(--v-theme-surface));
  border: 1px solid rgba(128,128,128,0.12);
  border-radius: 10px;
  transition: border-color 0.15s;
}
.card-base:hover { border-color: rgb(var(--v-theme-primary)); }

/* Snackbar overrides */
.v-snackbar__content { padding: 8px 14px !important; }
</style>
