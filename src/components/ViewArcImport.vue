<script>
import { ref } from "vue";

import { get, listArcByLastUpdated } from "@/api/client-v2.js";
import { watchAPI } from "@/api/service-util.js";
import ArcArticle from "@/api/arc-article.js";

export default {
  props: ["page"],
  setup(props) {
    const first = ref(true);
    const { apiState, fetch, computer } = watchAPI(
      () => ({ page: props.page, refresh: props.page === "0" && !first.value }),
      (params) => get(listArcByLastUpdated, params)
    );

    return {
      apiState,
      fetch,
      articles: computer((rawData) => {
        if (!rawData?.stories) {
          return [];
        }
        first.value = false;
        let { stories } = rawData;
        return stories.map((a) => new ArcArticle(a.raw_data));
      }),
      nextPage: computer((rawData) => {
        let page = rawData?.next_page;
        if (!page) return null;
        return {
          name: "arc-import",
          query: { page },
        };
      }),
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>Import from Arc • Spotlight PA Almanack</title>
  </MetaHead>
  <BulmaBreadcrumbs
    :links="[
      { name: 'Admin', to: { name: 'admin' } },
      {
        name: 'Import from Arc',
        to: { name: 'arc-import', query: { page: 0 } },
      },
    ]"
  />

  <h2 class="title">Import from Arc</h2>

  <SpinnerProgress :is-loading="apiState.isLoading.value" />
  <ErrorReloader :error="apiState.error.value" @reload="fetch" />
  <ul>
    <li v-for="article of articles" :key="article.id">
      {{ article.slug }}
    </li>
  </ul>

  <div class="buttons mt-5">
    <router-link
      v-if="nextPage"
      :to="nextPage"
      class="button is-primary has-text-weight-semibold"
    >
      Show Older Stories…
    </router-link>
  </div>
</template>
