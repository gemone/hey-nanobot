<template>
  <div class="d-flex flex-column" style="height: 100%;">
    <div class="page-header">
      <div class="d-flex align-center ga-2">
        <v-icon size="20" color="primary">mdi-cog-outline</v-icon>
        <span class="text-body-1 font-weight-bold">{{ t('config.title') }}</span>
      </div>
      <div class="actions d-flex ga-2">
        <v-btn variant="text" size="small" prepend-icon="mdi-code-tags" @click="formatConfig" style="color: #9898b0;">{{ t('config.format') }}</v-btn>
        <v-btn color="primary" size="small" variant="tonal" prepend-icon="mdi-content-save" @click="$emit('save', localConfig)">{{ t('config.save') }}</v-btn>
      </div>
    </div>
    <div class="flex-grow-1 pa-4">
      <textarea
        v-model="localConfig"
        spellcheck="false"
        class="config-editor-textarea"
      />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const props = defineProps<{ configJson: string }>()
defineEmits<{ (e: 'save', json: string): void }>()

const localConfig = ref(props.configJson)
watch(() => props.configJson, (v) => { localConfig.value = v })

function formatConfig() {
  try { localConfig.value = JSON.stringify(JSON.parse(localConfig.value), null, 2) } catch {}
}
</script>

<style scoped>
.config-editor-textarea {
  width: 100%; height: 100%;
  background: rgb(var(--v-theme-background));
  border: 1px solid rgba(128,128,128,0.12);
  border-radius: 8px;
  color: rgb(var(--v-theme-on-surface-variant));
  padding: 14px;
  font-family: 'SF Mono', 'Menlo', 'Monaco', monospace;
  font-size: 12px;
  line-height: 1.6;
  resize: none; outline: none; tab-size: 2;
}
.config-editor-textarea:focus { border-color: rgb(var(--v-theme-primary)); }
</style>
