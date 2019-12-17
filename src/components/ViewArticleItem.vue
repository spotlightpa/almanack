<script>
import APILoader from "./APILoader.vue";
import ArticleDetails from "./ArticleDetails.vue";
import { useAuth } from "@/api/auth.js";
import { useAPI } from "@/api/store.js";

export default {
  name: "ViewArticleItem",
  components: {
    APILoader,
    ArticleDetails,
  },
  props: {
    id: String,
  },
  setup() {
    let $auth = useAuth();
    let $api = useAPI();
    return {
      $auth,
      $api,
    };
  },
  computed: {
    article() {
      return this.$api.getByID(this.id);
    },
  },
};
</script>

<template>
  <APILoader role="editor">
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
