<script>
export default {
  props: {
    block: Object
  },
  computed: {
    credits() {
      return this.block.credits.by.map(v => v.name);
    }
  }
};
</script>

<template>
  <figure>
    <h2 class="title">Inline Image</h2>
    <div class="image is-128x128">
      <a :href="block.url" target="_blank" download><img :src="block.url"/></a>
    </div>
    <figcaption>
      <p><strong>Caption:</strong> {{ block.caption }}</p>
      <p><strong>Credit:</strong> {{ credits | commaand }}</p>
    </figcaption>
    <a
      :href="block.url"
      class="button is-danger has-text-weight-semibold"
      target="_blank"
      download
    >
      <span class="icon">
        <font-awesome-icon :icon="['fas', 'file-download']" />
      </span>
      <span> Download</span>
    </a>
    <details>
      <pre class="code">{{ block | json }}</pre>
    </details>
  </figure>
</template>
