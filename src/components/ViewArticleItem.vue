<script>
import APILoader from "./APILoader.vue";
import ArticleDetails from "./ArticleDetails.vue";
import { useAPI } from "@/api/hooks.js";

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
    let { articleFromID } = useAPI();
    return {
      article: articleFromID(() => props.id),
    };
  },
};
</script>

<template>
  <APILoader>
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
