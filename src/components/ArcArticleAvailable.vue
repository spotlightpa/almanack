<script>
import { reactive, toRefs } from "@vue/composition-api";
import { sendGAEvent } from "@/utils/google-analytics.js";
import { formatDate } from "@/utils/time-format.js";

import ArticleSlugLine from "./ArticleSlugLine.vue";
import ArticleWordCount from "./ArticleWordCount.vue";
import CopyTextarea from "./CopyTextarea.vue";
import CopyWithButton from "./CopyWithButton.vue";
import DOMInnerHTML from "./DOMInnerHTML.vue";
import ImageThumbnail from "./ImageThumbnail.vue";

export default {
  components: {
    ArticleSlugLine,
    ArticleWordCount,
    CopyTextarea,
    CopyWithButton,
    DOMInnerHTML,
    ImageThumbnail,
  },
  props: {
    article: { type: Object, required: true },
  },
  setup(props, { refs }) {
    let data = reactive({
      copied: false,
      viewHTML: false,
      articleHTML: "",
    });

    return {
      ...toRefs(data),

      embeds: props.article.embedComponents,

      showRichText() {
        data.viewHTML = false;
        sendGAEvent({
          eventCategory: "ArticleDetails interaction",
          eventAction: "Show Rich Text",
        });
      },
      copyRichText() {
        data.viewHTML = false;
        refs.copyRichText.copy();
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
        refs.copyHTML.copy();
        sendGAEvent({
          eventCategory: "ArticleDetails interaction",
          eventAction: "Copy HTML",
        });
      },

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

    <h2 class="title">Suggested Hed</h2>
    <CopyWithButton :value="article.headline" label="hed" />

    <h2 class="title">Suggested Description</h2>
    <CopyWithButton :value="article.description" label="description" />

    <h2 class="title">Byline</h2>
    <CopyWithButton :value="article.byline" label="byline" />

    <template v-if="article.featuredImage">
      <h2 class="title is-spaced">Featured Image</h2>
      <ImageThumbnail
        v-if="article.featuredImage"
        :url="article.featuredImage"
        :caption="article.featuredImageCaption"
        :credits="article.featuredImageCredits"
        class="block"
      />
    </template>

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
              <span>
                View Rich Text
              </span>
            </button>
            <button
              class="button is-primary has-text-weight-semibold"
              type="button"
              @click="copyRichText"
            >
              <span class="icon">
                <font-awesome-icon :icon="['far', 'copy']" />
              </span>
              <span>
                Copy Rich Text
              </span>
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
              <span>
                View HTML
              </span>
            </button>
            <button
              class="button is-primary has-text-weight-semibold"
              type="button"
              @click="copyHTML"
            >
              <span class="icon">
                <font-awesome-icon :icon="['far', 'copy']" />
              </span>
              <span>
                Copy HTML
              </span>
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
      ref="copyRichText"
      size="content height-50vh"
    >
      <component
        :is="block.component"
        v-for="(block, i) of article.contentComponents"
        ref="contentsEls"
        :key="i"
        :block="block.block"
      ></component>
    </CopyTextarea>

    <DOMInnerHTML @mounted="articleHTML = $event">
      <component
        :is="block.component"
        v-for="(block, i) of article.htmlComponents"
        :key="i"
        :block="block.block"
      ></component>
    </DOMInnerHTML>

    <CopyTextarea
      v-show="viewHTML"
      ref="copyHTML"
      size="is-small height-50vh"
      v-text="articleHTML"
    />

    <details class="block">
      <summary class="title">Budget details</summary>
      <p class="content">
        {{ article.budgetLine }}
      </p>
      <ArticleWordCount :article="article" />
    </details>
  </div>
</template>

<style>
.height-50vh {
  height: 50vh !important;
  overflow-y: scroll;
}
</style>
