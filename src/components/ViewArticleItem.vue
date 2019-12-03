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
    id: String
  },
  computed: {
    articleData() {
      return this.$api.getByID(this.id);
    }
  }
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
      <h1 class="title has-text-grey">
        <font-awesome-icon :icon="['far', 'newspaper']" />

        {{ article.slug }}
      </h1>
      <h2 class="title">
        Planned for
        {{ article.plannedDate | formatDate }}
      </h2>
      <template v-if="article.note">
        <h2 class="title is-stacked">
          Internal Note
        </h2>
        <p class="content has-margin-top-negative">
          {{ article.note }}
        </p>
      </template>

      <h2 class="title">Suggested Hed</h2>
      <BulmaCopyInput :value="article.headline" :rows="2"></BulmaCopyInput>

      <h2 class="title">Suggested Description</h2>
      <BulmaCopyInput :value="article.description" :rows="2"></BulmaCopyInput>

      <h2 class="title">Byline</h2>
      <BulmaCopyInput :value="article.authors | commaand"></BulmaCopyInput>

      <details class="block">
        <summary class="title">Budget details</summary>
        <p class="content">
          {{ article.budgetLine }}
        </p>
      </details>

      <pre class="code">{{ article.rawData | json }}</pre>
    </APIArticle>
  </APILoader>
</template>
