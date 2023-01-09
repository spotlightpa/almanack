<script setup>
import { get, getSharedArticle } from "@/api/client-v2.js";
import { watchAPI } from "@/api/service-util.js";
import SharedArticle from "@/api/shared-article.js";

const props = defineProps({
  id: String,
});

const { apiState, fetch, computedObj } = watchAPI(
  () => props.id,
  (id) => get(getSharedArticle, { id })
);
const article = computedObj((data) => new SharedArticle(data));
</script>

<template>
  <MetaHead>
    <title>Shared Article • Spotlight PA</title>
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

      <template v-if="article.isArc">
        <ArcArticleAvailable v-if="!article.isPreview" :article="article" />
        <ArcArticlePlanned v-else :article="article" />
      </template>
    </template>
  </APILoader>
</template>
