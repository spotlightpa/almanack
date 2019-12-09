<script>
export default {
  name: "TheAPI",
  props: {
    createAPIService: { type: Function, required: true },
  },
  data() {
    return { loading: true, feed: null, error: null };
  },
  computed: {
    service() {
      return this.createAPIService(this.$auth);
    },
    contents() {
      if (this.loading || this.error) {
        return [];
      }
      return (
        this.feed &&
        Array.from(this.feed.contents).sort(
          (a, b) =>
            a.planning.scheduling.planned_publish_date <=
            b.planning.scheduling.planned_publish_date
        )
      );
    },
  },
  methods: {
    getByID(id) {
      return this.contents.find(article => article._id === id);
    },
    getBySlug(slug) {
      return this.contents.find(article => article.slug === slug);
    },
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
    },
  },
};
</script>
