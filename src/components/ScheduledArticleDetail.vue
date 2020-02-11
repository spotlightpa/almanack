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
      saving: false,
      discardChanges() {},
    };
  },
};
</script>

<template>
  <div>
    <h2 class="title is-spaced">{{ article.id }} Scheduler View</h2>

    <BulmaFieldInput
      v-model="article.hed"
      label="Hed"
      help="Title as it appears at top of page"
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
    TK Kicker <br />
    TK Suppress Featured <br />
    TK Authors/bylines <br />
    <BulmaFieldInput v-model="article.body" label="Body"></BulmaFieldInput>

    <div class="field is-grouped">
      <div class="buttons">
        <button
          class="button is-success has-text-weight-semibold"
          :class="{ 'is-loading': saving }"
          :disabled="saving"
          @click.prevent="save"
        >
          Schedule
        </button>
        <button
          class="button is-light has-text-weight-semibold"
          :disabled="saving"
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
