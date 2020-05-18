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
  },
  data() {
    return {
      filterText: "",
      list1: [],
    };
  },
  computed: {
    filteredArticles() {
      if (!this.filterText) {
        return this.articles.slice(0, 10);
      }
      return this.articles
        .filter((a) => fuzzyMatch(a.filterableProps, this.filterText))
        .slice(0, 10);
    },
  },
  methods: {
    log: function (evt) {
      window.console.log(evt);
    },
    push(article) {
      this.list1.push(article);
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
        @change="log"
      >
        <a
          v-for="(article, i) of filteredArticles"
          :key="i"
          class="dropdown-item"
          @click="push(article)"
        >
          <span class="overflow">
            <b>{{ article.internal_id }}</b
            >: {{ article.hed }}
          </span>
        </a>
        <div slot="footer" class="dropdown-item">…</div>
      </draggable>
    </div>
    <div class="column is-half">
      <b-field label="Homepage featured article">
        <EditorsPicksDraggable v-model="list1" />
        Pin top story on homepage
      </b-field>
      <b-field label="Subfeatures stories">
        <EditorsPicksDraggable v-model="list1" />

        Bulleted items under the top story
      </b-field>
      <label class="checkbox">
        <input :value="true" type="checkbox" />
        Limit the number of subfeatured stories
      </label>
      <b-field>
        <b-numberinput
          class="has-margin-top-thin"
          :value="2"
          controls-position="compact"
          type="is-light"
        ></b-numberinput>
        Subfeatured story limit
      </b-field>
      <b-field label="Pinned stories">
        <EditorsPicksDraggable v-model="list1" />
        Pin stories at the top of homepage
      </b-field>

      <p>1: {{ list1 }}</p>
    </div>
  </div>
</template>

<style scoped>
.overflow {
  text-overflow: ellipsis;
  overflow-x: hidden;
  display: block;
}
</style>
