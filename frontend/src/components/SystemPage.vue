<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">ℹ️ System</h2>
    </div>
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
defineProps<{ info: Record<string, string> }>()
</script>
