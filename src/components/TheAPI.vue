<script>
import { Article } from "./APIArticle.js";

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
      return this.feed && Article.from(this.feed);
    },
  },
  methods: {
    getByID(id) {
      return this.contents.find(article => article.id === id);
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
