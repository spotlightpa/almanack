<script>
import { watchAPI } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";
import PageListItem from "@/api/spotlightpa-page-list-item.js";
import { useRoute } from "vue-router";

export default {
  props: { page: { default: "" } },
  setup(props) {
    const route = useRoute();

    let { listPages } = useClient();
    const { apiState, fetch, computedList, computedProp } = watchAPI(
      () => [props.page, route.meta.contentPath],
      ([page, path]) =>
        listPages({
          params: { page, path },
        })
    );

    return {
      apiState,
      fetch,
      title: route.meta.title,
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
    <title>{{ title }} • Spotlight PA Almanack</title>
  </MetaHead>

  <PageList
    :title="title"
    :page="page"
    :next-page="nextPage"
    :api-state="apiState"
    :reload="fetch"
    :pages="pages"
  ></PageList>
</template>
