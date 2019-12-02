<script>
import APILoader from "./APILoader.vue";
import APIArticle from "./APIArticle.vue";
import BulmaCopyInput from "./BulmaCopyInput.vue";

export default {
  name: "ViewArticleItem",
  components: {
    APILoader,
    APIArticle,
    BulmaCopyInput
  },
  props: {
    slug: String
  },
  computed: {
    articleData() {
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
      <h2 class="title">
        Embargoed for
        {{ article.plannedDate | formatDate }}
      </h2>

      <h2 class="title">Notes</h2>
      <p class="content">
        {{ article.budgetLine }}
      </p>

      <h2 class="title">Suggested Hed</h2>
      <BulmaCopyInput :value="article.headline"></BulmaCopyInput>

      <h2 class="title">Suggested Description</h2>
      <BulmaCopyInput :value="article.description"></BulmaCopyInput>

      <pre class="code">{{ article.rawData | json }}</pre>
    </APIArticle>
  </APILoader>
</template>
