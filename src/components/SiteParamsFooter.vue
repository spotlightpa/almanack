<script setup>
import { ref } from "vue";

import adImg from "@/assets/img/ad-footer.png";

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
    <summary class="title is-4">Footer ad</summary>
    <SiteParamsImageSet
      ref="desktopImageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="true"
      prop-name="ad-footer-desktop"
      label="Footer desktop ad is a rectangular ad at the bottom of the page"
      text="Show footer ad on desktop"
    />
    <SiteParamsImageSet
      ref="mobileImageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="true"
      prop-name="ad-footer-mobile"
      label="Footer mobile ad is a square ad at the bottom of the page"
      text="Show footer ad on mobile"
    />
    <img :src="adImg" class="ad-demo" style="" />
  </details>
</template>
