<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">{{ t('channel.title') }}</h2>
    </div>
    <v-row>
      <v-col v-for="(ch, name) in channels" :key="name" cols="12" sm="6" md="4">
        <v-card rounded="lg" class="pa-4" style="border: 1px solid #2a2a45;">
          <div class="d-flex align-center ga-2 mb-3">
            <v-icon size="18" color="primary">mdi-{{ channelIcon(name) }}</v-icon>
            <span class="text-body-2 font-weight-semibold text-capitalize">{{ name }}</span>
            <v-switch
              :model-value="ch.enabled"
              @update:model-value="$emit('toggle-channel', name, $event)"
              density="compact"
              hide-details
              color="primary"
              class="ml-auto"
            />
          </div>
          <template v-if="ch.enabled">
            <v-text-field
              v-for="field in getChannelFields(name)"
              :key="field.key"
              :model-value="ch[field.key] || ''"
              @update:model-value="$emit('update-field', name, field.key, $event)"
              :label="field.label"
              variant="outlined"
              density="compact"
              hide-details
              class="mb-2"
              :type="field.secret ? 'password' : 'text'"
            />
          </template>
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
import { useI18n } from 'vue-i18n'
const { t } = useI18n()

defineProps<{ channels: Record<string, any> }>()
defineEmits(['toggle-channel', 'update-field'])

function channelIcon(name: string): string {
  const icons: Record<string, string> = {
    telegram: 'send', discord: 'chat', qq: 'chat',
    slack: 'slack', feishu: 'bird', dingtalk: 'bell',
    wecom: 'briefcase', whatsapp: 'phone',
  }
  return icons[name] || 'message-text'
}

function getChannelFields(name: string) {
  const fields: Record<string, { key: string; label: string; secret?: boolean }[]> = {
    telegram: [{ key: 'token', label: 'Bot Token', secret: true }],
    discord: [{ key: 'token', label: 'Bot Token', secret: true }],
    qq: [{ key: 'app_id', label: 'App ID' }, { key: 'app_secret', label: 'App Secret', secret: true }],
    slack: [{ key: 'bot_token', label: 'Bot Token', secret: true }, { key: 'app_token', label: 'App Token', secret: true }],
    feishu: [{ key: 'app_id', label: 'App ID' }, { key: 'app_secret', label: 'App Secret', secret: true }],
    dingtalk: [{ key: 'client_id', label: 'Client ID' }, { key: 'client_secret', label: 'Client Secret', secret: true }],
    wecom: [{ key: 'corp_id', label: 'Corp ID' }, { key: 'agent_id', label: 'Agent ID' }, { key: 'secret', label: 'Secret', secret: true }],
    whatsapp: [{ key: 'token', label: 'Token', secret: true }, { key: 'phone_number_id', label: 'Phone Number ID' }],
  }
  return fields[name] || []
}
</script>
