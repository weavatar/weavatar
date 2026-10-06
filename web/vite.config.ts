import { execSync } from 'node:child_process'
import { fileURLToPath, URL } from 'node:url'

import { defineConfig, type Plugin } from 'vite'
import vue from '@vitejs/plugin-vue'
import UnoCSS from 'unocss/vite'

function versionFile(): Plugin {
  return {
    name: 'weavatar:version-file',
    apply: 'build',
    generateBundle() {
      let hash = 'unknown'
      try {
        hash = execSync('git rev-parse --short HEAD', { encoding: 'utf8' }).trim()
      } catch {
        // 不在 git 仓库里构建时保留 unknown
      }
      this.emitFile({ type: 'asset', fileName: 'version.json', source: JSON.stringify({ hash }) })
    }
  }
}

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [vue(), UnoCSS(), versionFile()],
  resolve: {
    alias: {
      '@': fileURLToPath(new URL('./src', import.meta.url))
    }
  }
})
