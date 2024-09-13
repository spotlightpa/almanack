import { computed, reactive, ref, toRefs, watch } from "vue";

import { makeState } from "@/api/service-util.js";
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
import maybeDate from "@/utils/maybe-date.js";

export class Page {
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
    this.suppressDate = this.frontmatter["suppress-date"] ?? false;
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
    if (dir === "news" || dir === "statecollege") {
      let date = this.publicationDate ?? new Date();
      let year = date.getFullYear();
      let month = (date.getMonth() + 1).toString().padStart(2, "0");
      dir = `${dir}/${year}/${month}`;
    }
    return new URL(`/${dir}/${slug}/`, "https://www.spotlightpa.org").href;
  }

  changeURL() {
    if (!this.isPublished) return;
    let oldURLPath = this.link.pathname;
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
      file_path: this.filePath,
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
        "suppress-date": this.suppressDate,
        pinned: this.isPinned,
        "no-index": this.noIndex,
        url: this.overrideURL,
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

function useAutocompletions() {
  const autocomplete = reactive({
    topics: [],
    series: [],
  });
  clientGet(listAllTopics).then(([data, err]) => {
    if (!err) {
      autocomplete.topics = data.topics || [];
    } else {
      console.warn(err);
    }
  });
  clientGet(listAllSeries).then(([data, err]) => {
    if (!err) {
      autocomplete.series = data.series || [];
    } else {
      console.warn(err);
    }
  });
  return autocomplete;
}

export function usePage(id) {
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
    ...toRefs(useAutocompletions()),

    fetch,
    post,
    page,

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
