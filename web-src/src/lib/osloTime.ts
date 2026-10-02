// Helpers for converting between "datetime-local" input values (which carry no
// timezone information) and the backend's RFC3339 timestamps. All maintenance
// scheduling is interpreted as Europe/Oslo wall-clock time regardless of the
// browser's local timezone.
const osloTimeZone = 'Europe/Oslo'

export const osloDateParts = (date: Date) => {
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: osloTimeZone,
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hourCycle: 'h23',
  }).formatToParts(date).reduce<Record<string, string>>((result, part) => {
    result[part.type] = part.value
    return result
  }, {})

  return parts
}

// Converts an ISO/UTC datetime string into a "datetime-local" input value
// (format: "YYYY-MM-DDTHH:mm") expressed in Oslo wall-clock time.
export const toDatetimeLocal = (datetime: string) => {
  const parts = osloDateParts(new Date(datetime))
  return `${parts.year}-${parts.month}-${parts.day}T${parts.hour}:${parts.minute}`
}

// Interprets a "datetime-local" input value (format: "YYYY-MM-DDTHH:mm") as
// Oslo wall-clock time and returns the equivalent UTC timestamp in
// milliseconds, correctly accounting for DST.
export const osloWallClockToTimestamp = (value: string): number => {
  const [datePart, timePart] = value.split('T')
  const [year, month, day] = datePart.split('-').map(Number)
  const [hour, minute] = timePart.split(':').map(Number)

  const wallClockAsUtc = Date.UTC(year, month - 1, day, hour, minute)
  const osloParts = osloDateParts(new Date(wallClockAsUtc))
  const osloClockAsUtc = Date.UTC(
    Number(osloParts.year),
    Number(osloParts.month) - 1,
    Number(osloParts.day),
    Number(osloParts.hour),
    Number(osloParts.minute),
  )
  return wallClockAsUtc - (osloClockAsUtc - wallClockAsUtc)
}

// Converts a "datetime-local" input value, interpreted as Oslo wall-clock
// time, into a valid RFC3339 string (e.g. "2026-10-01T13:30:00.000Z") for
// the backend's strict `time.Parse(time.RFC3339, ...)` parsing.
export const toServerDateTime = (value: string): string => {
  return new Date(osloWallClockToTimestamp(value)).toISOString()
}

export const nowLabel = () => {
  const parts = osloDateParts(new Date())
  return `${parts.year}-${parts.month}-${parts.day} ${parts.hour}:${parts.minute}:${parts.second}`
}
