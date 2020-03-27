<script>
import { reactive, computed, toRefs } from "@vue/composition-api";

import APILoader from "./APILoader.vue";

import { useClient } from "@/api/hooks.js";

export default {
  name: "ViewSpotlightPAArticles",
  components: {
    APILoader,
  },
  metaInfo: {
    title: "Spotlight PA Articles",
  },
  setup() {
    const articleProps = (article) => [
      article.internal_id,
      article.hed,
      ...article.authors,
    ];

    const fuzzyMatch = (str, substr) =>
      str.toLowerCase().indexOf(substr.toLowerCase()) >= 0;

    let state = reactive({
      isLoading: false,
      articles: [],
      filter: "",
      rawFilter: "",
      error: null,

      articleProps: computed(() =>
        Array.from(
          new Set(state.articles.flatMap((article) => articleProps(article)))
        ).sort()
      ),

      filteredArticles: computed(() => {
        if (!state.filter) {
          return state.articles;
        }
        return state.articles.filter((article) =>
          articleProps(article).some((prop) => fuzzyMatch(prop, state.filter))
        );
      }),
      filterOptions: computed(() =>
        state.articleProps.filter((prop) => fuzzyMatch(prop, state.rawFilter))
      ),
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
      state.articles = data.articles;
    }

    fetch();

    return {
      ...toRefs(state),
      fetch,
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
          <router-link exact :to="{ name: 'spotlightpa-articles' }">
            Spotlight PA Articles
          </router-link>
        </li>
      </ul>
    </nav>

    <h1 class="title">
      Spotlight PA Articles
    </h1>

    <APILoader
      :can-load="true"
      :is-loading="isLoading"
      :reload="fetch"
      :error="error"
    >
      <b-field label="">
        <b-autocomplete
          v-model="rawFilter"
          :data="filterOptions"
          placeholder="Filter articles"
          clearable
          @input="(option) => (filter = option)"
        >
          <template slot="empty">No results found</template>
        </b-autocomplete>
      </b-field>

      <table class="table">
        <thead>
          <tr>
            <th>Slug</th>
            <th>Hed</th>
            <th>Author</th>
            <th>Date</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="article of filteredArticles" :key="article.id">
            <td>
              <router-link
                :to="{ name: 'schedule', params: { id: article.arc_id } }"
              >
                {{ article.internal_id }}
              </router-link>
            </td>
            <td>{{ article.hed }}</td>
            <td>{{ article.authors | commaand }}</td>
            <td>{{ article.pub_date | formatDate }}</td>
          </tr>
        </tbody>
      </table>
    </APILoader>
  </div>
</template>
