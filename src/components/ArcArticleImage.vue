<script>
import ImageThumbnail from "./ImageThumbnail.vue";

export default {
  components: {
    ImageThumbnail,
  },
  props: {
    block: Object,
    n: Number,
  },
  computed: {
    credits() {
      return this.block.credits.by.map((v) => v.name);
    },
    imageURL() {
      // Some images haven't been published and can't be used
      if (this.block.url.match(/\/public\//)) {
        return this.block.url;
      }
      return this.block.additional_properties.resizeUrl;
    },
  },
};
</script>

<template>
  <div class="block">
    <h2 class="subtitle is-4 has-text-weight-semibold">
      Embed #{{ n }}: Inline Image
    </h2>
    <ImageThumbnail :url="imageURL" :caption="block.caption" :credits="credits">
    </ImageThumbnail>
  </div>
</template>
