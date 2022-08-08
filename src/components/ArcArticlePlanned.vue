<script>
import { formatDate } from "@/utils/time-format.js";

export default {
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
      <ArticleSlugLine :article="article" />
    </h1>
    <h2 class="title">Planned time</h2>
    <p class="content has-margin-top-negative">
      {{ formatDate(article.plannedDate) }}
    </p>
    <template v-if="article.note">
      <h2 class="title is-stacked">Publication Notes</h2>
      <p class="content has-margin-top-negative">
        {{ article.note }}
      </p>
    </template>

    <h2 class="title">Budget details</h2>
    <p class="content">
      {{ article.budgetLine }}
    </p>
    <ArticleWordCount :article="article" />

    <h2 class="title">Working Hed</h2>
    <NoCopyTextArea>{{ article.headline }}</NoCopyTextArea>

    <template v-if="article.description">
      <h2 class="title">Working Description</h2>
      <NoCopyTextArea>{{ article.description }}</NoCopyTextArea>
    </template>

    <h2 class="title">Byline</h2>
    <NoCopyTextArea>{{ article.byline }}</NoCopyTextArea>

    <template v-if="article.featuredImage">
      <h2 class="title">Working Image</h2>
      <div class="image max-256 has-background-grey-lighter has-margin-bottom">
        <img :src="article.featuredImage" />
      </div>
    </template>

    <h2 v-if="embeds.length === 1" class="title">Embed</h2>
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

.max-256 {
  max-height: 256px;
  max-width: 256px;
  min-height: 1rem;
  min-width: 1rem;
}
</style>
