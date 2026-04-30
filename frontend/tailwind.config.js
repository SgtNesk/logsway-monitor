/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,jsx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: [
          'Inter',
          '-apple-system',
          'BlinkMacSystemFont',
          'Segoe UI',
          'sans-serif',
        ],
      },
      colors: {
        healthy: '#22c55e',
        warning: '#eab308',
        critical: '#ef4444',
        offline: '#9ca3af',
      },
    },
  },
  plugins: [],
}
