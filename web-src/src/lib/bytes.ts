const BYTE_SIZE_UNITS = ["Bytes", "KB", "MB", "GB", "TB", "PB"];

export function convertBytes(bytes: number) {
  let count = 0;
  let unit = BYTE_SIZE_UNITS[0];
  let num = bytes;
  while (num.toString().split(".")[0].length > 3) {
    num = num / 1000;
    count++;
    unit = BYTE_SIZE_UNITS[count];
  }
  if (count === 0) {
    return `${num} ${unit}`;
  }
  return `${num.toFixed(2)} ${unit}`;
}
