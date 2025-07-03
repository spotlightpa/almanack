<script setup>
import { ref } from "vue";

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
    <summary class="title is-4">Homepage headwater ad</summary>
    <SiteParamsImageSet
      ref="desktopImageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="true"
      prop-name="ad-hp-headwater-desktop"
      label="Desktop headerwater ad is a rectangular ad before the homepage river"
      text="Show headwater ad on desktop"
    />
    <SiteParamsImageSet
      ref="mobileImageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="true"
      prop-name="ad-hp-headwater-mobile"
      label="Mobile headwater ad is a square ad before the homepage river"
      text="Show headwater ad on mobile"
    />
  </details>
</template>
