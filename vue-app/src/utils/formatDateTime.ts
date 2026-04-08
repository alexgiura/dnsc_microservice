import { appTimeZone } from '@/config/timezone'

/** Format API ISO timestamps in the app timezone (not the browser's local zone). */
export function formatDateTimeInAppTz(iso: string): string {
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString('ro-RO', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    timeZone: appTimeZone,
  })
}
