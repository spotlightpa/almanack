export default function maybeDate(obj, pathStr = "") {
  let d = obj;
  for (let prop of pathStr.split(".")) {
    if (!d) {
      break;
    }
    d = d[prop];
  }
  return d ? new Date(d) : null;
}
