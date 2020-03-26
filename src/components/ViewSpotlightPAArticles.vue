<script>
import { reactive, toRefs } from "@vue/composition-api";

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
    let apiState = reactive({
      isLoading: false,
      articles: [],
      error: null,
    });

    let { listSpotlightPAArticles } = useClient();

    async function fetch() {
      apiState.isLoading = true;
      let data;
      [data, apiState.error] = await listSpotlightPAArticles();
      apiState.isLoading = false;
      if (apiState.error) {
        return;
      }
      apiState.articles = data.articles;
    }

    fetch();

    return {
      ...toRefs(apiState),
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
          <tr v-for="article of articles" :key="article.id">
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
