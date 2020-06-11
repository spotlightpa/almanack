<script>
import { useAvailableArc } from "@/api/hooks.js";
import APILoader from "./APILoader.vue";
import ArticleDetails from "./ArticleDetails.vue";

export default {
  name: "ViewArcArticle",
  components: {
    APILoader,
    ArticleDetails,
  },
  metaInfo() {
    return {
      title: this.article ? this.article.slug : "Article",
    };
  },
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
  <APILoader :is-loading="isLoading" :reload="load" :error="error">
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
