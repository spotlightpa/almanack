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
  <figure>
    <div class="has-margin-bottom">
      <a
        :href="url"
        class="is-inline-flex max-256 has-background-grey-lighter"
        target="_blank"
        download
      >
        <img class="max-256" :src="url" />
      </a>
    </div>
    <figcaption>
      <p class="has-margin-bottom">
        <a
          :href="url"
          class="button is-danger has-text-weight-semibold"
          target="_blank"
          download
        >
          <span class="icon">
            <font-awesome-icon :icon="['fas', 'file-download']" />
          </span>
          <span>Download image</span>
        </a>
      </p>
      <p class="has-margin-bottom-thin">
        <strong>Caption:</strong>
      </p>
      <CopyWithButton :value="caption" label="caption" />
      <p class="has-margin-bottom-thin">
        <strong>Credit:</strong>
      </p>
      <CopyWithButton :value="correctedCredits" label="credit" />
    </figcaption>
  </figure>
</template>

<style scoped>
.max-256 {
  width: auto;
  height: auto;
  max-height: 256px;
  max-width: 256px;
  min-height: 1rem;
  min-width: 1rem;
}
</style>
