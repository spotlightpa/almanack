export default function getProp(obj, pathStr, { fallback = null } = {}) {
  for (let prop of pathStr.split(".")) {
    if (!obj) {
      break;
    }
    obj = obj[prop];
  }
  return obj ?? fallback;
}
