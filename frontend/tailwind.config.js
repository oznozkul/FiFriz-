/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      fontFamily: {
        sans: ["Inter", "ui-sans-serif", "system-ui", "sans-serif"],
      },
      colors: {
        ink: "#111827",
        muted: "#64748b",
        line: "#edf0f5",
        violet: {
          brand: "#7367f0",
          hover: "#6758e8",
          soft: "#f1efff",
          ring: "#e4e0ff",
        },
      },
      boxShadow: {
        panel: "0 16px 40px rgba(15, 23, 42, 0.055)",
        lift: "0 18px 44px rgba(15, 23, 42, 0.09)",
      },
    },
  },
  plugins: [],
};
