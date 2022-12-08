<script>
import { useClient } from "@/api/hooks.js";
import { watchAPI } from "@/api/service-util.js";
import SharedArticle from "@/api/shared-article.js";

export default {
  props: {
    id: String,
  },

  setup(props) {
    let { getSharedArticle } = useClient();
    const { apiState, fetch, computer } = watchAPI(
      () => props.id,
      (id) =>
        getSharedArticle({
          params: { id },
        })
    );

    return {
      apiState,
      fetch,
      article: computer((rawData) =>
        rawData ? new SharedArticle(rawData) : null
      ),
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>Article • Spotlight PA</title>
  </MetaHead>
  <APILoader
    :is-loading="apiState.isLoading.value"
    :reload="fetch"
    :error="apiState.error.value"
  >
    <template v-if="article">
      <MetaHead>
        <title>{{ article.slug }} • Spotlight PA</title>
      </MetaHead>

      <ArcArticleAvailable v-if="article.isAvailable" :article="article" />
      <ArcArticlePlanned v-else :article="article" />
    </template>
  </APILoader>
</template>
