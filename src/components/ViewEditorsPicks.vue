<script>
import { reactive, computed, toRefs } from "@vue/composition-api";

import { useClient } from "@/api/hooks.js";
import fuzzyMatch from "@/utils/fuzzy-match.js";

import EditorsPicks from "./EditorsPicks.vue";

export default {
  name: "ViewEditorsPicks",
  components: {
    EditorsPicks,
  },
  metaInfo: {
    title: "Homepage Editor",
  },
  setup() {
    const toArticle = (a) => ({
      ...a,
      filterableProps: `${a.internal_id} ${a.hed} ${a.authors.join(" ")}`,
    });

    let state = reactive({
      isLoading: false,
      articles: [],
      rawFilter: "",
      error: null,

      filterOptions: computed(() => {
        if (!state.rawFilter) {
          return state.articles;
        }
        return state.articles.filter((article) =>
          fuzzyMatch(article.display, state.rawFilter)
        );
      }),
    });

    let { listSpotlightPAArticles } = useClient();

    async function fetch() {
      state.isLoading = true;
      let data;
      [data, state.error] = await listSpotlightPAArticles();
      state.isLoading = false;
      if (state.error) {
        return;
      }
      state.articles = data.articles.map(toArticle);
    }

    fetch();

    return {
      ...toRefs(state),
      fetch,
      filterTags(text) {
        state.rawFilter = text;
      },
      tags: [],
    };
  },
};
</script>

<template>
  <div>
    <nav class="breadcrumb has-succeeds-separator" aria-label="breadcrumbs">
      <ul>
        <li>
          <router-link :to="{ name: 'admin' }">Admin</router-link>
        </li>
        <li class="is-active">
          <router-link exact :to="{ name: 'editors-picks' }">
            Homepage Editor
          </router-link>
        </li>
      </ul>
    </nav>

    <h1 class="title">
      Homepage Editor
    </h1>

    <details class="content">
      <summary>Help</summary>

      <p>Lorem ipsum</p>
    </details>

    <EditorsPicks :articles="articles" />

    <div class="buttons">
      <button type="button" class="button is-primary has-text-weight-semibold">
        Save
      </button>
      <button type="button" class="button is-light has-text-weight-semibold">
        Revert
      </button>
    </div>
  </div>
</template>
