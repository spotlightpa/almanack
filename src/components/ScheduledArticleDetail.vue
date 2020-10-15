<script>
import { computed, reactive } from "@vue/composition-api";

import { useClient } from "@/api/hooks.js";
import fuzzyMatch from "@/utils/fuzzy-match.js";
import { formatDate, formatTime } from "@/utils/time-format.js";

import BulmaField from "./BulmaField.vue";
import BulmaFieldInput from "./BulmaFieldInput.vue";
import CopyWithButton from "./CopyWithButton.vue";

export default {
  components: {
    BulmaField,
    BulmaFieldInput,
    CopyWithButton,
  },
  props: {
    article: { type: Object, required: true },
  },
  setup(props) {
    let { listAllTopics, listAllSeries } = useClient();
    let isPostDated = computed(() => new Date() - props.article.pubDate < 0);
    const autocomplete = reactive({
      topicsRaw: [],
      topicsFilter: "",
      topics: computed(() =>
        autocomplete.topicsFilter
          ? autocomplete.topicsRaw.filter((s) =>
              fuzzyMatch(s, autocomplete.topicsFilter)
            )
          : autocomplete.topicsRaw
      ),

      seriesRaw: [],
      seriesFilter: "",
      series: computed(() =>
        autocomplete.seriesFilter
          ? autocomplete.seriesRaw.filter((s) =>
              fuzzyMatch(s, autocomplete.seriesFilter)
            )
          : autocomplete.seriesRaw
      ),
    });

    listAllTopics().then(([data, err]) => {
      if (!err) {
        autocomplete.topicsRaw = data.topics || [];
      } else {
        // eslint-disable-next-line no-console
        console.warn(err);
      }
    });
    listAllSeries().then(([data, err]) => {
      if (!err) {
        autocomplete.seriesRaw = data.series || [];
      } else {
        // eslint-disable-next-line no-console
        console.warn(err);
      }
    });

    return {
      isPostDated,
      autocomplete,

      discardChanges() {
        if (window.confirm("Do you really want to discard all changes?")) {
          props.article.reset();
        }
      },
      schedulerPrimaryButtonText() {
        if (props.article.hasPublished) {
          return "Update live article";
        }
        if (isPostDated.value) {
          return props.article.scheduleFor
            ? "Update scheduled article"
            : "Schedule";
        }
        return "Post now";
      },
      schedulerSecondaryButtonText() {
        if (props.article.hasPublished) {
          return "Save without publishing changes";
        }
        if (isPostDated.value) {
          return props.article.scheduleFor
            ? "Unschedule"
            : "Save without scheduling";
        }
        return "Save without posting";
      },

      formatDate,
      formatTime,
    };
  },
};
</script>

<template>
  <div>
    <h2 class="title is-spaced">{{ article.id }} Scheduler</h2>
    <p class="content">
      <b
        >Article last synchronized with Arc at
        {{ formatDate(article.lastArcSync) }},
        {{ formatTime(article.lastArcSync) }}</b
      >
    </p>

    <div v-if="article.warnings.length" class="message is-warning">
      <div class="message-header">Warnings</div>
      <div class="message-body">
        <p
          v-for="warning of article.warnings"
          :key="warning"
          v-text="warning"
        ></p>
      </div>
    </div>

    <BulmaField
      v-slot="{ idForLabel }"
      label="Publication Date"
      help="Note that postdated articles will not be shown on the site before their publication time"
    >
      <b-datetimepicker
        :id="idForLabel"
        v-model="article.pubDate"
        icon="user-clock"
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
    />
    <b-field label="Topics">
      <b-taginput
        v-model="article.topics"
        :open-on-focus="true"
        :data="autocomplete.topics"
        attached
        @typing="autocomplete.topicsFilter = $event"
      ></b-taginput>
      Topics, e.g. Coronavirus
    </b-field>
    <b-field label="Series">
      <b-taginput
        v-model="article.series"
        :open-on-focus="true"
        :data="autocomplete.series"
        attached
        @typing="autocomplete.seriesFilter = $event"
      ></b-taginput>
      Series, e.g. “Legislative privilege 2020”
    </b-field>
    <BulmaFieldInput
      v-model="article.hed"
      label="Hed"
      help="Title as it appears at top of page"
    />
    <BulmaFieldInput
      v-model="article.subhead"
      label="Subhead"
      help="Appears below hed at top of page; AKA dek"
    />

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
    />

    <BulmaFieldInput
      v-model="article.linkTitle"
      label="Link to as"
      help="When linking to this page from another page, use this as the link title instead of hed"
    />

    <BulmaField
      v-slot="{ idForLabel }"
      label="Summary"
      required
      help="Shown in social share previews and search results"
    >
      <textarea
        :id="idForLabel"
        v-model="article.summary"
        class="textarea"
        rows="2"
      ></textarea>
    </BulmaField>

    <BulmaField
      v-slot="{ idForLabel }"
      label="Blurb"
      help="Short summary to appear in article rivers"
    >
      <textarea
        :id="idForLabel"
        v-model="article.blurb"
        class="textarea"
        rows="2"
      ></textarea>
      <p class="content is-small">
        <a
          href="#"
          class="has-text-info"
          @click.prevent="article.blurb = article.summary"
        >
          Copy from summary
        </a>
      </p>
    </BulmaField>

    <BulmaFieldInput v-model="article.imageURL" label="Image URL" />
    <picture v-if="article.imagePreviewURL" class="has-ratio">
      <img :src="article.imagePreviewURL" class="is-3x4" width="200" />
    </picture>
    <BulmaFieldInput
      v-model="article.imageDescription"
      label="Image description"
    />
    <BulmaFieldInput v-model="article.imageCredit" label="Image credit" />

    <b-field label="Image size">
      <b-select v-model="article.imageSize" expanded>
        <option
          v-for="size in ['inline', 'full', 'hidden']"
          :key="size"
          :value="size"
        >
          {{ size }}
        </option>
      </b-select>
    </b-field>

    <BulmaField v-slot="{ idForLabel }" label="Language">
      <div class="select is-fullwidth">
        <select :id="idForLabel" v-model="article.languageCode" class="select">
          <option value="">English</option>
          <option value="es">Spanish</option>
        </select>
      </div>
    </BulmaField>

    <b-field label="Slug">
      <b-input
        v-model="article.slug"
        :disabled="article.hasPublished"
        :readonly="article.hasPublished"
      />
    </b-field>
    <button
      class="block button is-small is-light has-text-weight-semibold"
      :disabled="article.hasPublished"
      @click.prevent="article.deriveSlug"
    >
      Derive slug from title
    </button>

    <CopyWithButton :value="article.pubURL" label="planned URL" />

    <div v-if="article.hasPublished && article.pubURL" class="buttons">
      <a
        :href="article.pubURL"
        class="button is-success has-text-weight-semibold"
        target="_blank"
      >
        <span class="icon is-size-6">
          <font-awesome-icon :icon="['fas', 'link']" />
        </span>
        <span> Open live URL </span>
      </a>
    </div>

    <BulmaField v-slot="{ idForLabel }" label="Article text">
      <textarea
        :id="idForLabel"
        v-model="article.body"
        class="textarea"
        rows="8"
      ></textarea>
    </BulmaField>

    <details class="field">
      <summary class="has-text-weight-semibold">Advanced options</summary>

      <BulmaFieldInput
        v-model="article.extendedKicker"
        label="Homepage extended kicker (e.g. Top News)"
      />

      <BulmaField label="Hide newsletters pop-up">
        <div>
          <label class="checkbox">
            <input v-model="article.modalExclude" type="checkbox" />
            Don't show newsletters modal screen on this page
          </label>
        </div>
      </BulmaField>

      <BulmaField label="No index">
        <div>
          <label class="checkbox">
            <input v-model="article.noIndex" type="checkbox" />
            Hide page from Google search results
          </label>
        </div>
      </BulmaField>

      <BulmaFieldInput v-model="article.overrideURL" label="Override URL" />

      <b-field label="URL Aliases">
        <b-taginput v-model="article.aliases"></b-taginput>
        Redirect these URLs to the story
      </b-field>

      <BulmaFieldInput v-model="article.layout" label="Layout override" />
    </details>

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
          @click.prevent="article.save({ schedule: true })"
          v-text="schedulerPrimaryButtonText()"
        />
        <button
          v-if="!article.hasPublished"
          class="button is-primary has-text-weight-semibold"
          :class="{ 'is-loading': article.isSaving }"
          :disabled="article.isSaving"
          @click.prevent="article.save({ schedule: false })"
          v-text="schedulerSecondaryButtonText()"
        />
        <button
          class="button is-light has-text-weight-semibold"
          :disabled="article.isSaving"
          @click.prevent="discardChanges"
        >
          Discard Changes
        </button>
        <button
          class="button is-light has-text-weight-semibold"
          :disabled="article.isSaving"
          @click.prevent="article.save({ refreshArc: true })"
        >
          Replace with Arc version
        </button>
      </div>
    </div>
    <p v-if="article.lastSaved" class="content">
      <b
        >Article last saved at {{ formatDate(article.lastSaved) }},
        {{ formatTime(article.lastSaved) }}</b
      >
    </p>
    <p v-if="article.scheduleFor" class="content">
      <b
        >Article is scheduled to publish at
        {{ formatDate(article.scheduleFor) }},
        {{ formatTime(article.scheduleFor) }}</b
      >
    </p>

    <div v-if="article.warnings.length" class="message is-warning">
      <div class="message-header">Warnings</div>
      <div class="message-body">
        <p
          v-for="warning of article.warnings"
          :key="warning"
          v-text="warning"
        ></p>
      </div>
    </div>
  </div>
</template>
