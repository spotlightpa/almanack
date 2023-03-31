<script setup>
import { ref } from "vue";

import { useAuth } from "@/api/hooks.js";
import { sendGAEvent } from "@/utils/google-analytics.js";
import { formatDate, formatDateTime } from "@/utils/time-format.js";

defineProps({
  article: Object,
});

const { isSpotlightPAUser } = useAuth();

const isShowingHTML = ref(false);
const richTextArea = ref(null);
const htmlArea = ref(null);

function showRichText() {
  isShowingHTML.value = false;
  sendGAEvent({
    eventCategory: "ArticleDetails interaction",
    eventAction: "Show Rich Text",
  });
}
function copyRichText() {
  isShowingHTML.value = false;
  richTextArea.value.copy();
  sendGAEvent({
    eventCategory: "ArticleDetails interaction",
    eventAction: "Copy Rich Text",
  });
}
function showHTML() {
  isShowingHTML.value = true;
  sendGAEvent({
    eventCategory: "ArticleDetails interaction",
    eventAction: "Show HTML",
  });
}
function copyHTML() {
  isShowingHTML.value = true;
  htmlArea.value.copy();
  sendGAEvent({
    eventCategory: "ArticleDetails interaction",
    eventAction: "Copy HTML",
  });
}
</script>

<template>
  <h1 class="title has-text-grey">
    <ArticleSlugLine :article="article" />
  </h1>
  <div
    v-if="isSpotlightPAUser && article.gdocs.warnings.length"
    class="message is-warning"
  >
    <div class="message-body">
      <p><strong>Warning:</strong></p>
      <p v-for="(w, i) of article.gdocs.warnings" :key="i" v-text="w"></p>
    </div>
  </div>
  <h2 class="mb-2 title">Budget details</h2>
  <p class="mb-2 content">
    {{ article.budgetLine }}
  </p>
  <ArticleWordCount :article="article" />

  <h2 v-if="article.isUnderEmbargo" class="title" style="color: red">
    Embargoed until {{ formatDateTime(article.embargoUntil) }}
  </h2>
  <template v-else-if="article.embargoUntil">
    <h2 class="mb-2 title">Embargo notes</h2>
    <p class="content">
      This article had been under embargoed until
      {{ formatDateTime(article.embargoUntil) }}. It is now available for
      publication.
    </p>
  </template>

  <h2 class="mb-2 title">Planned time</h2>
  <p class="content">
    {{ formatDate(article.publicationDate) }}
  </p>
  <template v-if="article.note">
    <h2 class="title is-stacked">Publication Notes</h2>
    <p class="content has-margin-top-negative">
      {{ article.note }}
    </p>
  </template>

  <h2 class="title">Suggested Hed</h2>
  <CopyWithButton :value="article.hed" label="hed" />

  <h2 class="title">Suggested Description</h2>
  <CopyWithButton :value="article.description" label="description" />

  <h2 class="title">Byline</h2>
  <CopyWithButton :value="article.byline" label="byline" />

  <template v-if="article.ledeImage">
    <h2 class="title is-spaced">Featured Image</h2>
    <ImageThumbnail
      :url="article.ledeImage"
      :caption="article.ledeImageCaption"
      :credits="article.ledeImageCredits"
      class="block"
    />
  </template>

  <h2 v-if="article.gdocs.embeds.length === 1" class="title">Embed</h2>
  <h2 v-if="article.gdocs.embeds.length > 1" class="title">
    Embeds: {{ article.gdocs.embeds.length }}
  </h2>

  <!-- Loop through embeds -->

  <div class="level">
    <div class="level-left">
      <div class="level-item">
        <div class="buttons has-addons">
          <button
            class="button is-light has-text-weight-semibold"
            type="button"
            @click="showRichText()"
          >
            <span class="icon">
              <font-awesome-icon :icon="['far', 'file-word']" />
            </span>
            <span> View Rich Text </span>
          </button>
          <button
            class="button is-primary has-text-weight-semibold"
            type="button"
            @click="copyRichText()"
          >
            <span class="icon">
              <font-awesome-icon :icon="['far', 'copy']" />
            </span>
            <span> Copy Rich Text </span>
          </button>
        </div>
      </div>
      <div class="level-item">
        <div class="buttons has-addons">
          <button
            class="button is-light has-text-weight-semibold"
            type="button"
            @click="showHTML()"
          >
            <span class="icon">
              <font-awesome-icon :icon="['far', 'file-code']" />
            </span>
            <span> View HTML </span>
          </button>
          <button
            class="button is-primary has-text-weight-semibold"
            type="button"
            @click="copyHTML()"
          >
            <span class="icon">
              <font-awesome-icon :icon="['far', 'copy']" />
            </span>
            <span> Copy HTML </span>
          </button>
        </div>
      </div>
    </div>
  </div>

  <CopyTextarea v-show="!isShowingHTML" ref="richTextArea" size="height-50vh">
    <div class="content" v-html="article.gdocs.rich_text" />
  </CopyTextarea>
  <CopyTextarea
    v-show="isShowingHTML"
    ref="htmlArea"
    size="is-small height-50vh"
  >
    <div class="pre-wrap" v-text="article.gdocs.raw_html"></div>
  </CopyTextarea>
</template>

<style>
.height-50vh {
  height: 50vh !important;
  overflow-y: scroll;
}
.pre-wrap {
  white-space: pre-wrap;
}
</style>
