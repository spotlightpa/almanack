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
          <ArticleSlugLine :article="article" />
        </h2>
        <template v-if="article.isArc">
          <p class="has-margin-top-negative">
            <strong>Byline:</strong>
            {{ article.arc.byline }}
          </p>
          <p>
            <strong>Planned time:</strong>
            {{ formatDate(article.publicationDate) }}
          </p>
          <p class="has-margin-top-thin content is-small">
            {{ article.budget }}
          </p>
        </template>
        <ArticleWordCount :article="article" />
      </div>
    </article>
  </nav>
</template>
