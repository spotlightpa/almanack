<script setup>
import { ref } from "vue";
import adImg from "@/assets/img/ad-rail.png";

defineProps({
  params: Object,
  fileProps: Object,
  open: Boolean,
  hideDemo: Boolean,
});

const imageSet = ref(null);

defineExpose({
  saveData() {
    return imageSet.value.saveData();
  },
});
</script>

<template>
  <details class="mt-4" :open="open || undefined">
    <summary class="title is-4">Rail top ad</summary>
    <SiteParamsImageSet
      ref="imageSet"
      :params="params"
      :file-props="fileProps"
      :show-width-height="false"
      :fixed-width="640"
      :fixed-height="535"
      prop-name="ad-rail-top"
      label="Rail top ad is 640x535 square at the top of the rail sidebar"
      text="Show sidebar top ad"
    />
    <img v-if="!hideDemo" :src="adImg" class="ad-demo" style="" />
  </details>
</template>
