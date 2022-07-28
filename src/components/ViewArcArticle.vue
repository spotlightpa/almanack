<script>
import { useAvailableArc } from "@/api/hooks.js";

export default {
  props: {
    id: String,
  },
  setup(props) {
    let { article, isLoading, load, error } = useAvailableArc(props.id);

    return {
      isLoading,
      load,
      error,
      article,
    };
  },
};
</script>

<template>
  <MetaHead>
    <title>Article • Spotlight PA</title>
  </MetaHead>
  <APILoader :is-loading="isLoading" :reload="load" :error="error">
    <template v-if="article">
      <MetaHead>
        <title>{{ article.slug }} • Spotlight PA</title>
      </MetaHead>

      <ArcArticleAvailable v-if="article.isAvailable" :article="article" />
      <ArcArticlePlanned v-else :article="article" />
    </template>
  </APILoader>
</template>
