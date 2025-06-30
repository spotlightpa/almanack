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
      :show-width-height="false"
      prop-name="ad-hp-river-mobile"
      label="Mobile river ad is a 300x250 square between every five items in the homepage river"
      text="Show river ads on mobile"
    />
  </details>
</template>
