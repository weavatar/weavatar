import { fileURLToPath, URL } from 'node:url'
import { includeIgnoreFile } from 'eslint/config'
import js from '@eslint/js'
import pluginVue from 'eslint-plugin-vue'
import { defineConfigWithVueTs, vueTsConfigs } from '@vue/eslint-config-typescript'
import skipFormatting from '@vue/eslint-config-prettier/skip-formatting'

export default defineConfigWithVueTs(
  includeIgnoreFile(fileURLToPath(new URL('.gitignore', import.meta.url))),
  pluginVue.configs['flat/essential'],
  js.configs.recommended,
  vueTsConfigs.base,
  vueTsConfigs.eslintRecommended,
  {
    files: ['**/*.{ts,mts,cts,tsx,vue}'],
    rules: {
      'no-unused-vars': 'off',
      'no-undef': 'off',
      '@typescript-eslint/no-unused-vars': 'warn'
    }
  },
  skipFormatting
)
