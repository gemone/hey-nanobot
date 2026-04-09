<template>
  <div class="d-flex flex-column" style="height: 100%;">
    <div class="page-header">
      <h2 class="text-body-1 font-weight-bold">⚙️ Config</h2>
      <div class="d-flex ga-2">
        <v-btn variant="text" size="small" prepend-icon="mdi-content-copy" @click="formatConfig">Format</v-btn>
        <v-btn color="primary" size="small" prepend-icon="mdi-content-save" @click="$emit('save', localConfig)">Save</v-btn>
      </div>
    </div>
    <div class="flex-grow-1 pa-4">
      <v-textarea
        v-model="localConfig"
        auto-grow
        variant="outlined"
        hide-details
        style="font-family: 'SF Mono', Monaco, Menlo, monospace; font-size: 12px; line-height: 1.6;"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{ configJson: string }>()
const emit = defineEmits<{ (e: 'save', json: string): void }>()

const localConfig = ref(props.configJson)
watch(() => props.configJson, (v) => { localConfig.value = v })

function formatConfig() {
  try {
    localConfig.value = JSON.stringify(JSON.parse(localConfig.value), null, 2)
  } catch {}
}
</script>
