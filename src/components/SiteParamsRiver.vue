<script setup>
import { ref } from "vue";

import adImg from "@/assets/img/ad-river.png";

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
    <summary class="title is-4">Homepage river ad</summary>
    <SiteParamsImageSet
      ref="desktopImageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="true"
      prop-name="ad-hp-river-desktop"
      label="Desktop river ads are a rectangular ad between every five items in the homepage river"
      text="Show river ads on desktop"
    />
    <SiteParamsImageSet
      ref="mobileImageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="true"
      prop-name="ad-hp-river-mobile"
      label="Mobile river ad is a square ad between every five items in the homepage river"
      text="Show river ads on mobile"
    />
    <img :src="adImg" class="ad-demo" style="" />
  </details>
</template>
