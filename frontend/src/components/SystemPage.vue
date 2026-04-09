<template>
  <div class="page-body">
    <div class="d-flex align-center ga-2 mb-4">
      <v-icon size="20" color="primary">mdi-information-outline</v-icon>
      <span class="text-body-1 font-weight-bold">{{ t('system.title') }}</span>
    </div>

    <!-- Nanobot Engine -->
    <div class="card-base pa-4 mb-4">
      <div class="d-flex align-center ga-2 mb-3">
        <v-icon size="16" color="primary">mdi-robot-outline</v-icon>
        <span class="text-caption font-weight-semibold" style="text-transform: uppercase; letter-spacing: 0.5px;">{{ t('system.nanobotEngine') }}</span>
      </div>
      <div class="info-rows">
        <div class="info-row">
          <span class="info-key">Path</span>
          <span class="info-val" style="font-family: 'SF Mono', monospace; font-size: 12px;">{{ nanobotInfo.path || t('system.notFound') }}</span>
        </div>
        <div class="info-row">
          <span class="info-key">{{ t('system.source') }}</span>
          <v-chip size="x-small" :color="sourceColor" variant="tonal">{{ sourceLabel }}</v-chip>
        </div>
        <div v-if="nanobotInfo.version" class="info-row">
          <span class="info-key">{{ t('system.version') }}</span>
          <span class="info-val">{{ nanobotInfo.version }}</span>
        </div>
      </div>
    </div>

    <!-- System Info -->
    <div class="card-base pa-4">
      <div class="d-flex align-center ga-2 mb-3">
        <v-icon size="16" color="primary">mdi-desktop-classic</v-icon>
        <span class="text-caption font-weight-semibold" style="text-transform: uppercase; letter-spacing: 0.5px;">System</span>
      </div>
      <div class="info-rows">
        <div v-for="(val, key) in info" :key="key" class="info-row">
          <span class="info-key">{{ key }}</span>
          <span class="info-val">{{ val }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const props = defineProps<{
  info: Record<string, string>
  nanobotInfo: { path: string; source: string; version: string; available: boolean }
}>()

const sourceLabel = computed(() => {
  const s = props.nanobotInfo?.source || 'none'
  return { standard: t('system.sourceStandard'), bundled: t('system.sourceBundled'), external: t('system.sourceExternal'), custom: t('system.sourceCustom'), none: t('system.sourceNone') }[s] || s
})
const sourceColor = computed(() => {
  const s = props.nanobotInfo?.source || 'none'
  return { standard: 'success', bundled: 'info', external: 'info', custom: 'warning', none: 'error' }[s] || 'grey'
})
</script>

<style scoped>
.info-rows { display: flex; flex-direction: column; gap: 8px; }
.info-row { display: flex; align-items: center; justify-content: space-between; padding: 4px 0; }
.info-key { font-size: 12px; color: #5a5a78; flex-shrink: 0; min-width: 70px; }
.info-val { font-size: 12px; color: #b0b0c8; text-align: right; word-break: break-all; }
</style>
