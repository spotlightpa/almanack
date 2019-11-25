<script>
export default {
  data() {
    return { loading: true, feed: null, error: null };
  },
  computed: {
    contents() {
      if (this.loading || this.error) {
        return [];
      }
      return this.feed.contents;
    }
  },
  async created() {
    [this.feed, this.error] = await this.$api.upcoming();
    this.loading = false;
  },
  methods: {
    async reload() {
      this.loading = true;
      [this.feed, this.error] = await this.$api.upcoming();
      this.loading = false;
    }
  }
};
</script>

<template>
  <div>Component</div>
</template>
