<script setup>
import { ref } from "vue";

import adImg from "@/assets/img/ad-featured.png";

defineProps({ params: Object, fileProps: Object });

const imageSet = ref(null);

defineExpose({
  saveData() {
    return imageSet.value.saveData();
  },
});
</script>

<template>
  <details class="mt-4">
    <summary class="title is-4">Homepage featured ad</summary>
    <SiteParamsImageSet
      ref="imageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="false"
      prop-name="ad-hp-featured"
      label="Featured ad is 600x500 square near the top of the homepage"
      text="Show homepage featured ad"
    />
    <img :src="adImg" class="ad-demo" style="" />
  </details>
</template>
