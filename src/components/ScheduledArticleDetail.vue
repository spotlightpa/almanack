<script>
import { watch } from "@vue/composition-api";

import BulmaFieldInput from "./BulmaFieldInput.vue";

export default {
  components: {
    BulmaFieldInput,
  },
  props: {
    article: { type: Object, required: true },
  },
  setup(props) {
    watch(() => {
      document.title = `Spotlight PA Almanack - ${props.article.id} Scheduler`;
    });
    return {
      discardChanges() {
        if (window.confirm("Do you really want to discard all changes?")) {
          props.article.reset();
        }
      },
    };
  },
};
</script>

<template>
  <div>
    <h2 class="title is-spaced">{{ article.id }} Scheduler</h2>

    <BulmaFieldInput
      v-model="article.kicker"
      label="Kicker"
      help="Small text appearing above the page headline, e.g. Health"
      :required="true"
    ></BulmaFieldInput>
    <BulmaFieldInput
      v-model="article.hed"
      label="Hed"
      help="Title as it appears at top of page"
      :required="true"
    ></BulmaFieldInput>
    <BulmaFieldInput
      v-model="article.subhead"
      label="Subhead"
      help="Appears below hed at top of page; AKA dek"
    ></BulmaFieldInput>
    <BulmaFieldInput
      v-model="article.linkTitle"
      label="Link to as"
      help="When linking to this page from another page, use this as the link title instead of hed"
    ></BulmaFieldInput>
    <BulmaFieldInput
      v-model="article.summary"
      label="Summary"
      help="Shown in social share previews and search results"
      :required="true"
    ></BulmaFieldInput>
    <BulmaFieldInput
      v-model="article.blurb"
      label="Blurb"
      help="Short summary to appear in article rivers"
    ></BulmaFieldInput>

    <BulmaFieldInput
      v-model="article.imageURL"
      label="Image URL"
    ></BulmaFieldInput>
    <BulmaFieldInput
      v-model="article.imageCaption"
      label="Image caption"
    ></BulmaFieldInput>
    <BulmaFieldInput
      v-model="article.imageCredit"
      label="Image credit"
    ></BulmaFieldInput>
    <BulmaFieldInput v-model="article.slug" label="Slug"></BulmaFieldInput>
    TK Dates <br />
    TK Image type <br />
    TK Suppress Featured <br />
    TK Authors/bylines <br />
    <BulmaFieldInput v-model="article.body" label="Body"></BulmaFieldInput>

    <div class="field is-grouped">
      <div class="buttons">
        <button
          class="button is-success has-text-weight-semibold"
          :class="{ 'is-loading': article.isSaving }"
          :disabled="article.isSaving"
          @click.prevent="article.save"
        >
          Schedule
        </button>
        <button
          class="button is-light has-text-weight-semibold"
          :disabled="article.isSaving"
          @click.prevent="discardChanges"
        >
          Discard Changes
        </button>
      </div>
    </div>

    <h1 class="title">Page preview:</h1>
    <hr />
    <!-- <PageBase :page="page" :is-preview="true" :menus="menus" /> -->
  </div>
</template>
