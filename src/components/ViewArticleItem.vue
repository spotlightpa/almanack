<script>
import APILoader from "./APILoader.vue";

export default {
  name: "ViewArticleItem",
  components: {
    APILoader
  },
  props: {
    slug: String
  },
  computed: {
    article() {
      return this.$api.contents.find(article => article.slug === this.slug);
    }
  }
};
</script>

<template>
  <APILoader role="editor">
    <h1 class="title">{{ slug }}</h1>
    <div v-if="!article" class="message is-warning">
      <p class="message-header">
        Not found
      </p>
      <p class="message-body">
        Article not found.
        <router-link :to="{ name: 'home' }">Go home</router-link>?
      </p>
    </div>
    <div v-else></div>
  </APILoader>
</template>
