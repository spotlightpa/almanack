export default function maybeDate(
  obj: unknown,
  pathStr = ""
): Date | null {
  let d: unknown = obj;
  for (const prop of pathStr.split(".")) {
    if (!d || typeof d !== "object") {
      break;
    }
    d = (d as Record<string, unknown>)[prop];
  }
  return d ? new Date(d as string | number) : null;
}
