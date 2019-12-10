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
    <h2 class="subtitle is-4">
      Embed #{{ n }}: {{ block.raw_oembed.provider_name }} Embed
    </h2>
    <h3 class="subtitle is-5">
      Source:
      <a :href="block.raw_oembed.url" target="_blank">
        <span class="is-size-6">
          <font-awesome-icon class="" :icon="['fas', 'link']" />
        </span>
        {{ block.raw_oembed.url }}</a
      >
    </h3>
    <CopyWithButton
      :value="block.raw_oembed.html"
      :rows="4"
      size="is-small"
      label="Code"
    ></CopyWithButton>
  </figure>
</template>
