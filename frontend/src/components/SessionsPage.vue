<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">{{ t('sessions.title') }}</h2>
    </div>
    <div v-if="sessions.length === 0" class="text-center py-12">
      <div style="font-size: 48px;" class="mb-3">📂</div>
      <p class="text-body-2 text-medium-emphasis">{{ t('sessions.noSessions') }}</p>
    </div>
    <v-list v-else bg-color="transparent" density="compact" rounded="lg" style="border: 1px solid #2a2a45;">
      <v-list-item v-for="s in sessions" :key="s.key" @click="$emit('open-in-finder', s.path)">
        <template v-slot:prepend>
          <v-icon size="16" color="primary">mdi-file-document-outline</v-icon>
        </template>
        <v-list-item-title class="text-body-2">{{ s.key }}</v-list-item-title>
        <v-list-item-subtitle class="text-caption text-medium-emphasis">
          {{ s.updated_at || s.created_at }}
        </v-list-item-subtitle>
        <template v-slot:append>
          <v-btn icon size="x-small" variant="text">
            <v-icon size="14">mdi-open-in-new</v-icon>
          </v-btn>
        </template>
      </v-list-item>
    </v-list>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()

defineProps<{ sessions: any[] }>()
defineEmits(['open-in-finder'])
</script>
