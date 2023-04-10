<script setup>
import { computed } from "vue";

import imgproxyURL from "@/api/imgproxy-url.js";

const props = defineProps({
  path: { type: String, default: "" },
  caption: { type: String, default: "" },
  credits: { type: String, default: "" },
});

const imageURL = computed(() =>
  imgproxyURL(props.path, {
    width: 512,
    height: (512 * 3) / 4,
    extension: "webp",
  })
);

// TODO: Fix SSR for editors
const downloadURL = computed(
  () => "/ssr/download-image?src=" + encodeURIComponent(props.path)
);
</script>

<template>
  <ThumbnailImage
    :image-url="imageURL"
    :download-url="downloadURL"
    :caption="caption"
    :credits="credits"
  />
</template>
