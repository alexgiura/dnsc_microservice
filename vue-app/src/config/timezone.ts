/** IANA timezone for UI date/time (match backend APP_TIMEZONE). */
export const appTimeZone =
  (import.meta.env.VITE_APP_TIMEZONE as string | undefined)?.trim() || 'Europe/Bucharest'
