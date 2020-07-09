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
      let srcURL = this.block.url;
      // Some images haven't been published and can't be used
      if (!this.block.url.match(/\/public\//)) {
        srcURL = this.block.additional_properties.resizeUrl;
      }
      if (!srcURL) {
        return "";
      }
      return `/api/proxy-image/${window.btoa(srcURL)}`;
    },
  },
};
</script>

<template>
  <div class="block">
    <h2 class="subtitle is-4 has-text-weight-semibold">
      Embed #{{ n }}: Inline Image
    </h2>
    <ImageThumbnail
      :url="imageURL"
      :caption="block.caption"
      :credits="credits"
    />
  </div>
</template>
