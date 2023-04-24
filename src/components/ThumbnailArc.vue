<script>
import commaAnd from "@/utils/comma-and.js";
export default {
  name: "ImageThumbnail",
  props: {
    url: String,
    caption: String,
    credits: Array,
  },
  computed: {
    correctedCredits() {
      return commaAnd(
        this.credits.map((credit) =>
          credit.replace(
            /\b(staff( photographer)?)\b/gi,
            "Philadelphia Inquirer"
          )
        )
      );
    },
  },
};
</script>

<template>
  <ThumbnailImage
    :image-url="url"
    :download-url="url"
    :caption="caption"
    :credit="correctedCredits"
  />
</template>
