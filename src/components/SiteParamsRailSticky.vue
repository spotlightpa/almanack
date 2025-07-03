<script setup>
import { ref } from "vue";

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
    <summary class="title is-4">Rail sticky ad</summary>
    <SiteParamsImageSet
      ref="imageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="false"
      prop-name="ad-rail-sticky"
      label="Sidebar sticky ad is a 640x535 square at the bottom of the rail sidebar that sticks to the screen on scroll"
      text="Show sidebar sticky ad"
    />
  </details>
</template>
