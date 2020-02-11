<script>
import { computed, watch } from "@vue/composition-api";

import BulmaField from "./BulmaField.vue";
import BulmaFieldInput from "./BulmaFieldInput.vue";

export default {
  components: {
    BulmaField,
    BulmaFieldInput,
  },
  props: {
    article: { type: Object, required: true },
  },
  setup(props) {
    watch(() => {
      document.title = `Spotlight PA Almanack - ${props.article.id} Scheduler`;
    });

    let isPostDated = computed(() => new Date() - props.article.pubDate < 0);

    return {
      isPostDated,
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
    <p class="content">
      <b>Article last synchronized with Arc at {{ article.lastArcSync }}</b>
    </p>

    <BulmaField
      v-slot="{ idForLabel }"
      label="Publication Date"
      help="Note that postdated articles will not be shown on the site before their publication time"
    >
      <b-datetimepicker
        :id="idForLabel"
        v-model="article.pubDate"
        icon="calendar-alt"
        :timepicker="{ hourFormat: '12' }"
      >
      </b-datetimepicker>
      <p class="content is-small">
        <a
          href="#"
          class="has-text-info"
          @click.prevent="article.pubDate = new Date()"
        >
          Set to now
        </a>
      </p>
    </BulmaField>
    <p v-if="isPostDated" class="content has-text-primary is-small">
      Article publication date is in the future.
    </p>
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

    <b-field label="Authors">
      <b-taginput
        v-model="article.authors"
        attached
        allow-duplicates
      ></b-taginput>
      Full name as listed in data profile
    </b-field>
    <BulmaFieldInput
      v-model="article.byline"
      label="Byline"
      help="If present, overrides the byline created from authors list"
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

    TK Image type <br />

    <BulmaFieldInput v-model="article.slug" label="Slug"></BulmaFieldInput>
    <BulmaField label="Suppress in featured slot">
      <div>
        <label class="checkbox">
          <input v-model="article.suppressFeatured" type="checkbox" />
          Don't make this the top story on the homepage
        </label>
      </div>
    </BulmaField>

    <BulmaField v-slot="{ idForLabel }" label="Article text">
      <textarea
        :id="idForLabel"
        v-model="article.body"
        class="textarea"
        rows="8"
      ></textarea>
    </BulmaField>

    <div v-if="article.saveError" class="message is-danger">
      <p class="message-header">{{ article.saveError.name }}</p>
      <p class="message-body">{{ article.saveError.message }}</p>
    </div>

    <div class="field is-grouped">
      <div class="buttons">
        <button
          class="button is-success has-text-weight-semibold"
          :class="{ 'is-loading': article.isSaving }"
          :disabled="article.isSaving"
          @click.prevent="article.schedule"
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
        <button
          class="button is-light has-text-weight-semibold"
          :disabled="true"
        >
          Replace with Arc version
        </button>
      </div>
    </div>
    <p v-if="article.scheduleFor" class="content">
      <b>Article is scheduled to publish at {{ article.scheduleFor }}</b>
    </p>
  </div>
</template>
