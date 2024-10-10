export default function imageURL(
  filepath,
  {
    width = 400,
    height = 300,
    extension = "jpeg",
    gravity = "",
    quality = 75,
  } = {}
) {
  if (!filepath) {
    return "";
  }
  let baseURL = "https://images.data.spotlightpa.org";
  let signature = "insecure";
  let resizing_type = "fill";
  let enlarge = "1";
  let encoded_source_url = btoa(filepath);

  gravity = gravity || "sm";
  width = Math.round(width);
  height = Math.round(height);
  quality = Math.round(quality);
  let qualityStr = quality ? `/q:${quality}` : "";
  return `${baseURL}/${signature}/rt:${resizing_type}/w:${width}/h:${height}/g:${gravity}/el:${enlarge}${qualityStr}/${encoded_source_url}.${extension}`;
}
