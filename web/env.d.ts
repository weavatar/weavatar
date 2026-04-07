/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare module 'vue-cropper' {
  import type { DefineComponent } from 'vue'
  export const VueCropper: DefineComponent<any, any, any>
}
