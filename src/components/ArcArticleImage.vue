<script>
export default {
  props: {
    block: Object,
    n: Number,
  },
  computed: {
    credits() {
      return this.block.credits?.by?.map?.((v) => v.name) ?? [];
    },
    imageURL() {
      let srcURL = this.block.url;
      // Some images haven't been published and can't be used
      let pubURL = this.block?.additional_properties?.resizeUrl;
      if (!srcURL.match(/\/public\//) && pubURL) {
        srcURL = pubURL;
      }
      if (!srcURL) {
        return "";
      }
      return `/api/arc-image?${new URLSearchParams({ src_url: srcURL })}`;
    },
  },
};
</script>

<template>
  <div class="block">
    <h2 class="subtitle is-4 has-text-weight-semibold">
      Embed #{{ n }}: Inline Image
    </h2>
    <ThumbnailArc
      :url="imageURL"
      :caption="block.caption"
      :credits="credits"
    ></ThumbnailArc>
  </div>
</template>
