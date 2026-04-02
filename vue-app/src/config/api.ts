/** API base URL – setează VITE_API_BASE_URL în env */
export const apiBaseUrl =
  // Dacă VITE_API_BASE_URL e gol (sau nu există), folosim apeluri relative (ex: /api/domains)
  // astfel încât Nginx (din containerul FE) să poată proxya către BE.
  (import.meta.env.VITE_API_BASE_URL as string)?.replace(/\/$/, '') || ''
