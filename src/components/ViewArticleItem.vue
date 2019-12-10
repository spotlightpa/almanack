<script>
import APILoader from "./APILoader.vue";
import APIArticle from "./APIArticle.vue";
import ArticleDetails from "./ArticleDetails.vue";

export default {
  name: "ViewArticleItem",
  components: {
    APILoader,
    APIArticle,
    ArticleDetails,
  },
  props: {
    id: String,
  },
  computed: {
    articleData() {
      return this.$api.getByID(this.id);
    },
  },
};
</script>

<template>
  <APILoader role="editor">
    <div v-if="!articleData" class="message is-warning">
      <p class="message-header">
        Not found
      </p>
      <p class="message-body">
        Article not found.
        <router-link :to="{ name: 'home' }">Go home</router-link>?
      </p>
    </div>
    <APIArticle v-else v-slot="{ article }" :data="articleData">
      <ArticleDetails :article="article"></ArticleDetails>
    </APIArticle>
  </APILoader>
</template>
