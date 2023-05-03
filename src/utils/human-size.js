export default function humanSize(size) {
  if (size < 1024) {
    return `${size} B`;
  }
  const units = ["B", "KB", "MB", "GB", "TB", "PB"];
  let unit;
  for (unit of units) {
    if (size < 1024) {
      break;
    }
    size /= 1024;
  }
  if (size < 10) {
    return `${size.toFixed(1)} ${unit}`;
  }
  return `${size.toFixed(0)} ${unit}`;
}
