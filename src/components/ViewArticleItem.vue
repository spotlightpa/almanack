<script>
import { watch } from "@vue/composition-api";

import { getAvailableArticle } from "@/api/hooks.js";
import APILoader from "./APILoader.vue";
import ArticleDetails from "./ArticleDetails.vue";

export default {
  name: "ViewArticleItem",
  components: {
    APILoader,
    ArticleDetails,
  },
  props: {
    id: String,
  },
  setup(props) {
    let { article, canLoad, isLoading, fetch, error } = getAvailableArticle(
      props.id
    );

    watch(() => {
      if (article.value?.slug) {
        document.title = `Spotlight PA Almanack - ${article.value?.slug}`;
      }
    });

    return {
      canLoad,
      isLoading,
      fetch,
      error,
      article,
    };
  },
};
</script>

<template>
  <APILoader
    :can-load="canLoad"
    :is-loading="isLoading"
    :reload="fetch"
    :error="error"
  >
    <div v-if="!article" class="message is-warning">
      <p class="message-header">
        Not found
      </p>
      <p class="message-body">
        Article not found.
        <router-link :to="{ name: 'home' }">Go home</router-link>?
      </p>
    </div>
    <ArticleDetails v-else :article="article"></ArticleDetails>
  </APILoader>
</template>
