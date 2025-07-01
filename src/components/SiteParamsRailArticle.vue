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
    <summary class="title is-4">Article rail top</summary>
    <SiteParamsImageSet
      ref="imageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="false"
      prop-name="ad-art-rail"
      label="Sidebar top ad is a 300x250 ad on the article rail sidebar"
      text="Show sidebar top ad"
    />
  </details>
</template>
