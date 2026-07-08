export function toAbs(relurl: string): string {
  if (!relurl || !URL.canParse(relurl, "https://www.spotlightpa.org")) {
    return relurl;
  }
  return new URL(relurl, "https://www.spotlightpa.org").href;
}

export function toRel(url: string): string {
  if (!url) {
    return "";
  }
  let u: URL;
  try {
    u = new URL(url, "https://www.spotlightpa.org");
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
