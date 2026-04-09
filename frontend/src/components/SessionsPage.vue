<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">📂 Sessions</h2>
      <v-chip size="small" variant="tonal" color="primary">{{ sessions.length }} sessions</v-chip>
    </div>
    <v-list bg-color="transparent" density="compact" rounded="lg">
      <v-list-item
        v-for="s in sessions"
        :key="s.key"
        @click="$emit('open-in-finder', s.path || s.key)"
        rounded="lg"
        class="mb-1"
      >
        <template v-slot:prepend>
          <v-icon size="18" class="mr-2">mdi-file-document-outline</v-icon>
        </template>
        <v-list-item-title class="text-body-2">{{ s.key }}</v-list-item-title>
        <v-list-item-subtitle class="text-caption text-medium-emphasis">
          {{ s.channel || 'unknown' }} · {{ s.messages || 0 }} msgs
        </v-list-item-subtitle>
        <template v-slot:append>
          <v-btn icon size="x-small" variant="text">
            <v-icon size="14">mdi-open-in-new</v-icon>
          </v-btn>
        </template>
      </v-list-item>
    </v-list>
    <div v-if="!sessions.length" class="text-center py-12 text-body-2 text-medium-emphasis">
      No sessions yet
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ sessions: any[] }>()
defineEmits(['open-in-finder'])
</script>
