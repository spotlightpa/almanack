<script>
export default {
  props: {
    block: Object,
    n: Number
  },
  computed: {
    credits() {
      return this.block.credits.by.map(v => v.name);
    }
  }
};
</script>

<template>
  <figure class="block">
    <h2 class="subtitle is-4">Embed #{{ n }}: Inline Image</h2>
    <div class="image is-128x128">
      <a :href="block.url" target="_blank" download><img :src="block.url"/></a>
    </div>
    <figcaption>
      <p class="has-margin-bottom-thin">
        <strong>Caption:</strong> {{ block.caption }}
      </p>
      <p class="has-margin-bottom-thin">
        <strong>Credit:</strong> {{ credits | commaand }}
      </p>
      <p>
        <a
          :href="block.url"
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
