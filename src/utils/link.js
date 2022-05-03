export function toAbs(relurl) {
  if (!relurl) {
    return "";
  }
  return new URL(relurl, "https://www.spotlightpa.org").href;
}

export function toRel(url) {
  if (!url) {
    return "";
  }
  let u;
  try {
    u = new URL(url);
  } catch (e) {
    return url;
  }
  if (
    u.hostname === "www.spotlightpa.org" ||
    u.hostname === "spotlightpa.org"
  ) {
    return u.pathname;
  }
  return url;
}
