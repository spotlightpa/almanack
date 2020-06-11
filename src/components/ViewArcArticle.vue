<script>
import { useAvailableArc } from "@/api/hooks.js";
import APILoader from "./APILoader.vue";
import ArcArticleAvailable from "./ArcArticleAvailable.vue";

export default {
  name: "ViewArcArticle",
  components: {
    APILoader,
    ArcArticleAvailable,
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
    <template v-if="article">
      <ArcArticleAvailable
        v-if="article.isAvailable"
        :article="article"
      ></ArcArticleAvailable>
    </template>
  </APILoader>
</template>
