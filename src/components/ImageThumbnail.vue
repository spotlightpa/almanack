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
};
</script>

<template>
  <figure>
    <div class="image is-128x128">
      <a :href="url" target="_blank" download><img :src="url"/></a>
    </div>
    <figcaption>
      <p class="has-margin-bottom-thin">
        <strong>Caption:</strong>
      </p>
      <CopyWithButton :value="caption" label="caption"></CopyWithButton>
      <p class="has-margin-bottom-thin">
        <strong>Credit:</strong>
      </p>
      <CopyWithButton
        :value="credits | commaand"
        label="credit"
      ></CopyWithButton>
      <p>
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
    </figcaption>
  </figure>
</template>
