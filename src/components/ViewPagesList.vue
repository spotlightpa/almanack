<script>
import { watchAPI } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";
import PageListItem from "@/api/spotlightpa-page-list-item.js";
import { useRoute } from "vue-router";
import { computed } from "vue";

const CONFIG = {
  "news-pages": {
    path: "content/news/",
    title: "Spotlight PA News Pages",
  },
  "statecollege-pages": {
    path: "content/statecollege/",
    title: "State College Articles",
  },
  "berks-pages": {
    path: "content/berks/",
    title: "Berks County Articles",
  },
  "sponsored-pages": {
    path: "content/sponsored/",
    title: "Sponsored Content",
  },
};

export default {
  props: { page: { default: "" } },
  setup(props) {
    const route = useRoute();
    const config = computed(() => CONFIG[route.name] || CONFIG["news-pages"]);

    let { listPages } = useClient();
    const { apiState, fetch, computedList, computedProp } = watchAPI(
      () => [props.page, config.value.path],
      ([page, path]) =>
        listPages({
          params: { page, path },
        })
    );

    return {
      apiState,
      fetch,
      config,
      pages: computedList("pages", (page) => new PageListItem(page)),
      nextPage: computedProp("next_page", (page) => ({
        name: route.name,
        query: { page },
      })),
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>{{ config.title }} • Spotlight PA Almanack</title>
  </MetaHead>

  <PageList
    :title="config.title"
    :page="page"
    :next-page="nextPage"
    :api-state="apiState"
    :reload="fetch"
    :pages="pages"
  ></PageList>
</template>
