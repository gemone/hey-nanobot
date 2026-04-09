<template>
  <div class="page-body">
    <div class="d-flex align-center justify-space-between mb-5">
      <h2 class="text-body-1 font-weight-bold">🔑 Providers</h2>
    </div>
    <v-row>
      <v-col v-for="(pv, name) in providers" :key="name" cols="12" sm="6" md="4" lg="3">
        <v-card rounded="lg" class="pa-4" style="border: 1px solid #2a2a45;" @mouseenter="$event.currentTarget.style.borderColor='#6c5ce7'" @mouseleave="$event.currentTarget.style.borderColor='#2a2a45'">
          <div class="d-flex align-center ga-2 mb-3">
            <v-icon size="18" color="primary">mdi-cloud-outline</v-icon>
            <span class="text-body-2 font-weight-semibold text-capitalize">{{ name }}</span>
            <v-chip v-if="pv.api_key" size="x-small" color="success" variant="tonal" class="ml-auto">✓</v-chip>
          </div>
          <v-text-field
            :model-value="pv.api_key || ''"
            @update:model-value="$emit('set-key', name, $event)"
            label="API Key"
            variant="outlined"
            density="compact"
            hide-details
            type="password"
            placeholder="sk-..."
          />
        </v-card>
      </v-col>
    </v-row>
  </div>
</template>

<script setup lang="ts">
defineProps<{ providers: Record<string, any> }>()
defineEmits(['set-key'])
</script>
