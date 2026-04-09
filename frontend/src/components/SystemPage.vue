<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">{{ t('system.title') }}</h2>
    </div>

    <!-- Nanobot Engine -->
    <v-card rounded="lg" class="mb-4" style="border: 1px solid #2a2a45;">
      <v-card-title class="text-body-2 font-weight-bold px-4 pt-3 pb-1">
        <v-icon size="16" class="mr-2">mdi-robot-outline</v-icon>
        {{ t('system.nanobotEngine') }}
      </v-card-title>
      <v-list bg-color="transparent" density="compact" class="pb-2">
        <v-list-item>
          <template v-slot:prepend><v-icon size="16" class="mr-3">mdi-file-outline</v-icon></template>
          <v-list-item-title class="text-body-2">
            <strong>{{ t('system.path') }}</strong>: <span class="text-medium-emphasis">{{ nanobotInfo.path || t('system.notFound') }}</span>
          </v-list-item-title>
        </v-list-item>
        <v-list-item>
          <template v-slot:prepend><v-icon size="16" class="mr-3">mdi-source-branch</v-icon></template>
          <v-list-item-title class="text-body-2">
            <strong>{{ t('system.source') }}</strong>:
            <v-chip size="x-small" :color="sourceColor" variant="tonal" class="ml-1">
              {{ sourceLabel }}
            </v-chip>
          </v-list-item-title>
        </v-list-item>
        <v-list-item v-if="nanobotInfo.version">
          <template v-slot:prepend><v-icon size="16" class="mr-3">mdi-tag-outline</v-icon></template>
          <v-list-item-title class="text-body-2">
            <strong>{{ t('system.version') }}</strong>: <span class="text-medium-emphasis">{{ nanobotInfo.version }}</span>
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
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

const props = defineProps<{
  info: Record<string, string>
  nanobotInfo: { path: string; source: string; version: string; available: boolean }
}>()

const sourceLabel = computed(() => {
  const s = props.nanobotInfo?.source || 'none'
  const map: Record<string, string> = {
    standard: t('system.sourceStandard'),
    bundled: t('system.sourceBundled'),
    external: t('system.sourceExternal'),
    custom: t('system.sourceCustom'),
    none: t('system.sourceNone'),
  }
  return map[s] || s
})

const sourceColor = computed(() => {
  const s = props.nanobotInfo?.source || 'none'
  const map: Record<string, string> = {
    standard: 'success',
    bundled: 'info',
    external: 'info',
    custom: 'warning',
    none: 'error',
  }
  return map[s] || 'grey'
})
</script>
