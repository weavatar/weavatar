import { defineConfig, presetIcons, presetWind4, transformerDirectives } from 'unocss'

export default defineConfig({
  content: {
    pipeline: {
      include: [/\.(vue|ts)($|\?)/],
      exclude: ['node_modules', '.git', '.github', '.vscode', 'build', 'dist', 'public', 'types']
    }
  },
  presets: [
    presetWind4({
      dark: 'class',
      preflights: { reset: true, theme: 'on-demand' }
    }),
    presetIcons({
      scale: 1.1,
      extraProperties: {
        display: 'inline-block',
        'vertical-align': 'middle',
        'flex-shrink': '0'
      },
      collections: {
        lucide: () => import('@iconify-json/lucide/icons.json').then((i) => i.default)
      }
    })
  ],
  transformers: [transformerDirectives()],
  shortcuts: {
    // 布局
    wrap: 'mx-auto w-full max-w-7xl px-6 sm:px-10',
    section: 'py-20 sm:py-28',
    // 表面
    surface: 'bg-elev border border-line rounded-2xl',
    'surface-muted': 'bg-muted border border-line rounded-2xl',
    // 文字
    eyebrow: 'inline-flex items-center gap-2.5 text-xs font-600 tracking-[0.06em] text-brand',
    'display-text': 'font-700 tracking-normal leading-[1.18]',
    // 按钮
    btn: 'inline-flex items-center justify-center gap-2 h-11 px-5 rounded-full text-sm font-600 transition-all duration-200 select-none cursor-pointer whitespace-nowrap disabled:opacity-50 disabled:cursor-not-allowed disabled:hover:translate-y-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-brand/40',
    'btn-primary':
      'btn bg-brand text-white hover:bg-brand-hover active:bg-brand-pressed shadow-brand hover:-translate-y-0.5',
    'btn-secondary':
      'btn bg-elev text-fg border border-line-strong hover:bg-muted hover:border-fg/25 hover:-translate-y-0.5',
    'btn-ghost': 'btn bg-transparent text-fg2 hover:text-fg hover:bg-fg/6',
    'btn-danger': 'btn bg-red-500/10 text-red-600 dark:text-red-400 hover:bg-red-500/16',
    'btn-sm': 'h-9 px-4 text-[13px]',
    'btn-lg': 'h-12 px-7 text-[15px]',
    'btn-icon': 'px-0 w-10 h-10',
    // 链接
    link: 'text-brand hover:underline underline-offset-4 decoration-brand/40',
    'nav-link':
      'px-3 py-2 rounded-lg text-sm font-500 text-fg2 hover:text-fg hover:bg-fg/6 transition-colors',
    // 小标签
    chip: 'inline-flex items-center gap-1.5 h-7 px-2.5 rounded-full text-xs font-500 bg-muted border border-line text-fg2',
    'chip-brand': 'chip bg-brand/10 border-brand/20 text-brand',
    kbd: 'inline-block px-1.5 py-0.5 rounded-md bg-muted border border-line text-[0.85em] font-mono text-fg',
    // 表单
    'field-label': 'block text-sm font-600 text-fg mb-2'
  },
  theme: {
    colors: {
      brand: {
        DEFAULT: '#0a7aff',
        hover: '#2b8dff',
        pressed: '#0062d6'
      },
      bg: 'var(--wa-bg)',
      elev: 'var(--wa-elev)',
      muted: 'var(--wa-muted)',
      fg: 'var(--wa-fg)',
      fg2: 'var(--wa-fg2)',
      fg3: 'var(--wa-fg3)',
      line: {
        DEFAULT: 'var(--wa-line)',
        strong: 'var(--wa-line-strong)'
      }
    },
    font: {
      sans: 'var(--wa-font-sans)',
      mono: 'var(--wa-font-mono)'
    },
    shadow: {
      card: '0 1px 2px rgba(0,0,0,.04), 0 8px 24px -12px rgba(0,0,0,.12)',
      'card-hover': '0 1px 2px rgba(0,0,0,.04), 0 16px 40px -16px rgba(0,0,0,.18)',
      brand: '0 1px 2px rgba(10,122,255,.3), 0 8px 24px -8px rgba(10,122,255,.5)',
      glow: '0 0 0 1px rgba(10,122,255,.25), 0 12px 40px -12px rgba(10,122,255,.45)'
    },
    animation: {
      keyframes: {
        'scroll-up': '{from{transform:translateY(0)}to{transform:translateY(-50%)}}',
        'scroll-down': '{from{transform:translateY(-50%)}to{transform:translateY(0)}}',
        marquee: '{from{transform:translateX(0)}to{transform:translateX(-50%)}}',
        float: '{0%,100%{transform:translateY(0)}50%{transform:translateY(-8px)}}',
        dash: '{to{stroke-dashoffset:-24}}',
        'pulse-ring': '{0%{transform:scale(1);opacity:.6}100%{transform:scale(2.2);opacity:0}}'
      },
      durations: {
        'scroll-up': '40s',
        'scroll-down': '36s',
        marquee: '45s',
        float: '6s',
        dash: '1.2s',
        'pulse-ring': '1.8s'
      },
      timingFns: {
        'scroll-up': 'linear',
        'scroll-down': 'linear',
        marquee: 'linear',
        float: 'ease-in-out',
        dash: 'linear',
        'pulse-ring': 'ease-out'
      },
      counts: {
        'scroll-up': 'infinite',
        'scroll-down': 'infinite',
        marquee: 'infinite',
        float: 'infinite',
        dash: 'infinite',
        'pulse-ring': 'infinite'
      }
    }
  }
})
