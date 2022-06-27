<script>
import { computed, watch } from "@vue/composition-api";

import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";
import PageListItem from "@/api/spotlightpa-page-list-item.js";

export default {
  props: ["page"],
  metaInfo: {
    title: "Spotlight PA News Pages",
  },
  setup(props) {
    let { listNewsPages } = useClient();
    let { apiStateRefs, exec } = makeState();
    const { rawData } = apiStateRefs;
    const fetch = (page) => exec(() => listNewsPages(page));
    watch(() => props.page, fetch, {
      immediate: true,
    });
    return {
      apiStateRefs,
      fetch,
      pages: PageListItem.from(rawData),
      nextPage: computed(() => {
        let param = rawData.value?.next_page;
        if (!param) return null;
        return {
          name: "news-pages",
          query: {
            page: param,
          },
        };
      }),
    };
  },
};
</script>

<template>
  <PageList
    title="Spotlight PA News Pages"
    :page="page"
    :next-page="nextPage"
    :api-state="apiStateRefs"
    :reload="fetch"
    :pages="pages"
  />
</template>
