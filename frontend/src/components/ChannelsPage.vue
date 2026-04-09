<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">🔗 Channels</h2>
    </div>
    <v-row>
      <v-col v-for="(ch, name) in channels" :key="name" cols="12" sm="6" md="4" lg="3">
        <v-card rounded="lg" class="pa-4" style="border: 1px solid #2a2a45;" @mouseenter="$event.currentTarget.style.borderColor='#6c5ce7'" @mouseleave="$event.currentTarget.style.borderColor='#2a2a45'">
          <div class="d-flex align-center justify-space-between mb-3">
            <div class="d-flex align-center ga-2">
              <span style="font-size: 18px;">{{ channelIcon(name) }}</span>
              <span class="text-body-2 font-weight-semibold text-capitalize">{{ name }}</span>
            </div>
            <v-switch
              :model-value="ch.enabled"
              @update:model-value="$emit('toggle-channel', name, $event)"
              density="compact"
              hide-details
              color="success"
            />
          </div>
          <div class="d-flex flex-column ga-2">
            <v-text-field
              v-for="(val, field) in filterFields(ch)"
              :key="field"
              :model-value="val"
              @update:model-value="$emit('update-field', name, field, $event)"
              :label="field"
              variant="outlined"
              density="compact"
              hide-details
              :type="field.includes('secret') || field.includes('token') || field.includes('key') ? 'password' : 'text'"
            />
          </div>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
defineProps<{ channels: Record<string, any> }>()
defineEmits(['toggle-channel', 'update-field'])

function channelIcon(name: string) {
  const m: Record<string, string> = { telegram: '✈️', discord: '🎮', slack: '💬', qq: '🐧', wecom: '💼', feishu: '🐦', dingtalk: '🔔', whatsapp: '📱', email: '📧', matrix: '🟢' }
  return m[name] || '📡'
}

function filterFields(ch: any) {
  const { enabled, ...rest } = ch
  return rest
}
</script>
