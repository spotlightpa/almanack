<script>
import { formatDate } from "@/utils/time-format.js";

import ArticleSlugLine from "./ArticleSlugLine.vue";
import ArticleWordCount from "./ArticleWordCount.vue";
import NoCopyTextArea from "./NoCopyTextArea.vue";

export default {
  components: {
    ArticleSlugLine,
    ArticleWordCount,
    NoCopyTextArea,
  },
  props: {
    article: { type: Object, required: true },
  },
  setup(props) {
    return {
      embeds: props.article.embedComponents,
      formatDate,
    };
  },
};
</script>

<template>
  <div>
    <h1 class="title has-text-grey">
      <ArticleSlugLine :article="article"></ArticleSlugLine>
    </h1>
    <h2 class="title">
      Planned time
    </h2>
    <p class="content has-margin-top-negative">
      {{ formatDate(article.plannedDate) }}
    </p>
    <template v-if="article.note">
      <h2 class="title is-stacked">
        Publication Notes
      </h2>
      <p class="content has-margin-top-negative">
        {{ article.note }}
      </p>
    </template>

    <h2 class="title">Budget details</h2>
    <p class="content">
      {{ article.budgetLine }}
    </p>
    <ArticleWordCount :article="article"></ArticleWordCount>

    <h2 class="title">Working Hed</h2>
    <NoCopyTextArea v-text="article.headline" />

    <template v-if="article.description">
      <h2 class="title">Working Description</h2>
      <NoCopyTextArea v-text="article.description" />
    </template>

    <h2 class="title">Byline</h2>
    <NoCopyTextArea v-text="article.byline" />

    <h2 v-if="embeds.length === 1" class="title">
      Embed
    </h2>
    <h2 v-if="embeds.length > 1" class="title">Embeds: {{ embeds.length }}</h2>

    <component
      :is="component"
      v-for="{ block, component, n } of embeds"
      :key="n"
      :block="block"
      :n="n"
    ></component>

    <h2 class="title">Draft text</h2>
    <NoCopyTextArea class="content height-50vh">
      <component
        :is="block.component"
        v-for="(block, i) of article.contentComponents"
        ref="contentsEls"
        :key="i"
        :block="block.block"
      ></component>
    </NoCopyTextArea>
  </div>
</template>

<style scoped>
.height-50vh {
  height: 50vh !important;
  overflow-y: scroll;
}
</style>
