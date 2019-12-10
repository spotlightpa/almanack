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
      return this.block.credits.by.map(v => v.name);
    },
  },
};
</script>

<template>
  <div class="block">
    <h2 class="subtitle is-4">Embed #{{ n }}: Inline Image</h2>
    <ImageThumbnail
      :url="block.url"
      :caption="block.caption"
      :credits="credits"
    >
    </ImageThumbnail>
  </div>
</template>
