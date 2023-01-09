<script>
import { watchAPI } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";
import PageListItem from "@/api/spotlightpa-page-list-item.js";

export default {
  props: { page: { default: "" } },
  setup(props) {
    let { listPages } = useClient();
    const { apiState, fetch, computedList, computedProp } = watchAPI(
      () => props.page,
      (page) =>
        listPages({
          params: { page, path: "content/news/" },
        })
    );

    return {
      apiState,
      fetch,
      pages: computedList("pages", (page) => new PageListItem(page)),
      nextPage: computedProp("next_page", (page) => ({
        name: "news-pages",
        query: { page },
      })),
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>Spotlight PA News Pages • Spotlight PA Almanack</title>
  </MetaHead>

  <PageList
    title="Spotlight PA News Pages"
    :page="page"
    :next-page="nextPage"
    :api-state="apiState"
    :reload="fetch"
    :pages="pages"
  />
</template>
