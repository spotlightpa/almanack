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
    <h2 class="subtitle is-4">Embed #{{ n }}: Raw HTML</h2>
    <CopyWithButton
      :value="block.content"
      :rows="4"
      size="is-small"
      label="Code"
    ></CopyWithButton>
  </figure>
</template>
