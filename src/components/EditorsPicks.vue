<script>
import draggable from "vuedraggable";

import fuzzyMatch from "@/utils/fuzzy-match.js";

export default {
  name: "EditorsPicks",
  components: {
    draggable,
  },
  props: {
    pages: Array,
    editorsPicks: Object,
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
  methods: {
    push(article) {
      this.editorsPicks.featuredStories.push(article);
    },
  },
};
</script>

<template>
  <div class="columns">
    <div class="column is-half">
      <b-field label="Recent Spotlight PA Articles">
        <input
          v-model="filterText"
          class="input"
          placeholder="Filter articles"
        />
      </b-field>

      <draggable
        class="dropdown-content"
        :value="filteredPages"
        :sort="false"
        :group="{ name: 'articles', pull: 'clone', put: false }"
        ghost-class="is-info"
        chosen-class="is-active"
      >
        <a
          v-for="(article, i) of filteredPages.slice(0, 10)"
          :key="i"
          class="dropdown-item select-none"
          @click="push(article)"
        >
          <span class="overflow">
            <b>{{ article.internalID }}</b
            >: {{ article.hed }}
          </span>
        </a>
        <template v-slot:footer>
          <div v-if="filteredPages.length > 10" class="dropdown-item">
            More results hidden…
          </div>
          <div v-if="filteredPages.length === 0" class="dropdown-item">
            No results
          </div>
        </template>
      </draggable>
    </div>
    <div class="column is-half">
      <b-field label="Homepage featured article">
        <EditorsPicksDraggable v-model="editorsPicks.featuredStories" />
        Pin top story on homepage
      </b-field>
      <b-field label="Subfeatures stories">
        <EditorsPicksDraggable v-model="editorsPicks.subfeatures" />

        Bulleted items under the top story
      </b-field>
      <b-field label="Pinned stories">
        <EditorsPicksDraggable v-model="editorsPicks.topSlots" />
        Pin stories at the top of homepage
      </b-field>
      <b-field label="Editor's Pick sidebar">
        <EditorsPicksDraggable v-model="editorsPicks.sidebarPicks" />
        Pin story in the homepage sidebar under Most Popular
      </b-field>
    </div>
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
