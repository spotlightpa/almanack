<script>
import BulmaCopyInput from "./BulmaCopyInput.vue";

export default {
  components: {
    BulmaCopyInput,
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
    <BulmaCopyInput
      :value="block.content"
      :rows="4"
      size="is-small"
      label="Code"
    ></BulmaCopyInput>
  </figure>
</template>
