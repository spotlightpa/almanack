<script>
import ArticleSlugLine from "./ArticleSlugLine.vue";
import ArticleWordCount from "./ArticleWordCount.vue";

export default {
  name: "ArticleList",
  components: {
    ArticleSlugLine,
    ArticleWordCount,
  },
  props: {
    articles: Array,
    title: String,
  },
};
</script>
<template>
  <nav class="panel is-black">
    <h1 class="panel-heading" v-text="title"></h1>
    <article v-for="article of articles" :key="article.id" class="panel-block">
      <div class="control">
        <h2 class="title is-spaced is-3">
          <ArticleSlugLine :article="article"></ArticleSlugLine>
        </h2>

        <p class="has-margin-top-negative">
          <strong>Byline:</strong>
          {{ article.authors | commaand }}
        </p>
        <p>
          <strong>Planned time:</strong>
          {{ article.plannedDate | formatDate }}
        </p>
        <p class="has-margin-top-thin content is-small">
          {{ article.budgetLine }}
        </p>

        <ArticleWordCount :article="article"></ArticleWordCount>
      </div>
    </article>
  </nav>
</template>
