export default function sanitizeText(text: string | null | undefined): string {
  const el = document.createElement("div");
  el.innerText = text ?? "";
  let html = el.innerHTML;
  html = html
    .replace(/&lt;strong&gt;/g, "<strong>")
    .replace(/&lt;\/strong&gt;/g, "</strong>")
    .replace(/&lt;em&gt;/g, "<em>")
    .replace(/&lt;\/em&gt;/g, "</em>")
    .replace(/&lt;b&gt;/g, "<strong>")
    .replace(/&lt;\/b&gt;/g, "</strong>")
    .replace(/&lt;i&gt;/g, "<em>")
    .replace(/&lt;\/i&gt;/g, "</em>")
    .replace(/<br>/g, "\n");
  el.innerHTML = html;
  return el.innerHTML;
}
