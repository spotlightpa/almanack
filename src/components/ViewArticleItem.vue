<script>
import APILoader from "./APILoader.vue";
import BulmaCopyInput from "./BulmaCopyInput.vue";

export default {
  name: "ViewArticleItem",
  components: {
    APILoader,
    BulmaCopyInput
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
    <h1 class="title">
      <font-awesome-icon :icon="['far', 'newspaper']" />

      {{ slug }}
    </h1>
    <div v-if="!article" class="message is-warning">
      <p class="message-header">
        Not found
      </p>
      <p class="message-body">
        Article not found.
        <router-link :to="{ name: 'home' }">Go home</router-link>?
      </p>
    </div>
    <div v-else>
      <h2 class="title">
        Embargoed for
        {{ article.planning.scheduling.planned_publish_date | formatDate }}
      </h2>

      <h2 class="title">Notes</h2>
      <p class="content">
        {{ article.planning.budget_line }}
      </p>

      <h2 class="title">Suggested Hed</h2>
      <BulmaCopyInput :value="article.headlines.basic"></BulmaCopyInput>

      <h2 class="title">Suggested Description</h2>
      <BulmaCopyInput :value="article.description.basic"></BulmaCopyInput>

      <pre class="code">{{ article | json }}</pre>
    </div>
  </APILoader>
</template>
