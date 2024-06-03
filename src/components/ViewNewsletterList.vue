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
          params: { page, path: "content/newsletters/" },
        })
    );

    return {
      apiState,
      fetch,
      pages: computedList("pages", (page) => new PageListItem(page)),
      nextPage: computedProp("next_page", (page) => ({
        name: "newsletters",
        query: { page },
      })),
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>Newsletter Pages • Spotlight PA</title>
  </MetaHead>
  <PageList
    title="Newsletter Pages"
    :page="page"
    :next-page="nextPage"
    :api-state="apiState"
    :reload="fetch"
    :pages="pages"
  ></PageList>
</template>
