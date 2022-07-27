<script>
import draggable from "vuedraggable";

import fuzzyMatch from "@/utils/fuzzy-match.js";

export default {
  components: {
    draggable,
  },
  props: {
    pages: Array,
  },
  data() {
    return {
      filterText: "",
    };
  },
  computed: {
    filteredPages() {
      if (!this.filterText) {
        return this.pages;
      }
      return this.pages.filter((p) =>
        fuzzyMatch(p.filterableProps, this.filterText)
      );
    },
  },
  methods: {},
};
</script>

<template>
  <div>
    <BulmaField label="Spotlight PA Pages">
      <input v-model="filterText" class="input" placeholder="Filter pages" />
    </BulmaField>

    <draggable
      class="dropdown-content"
      :model-value="filteredPages"
      :sort="false"
      :group="{ name: 'articles', pull: 'clone', put: false }"
      ghost-class="is-info"
      chosen-class="is-active"
    >
      <a
        v-for="(article, i) of filteredPages.slice(0, 10)"
        :key="i"
        class="dropdown-item select-none"
        @click="$emit('select', article)"
      >
        <span class="overflow">
          <b>{{ article.internalID }}</b
          >: {{ article.hed }}
        </span>
      </a>
      <template v-slot:footer>
        <div v-if="pages.length === 0" class="dropdown-item">Loading…</div>
        <div v-else-if="filteredPages.length > 10" class="dropdown-item">
          More results hidden…
        </div>
        <div v-else-if="filteredPages.length === 0" class="dropdown-item">
          No results
        </div>
      </template>
    </draggable>
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
