/// <reference types="vite/client" />

declare module '*.vue' {
    import type {DefineComponent} from 'vue'
    const component: DefineComponent<{}, {}, any>
    export default component
}

// Wails bindings
interface WailsApp {
  GetServerURL: () => Promise<string>
}

interface WailsMain {
  App: WailsApp
}

interface WailsGo {
  main: WailsMain
}

declare global {
  interface Window {
    go: WailsGo
  }
}

export {}
