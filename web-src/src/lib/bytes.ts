const BYTE_SIZE_UNITS = ["Bytes", "KB", "MB", "GB", "TB", "PB"];

export function convertBytes(bytes: number | null | undefined) {
  if (bytes == null) return "N/A";
  let count = 0;
  let num = bytes;
  while (Math.floor(num) >= 1000 && count < BYTE_SIZE_UNITS.length - 1) {
    num = num / 1000;
    count++;
  }
  if (count === 0) {
    return `${num} ${BYTE_SIZE_UNITS[0]}`;
  }
  return `${num.toFixed(2)} ${BYTE_SIZE_UNITS[count]}`;
}
