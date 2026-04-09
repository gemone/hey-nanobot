<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">ℹ️ System</h2>
    </div>

    <!-- Nanobot Engine -->
    <v-card rounded="lg" class="mb-4" style="border: 1px solid #2a2a45;">
      <v-card-title class="text-body-2 font-weight-bold px-4 pt-3 pb-1">
        <v-icon size="16" class="mr-2">mdi-robot-outline</v-icon>
        Nanobot Engine
      </v-card-title>
      <v-list bg-color="transparent" density="compact" class="pb-2">
        <v-list-item>
          <template v-slot:prepend><v-icon size="16" class="mr-3">mdi-file-outline</v-icon></template>
          <v-list-item-title class="text-body-2">
            <strong>Path</strong>: <span class="text-medium-emphasis">{{ nanobotInfo.path || 'Not found' }}</span>
          </v-list-item-title>
        </v-list-item>
        <v-list-item>
          <template v-slot:prepend><v-icon size="16" class="mr-3">mdi-source-branch</v-icon></template>
          <v-list-item-title class="text-body-2">
            <strong>Source</strong>:
            <v-chip size="x-small" :color="sourceColor" variant="tonal" class="ml-1">
              {{ sourceLabel }}
            </v-chip>
          </v-list-item-title>
        </v-list-item>
        <v-list-item v-if="nanobotInfo.version">
          <template v-slot:prepend><v-icon size="16" class="mr-3">mdi-tag-outline</v-icon></template>
          <v-list-item-title class="text-body-2">
            <strong>Version</strong>: <span class="text-medium-emphasis">{{ nanobotInfo.version }}</span>
          </v-list-item-title>
        </v-list-item>
      </v-list>
    </v-card>

    <!-- System Info -->
    <v-card rounded="lg" style="border: 1px solid #2a2a45;">
      <v-list bg-color="transparent" density="compact">
        <v-list-item v-for="(val, key) in info" :key="key">
          <template v-slot:prepend>
            <v-icon size="16" class="mr-3">mdi-circle-small</v-icon>
          </template>
          <v-list-item-title class="text-body-2">
            <strong>{{ key }}</strong>: <span class="text-medium-emphasis">{{ val }}</span>
          </v-list-item-title>
        </v-list-item>
      </v-list>
    </v-card>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { Ref } from 'vue'

const props = defineProps<{
  info: Record<string, string>
  nanobotInfo: { path: string; source: string; version: string; available: boolean }
}>()

const sourceLabel = computed(() => {
  const s = props.nanobotInfo?.source || 'none'
  const map: Record<string, string> = {
    bundled: '📦 Built-in',
    external: '🌐 External',
    custom: '🔧 Custom',
    none: '❌ Not Found',
  }
  return map[s] || s
})

const sourceColor = computed(() => {
  const s = props.nanobotInfo?.source || 'none'
  const map: Record<string, string> = {
    bundled: 'success',
    external: 'info',
    custom: 'warning',
    none: 'error',
  }
  return map[s] || 'grey'
})
</script>
