<script>
import { formatDate } from "@/utils/time-format.js";

export default {
  name: "ArticleList",
  props: {
    articles: Array,
    title: String,
  },
  methods: { formatDate },
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
          {{ article.isArc ? article.arc.byline : article.byline }}
        </p>
        <p>
          <strong>Planned time:</strong>
          {{ formatDate(article.publicationDate) }}
        </p>
        <p class="has-margin-top-thin content is-small">
          {{ article.budget }}
        </p>

        <ArticleWordCount :article="article"></ArticleWordCount>
      </div>
    </article>
  </nav>
</template>
