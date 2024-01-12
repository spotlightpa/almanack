<script>
import draggable from "vuedraggable";

import { useClient } from "@/api/client.js";

import { Page } from "@/api/spotlightpa-page.js";

const { listPagesByFTS } = useClient();

export default {
  components: {
    draggable,
  },
  data() {
    return {
      query: "",
      rawPages: [],
      error: null,
      loading: false,
      timeout: null,
      currentRequest: 0,
    };
  },
  computed: {
    pages() {
      return this.rawPages.map((data) => new Page(data));
    },
  },
  watch: {
    async query(val) {
      this.loading = true;
      window.clearTimeout(this.timeout);
      this.timeout = window.setTimeout(() => {
        this.fetch(val);
      }, 300);
    },
  },
  created() {
    this.loading = true;
    this.fetch(this.query);
  },
  methods: {
    async fetch(query) {
      let token = ++this.currentRequest;
      const select = "-body";
      let [data, err] = await listPagesByFTS({ params: { query, select } });
      if (this.currentRequest !== token) {
        // eslint-disable-next-line no-console
        console.warn("discarding stale result");
        return;
      }
      this.loading = false;
      if (err) {
        this.error = err;
      } else {
        this.error = null;
        this.rawPages = data ?? [];
      }
    },
  },
};
</script>

<template>
  <div>
    <BulmaFieldInput
      v-model="query"
      placeholder="Filter pages"
      label="Spotlight PA Pages"
    />
    <draggable
      class="dropdown-content"
      :model-value="pages"
      item-key="id"
      :sort="false"
      :group="{ name: 'articles', pull: 'clone', put: false }"
      ghost-class="is-info"
      chosen-class="is-active"
    >
      <template #item="{ element: article }">
        <a
          class="dropdown-item select-none"
          @click="$emit('select-page', article)"
        >
          <span class="overflow">
            <b>{{ article.internalID }}</b
            >: {{ article.title }}
          </span>
        </a>
      </template>
      <template v-slot:footer>
        <div v-if="loading" class="dropdown-item">Loading…</div>
        <div v-else-if="rawPages.length === 0" class="dropdown-item">
          No results found
        </div>
        <div v-else-if="rawPages.length > 15" class="dropdown-item">
          More results hidden…
        </div>
      </template>
    </draggable>
    <ErrorSimple class="mt-4" :error="error" />
  </div>
</template>

<style scoped>
.overflow {
  text-overflow: ellipsis;
  overflow-x: hidden;
  display: block;
}
.select-none {
  cursor: grab;
  user-select: none;
}
</style>
