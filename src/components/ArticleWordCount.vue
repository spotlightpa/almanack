<script setup>
import { computed } from "vue";
import { intcomma } from "journalize";

const props = defineProps({
  article: {
    type: Object,
    required: true,
  },
});

const wordCounts = computed(() => {
  let article = props.article;
  if (article.isArc) {
    return {
      actual: intcomma(article.arc.actualWordCount),
      lines: intcomma(article.arc.actualLineCount),
      inches: intcomma(article.arc.actualInchCount),
    };
  }
  let wc = article.gdocs?.word_count;
  if (wc) {
    return {
      actual: intcomma(wc),
      lines: intcomma(Math.round(wc / 30)),
      inches: intcomma(Math.ceil(wc / 30 / 8)),
    };
  }
  return {
    actual: "",
    lines: "",
    inches: "",
  };
});
</script>

<template>
  <div class="level is-mobile is-clipped">
    <div class="level-left">
      <p v-if="wordCounts.actual" class="level-item is-hidden-mobile">
        <span>
          <strong>Word Count:</strong>
          {{ wordCounts.actual }}
        </span>
      </p>
      <p class="level-item is-hidden-mobile">
        <span>
          <strong>Lines:</strong>
          {{ wordCounts.lines }}
        </span>
      </p>
      <p class="level-item is-hidden-mobile">
        <span>
          <strong>Column inches:</strong>
          {{ wordCounts.inches }}
        </span>
      </p>
    </div>
  </div>
</template>
