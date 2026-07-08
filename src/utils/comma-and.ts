export default function commaAnd(a: readonly { toString(): string }[]): string {
  if (a.length === 0) {
    return "";
  }
  const ss = a.map((item) => item.toString());
  if (ss.length < 3) {
    return ss.join(" and ");
  }
  const commas = ss.slice(0, -1).join(", ");
  return `${commas} and ${ss[ss.length - 1]}`;
}
