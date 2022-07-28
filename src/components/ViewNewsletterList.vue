<script>
import { watchAPI } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";
import PageListItem from "@/api/spotlightpa-page-list-item.js";

export default {
  props: { page: { default: "" } },
  setup(props) {
    let { listPages } = useClient();
    const { apiState, fetch, computer } = watchAPI(
      () => props.page,
      (page) =>
        listPages({
          params: { page, path: "content/newsletters/" },
        })
    );

    return {
      apiState,
      fetch,
      pages: computer((rawData) => (rawData ? PageListItem.from(rawData) : [])),
      nextPage: computer((rawData) => {
        let page = rawData?.next_page;
        if (!page) return null;
        return {
          name: "newsletters",
          query: { page },
        };
      }),
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
  />
</template>
