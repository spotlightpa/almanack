<script>
import CopyWithButton from "./CopyWithButton.vue";

export default {
  components: {
    CopyWithButton,
  },
  props: {
    block: Object,
    n: Number,
  },
};
</script>

<template>
  <figure class="block">
    <h2 class="subtitle is-4 has-text-weight-semibold">
      Embed #{{ n }}: Raw HTML
    </h2>
    <CopyWithButton
      :value="block.content"
      :rows="4"
      size="is-small is-clipped"
      label="Code"
    />
  </figure>
</template>
