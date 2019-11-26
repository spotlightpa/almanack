<script>
export default {
  props: {
    service: { type: Object, required: true }
  },
  data() {
    return { loading: true, feed: null, error: null };
  },
  computed: {
    contents() {
      if (this.loading || this.error) {
        return [];
      }
      return this.feed && this.feed.contents;
    }
  },
  methods: {
    async load() {
      if (!this.loading) {
        return;
      }
      [this.feed, this.error] = await this.service.upcoming();
      this.loading = false;
    },
    async reload() {
      this.loading = true;
      [this.feed, this.error] = await this.service.upcoming();
      this.loading = false;
    }
  }
};
</script>
