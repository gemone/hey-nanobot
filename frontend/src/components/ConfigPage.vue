<template>
  <div style="display:flex; flex-direction:column; height:100%">
    <div class="page-header">
      <h2>⚙️ Config</h2>
      <div class="actions">
        <button class="btn btn-ghost btn-sm" @click="formatJson">📐 Format</button>
        <button class="btn btn-primary btn-sm" @click="save" :disabled="saving">
          {{ saving ? '⏳' : '💾' }} Save
        </button>
      </div>
    </div>
    <div class="page-body" style="display:flex; flex-direction:column; height:calc(100% - 52px)">
      <div class="config-editor" style="flex:1; display:flex; flex-direction:column">
        <textarea v-model="localConfig" spellcheck="false"
          @keydown.ctrl.s.prevent="save" @keydown.meta.s.prevent="save" />
        <div v-if="error" style="color:var(--red); font-size:11px; padding:6px 0">❌ {{ error }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { SaveConfig } from '../../wailsjs/go/main/App'

const props = defineProps<{ configJson: string }>()
const emit = defineEmits(['save'])
const localConfig = ref(props.configJson)
const saving = ref(false)
const error = ref('')

watch(() => props.configJson, (val) => { localConfig.value = val })

function formatJson() {
  try {
    const parsed = JSON.parse(localConfig.value)
    localConfig.value = JSON.stringify(parsed, null, 2)
    error.value = ''
  } catch (e) { error.value = 'Invalid JSON: ' + e }
}

async function save() {
  error.value = ''
  try { JSON.parse(localConfig.value) } catch (e) { error.value = 'Invalid JSON: ' + e; return }
  saving.value = true
  try { await SaveConfig(localConfig.value); emit('save', localConfig.value) }
  catch (e) { error.value = String(e) }
  saving.value = false
}
</script>
