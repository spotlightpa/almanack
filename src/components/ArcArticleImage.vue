<script setup>
import { computed } from "vue";

const props = defineProps({
  block: Object,
  n: Number,
});

const credits = computed(() => props.block.credits?.by?.map?.((v) => v.name) ?? []);

const imageURL = computed(() => {
  let srcURL = props.block.url;
  // Some images haven't been published and can't be used
  let pubURL = props.block?.additional_properties?.resizeUrl;
  if (!srcURL.match(/\/public\//) && pubURL) {
    srcURL = pubURL;
  }
  if (!srcURL) {
    return "";
  }
  return `/api/arc-image?${new URLSearchParams({ src_url: srcURL })}`;
});
</script>

<template>
  <div class="block">
    <h2 class="subtitle is-4 has-text-weight-semibold">
      Embed #{{ n }}: Inline Image
    </h2>
    <ThumbnailArc
      :url="imageURL"
      :caption="block.caption"
      :credits="credits"
    ></ThumbnailArc>
  </div>
</template>
