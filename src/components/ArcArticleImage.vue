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
      return "";
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
