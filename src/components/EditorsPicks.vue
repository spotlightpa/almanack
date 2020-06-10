<script>
import draggable from "vuedraggable";
import fuzzyMatch from "@/utils/fuzzy-match.js";
import EditorsPicksDraggable from "./EditorsPicksDraggable.vue";

export default {
  name: "EditorsPicks",
  components: {
    draggable,
    EditorsPicksDraggable,
  },
  props: {
    articles: Array,
    editorsPicks: Object,
  },
  data() {
    return {
      filterText: "",
    };
  },
  computed: {
    filteredArticles() {
      if (!this.filterText) {
        return this.articles;
      }
      return this.articles.filter((a) =>
        fuzzyMatch(a.filterableProps, this.filterText)
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
        :value="filteredArticles"
        :sort="false"
        :group="{ name: 'articles', pull: 'clone', put: false }"
        ghost-class="is-info"
        chosen-class="is-active"
      >
        <a
          v-for="(article, i) of filteredArticles.slice(0, 10)"
          :key="i"
          class="dropdown-item select-none"
          @click="push(article)"
        >
          <span class="overflow">
            <b>{{ article.internal_id }}</b
            >: {{ article.hed }}
          </span>
        </a>
        <div
          v-if="filteredArticles.length > 10"
          slot="footer"
          class="dropdown-item"
        >
          More results hidden…
        </div>
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
      <label class="checkbox">
        <input v-model="editorsPicks.limitSubfeatures" type="checkbox" />
        Limit the number of subfeatured stories
      </label>
      <b-field v-if="editorsPicks.limitSubfeatures">
        <b-numberinput
          v-model="editorsPicks.subfeaturesLimit"
          class="has-margin-top-thin"
          controls-position="compact"
          min="0"
          type="is-light"
        ></b-numberinput>
        Subfeatured story limit
      </b-field>
      <b-field label="Pinned stories">
        <EditorsPicksDraggable v-model="editorsPicks.topSlots" />
        Pin stories at the top of homepage
      </b-field>
      <b-field label="Editor's Picks sidebar">
        <EditorsPicksDraggable v-model="editorsPicks.sidebarPicks" />
        Pin stories in the homepage sidebar under Most Popular
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
