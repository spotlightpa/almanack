<script setup>
import { ref } from "vue";

import adImg from "@/assets/img/ad-breaker.png";

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
    <summary class="title is-4">Homepage breaker ad</summary>
    <SiteParamsImageSet
      ref="desktopImageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="true"
      prop-name="ad-hp-breaker-desktop"
      label="Desktop breaker ad is a rectangular ad after the featured slot on the homepage"
      text="Show breaker ad on desktop"
    />
    <SiteParamsImageSet
      ref="mobileImageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="true"
      prop-name="ad-hp-breaker-mobile"
      label="Mobile breaker ad is a square ad after the featured slot on the homepage"
      text="Show breaker ad on mobile"
    />
    <img :src="adImg" class="ad-demo" style="" />
  </details>
</template>
