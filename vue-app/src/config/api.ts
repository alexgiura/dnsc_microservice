/** API base URL – setează VITE_API_URL în .env */
export const apiBaseUrl =
  (import.meta.env.VITE_API_URL as string)?.replace(/\/$/, '') || 'http://178.62.245.152:8080'
