<script setup>
import { ref } from "vue";

import adImg from "@/assets/img/ad-header.png";

defineProps({ params: Object, fileProps: Object });

const desktopImageSet = ref(null);
const mobileImageSet = ref(null);

defineExpose({
  saveData() {
    return {
      ...desktopImageSet.value.saveData(),
      ...mobileImageSet.value.saveData(),
    };
  },
});
</script>

<template>
  <details class="mt-4">
    <summary class="title is-4">Header ad</summary>
    <SiteParamsImageSet
      ref="desktopImageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="true"
      prop-name="ad-header-desktop"
      label="Header desktop ad is a rectangular ad at the start of the page"
      text="Show header ad on desktop"
    />
    <SiteParamsImageSet
      ref="mobileImageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="true"
      prop-name="ad-header-mobile"
      label="Header mobile ad is a square ad at the start of the page"
      text="Show header ad on mobile"
    />
    <img :src="adImg" class="ad-demo" style="" />
  </details>
</template>
