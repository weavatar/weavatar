<template>
  <n-config-provider
    :locale="zhCN"
    :date-locale="dateZhCN"
    :theme="theme"
    :theme-overrides="themeOverrides"
    inline-theme-disabled
  >
    <n-loading-bar-provider :loading-bar-style="{ loading: { height: '2px' } }">
      <n-message-provider placement="top" :max="3">
        <n-notification-provider>
          <n-dialog-provider>
            <app-provider>
              <router-view />
            </app-provider>
          </n-dialog-provider>
        </n-notification-provider>
      </n-message-provider>
    </n-loading-bar-provider>
  </n-config-provider>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import {
  darkTheme,
  dateZhCN,
  NConfigProvider,
  NDialogProvider,
  NLoadingBarProvider,
  NMessageProvider,
  NNotificationProvider,
  zhCN,
  type GlobalThemeOverrides
} from 'naive-ui'
import AppProvider from '@/components/AppProvider.vue'
import { useTheme } from '@/composables/useTheme'

const { isDark } = useTheme()
const theme = computed(() => (isDark.value ? darkTheme : null))

const common = {
  fontFamily: 'var(--wa-font-sans)',
  fontFamilyMono: 'var(--wa-font-mono)',
  primaryColor: '#0a7aff',
  primaryColorHover: '#2b8dff',
  primaryColorPressed: '#0062d6',
  primaryColorSuppl: '#2b8dff',
  infoColor: '#0a7aff',
  infoColorHover: '#2b8dff',
  infoColorPressed: '#0062d6',
  infoColorSuppl: '#2b8dff',
  borderRadius: '10px',
  borderRadiusSmall: '8px',
  heightSmall: '32px',
  heightMedium: '40px',
  heightLarge: '44px',
  fontSize: '14px',
  fontSizeMedium: '14px',
  fontSizeLarge: '15px',
  fontWeightStrong: '600'
}

const components: GlobalThemeOverrides = {
  Button: {
    borderRadiusTiny: '999px',
    borderRadiusSmall: '999px',
    borderRadiusMedium: '999px',
    borderRadiusLarge: '999px',
    fontWeight: '600',
    paddingMedium: '0 18px',
    paddingLarge: '0 24px'
  },
  Card: { borderRadius: '18px', paddingMedium: '24px' },
  Input: { borderRadius: '10px' },
  Tag: { borderRadius: '999px' },
  Pagination: { itemBorderRadius: '10px' },
  Dropdown: { borderRadius: '14px', padding: '6px' },
  Dialog: { borderRadius: '18px' },
  Popover: { borderRadius: '12px' }
}

const lightOverrides: GlobalThemeOverrides = {
  ...components,
  common: {
    ...common,
    bodyColor: '#fafafb',
    cardColor: '#ffffff',
    modalColor: '#ffffff',
    popoverColor: '#ffffff',
    borderColor: 'rgba(15,23,42,0.12)',
    dividerColor: 'rgba(15,23,42,0.08)',
    textColorBase: '#0b0c10',
    textColor1: '#0b0c10',
    textColor2: '#2b2f3a',
    textColor3: '#6b7280',
    placeholderColor: '#9ca3af'
  }
}

const darkOverrides: GlobalThemeOverrides = {
  ...components,
  common: {
    ...common,
    bodyColor: '#0a0b0e',
    cardColor: '#131419',
    modalColor: '#131419',
    popoverColor: '#1a1c22',
    inputColor: '#0f1014',
    inputColorDisabled: '#16181d',
    borderColor: 'rgba(255,255,255,0.14)',
    dividerColor: 'rgba(255,255,255,0.08)',
    hoverColor: 'rgba(255,255,255,0.08)'
  }
}

const themeOverrides = computed(() => (isDark.value ? darkOverrides : lightOverrides))
</script>
