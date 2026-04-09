<template>
  <div>
    <div class="page-header">
      <h2>🔗 Channels</h2>
    </div>
    <div class="page-body">
      <div class="channel-grid">
        <div v-for="(config, name) in channels" :key="name" class="channel-card">
          <div class="ch-header">
            <span class="ch-name">
              <span class="ch-icon">{{ getChannelIcon(name) }}</span>
              {{ name }}
            </span>
            <label class="toggle">
              <input type="checkbox" :checked="config.enabled"
                @change="$emit('toggle-channel', name, ($event.target as HTMLInputElement).checked)" />
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div class="ch-fields">
            <div v-for="field in getEditableFields(name, config)" :key="field.key" class="field-row">
              <label>{{ field.label }}</label>
              <input class="form-input" style="flex:1; padding:4px 8px; font-size:11px;"
                :type="field.secret ? 'password' : 'text'"
                :value="field.value" :placeholder="field.placeholder"
                @change="$emit('update-field', name, field.key, ($event.target as HTMLInputElement).value)" />
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
defineProps<{ channels: Record<string, any> }>()
defineEmits(['toggle-channel', 'update-field'])

function getChannelIcon(name: string): string {
  const icons: Record<string, string> = {
    telegram: '✈️', discord: '🎮', slack: '💼', qq: '🐧',
    wecom: '🏢', feishu: '🪽', dingtalk: '🔔', whatsapp: '📱',
    email: '📧', mochat: '🫧',
  }
  return icons[name] || '📡'
}

interface FieldDef { key: string; label: string; value: string; placeholder: string; secret: boolean }

function getEditableFields(name: string, config: any): FieldDef[] {
  const defs: Record<string, { key: string; label: string; secret?: boolean; placeholder?: string }[]> = {
    telegram: [
      { key: 'token', label: 'Token', secret: true, placeholder: '123456:ABC-DEF...' },
      { key: 'proxy', label: 'Proxy', placeholder: 'socks5://...' },
    ],
    discord: [{ key: 'token', label: 'Token', secret: true }],
    slack: [
      { key: 'botToken', label: 'Bot Token', secret: true },
      { key: 'appToken', label: 'App Token', secret: true },
    ],
    qq: [
      { key: 'appId', label: 'App ID' },
      { key: 'secret', label: 'Secret', secret: true },
    ],
    wecom: [
      { key: 'botId', label: 'Bot ID' },
      { key: 'secret', label: 'Secret', secret: true },
    ],
    feishu: [
      { key: 'appId', label: 'App ID' },
      { key: 'appSecret', label: 'App Secret', secret: true },
    ],
    dingtalk: [
      { key: 'clientId', label: 'Client ID' },
      { key: 'clientSecret', label: 'Secret', secret: true },
    ],
    email: [
      { key: 'imapHost', label: 'IMAP Host' },
      { key: 'imapUsername', label: 'User' },
      { key: 'imapPassword', label: 'Pass', secret: true },
      { key: 'smtpHost', label: 'SMTP Host' },
    ],
    whatsapp: [
      { key: 'bridgeUrl', label: 'Bridge URL' },
      { key: 'bridgeToken', label: 'Token', secret: true },
    ],
    mochat: [{ key: 'clawToken', label: 'Claw Token', secret: true }],
  }
  return (defs[name] || []).map(d => ({
    key: d.key, label: d.label,
    value: String(config[d.key] || ''),
    placeholder: d.placeholder || '',
    secret: d.secret || false,
  }))
}
</script>
