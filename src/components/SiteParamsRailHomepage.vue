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
    <summary class="title is-4">Homepage rail ad</summary>
    <SiteParamsImageSet
      ref="imageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="false"
      prop-name="ad-hp-rail"
      label="Rail ad is 640x535 rectangle at the top of the homepage river sidebar"
      text="Show rail ad"
    />
  </details>
</template>
