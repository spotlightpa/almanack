<script>
import { computed, reactive, ref, toRefs, watch } from "vue";

import { makeState, watchAPI } from "@/api/service-util.js";
import {
  get as clientGet,
  post as clientPost,
  getPage,
  listAllTopics,
  listAllSeries,
  listImages,
  postPage,
  postPageRefresh,
} from "@/api/client-v2.js";
import { processGDocsDoc } from "@/api/gdocs.js";
import imgproxyURL from "@/api/imgproxy-url.js";
import { toRel } from "@/utils/link.js";
import maybeDate from "@/utils/maybe-date.js";

import { formatDateTime } from "@/utils/time-format.js";

class Page {
  constructor(data) {
    this.id = data["id"] ?? "";
    this.body = data["body"] ?? "";
    this.frontmatter = data["frontmatter"] ?? {};
    this.filePath = data["file_path"] ?? "";
    this.urlPath = data["url_path"] ?? "";
    this.sourceType = data["source_type"] ?? "";
    this.sourceID = data["source_id"] ?? "";
    this.createdAt = data["created_at"] ?? "";
    this.publicationDate = maybeDate(this.frontmatter, "published");
    this.updatedAt = maybeDate(data, "updated_at");
    this.lastPublished = maybeDate(data, "last_published");
    this.scheduleFor = maybeDate(data, "schedule_for");
    this.eventDate = maybeDate(this.frontmatter, "event-date");
    this.eventTitle = this.frontmatter["event-title"] ?? "";
    this.eventURL = this.frontmatter["event-url"] ?? "";
    this.arcID = this.frontmatter["arc-id"] ?? "";
    this.kicker = this.frontmatter["kicker"] ?? "";
    this.title = this.frontmatter["title"] ?? "";
    this.internalID = this.frontmatter["internal-id"] ?? "";
    this.linkTitle = this.frontmatter["linktitle"] ?? "";
    this.titleTag = this.frontmatter["title-tag"] ?? "";
    this.ogTitle = this.frontmatter["og-title"] ?? "";
    this.twitterTitle = this.frontmatter["twitter-title"] ?? "";
    this.authors = this.frontmatter["authors"] ?? [];
    this.byline = this.frontmatter["byline"] ?? "";
    this.summary = this.frontmatter["description"] ?? "";
    this.blurb = this.frontmatter["blurb"] ?? "";
    this.topics = this.frontmatter["topics"] ?? [];
    this.series = this.frontmatter["series"] ?? [];
    this.appImage = this.frontmatter["app-image"] ?? "";
    this.appImageGravity = this.frontmatter["app-image-gravity"] ?? "";
    this.appImageDescription = this.frontmatter["app-image-description"] ?? "";
    this.appImageCredit = this.frontmatter["app-image-credit"] ?? "";
    this.image = this.frontmatter["image"] ?? "";
    this.imageGravity = this.frontmatter["image-gravity"] ?? "";
    this.imageDescription = this.frontmatter["image-description"] ?? "";
    this.imageCaption = this.frontmatter["image-caption"] ?? "";
    this.imageCredit = this.frontmatter["image-credit"] ?? "";
    this.imageSize = this.frontmatter["image-size"] ?? "";
    this.languageCode = this.frontmatter["language-code"] ?? "";
    this.slug = this.frontmatter["slug"] ?? "";
    this.extendedKicker = this.frontmatter["extended-kicker"] ?? "";
    this.modalExclude = this.frontmatter["modal-exclude"] ?? false;
    this.suppressAds = this.frontmatter["suppress-ads"] ?? false;
    this.suppressDate = this.frontmatter["suppress-date"] ?? false;
    this.isDraft = this.frontmatter["draft"] ?? false;
    this.isPinned = this.frontmatter["pinned"] ?? false;
    this.noIndex = this.frontmatter["no-index"] ?? null;
    this.overrideURL = this.frontmatter["url"] ?? "";
    this.aliases = this.frontmatter["aliases"] ?? [];
    this.layout = this.frontmatter["layout"] ?? "";
    this.feedExclude = this.frontmatter["feed-exclude"] ?? false;
    this.contentSource = this.frontmatter["content-source"] ?? "";

    // not a getter so it won't react to changes
    this.status = "pub";
    if (!this.lastPublished) {
      this.status = this.scheduleFor ? "sked" : "none";
    }
    this.shouldUpdateURLPath = false;
  }

  get isPublished() {
    return !!this.lastPublished;
  }

  get isFutureDated() {
    return this.publicationDate && this.publicationDate > new Date();
  }

  get statusVerbose() {
    return {
      pub: "published",
      sked: "scheduled to be published",
      none: "unpublished",
    }[this.status];
  }

  get isGDoc() {
    return this.sourceType === "gdocs";
  }

  get link() {
    if (this.urlPath) {
      return new URL(this.urlPath, "https://www.spotlightpa.org").href;
    }
    if (this.overrideURL) {
      return new URL(this.overrideURL, "https://www.spotlightpa.org").href;
    }
    let [, dir, fname] = this.filePath.match(/^content\/(.+)\/([^/]+)\.md/);
    let slug = this.slug || fname;
    if (
      dir === "news" ||
      dir === "statecollege" ||
      dir === "berks" ||
      dir === "sponsored"
    ) {
      let date = this.publicationDate ?? new Date();
      let year = date.getFullYear();
      let month = (date.getMonth() + 1).toString().padStart(2, "0");
      dir = `${dir}/${year}/${month}`;
    }
    return new URL(`/${dir}/${slug}/`, "https://www.spotlightpa.org").href;
  }

  changeURL() {
    if (!this.isPublished) return;
    let oldURLPath = new URL(this.link).pathname;
    let message = `Are you sure you want to change the URL? Current URL is ${oldURLPath}. Changing the URL will automatically add a redirect from the old URL to a new one. Please enter new URL below.`;
    let newURLPath = window.prompt(message, oldURLPath);
    if (!newURLPath || newURLPath === oldURLPath) return;
    let newURL = new URL(newURLPath, "https://www.spotlightpa.org");
    newURLPath = newURL.pathname;
    this.aliases.push(oldURLPath);
    this.overrideURL = newURLPath;
    this.urlPath = newURLPath;
    this.shouldUpdateURLPath = true;
  }

  getImagePreviewURL(options) {
    if (!this.image || this.image.match(/^http/)) {
      return "";
    }
    return imgproxyURL(this.image, options);
  }

  getAppImagePreviewURL() {
    if (!this.appImage || this.appImage.match(/^http/)) {
      return this.getImagePreviewURL({
        width: 400,
        height: 500,
        gravity: this.imageGravity,
      });
    }
    return imgproxyURL(this.appImage, {
      width: 400,
      height: 500,
      gravity: this.appImageGravity,
    });
  }

  get imagePreviewURL() {
    return this.getImagePreviewURL();
  }

  get arcURL() {
    if (!this.arcID) {
      return "";
    }
    return `https://pmn.arcpublishing.com/composer/edit/${this.arcID}/`;
  }

  get gdocsURL() {
    return !this.isGDoc
      ? ""
      : `https://docs.google.com/document/d/${this.sourceID}/edit`;
  }

  get sharedViewRoute() {
    return {
      name: "shared-article-redirect-from-page",
      query: { id: this.sourceID, source_type: this.sourceType },
    };
  }

  get sharedAdminRoute() {
    return {
      name: "shared-article-admin-redirect-from-page",
      query: { id: this.sourceID, source_type: this.sourceType },
    };
  }

  get mainTopic() {
    return this.topics[0] || "";
  }

  get parentPage() {
    if (this.filePath.match(/content\/statecollege/)) {
      return {
        name: "State College Pages",
        to: { name: "statecollege-pages" },
      };
    }
    return { name: "Spotlight PA Pages", to: { name: "news-pages" } };
  }

  toJSON() {
    return {
      id: this.id,
      set_frontmatter: true,
      frontmatter: {
        // preserve unknown props
        ...this.frontmatter,
        // copy others
        published: this.publicationDate,
        "event-date": this.eventDate,
        "event-title": this.eventTitle,
        "event-url": this.eventURL,
        kicker: this.kicker || this.mainTopic,
        title: this.title,
        "internal-id": this.internalID,
        linktitle: this.linkTitle,
        "title-tag": this.titleTag,
        "og-title": this.ogTitle,
        "twitter-title": this.twitterTitle,
        authors: this.authors,
        byline: this.byline,
        description: this.summary,
        blurb: this.blurb,
        topics: this.topics,
        series: this.series,
        "app-image": this.appImage,
        "app-image-gravity": this.appImageGravity,
        "app-image-description": this.appImageDescription,
        "app-image-credit": this.appImageCredit,
        image: this.image,
        "image-gravity": this.imageGravity,
        "image-description": this.imageDescription,
        "image-caption": this.imageCaption,
        "image-credit": this.imageCredit,
        "image-size": this.imageSize,
        "language-code": this.languageCode,
        slug: this.slug,
        "extended-kicker": this.extendedKicker,
        "modal-exclude": this.modalExclude,
        "suppress-ads": this.suppressAds,
        "suppress-date": this.suppressDate,
        draft: this.isDraft,
        pinned: this.isPinned,
        "no-index": this.noIndex,
        url: toRel(this.overrideURL),
        aliases: this.aliases,
        layout: this.layout,
        "feed-exclude": this.feedExclude,
        "content-source": this.contentSource,
      },
      set_body: true,
      body: this.body,
      set_schedule_for: true,
      schedule_for: this.scheduleFor,
      // leave blank to prevent changes by default
      url_path: this.shouldUpdateURLPath ? this.urlPath : "",
      set_last_published: false,
    };
  }
}

function useTopics() {
  const { computedList } = watchAPI(
    () => null,
    () => clientGet(listAllTopics)
  );
  return computedList("pages", (obj) =>
    obj.file_path?.replace(/^content\/topics\/(.+)\/_index\.md$/, "$1")
  );
}

function useSeries() {
  const { computedObj } = watchAPI(
    () => null,
    () => clientGet(listAllSeries)
  );
  return computedObj((obj) => obj.series);
}

function usePage(id) {
  const { apiState, exec } = makeState();

  const fetch = (id) =>
    exec(() =>
      clientGet(getPage, { by: "id", value: id, refresh_content_store: true })
    );
  const post = (page) => exec(() => clientPost(postPage, page));

  const page = computed(() =>
    apiState.rawData ? reactive(new Page(apiState.rawData)) : null
  );

  watch(() => id.value, fetch, {
    immediate: true,
  });

  const { apiState: imageState, exec: execImage } = makeState();
  execImage(() => clientGet(listImages));

  return {
    showScheduler: ref(false),

    ...toRefs(apiState),
    topics: useTopics(),
    series: useSeries(),
    fetch,
    post,
    page,

    deriveSlug() {
      page.value.slug = page.value.title
        .toLowerCase()
        .replace(/\b(the|an?)\b/g, " ")
        .replace(/\bpa\b/g, "pennsylvania")
        .replace(/'/g, "\u2019")
        .replace(/.?'s/g, "s")
        .replace(/\W+/g, " ")
        .trim()
        .replace(/ /g, "-");
    },
    discardChanges() {
      if (window.confirm("Do you really want to discard all changes?")) {
        fetch(id.value);
      }
    },
    publishNow(formEl) {
      if (
        !page.value.isPublished &&
        !window.confirm("Are you sure you want to publish this now?")
      ) {
        return;
      }
      if (!formEl.reportValidity()) {
        return;
      }
      page.value.scheduleFor = new Date();
      return post(page.value);
    },
    updateSchedule(formEl) {
      if (!formEl.reportValidity()) {
        return;
      }
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
    refreshFromSource({ metadata } = {}) {
      return exec(async () => {
        if (page.value.isGDoc) {
          let [, err] = await processGDocsDoc(page.value.sourceID);
          if (err) {
            return [null, err];
          }
        }
        return clientPost(postPageRefresh, {
          id: id.value,
          refresh_metadata: metadata,
        });
      });
    },
    imageState,
    images: computed(() =>
      !imageState.rawData ? [] : imageState.rawData.images
    ),
    setImageProps(image) {
      page.value.image = image.path;
      page.value.imageDescription = image.description;
      page.value.imageCredit = image.credit;
      page.value.imageGravity = "";
    },
  };
}

export default {
  props: {
    id: String,
  },
  setup(props) {
    const { id } = toRefs(props);
    const pageData = usePage(id);
    return {
      parentPage: computed(() => {
        if (!pageData.page.value) {
          return { name: "Spotlight PA Pages", to: { name: "news-pages" } };
        }
        return pageData.page.value.parentPage;
      }),
      ...pageData,
      formatDateTime,
      title: computed(() => {
        if (!pageData.page.value) {
          return `Spotlight PA Page ${id.value}`;
        }
        return "Page " + (pageData.page.value.internalID || "Untitled");
      }),
    };
  },
};
</script>

<template>
  <div>
    <MetaHead>
      <title>{{ title }} • Spotlight PA Almanack</title>
    </MetaHead>
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        parentPage,
        { name: title, to: { name: 'news-page', params: { id } } },
      ]"
    ></BulmaBreadcrumbs>

    <h1 class="mb-2 is-spaced title">
      {{ title }}
    </h1>
    <h2 class="subtitle">
      <span class="tags">
        <TagStatus v-if="page" :status="page.status"></TagStatus>
        <router-link
          v-if="page && page.sharedAdminRoute"
          class="tag is-light has-text-weight-semibold"
          :to="page.sharedAdminRoute"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'sliders']"></font-awesome-icon>
          </span>
          <span>Sharing admin</span>
        </router-link>
        <router-link
          v-if="page && page.sharedViewRoute"
          class="tag is-light has-text-weight-semibold"
          :to="page.sharedViewRoute"
        >
          <span class="icon is-size-6">
            <font-awesome-icon
              :icon="['fas', 'file-invoice']"
            ></font-awesome-icon>
          </span>
          <span>Partner view</span>
        </router-link>
        <a
          v-if="page && page.arcURL"
          class="tag is-light has-text-weight-semibold"
          :href="page.arcURL"
          target="_blank"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'link']"></font-awesome-icon>
          </span>
          <span>Arc view</span>
        </a>
        <TagLink
          v-if="page && page.isGDoc"
          :href="page.gdocsURL"
          :icon="['fas', 'link']"
        >
          Google Docs
        </TagLink>
        <a
          v-if="page && page.status === 'pub' && page.link"
          :href="page.link"
          class="tag is-primary has-text-weight-semibold"
          target="_blank"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'link']"></font-awesome-icon>
          </span>
          <span>Live URL</span>
        </a>
      </span>
    </h2>

    <form v-if="page" ref="form">
      <BulmaDateTime
        v-model="page.publicationDate"
        label="Publication Date"
        :icon="['fas', 'user-clock']"
        :disabled="page.isPublished"
        help="Page will be listed on the site under this date"
      >
        <p class="content is-small">
          <a
            v-show="!page.isPublished"
            href="#"
            class="has-text-info"
            @click.prevent="page.publicationDate = new Date()"
          >
            Set to now
          </a>
        </p>
      </BulmaDateTime>

      <p v-if="page.isFutureDated" class="content has-text-warning is-small">
        Article publication date is in the future.
      </p>

      <BulmaAutocompleteArray
        id="topics"
        v-model="page.topics"
        label="Topics"
        :options="topics"
        help="Topics are open-ended collections, e.g. “Events”, “Coronavirus”"
      ></BulmaAutocompleteArray>

      <div v-show="page.topics.includes('Events')">
        <BulmaDateTime
          v-model="page.eventDate"
          label="Event Date"
          help="If present, the events landing page will show this date for the event"
          icon="user-clock"
        ></BulmaDateTime>
        <BulmaFieldInput
          v-model="page.eventTitle"
          label="Name of Event"
          help="Shown in search results"
        ></BulmaFieldInput>
        <BulmaFieldInput
          v-model="page.eventURL"
          label="Registration link"
          type="url"
          help="Shown in search results"
        ></BulmaFieldInput>
      </div>

      <BulmaAutocompleteArray
        v-model="page.series"
        label="Series"
        :options="series"
        help="Series are limited-time collections, e.g. “Legislative privilege 2020”"
      ></BulmaAutocompleteArray>

      <BulmaFieldInput
        v-model="page.extendedKicker"
        placeholder="Top News"
        label="Homepage extended eyebrow (e.g. “Top News” if blank)"
      ></BulmaFieldInput>

      <BulmaFieldInput
        id="eyebrow"
        v-model="page.kicker"
        label="Eyebrow"
        help="Small text appearing above the page hed"
        :placeholder="page.mainTopic"
        autocomplete="off"
        @focusout="
          if (!page.kicker) {
            page.kicker = page.mainTopic;
          }
        "
      ></BulmaFieldInput>
      <BulmaCharLimit
        :warn="15"
        :max="20"
        :value="page.kicker"
        class="mt-1 mb-4"
      ></BulmaCharLimit>

      <BulmaFieldCheckbox v-model="page.isPinned" label="Pin article">
        Pin article to the top of topic and series landing pages
      </BulmaFieldCheckbox>

      <BulmaFieldInput
        id="hed"
        v-model="page.title"
        label="Hed"
        help="Hed on the page and the default value for link title, SEO title, and share titles"
        :required="true"
      ></BulmaFieldInput>
      <BulmaCharLimit
        :warn="90"
        :max="100"
        :value="page.title"
        class="mt-1 mb-4"
      ></BulmaCharLimit>

      <BulmaFieldInput
        v-model="page.linkTitle"
        label="Link to as"
        help="When linking to this page from the homepage or an article list, use this as the link title instead of the hed"
      ></BulmaFieldInput>

      <BulmaFieldInput
        id="seo"
        v-model="page.titleTag"
        label="SEO Hed"
        help="If set, this is the title seen by search engines"
      ></BulmaFieldInput>
      <BulmaCharLimit
        :warn="40"
        :max="55"
        :value="page.titleTag"
        class="mt-1 mb-4"
      ></BulmaCharLimit>

      <BulmaFieldInput
        id="facebook"
        v-model="page.ogTitle"
        label="FaceBook Hed"
        help="If set, this overrides the SEO hed on Facebook"
      ></BulmaFieldInput>
      <BulmaCharLimit
        :warn="60"
        :max="80"
        :value="page.ogTitle"
        class="mt-1 mb-4"
      ></BulmaCharLimit>

      <BulmaFieldInput
        id="twitter"
        v-model="page.twitterTitle"
        label="Twitter Hed"
        help="If set, this overrides the SEO hed on Twitter"
      ></BulmaFieldInput>
      <BulmaCharLimit
        :warn="60"
        :max="70"
        :value="page.twitterTitle"
        class="mt-1 mb-4"
      ></BulmaCharLimit>

      <BulmaAutocompleteArray
        v-model="page.authors"
        label="Authors"
        help="Adds links to and from each listed author page"
        :options="[]"
      ></BulmaAutocompleteArray>

      <BulmaFieldInput
        v-model="page.byline"
        label="Byline"
        help="If present, overrides the byline created from authors list"
      ></BulmaFieldInput>

      <BulmaTextarea
        id="description"
        v-model="page.summary"
        label="SEO Description"
        help="Shown in social share previews and search results"
      ></BulmaTextarea>
      <BulmaCharLimit
        :warn="135"
        :max="150"
        :value="page.summary"
        class="mt-1 mb-4"
      ></BulmaCharLimit>

      <BulmaTextarea
        id="blurb"
        v-model="page.blurb"
        label="Blurb"
        help="Short summary to appear in article rivers"
      ></BulmaTextarea>
      <BulmaCharLimit
        :warn="190"
        :max="200"
        :value="page.blurb"
        class="mt-1 mb-4"
      ></BulmaCharLimit>

      <PickerImages
        :images="images"
        @select-image="setImageProps($event)"
      ></PickerImages>

      <BulmaField
        label="Photo ID"
        help="Image is shown in article rivers and on social media"
        v-slot="{ idForLabel }"
      >
        <div class="is-flex">
          <input :id="idForLabel" v-model="page.image" class="input" />
          <BulmaPaste @paste="page.image = $event"></BulmaPaste>
        </div>
      </BulmaField>

      <div v-if="page.imagePreviewURL" class="is-flex">
        <div>
          <picture class="has-ratio">
            <img
              :src="
                page.getImagePreviewURL({
                  width: 400,
                  height: 267,
                  gravity: page.imageGravity,
                })
              "
              class="is-3x2 border-thick"
              width="200"
            />
          </picture>
          <p class="has-text-centered">3 x 2</p>
        </div>
        <div class="ml-5">
          <picture class="has-ratio">
            <img
              :src="
                page.getImagePreviewURL({
                  width: 400,
                  height: (400 * 9) / 16,
                  gravity: page.imageGravity,
                })
              "
              class="is-16x9 border-thick"
              width="200"
            />
          </picture>
          <p class="has-text-centered">16 x 9</p>
        </div>
        <div class="ml-5">
          <picture class="has-ratio" style="aspect-ratio: 5/4">
            <img
              :src="page.getAppImagePreviewURL()"
              class="border-thick"
              width="200"
            />
          </picture>
          <p class="has-text-centered">4 x 5</p>
        </div>
      </div>

      <BulmaField label="Image focus">
        <div class="control is-expanded">
          <span class="select is-fullwidth">
            <select v-model="page.imageGravity">
              <option
                v-for="[val, desc] in [
                  ['', 'Auto'],
                  ['we', 'Left'],
                  ['no', 'Top'],
                  ['ea', 'Right'],
                  ['so', 'Bottom'],
                  ['ce', 'Center'],
                ]"
                :key="val"
                :value="val"
              >
                {{ desc }}
              </option>
            </select>
          </span>
        </div>
      </BulmaField>

      <BulmaTextarea
        id="alt"
        v-model="page.imageDescription"
        label="SEO Image Alt Text"
      ></BulmaTextarea>
      <BulmaCharLimit
        :warn="100"
        :max="120"
        :value="page.imageDescription"
        class="mt-1 mb-4"
      ></BulmaCharLimit>

      <BulmaFieldInput
        v-model="page.imageCredit"
        label="Image credit"
      ></BulmaFieldInput>

      <BulmaTextarea
        id="caption"
        v-model="page.imageCaption"
        label="Image Caption"
        help="If set, captions appear as an overlay on top of the image on the article page"
      ></BulmaTextarea>

      <BulmaField label="Image size">
        <div class="control is-expanded">
          <span class="select is-fullwidth">
            <select v-model="page.imageSize">
              <option
                v-for="[val, desc] in [
                  ['inline', 'Normal'],
                  ['hidden', 'Hidden'],
                  ['wide', 'Suppress Right Rail'],
                ]"
                :key="val"
                :value="val"
              >
                {{ desc }}
              </option>
            </select>
          </span>
        </div>
      </BulmaField>

      <details class="my-5">
        <summary>Mobile App Photo Override</summary>
        <BulmaField
          label="Mobile app photo ID"
          help="If present, overrides default photo in the mobile app."
          v-slot="{ idForLabel }"
        >
          <div class="is-flex">
            <input :id="idForLabel" v-model="page.appImage" class="input" />
            <BulmaPaste @paste="page.appImage = $event"></BulmaPaste>
          </div>
        </BulmaField>

        <div v-if="page.appImage" class="is-flex">
          <div>
            <picture class="has-ratio" style="aspect-ratio: 5/4">
              <img
                :src="page.getAppImagePreviewURL()"
                class="border-thick"
                width="200"
              />
            </picture>
            <p class="has-text-centered">4 x 5</p>
          </div>
        </div>

        <BulmaField label="Mobile app image focus">
          <div class="control is-expanded">
            <span class="select is-fullwidth">
              <select v-model="page.appImageGravity">
                <option
                  v-for="[val, desc] in [
                    ['', 'Auto'],
                    ['we', 'Left'],
                    ['no', 'Top'],
                    ['ea', 'Right'],
                    ['so', 'Bottom'],
                    ['ce', 'Center'],
                  ]"
                  :key="val"
                  :value="val"
                >
                  {{ desc }}
                </option>
              </select>
            </span>
          </div>
        </BulmaField>

        <BulmaTextarea
          v-model="page.appImageDescription"
          label="Mobile app image alt text"
        ></BulmaTextarea>

        <BulmaFieldInput
          class="mb-6"
          v-model="page.appImageCredit"
          label="Mobile app image credit"
        ></BulmaFieldInput>
      </details>

      <BulmaFieldInput
        v-model="page.slug"
        label="URL keywords slug"
        :disabled="page.isPublished || null"
        :readonly="page.isPublished || null"
        :required="true"
      ></BulmaFieldInput>
      <button
        class="block button is-small is-light has-text-weight-semibold"
        type="button"
        :disabled="page.isPublished || null"
        @click.prevent="deriveSlug"
      >
        Derive keywords from title
      </button>

      <CopyWithButton
        v-if="page.link"
        :value="page.link"
        label="Page URL"
      ></CopyWithButton>

      <div v-if="page.isPublished && page.link" class="buttons">
        <a
          :href="page.link"
          class="button is-success has-text-weight-semibold"
          target="_blank"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'link']"></font-awesome-icon>
          </span>
          <span> Open live URL </span>
        </a>
        <button
          class="button is-light has-text-weight-semibold"
          type="button"
          @click="page.changeURL()"
        >
          Change URL
        </button>
      </div>

      <BulmaTextarea
        v-model="page.body"
        label="Content"
        :rows="8"
      ></BulmaTextarea>

      <BulmaField help="Remember to save pages after refreshing">
        <div class="buttons">
          <button
            v-if="page.isGDoc"
            class="block button is-warning is-small has-text-weight-semibold"
            :class="{ 'is-loading': isLoadingThrottled }"
            type="button"
            @click.prevent="refreshFromSource({ metadata: false })"
          >
            Refresh content from Google Docs
          </button>
          <button
            v-if="page.isGDoc"
            class="block button is-warning is-small has-text-weight-semibold"
            :class="{ 'is-loading': isLoadingThrottled }"
            type="button"
            @click.prevent="refreshFromSource({ metadata: true })"
          >
            Refresh content and metadata
          </button>
          <a
            v-if="page.isGDoc"
            class="block button is-primary is-small has-text-weight-semibold"
            :href="page.gdocsURL"
            target="_blank"
          >
            Open Google Doc
          </a>
        </div>
      </BulmaField>

      <details class="field">
        <summary class="has-text-weight-semibold">Advanced options</summary>

        <BulmaField label="Content Source">
          <div class="control is-expanded">
            <span class="select is-fullwidth">
              <select v-model="page.contentSource">
                <option value="">Spotlight PA</option>
                <option value="Associated Press">Associated Press</option>
                <option value="Inside Climate News">Inside Climate News</option>
              </select>
            </span>
          </div>
        </BulmaField>
        <BulmaFieldCheckbox
          v-model="page.feedExclude"
          label="Exclude from mobile app feed"
        >
          Don't show article in mobile app list of articles
        </BulmaFieldCheckbox>

        <BulmaField v-slot="{ idForLabel }" label="Language">
          <div class="select is-fullwidth">
            <select :id="idForLabel" v-model="page.languageCode" class="select">
              <option value="">English</option>
              <option value="es">Spanish</option>
            </select>
          </div>
        </BulmaField>

        <BulmaFieldCheckbox
          v-model="page.suppressDate"
          label="Evergreen content"
        >
          Don't show date on page
        </BulmaFieldCheckbox>

        <BulmaFieldCheckbox v-model="page.modalExclude" label="Hide pop-up ads">
          Don't trigger Wisepops and newsletter modal screens on this article
        </BulmaFieldCheckbox>

        <BulmaFieldCheckbox v-model="page.noIndex" label="No index">
          Hide page from Google search results and homepage river
        </BulmaFieldCheckbox>

        <BulmaFieldCheckbox v-model="page.suppressAds" label="Suppress ads">
          Hide ads from header, footer, and sidebar of page
        </BulmaFieldCheckbox>

        <BulmaFieldInput
          v-model="page.overrideURL"
          label="Override URL path"
        ></BulmaFieldInput>

        <BulmaAutocompleteArray
          v-model="page.aliases"
          label="URL Aliases"
          help="Redirect these URLs to the story"
          :options="[]"
        ></BulmaAutocompleteArray>

        <BulmaField v-slot="{ idForLabel }" label="Layout override">
          <input v-model="page.layout" class="input" :list="idForLabel" />
          <datalist :id="idForLabel">
            <option value="blank"></option>
            <option value="featured"></option>
          </datalist>
        </BulmaField>

        <BulmaFieldCheckbox
          v-model="page.isDraft"
          label="Unpublish as draft (for emergency use only!!)"
        >
          Emergency unpublish story by marking as draft.
          <span class="help is-danger">
            Warning: Unpublishing makes the URL a 404 error, but content can
            still be found in public archives.
          </span>
        </BulmaFieldCheckbox>
      </details>

      <BulmaWarnings
        :values="[
          [page.topics.length < 1, '#topics', 'Page topic is unset'],
          [page.kicker.length < 1, '#eyebrow', 'Eyebrow is unset'],
          [page.kicker.length > 20, '#eyebrow', 'Eyebrow is long'],
          [page.title.length < 1, '#hed', 'Hed is unset'],
          [page.title.length > 100, '#hed', 'Hed is long'],
          [page.titleTag.length < 1, '#seo', 'SEO hed is unset'],
          [page.titleTag.length > 55, '#seo', 'SEO hed is long'],
          [page.ogTitle.length > 80, '#facebook', 'Facebook hed is long'],
          [page.twitterTitle.length > 70, '#twitter', 'Twitter hed is long'],
          [page.summary.length < 1, '#description', 'SEO description is unset'],
          [
            page.summary.length > 150,
            '#description',
            'SEO description is long',
          ],
          [
            page.imageDescription.length < 1,
            '#alt',
            'Image description is unset',
          ],
          [
            page.imageDescription.length > 120,
            '#alt',
            'Image description is long',
          ],
        ]"
      ></BulmaWarnings>

      <p class="my-4 has-text-weight-semibold">
        Page is {{ page.statusVerbose
        }}<template v-if="page.status == 'sked'">
          at {{ formatDateTime(page.scheduleFor) }}</template
        >.
      </p>

      <div v-if="page.status !== 'pub'" class="field mb-5">
        <BulmaDateTime
          v-model="page.scheduleFor"
          :label="
            page.status === 'sked'
              ? `Scheduled for ${formatDateTime(page.scheduleFor)}`
              : `Schedule for`
          "
          icon="user-clock"
        >
          <p v-if="page.isFutureDated" class="mt-2 content is-small">
            <a
              href="#"
              class="has-text-info"
              @click.prevent="page.scheduleFor = page.publicationDate"
            >
              Schedule for publication date
            </a>
          </p>
        </BulmaDateTime>
      </div>
      <div class="field">
        <div class="buttons">
          <button
            class="button is-success has-text-weight-semibold"
            :disabled="isLoading || null"
            type="button"
            @click="publishNow($refs.form)"
          >
            {{ page.status === "pub" ? "Update page" : "Publish now" }}
          </button>
          <button
            v-if="page.status !== 'pub'"
            class="button is-warning has-text-weight-semibold"
            :disabled="
              isLoading ||
              !page.scheduleFor ||
              page.scheduleFor < new Date() ||
              null
            "
            type="button"
            @click="updateSchedule($refs.form)"
          >
            {{
              page.status === "none" ? "Schedule to publish" : "Save changes"
            }}
          </button>
          <button
            v-if="page.status === 'sked'"
            class="button is-danger has-text-weight-semibold"
            :disabled="isLoading || null"
            type="button"
            @click="updateOnly"
          >
            Unschedule
          </button>
          <button
            v-if="page.status === 'none'"
            class="button is-light has-text-weight-semibold"
            :disabled="isLoading || null"
            type="button"
            @click="updateOnly"
          >
            Save without publishing
          </button>

          <button
            class="button is-light has-text-weight-semibold"
            :disabled="isLoading || null"
            type="button"
            @click="discardChanges"
          >
            Discard Changes
          </button>
        </div>
      </div>
    </form>

    <SpinnerProgress :is-loading="isLoadingThrottled"></SpinnerProgress>
    <div class="my-5">
      <ErrorReloader :error="error" @reload="fetch(id)"></ErrorReloader>
    </div>
  </div>
</template>

<style scoped>
.border-thick {
  border: 2px solid #ccc;
}
</style>
