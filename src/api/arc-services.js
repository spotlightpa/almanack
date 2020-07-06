import { toRefs, computed, ref, watch } from "@vue/composition-api";

import ArcArticle from "./arc-article.js";
import { makeState } from "./service-util.js";
import { useClient } from "./client.js";

export function useListAvailableArc() {
  let { listAvailableArc } = useClient();
  let { apiState, exec } = makeState();

  const load = () => exec(listAvailableArc);
  load();

  return {
    ...toRefs(apiState),
    load,

    articles: computed(() =>
      apiState.isLoading || apiState.error || !apiState.rawData
        ? []
        : ArcArticle.from(apiState.rawData)
    ),
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

export function useListAnyArc(page) {
  let { listAnyArc, listRefreshArc } = useClient();
  let { apiState, exec } = makeState();

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

  if (!page) {
    actions.loadAndRefresh();
  } else {
    actions.load();
  }

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
