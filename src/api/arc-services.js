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

export function useListAnyArc() {
  let { listAnyArc, listRefreshArc } = useClient();
  let { apiState, exec } = makeState();
  let page = "0";

  const actions = {
    load: () => exec(() => listAnyArc(page)),
    loadAndRefresh: () => exec(listRefreshArc),
    loadNextPage: () => {
      page = apiState.rawData.next_page;
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

  actions.loadAndRefresh();

  return {
    ...toRefs(apiState),
    ...actions,
    articles,

    hasNextPage: computed(() =>
      apiState.isLoading || apiState.error || !apiState.rawData
        ? false
        : !!apiState.rawData.next_page
    ),
  };
}
