/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["../internal/server/templates/*.{templ,go}"],
  theme: {
    extend: {},
    fontFamily: {
      sans: ["'Motiva Sans'","'Helvetica Neue'", "Helvetica", "Arial", "sans-serif"],
  },},
  plugins: [],
}

