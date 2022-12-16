<script>
import { intcomma } from "journalize";

export default {
  props: {
    article: {
      type: Object,
      required: true,
    },
  },
  methods: { intcomma },
};
</script>

<template>
  <div v-if="article.isArc" class="level is-mobile is-clipped">
    <div class="level-left">
      <p v-if="article.arc.plannedWordCount" class="level-item">
        <span>
          <strong>Planned Word Count:</strong>
          {{ intcomma(article.arc.plannedWordCount) }}
        </span>
      </p>

      <p v-if="article.arc.actualWordCount" class="level-item is-hidden-mobile">
        <span>
          <strong>Word Count:</strong>
          {{ intcomma(article.arc.actualWordCount) }}
        </span>
      </p>
      <p class="level-item is-hidden-mobile">
        <span>
          <strong>Lines:</strong>
          {{ article.arc.actualLineCount }}
        </span>
      </p>
      <p class="level-item is-hidden-mobile">
        <span>
          <strong>Column inches:</strong>
          {{ article.arc.actualInchCount }}
        </span>
      </p>
    </div>
  </div>
</template>
