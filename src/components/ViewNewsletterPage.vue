<script>
import Vue from "vue";
import { computed, ref, toRefs, watch } from "@vue/composition-api";

import { formatDateTime } from "@/utils/time-format.js";

import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";
import imgproxyURL from "@/api/imgproxy-url.js";

import BulmaAutocompleteArray from "./BulmaAutocompleteArray.vue";
import BulmaField from "./BulmaField.vue";
import BulmaFieldInput from "./BulmaFieldInput.vue";

class Page {
  constructor(data) {
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

    // not a getter so it won't react to changes
    this.status = "pub";
    if (!this.lastPublished) {
      this.status = this.scheduleFor ? "sked" : "none";
    }

    Vue.observable(this);
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

  get statusVerbose() {
    return {
      pub: "published",
      sked: "scheduled to be published",
      none: "unpublished",
    }[this.status];
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

  toJSON() {
    return {
      file_path: this.filePath,
      set_frontmatter: true,
      frontmatter: {
        // preserve unknown props
        ...this.frontmatter,
        // copy others
        published: this.publishedAt,
        kicker: this.kicker,
        title: this.title,
        "internal-id": this.internalID,
        linktitle: this.linkTitle,
        "title-tag": this.titleTag,
        authors: this.authors,
        byline: this.byline,
        description: this.summary,
        blurb: this.blurb,
        topics: this.topics,
        series: this.series,
        image: this.image,
        "image-description": this.imageDescription,
        "image-credit": this.imageCredit,
        "image-size": this.imageSize,
        "language-code": this.languageCode,
        slug: this.slug,
        "extended-kicker": this.extendedKicker,
        "modal-exclude": this.modalExclude,
        "no-index": this.noIndex,
        url: this.overrideURL,
        aliases: this.aliases,
        layout: this.layout,
      },
      set_body: false, // todo
      body: "",
      set_schedule_for: true,
      schedule_for: {
        Valid: !!this.scheduleFor,
        Time: this.scheduleFor,
      },
      url_path: "", // leave blank to prevent changes
      set_last_published: false,
    };
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
    const showProgress = ref(false);
    const toggleProgress = () => {
      showProgress.value = true;
      window.setTimeout(() => {
        showProgress.value = false;
      }, 1000);
    };

    const { getPage, postPage, listImages } = useClient();
    const { apiState, exec } = makeState();

    const fetch = (id) => {
      toggleProgress();
      return exec(() => getPage(id));
    };
    const post = (page) => {
      toggleProgress();
      return exec(() => postPage(page));
    };
    const page = computed(() =>
      apiState.rawData ? new Page(apiState.rawData) : null
    );

    watch(() => props.id, fetch, {
      immediate: true,
    });

    const { apiState: imageState, exec: execImage } = makeState();
    execImage(() => listImages());

    return {
      showProgress,
      showScheduler: ref(false),

      ...toRefs(apiState),
      formatDateTime,

      fetch,
      post,
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
          .replace(/’/g, "'")
          .replace(/.?'s/g, "s")
          .replace(/\W+/g, " ")
          .trim()
          .replace(/ /g, "-");
      },
      discardChanges() {
        if (window.confirm("Do you really want to discard all changes?")) {
          fetch(props.id);
        }
      },
      publishNow() {
        page.value.scheduleFor = new Date();
        return post(page.value);
      },
      updateSchedule() {
        const msg =
          "Scheduled publication date is in the past. Do you want to publish now?";
        let isPostDated = page.value.scheduleFor - new Date() > 0;
        if (!isPostDated && !window.confirm(msg)) {
          return;
        }
        return post(page.value);
      },
      updateOnly() {
        page.value.scheduleFor = null;
        return post(page.value);
      },

      imageState,
      images: computed(() =>
        !imageState.rawData ? [] : imageState.rawData.images
      ),
      imgproxyURL,
      setImageProps(image) {
        page.value.image = image.path;
        page.value.imageDescription = image.description;
        page.value.imageCredit = image.credit;
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

    <h1 class="title">
      {{ title }}

      <a
        v-if="page && page.status === 'pub' && page.link"
        :href="page.link"
        class="is-size-6"
        target="_blank"
      >
        <span class="icon is-size-6">
          <font-awesome-icon :icon="['fas', 'link']" />
        </span>
        <span>Open live URL</span>
      </a>
    </h1>

    <div v-if="page">
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

      <BulmaField v-if="images.length" label="Choose from recent photos">
        <div class="textarea" style="height: 225px; overflow-y: auto">
          <table class="table is-striped is-narrow is-fullwidth">
            <tbody>
              <tr v-for="image in images" :key="image.id">
                <a
                  class="is-flex-tablet p-1 has-text-black"
                  @click="setImageProps(image)"
                >
                  <div
                    class="mr-2 is-flex-shrink-0 is-clipped"
                    style="width: 128px"
                  >
                    <picture class="has-ratio">
                      <img
                        class="is-3x4"
                        :src="
                          imgproxyURL(image.path, {
                            width: 256,
                            height: 192,
                            extension: 'webp',
                          })
                        "
                        :alt="image.path"
                        loading="lazy"
                      />
                    </picture>
                  </div>
                  <div>
                    <div class="clamped-3">
                      {{ image.description }}
                      <template v-if="image.credit">
                        ({{ image.credit }})
                      </template>
                    </div>
                  </div>
                </a>
              </tr>
            </tbody>
          </table>
        </div>
        <p>
          <router-link :to="{ name: 'uploader' }" target="_blank">
            Manage photos
          </router-link>
        </p>
      </BulmaField>

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

      <p class="my-4 has-text-weight-semibold">
        Page is {{ page.statusVerbose
        }}<template v-if="page.status == 'sked'">
          at {{ formatDateTime(page.scheduleFor) }}</template
        >.
      </p>

      <div v-if="page.status !== 'pub'" class="field mb-5">
        <BulmaField
          v-slot="{ idForLabel }"
          label="Schedule For"
          help="Page will be automatically published at this time if set"
        >
          <b-datetimepicker
            :id="idForLabel"
            v-model="page.scheduleFor"
            icon="user-clock"
            :datetime-formatter="formatDateTime"
            locale="en-US"
          />
        </BulmaField>
      </div>
      <div class="field">
        <div class="buttons">
          <button
            class="button is-success has-text-weight-semibold"
            :disabled="isLoading"
            type="button"
            @click="publishNow"
          >
            {{ page.status === "pub" ? "Update page" : "Publish now" }}
          </button>
          <button
            v-if="page.status !== 'pub'"
            class="button is-warning has-text-weight-semibold"
            :disabled="isLoading || !page.scheduleFor"
            type="button"
            @click="updateSchedule"
          >
            {{
              page.status === "none" ? "Schedule to publish" : "Save changes"
            }}
          </button>
          <button
            v-if="page.status === 'sked'"
            class="button is-danger has-text-weight-semibold"
            :disabled="isLoading"
            type="button"
            @click="updateOnly"
          >
            Unschedule
          </button>
          <button
            v-if="page.status === 'none'"
            class="button is-light has-text-weight-semibold"
            :disabled="isLoading"
            type="button"
            @click="updateOnly"
          >
            Save without publishing
          </button>

          <button
            class="button is-light has-text-weight-semibold"
            :disabled="isLoading"
            type="button"
            @click="discardChanges"
          >
            Discard Changes
          </button>
        </div>
      </div>
    </div>

    <progress
      v-if="isLoading || showProgress"
      class="my-5 progress is-large is-warning"
      max="100"
    >
      Loading…
    </progress>

    <div v-if="error" class="message is-danger">
      <div class="message-header">{{ error.name }}</div>
      <div class="message-body">
        <p class="content">{{ error.message }}</p>
        <div class="buttons">
          <button
            class="button is-danger has-text-weight-semibold"
            type="button"
            @click="fetch(id)"
          >
            Reload?
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.clamped-3 {
  display: -webkit-box;
  -webkit-box-orient: vertical;
  -webkit-line-clamp: 3;
  overflow: hidden;
}
</style>
