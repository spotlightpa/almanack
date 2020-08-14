import { toRefs, computed, ref, watch } from "@vue/composition-api";

import ArcArticle from "./arc-article.js";
import { makeState } from "./service-util.js";
import { useClient } from "./client.js";

export function useListAvailableArc(pageCB) {
  let page;
  let { listAvailableArc } = useClient();
  let { apiState, exec } = makeState();

  const load = () => exec(() => listAvailableArc(page));

  watch(
    pageCB,
    (newVal) => {
      page = newVal;
      load();
    },
    {
      immediate: true,
    }
  );

  return {
    ...toRefs(apiState),
    load,

    loadPage(newPage) {
      page = newPage;
      return load();
    },
    articles: computed(() =>
      apiState.isLoading || apiState.error || !apiState.rawData
        ? []
        : ArcArticle.from(apiState.rawData)
    ),
    nextPage: computed(() => {
      if (apiState.isLoading || apiState.error || !apiState.rawData) {
        return null;
      }
      if (!apiState.rawData.next_page) {
        return null;
      }
      return {
        name: "articles",
        query: {
          page: apiState.rawData.next_page,
        },
      };
    }),
  };
}

export function useAvailableArc(id) {
  let { getAvailableArc } = useClient();
  let { apiState, exec } = makeState();

  const load = () => exec(() => getAvailableArc(id));
  load();

  return {
    ...toRefs(apiState),
    load,

    article: computed(() =>
      apiState.isLoading || apiState.error || !apiState.rawData
        ? null
        : new ArcArticle(apiState.rawData)
    ),
  };
}

export function useListAnyArc(pageCB) {
  let { listAnyArc, listRefreshArc } = useClient();
  let { apiState, exec } = makeState();

  let page;

  const actions = {
    load: () => exec(() => listAnyArc(page)),
    loadAndRefresh: () => exec(listRefreshArc),
    loadPage: (newPage) => {
      page = newPage;
      return actions.load();
    },
  };

  const articles = ref([]);
  watch(
    () => !apiState.isLoading && !apiState.error && apiState.rawData,
    (hasData) => {
      if (hasData) {
        articles.value = ArcArticle.from(apiState.rawData);
      }
    }
  );

  watch(
    pageCB,
    (newVal, oldVal) => {
      page = newVal;
      if (newVal !== oldVal) {
        articles.value = [];
        apiState.didLoad = false;
      }
      if (!page) {
        actions.loadAndRefresh();
      } else {
        actions.load();
      }
    },
    {
      immediate: true,
    }
  );

  return {
    ...toRefs(apiState),
    ...actions,
    articles,

    nextPage: computed(() => {
      if (apiState.isLoading || apiState.error || !apiState.rawData) {
        return null;
      }
      if (!apiState.rawData.next_page) {
        return null;
      }
      return {
        name: "admin",
        query: {
          page: apiState.rawData.next_page,
        },
      };
    }),
  };
}
