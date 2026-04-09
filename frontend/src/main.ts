import { createApp } from 'vue'
import App from './App.vue'

// Vuetify
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
import { createVuetify, ThemeDefinition } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'

// i18n
import i18n from './i18n'

const darkTheme: ThemeDefinition = {
  dark: true,
  colors: {
    background: '#0f0f1a',
    surface: '#161628',
    'surface-variant': '#222240',
    'on-surface-variant': '#9898b0',
    primary: '#6c5ce7',
    'primary-darken-1': '#5a4bd6',
    secondary: '#00cec9',
    error: '#ff6b6b',
    info: '#74b9ff',
    success: '#00cec9',
    warning: '#fdcb6e',
  },
}

const lightTheme: ThemeDefinition = {
  dark: false,
  colors: {
    background: '#f5f5fa',
    surface: '#ffffff',
    'surface-variant': '#eaeaf2',
    'on-surface-variant': '#5a5a78',
    primary: '#6c5ce7',
    'primary-darken-1': '#5a4bd6',
    secondary: '#00b3ae',
    error: '#e74c3c',
    info: '#3498db',
    success: '#00b3ae',
    warning: '#f39c12',
  },
}

const vuetify = createVuetify({
  components,
  directives,
  theme: {
    defaultTheme: 'dark',
    themes: {
      dark: darkTheme,
      light: lightTheme,
    },
  },
})

const app = createApp(App)
app.use(vuetify)
app.use(i18n)
app.mount('#app')
