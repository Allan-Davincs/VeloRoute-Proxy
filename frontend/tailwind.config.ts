/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        background: '#0f1117',
        card: '#1a1d27',
        border: '#2a2d3e',
        primary: '#6366f1',
        success: '#22c55e',
        error: '#ef4444',
        warning: '#f59e0b',
        'text-primary': '#f1f5f9',
        'text-muted': '#64748b',
      },
      fontFamily: {
        sans: ['Inter', 'system-ui', 'sans-serif'],
        mono: ['JetBrains Mono', 'monospace'],
      },
    },
  },
  plugins: [],
}
