<script>
import CopyWithButton from "./CopyWithButton.vue";

export default {
  name: "ImageThumbnail",
  components: {
    CopyWithButton,
  },
  props: {
    url: String,
    caption: String,
    credits: Array,
  },
  computed: {
    correctedCredits() {
      return this.credits.map((credit) =>
        credit.replace(/\b(staff( photographer)?)\b/gi, "Philadelphia Inquirer")
      );
    },
  },
};
</script>

<template>
  <figure>
    <div class="image max-256 has-background-grey-lighter has-margin-bottom">
      <a :href="url" target="_blank" download><img :src="url" /></a>
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
      <CopyWithButton :value="caption" label="caption"></CopyWithButton>
      <p class="has-margin-bottom-thin">
        <strong>Credit:</strong>
      </p>
      <CopyWithButton
        :value="correctedCredits | commaand"
        label="credit"
      ></CopyWithButton>
    </figcaption>
  </figure>
</template>

<style scoped>
.max-256 {
  max-height: 256px;
  max-width: 256px;
  min-height: 1rem;
  min-width: 1rem;
}
</style>
