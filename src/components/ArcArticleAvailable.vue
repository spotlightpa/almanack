<script>
import { reactive, ref, toRefs } from "vue";
import { sendGAEvent } from "@/utils/google-analytics.js";
import { formatDate, formatDateTime } from "@/utils/time-format.js";

export default {
  props: {
    article: { type: Object, required: true },
  },
  setup(props) {
    const htmlArea = ref();
    const richTextArea = ref();

    const data = reactive({
      copied: false,
      viewHTML: false,
      articleHTML: "",
    });

    return {
      htmlArea,
      richTextArea,

      ...toRefs(data),

      embeds: props.article.arc.embedComponents,

      showRichText() {
        data.viewHTML = false;
        sendGAEvent({
          eventCategory: "ArticleDetails interaction",
          eventAction: "Show Rich Text",
        });
      },
      copyRichText() {
        data.viewHTML = false;
        richTextArea.value.copy();
        sendGAEvent({
          eventCategory: "ArticleDetails interaction",
          eventAction: "Copy Rich Text",
        });
      },
      showHTML() {
        data.viewHTML = true;
        sendGAEvent({
          eventCategory: "ArticleDetails interaction",
          eventAction: "Show HTML",
        });
      },
      copyHTML() {
        data.viewHTML = true;
        htmlArea.value.copy();
        sendGAEvent({
          eventCategory: "ArticleDetails interaction",
          eventAction: "Copy HTML",
        });
      },

      formatDate,
      formatDateTime,
    };
  },
};
</script>

<template>
  <div>
    <h1 class="title has-text-grey">
      <ArticleSlugLine :article="article" />
    </h1>
    <h2 class="mb-2 title">Budget details</h2>
    <p class="mb-2 content">
      {{ article.arc.budgetLine }}
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
      {{ formatDate(article.arc.plannedDate) }}
    </p>
    <template v-if="article.note">
      <h2 class="title is-stacked">Publication Notes</h2>
      <p class="content has-margin-top-negative">
        {{ article.note }}
      </p>
    </template>

    <h2 class="title">Suggested Hed</h2>
    <CopyWithButton :value="article.arc.headline" label="hed" />

    <h2 class="title">Suggested Description</h2>
    <CopyWithButton :value="article.arc.description" label="description" />

    <h2 class="title">Byline</h2>
    <CopyWithButton :value="article.arc.byline" label="byline" />

    <template v-if="article.arc.featuredImage">
      <h2 class="title is-spaced">Featured Image</h2>
      <ImageThumbnail
        :url="article.arc.featuredImage"
        :caption="article.arc.featuredImageCaption"
        :credits="article.arc.featuredImageCredits"
        class="block"
      />
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

    <div class="level">
      <div class="level-left">
        <div class="level-item">
          <div class="buttons has-addons">
            <button
              class="button is-light has-text-weight-semibold"
              type="button"
              @click="showRichText"
            >
              <span class="icon">
                <font-awesome-icon :icon="['far', 'file-word']" />
              </span>
              <span> View Rich Text </span>
            </button>
            <button
              class="button is-primary has-text-weight-semibold"
              type="button"
              @click="copyRichText"
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
              @click="showHTML"
            >
              <span class="icon">
                <font-awesome-icon :icon="['far', 'file-code']" />
              </span>
              <span> View HTML </span>
            </button>
            <button
              class="button is-primary has-text-weight-semibold"
              type="button"
              @click="copyHTML"
            >
              <span class="icon">
                <font-awesome-icon :icon="['far', 'copy']" />
              </span>
              <span> Copy HTML </span>
            </button>
          </div>
        </div>
        <div class="level-item">
          <transition name="fade">
            <div
              v-if="copied"
              class="tag is-rounded is-success is-light has-text-weight-semibold"
            >
              Copied
            </div>
          </transition>
        </div>
      </div>
    </div>

    <CopyTextarea
      v-show="!viewHTML"
      ref="richTextArea"
      size="content height-50vh"
    >
      <component
        :is="block.component"
        v-for="(block, i) of article.arc.contentComponents"
        :key="i"
        :block="block.block"
      ></component>
    </CopyTextarea>

    <DOMInnerHTML @mounted="articleHTML = $event">
      <component
        :is="block.component"
        v-for="(block, i) of article.arc.htmlComponents"
        :key="i"
        :block="block.block"
      ></component>
    </DOMInnerHTML>

    <CopyTextarea
      v-show="viewHTML"
      ref="htmlArea"
      size="is-small height-50vh"
      >{{ articleHTML }}</CopyTextarea
    >
  </div>
</template>

<style>
.height-50vh {
  height: 50vh !important;
  overflow-y: scroll;
}
</style>
