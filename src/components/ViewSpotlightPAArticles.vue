<script>
import { reactive, computed, toRefs } from "@vue/composition-api";

import { useClient, makeState } from "@/api/hooks.js";
import commaAnd from "@/utils/comma-and.js";
import { formatDate } from "@/utils/time-format.js";
import fuzzyMatch from "@/utils/fuzzy-match.js";

import APILoader from "./APILoader.vue";

const articleProps = (article) => [
  article.internal_id,
  article.hed,
  ...article.authors,
];

export default {
  name: "ViewSpotlightPAArticles",
  components: {
    APILoader,
  },
  metaInfo: {
    title: "Spotlight PA Articles",
  },
  setup() {
    let { listSpotlightPAArticles } = useClient();
    let { apiState, exec } = makeState();

    let state = reactive({
      articles: computed(() =>
        apiState.rawData ? apiState.rawData.articles : []
      ),
      rawFilter: "",

      articleProps: computed(() =>
        Array.from(
          new Set(state.articles.flatMap((article) => articleProps(article)))
        ).sort()
      ),

      filteredArticles: computed(() => {
        if (!state.rawFilter) {
          return state.articles;
        }
        return state.articles.filter((article) =>
          articleProps(article).some((prop) =>
            fuzzyMatch(prop, state.rawFilter)
          )
        );
      }),
      filterOptions: computed(() =>
        state.articleProps.filter((prop) => fuzzyMatch(prop, state.rawFilter))
      ),
    });

    const fetch = () => exec(listSpotlightPAArticles);

    fetch();

    return {
      ...toRefs(apiState),
      ...toRefs(state),
      fetch,
      commaAnd,
      formatDate,
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

    <APILoader :is-loading="isLoading" :reload="fetch" :error="error">
      <b-field label="">
        <b-autocomplete
          v-model="rawFilter"
          :data="filterOptions"
          placeholder="Filter articles"
          clearable
        >
          <template slot="empty">No results found</template>
        </b-autocomplete>
      </b-field>

      <table class="table is-striped is-fullwidth">
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
            <td>{{ commaAnd(article.authors) }}</td>
            <td>{{ formatDate(article.pub_date) }}</td>
          </tr>
        </tbody>
      </table>
    </APILoader>
  </div>
</template>
