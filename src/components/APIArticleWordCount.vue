<script>
export default {
  props: {
    article: {
      type: Object,
      required: true,
    },
  },
};
</script>

<template>
  <div class="level is-mobile is-clipped">
    <div class="level-left">
      <p v-if="article.plannedWordCount" class="level-item">
        <span>
          <strong>Planned Word Count:</strong>
          {{ article.plannedWordCount | intcomma }}
        </span>
      </p>

      <p v-if="article.actualWordCount" class="level-item is-hidden-mobile">
        <span>
          <strong>Word Count:</strong>
          {{ article.actualWordCount | intcomma }}
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
