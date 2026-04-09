import { createApp } from 'vue'
import App from './App.vue'

// Vuetify
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

// i18n
import i18n from './i18n'

const vuetify = createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'dark',
    themes: {
      dark: {
        dark: true,
        colors: {
          background: '#0f0f1a',
          surface: '#1a1a30',
          'surface-variant': '#252540',
          'on-surface-variant': '#9898b0',
          primary: '#6c5ce7',
          'primary-darken-1': '#5a4bd6',
          secondary: '#00cec9',
          error: '#ff6b6b',
          info: '#74b9ff',
          success: '#00cec9',
          warning: '#fdcb6e',
          border: '#2a2a45',
        },
      },
    },
  },
})

const app = createApp(App)
app.use(vuetify)
app.use(i18n)
app.mount('#app')
