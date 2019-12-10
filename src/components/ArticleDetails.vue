<script>
import DOMInnerHTML from "./DOMInnerHTML.vue";
import APIArticleSlugLine from "./APIArticleSlugLine.vue";
import CopyTextarea from "./CopyTextarea.vue";
import CopyWithButton from "./CopyWithButton.vue";

export default {
  components: {
    DOMInnerHTML,
    APIArticleSlugLine,
    CopyTextarea,
    CopyWithButton,
  },
  props: {
    article: { type: Object, required: true },
  },
  data() {
    return {
      copied: false,
      viewHTML: false,
      articleHTML: "",
    };
  },
  computed: {
    embeds() {
      return this.article.embedComponents;
    },
  },
};
</script>

<template>
  <div>
    <h1 class="title has-text-grey">
      <APIArticleSlugLine :article="article"></APIArticleSlugLine>
    </h1>
    <h2 class="title">
      Planned for
      {{ article.plannedDate | formatDate }}
    </h2>
    <template v-if="article.note">
      <h2 class="title is-stacked">
        Internal Note
      </h2>
      <p class="content has-margin-top-negative">
        {{ article.note }}
      </p>
    </template>

    <h2 class="title">Suggested Hed</h2>
    <CopyWithButton :value="article.headline" :rows="2"></CopyWithButton>

    <h2 class="title">Suggested Description</h2>
    <CopyWithButton :value="article.description" :rows="2"></CopyWithButton>

    <h2 class="title">Byline</h2>
    <CopyWithButton :value="article.byline"></CopyWithButton>

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
              @click="viewHTML = false"
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
              @click="
                viewHTML = false;
                $refs.copyRichText.copy();
              "
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
              @click="viewHTML = true"
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
              @click="
                viewHTML = true;
                $refs.copyHTML.copy();
              "
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
    ></CopyTextarea>

    <details class="block">
      <summary class="title">Budget details</summary>
      <p class="content">
        {{ article.budgetLine }}
      </p>
    </details>
    <details class="block">
      <summary class="title">Raw JSON</summary>
      <pre class="code">{{ article.rawData | json }}</pre>
    </details>
  </div>
</template>

<style>
.height-50vh {
  height: 50vh !important;
  overflow-y: scroll;
}
</style>
