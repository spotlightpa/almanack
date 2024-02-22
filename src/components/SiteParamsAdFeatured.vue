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
  <details class="mt-4" open>
    <summary class="title is-4">Featured ad</summary>
    <SiteParamsImageSet
      ref="imageSet"
      :params="params"
      :file-props="fileProps"
      prop-name="ad-hp-featured"
      label="Featured ad is 250x300 square near the top of the homepage"
      text="Show featured ad on homepage"
    />
  </details>
</template>
