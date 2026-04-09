<template>
  <div>
    <div class="page-header">
      <h2>🔑 Providers</h2>
    </div>
    <div class="page-body">
      <div class="provider-grid">
        <div v-for="(config, name) in providers" :key="name" class="provider-card">
          <div class="pv-name">
            <span class="pv-dot" :class="config.apiKey ? 'on' : 'off'"></span>
            {{ name }}
          </div>
          <div class="form-group" style="margin-bottom:6px">
            <input class="form-input" :type="showKeys[name] ? 'text' : 'password'"
              :value="config.apiKey || ''" :placeholder="config.apiKey ? '••••••' : 'API Key'"
              style="font-size:11px; padding:5px 8px;"
              @change="$emit('set-key', name, ($event.target as HTMLInputElement).value)" />
          </div>
          <div style="display:flex; justify-content:flex-end">
            <button class="btn btn-ghost btn-xs" @click="toggleShow(name)">
              {{ showKeys[name] ? '🙈 Hide' : '👁️' }}
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
defineProps<{ providers: Record<string, any> }>()
defineEmits(['set-key'])
const showKeys = ref<Record<string, boolean>>({})
function toggleShow(name: string) { showKeys.value[name] = !showKeys.value[name] }
</script>
