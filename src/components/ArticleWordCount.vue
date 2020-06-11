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
  <div class="level is-mobile is-clipped">
    <div class="level-left">
      <p v-if="article.plannedWordCount" class="level-item">
        <span>
          <strong>Planned Word Count:</strong>
          {{ intcomma(article.plannedWordCount) }}
        </span>
      </p>

      <p v-if="article.actualWordCount" class="level-item is-hidden-mobile">
        <span>
          <strong>Word Count:</strong>
          {{ intcomma(article.actualWordCount) }}
        </span>
      </p>
      <p class="level-item is-hidden-mobile">
        <span>
          <strong>Lines:</strong>
          {{ article.actualLineCount }}
        </span>
      </p>
      <p class="level-item is-hidden-mobile">
        <span>
          <strong>Column inches:</strong>
          {{ article.actualInchCount }}
        </span>
      </p>
    </div>
  </div>
</template>
