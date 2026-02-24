<script setup>
import { computed } from "vue";
import commaAnd from "@/utils/comma-and.js";

const props = defineProps({
  url: String,
  caption: String,
  credits: Array,
});

const correctedCredits = computed(() =>
  commaAnd(
    props.credits.map((credit) =>
      credit.replace(
        /\b(staff( photographer)?)\b/gi,
        "Philadelphia Inquirer"
      )
    )
  )
);
</script>

<template>
  <ThumbnailImage
    :image-url="url"
    :download-url="url"
    :caption="caption"
    :credit="correctedCredits"
  ></ThumbnailImage>
</template>
