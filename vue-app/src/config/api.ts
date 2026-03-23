/** API base URL – setează VITE_API_BASE_URL în env */
export const apiBaseUrl =
  (import.meta.env.VITE_API_BASE_URL as string)?.replace(/\/$/, '') || 'http://localhost:8080'
