<script setup>
import { computed } from "vue";

import imgproxyURL from "@/api/imgproxy-url.js";

const props = defineProps({
  path: { type: String, default: "" },
  caption: { type: String, default: "" },
  credits: { type: String, default: "" },
});

const src = computed(() =>
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
  <figure>
    <div class="has-margin-bottom">
      <a
        :href="downloadURL"
        class="is-inline-flex max-256 has-background-grey-lighter"
        target="_blank"
        download
      >
        <img class="max-256" :src="src" width="256" height="192" />
      </a>
    </div>
    <figcaption>
      <p class="has-margin-bottom">
        <a
          :href="downloadURL"
          class="button is-danger has-text-weight-semibold"
          target="_blank"
          download
        >
          <span class="icon">
            <font-awesome-icon :icon="['fas', 'file-download']" />
          </span>
          <span>Download image</span>
        </a>
      </p>
      <p class="has-margin-bottom-thin">
        <strong>Caption:</strong>
      </p>
      <CopyWithButton :value="caption" label="caption" />
      <p class="has-margin-bottom-thin">
        <strong>Credit:</strong>
      </p>
      <CopyWithButton :value="credits" label="credit" />
    </figcaption>
  </figure>
</template>

<style scoped>
.max-256 {
  width: auto;
  height: auto;
  max-height: 256px;
  max-width: 256px;
  min-height: 1rem;
  min-width: 1rem;
}
</style>
