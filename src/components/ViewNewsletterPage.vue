<script>
import Vue from "vue";
import { computed, toRefs, watch } from "@vue/composition-api";

import { formatDateTime } from "@/utils/time-format.js";

import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";
import imgproxyURL from "@/api/imgproxy-url.js";

import BulmaAutocompleteArray from "./BulmaAutocompleteArray.vue";
import BulmaField from "./BulmaField.vue";
import BulmaFieldInput from "./BulmaFieldInput.vue";

class Page {
  constructor(data) {
    this.init(data);
    Vue.observable(this);
  }

  init(data) {
    this.id = data["id"] ?? "";
    this.body = data["body"] ?? "";
    this.frontmatter = data["frontmatter"] ?? {};
    this.filePath = data["file_path"] ?? "";
    this.urlPath = data["url_path"]?.String ?? "";
    this.createdAt = data["created_at"] ?? "";
    this.publishedAt = Page.getDate(this.frontmatter, "published");
    this.updatedAt = Page.getDate(data, "updated_at");
    this.lastPublished = Page.getNullableDate(data, "last_published");
    this.scheduleFor = Page.getNullableDate(data, "schedule_for");
    this.kicker = this.frontmatter["kicker"] ?? "";
    this.title = this.frontmatter["title"] ?? "";
    this.internalID = this.frontmatter["internal-id"] ?? "";
    this.linkTitle = this.frontmatter["linktitle"] ?? "";
    this.titleTag = this.frontmatter["title-tag"] ?? "";
    this.authors = this.frontmatter["authors"] ?? [];
    this.byline = this.frontmatter["byline"] ?? "";
    this.summary = this.frontmatter["description"] ?? "";
    this.blurb = this.frontmatter["blurb"] ?? "";
    this.topics = this.frontmatter["topics"] ?? [];
    this.series = this.frontmatter["series"] ?? [];
    this.image = this.frontmatter["image"] ?? "";
    this.imageDescription = this.frontmatter["image-description"] ?? "";
    this.imageCredit = this.frontmatter["image-credit"] ?? "";
    this.imageSize = this.frontmatter["image-size"] ?? "";
    this.languageCode = this.frontmatter["language-code"] ?? "";
    this.slug = this.frontmatter["slug"] ?? "";
    this.extendedKicker = this.frontmatter["extended-kicker"] ?? "";
    this.modalExclude = this.frontmatter["modal-exclude"] ?? "";
    this.noIndex = this.frontmatter["no-index"] ?? "";
    this.overrideURL = this.frontmatter["url"] ?? "";
    this.aliases = this.frontmatter["aliases"] ?? [];
    this.layout = this.frontmatter["layout"] ?? "";
  }

  static getDate(data, prop) {
    let date = data[prop] ?? null;
    return date && new Date(date);
  }

  static getNullableDate(data, prop) {
    return data[prop]?.Valid ? new Date(data[prop].Time) : null;
  }

  get isPublished() {
    return !!this.lastPublished;
  }

  get link() {
    if (!this.isPublished || !this.urlPath) {
      return "";
    }

    return new URL(this.urlPath, "https://www.spotlightpa.org").href;
  }

  get imagePreviewURL() {
    if (!this.image || this.image.match(/^http/)) {
      return "";
    }
    return imgproxyURL(this.image);
  }
}

export default {
  name: "ViewNewsletterPage",
  components: {
    BulmaAutocompleteArray,
    BulmaField,
    BulmaFieldInput,
  },
  props: {
    id: String,
  },
  metaInfo() {
    return {
      title: this.title,
    };
  },
  setup(props) {
    let { getPage } = useClient();
    let { apiState, exec } = makeState();

    const fetch = (id) => exec(() => getPage(id));

    watch(() => props.id, fetch, {
      immediate: true,
    });
    const page = computed(() =>
      apiState.rawData ? new Page(apiState.rawData) : null
    );

    return {
      ...toRefs(apiState),
      formatDateTime,

      fetch,
      page,
      title: computed(() => {
        if (!page.value) {
          return `Newsletter Page ${props.id}`;
        }
        return "Page " + (page.value.internalID || "Untitled");
      }),
      deriveSlug() {
        page.value.slug = page.value.title
          .toLowerCase()
          .replace(/\b(the|an?)\b/g, " ")
          .replace(/\bpa\b/g, "pennsylvania")
          .replace(/\W+/g, " ")
          .trim()
          .replace(/ /g, "-");
      },
    };
  },
};
</script>

<template>
  <div>
    <nav class="breadcrumb has-succeeds-separator" aria-label="breadcrumbs">
      <ul>
        <li>
          <router-link :to="{ name: 'admin' }">Admin</router-link>
        </li>
        <li>
          <router-link exact :to="{ name: 'newsletters' }">
            Newsletter Pages
          </router-link>
        </li>
        <li class="is-active">
          <router-link exact :to="{ name: 'newsletter-page', params: { id } }">
            {{ title }}
          </router-link>
        </li>
      </ul>
    </nav>

    <h1 class="title" v-text="title" />

    <progress
      v-if="!didLoad && isLoading"
      class="progress is-large is-warning"
      max="100"
    >
      Loading…
    </progress>

    <div v-if="didLoad">
      <BulmaField
        v-slot="{ idForLabel }"
        label="Publication Date"
        help="Page will be listed on the site under this date"
      >
        <b-datetimepicker
          :id="idForLabel"
          v-model="page.publishedAt"
          icon="user-clock"
          :datetime-formatter="formatDateTime"
          locale="en-US"
        />
        <p class="content is-small">
          <a
            href="#"
            class="has-text-info"
            @click.prevent="page.publishedAt = new Date()"
          >
            Set to now
          </a>
        </p>
      </BulmaField>
      <p
        v-if="page.publishedAt - new Date() > 0"
        class="content has-text-primary is-small"
      >
        Article publication date is in the future.
      </p>

      <BulmaFieldInput
        v-model="page.kicker"
        label="Eyebrow"
        help="Small text appearing above the page hed"
        :required="true"
      />
      <!-- todo: topics and series -->
      <BulmaFieldInput
        v-model="page.title"
        label="Title"
        help="Default value for title tag, link title, and share title"
      />

      <BulmaFieldInput
        v-model="page.linkTitle"
        label="Link to as"
        help="When linking to this page from another page, use this as the link title instead of regular title"
      />

      <BulmaFieldInput
        v-model="page.titleTag"
        label="Title tag"
        help="Use this in the page title bar rather than the regular title"
      />

      <BulmaAutocompleteArray
        v-model="page.authors"
        label="Authors"
        help="Full name as listed in data profile"
        :options="[]"
      />

      <BulmaFieldInput
        v-model="page.byline"
        label="Byline"
        help="If present, overrides the byline created from authors list"
      />

      <BulmaField
        v-slot="{ idForLabel }"
        label="Summary"
        required
        help="Shown in social share previews and search results"
      >
        <textarea
          :id="idForLabel"
          v-model="page.summary"
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
          v-model="page.blurb"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>

      <BulmaAutocompleteArray
        v-model="page.topics"
        label="Topics"
        :options="[]"
        help="Topics are open-ended collections, e.g. “Events”, “Coronavirus”"
      />

      <BulmaAutocompleteArray
        v-model="page.series"
        label="Series"
        :options="[]"
        help="Series are limited-time collections, e.g. “Legislative privilege 2020”"
      />

      <BulmaFieldInput
        v-model="page.image"
        label="Photo ID"
        help="Image is shown in article rivers and on social media"
      />
      <picture v-if="page.imagePreviewURL" class="has-ratio">
        <img :src="page.imagePreviewURL" class="is-3x4" width="200" />
      </picture>

      <!-- todo: search for image -->

      <BulmaFieldInput
        v-model="page.imageDescription"
        label="Image description"
      />
      <BulmaFieldInput v-model="page.imageCredit" label="Image credit" />

      <b-field label="Image size">
        <b-select v-model="page.imageSize" expanded>
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
          <select :id="idForLabel" v-model="page.languageCode" class="select">
            <option value="">English</option>
            <option value="es">Spanish</option>
          </select>
        </div>
      </BulmaField>

      <b-field label="URL keywords slug">
        <b-input
          v-model="page.slug"
          :disabled="page.isPublished"
          :readonly="page.isPublished"
        />
      </b-field>
      <button
        class="block button is-small is-light has-text-weight-semibold"
        type="button"
        :disabled="page.isPublished"
        @click.prevent="deriveSlug"
      >
        Derive slug from title
      </button>

      <div v-if="page.link" class="buttons">
        <a
          :href="page.link"
          class="button is-success has-text-weight-semibold"
          target="_blank"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'link']" />
          </span>
          <span> Open live URL </span>
        </a>
      </div>
      <details class="field">
        <summary class="has-text-weight-semibold">Advanced options</summary>

        <BulmaFieldInput
          v-model="page.extendedKicker"
          label="Homepage extended kicker (e.g. Top News)"
        />

        <BulmaField label="Hide newsletters pop-up">
          <div>
            <label class="checkbox">
              <input v-model="page.modalExclude" type="checkbox" />
              Don't show newsletters modal screen on this page
            </label>
          </div>
        </BulmaField>

        <BulmaField label="No index">
          <div>
            <label class="checkbox">
              <input v-model="page.noIndex" type="checkbox" />
              Hide page from Google search results
            </label>
          </div>
        </BulmaField>

        <BulmaFieldInput v-model="page.overrideURL" label="Override URL" />

        <BulmaAutocompleteArray
          v-model="page.aliases"
          label="URL Aliases"
          help="Redirect these URLs to the story"
          :options="[]"
        />

        <BulmaFieldInput v-model="page.layout" label="Layout override" />
      </details>
    </div>

    <div v-if="error" class="message is-danger">
      <div class="message-header">{{ error.name }}</div>
      <div class="message-body">
        <p class="content">{{ error.message }}</p>
        <div class="buttons">
          <button
            class="button is-danger has-text-weight-semibold"
            type="button"
            @click="reload"
          >
            Reload?
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
