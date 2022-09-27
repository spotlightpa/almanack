import { computed, reactive, ref, toRefs, watch } from "vue";

import { makeState } from "@/api/service-util.js";
import { useClient } from "@/api/client.js";
import imgproxyURL from "@/api/imgproxy-url.js";

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
    this.publishedAt = Page.getDate(this.frontmatter, "published");
    this.updatedAt = Page.getDate(data, "updated_at");
    this.lastPublished = Page.getDate(data, "last_published");
    this.scheduleFor = Page.getDate(data, "schedule_for");
    this.eventDate = Page.getDate(this.frontmatter, "event-date");
    this.arcID = this.frontmatter["arc-id"] ?? "";
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
    this.imageGravity = this.frontmatter["image-gravity"] ?? "";
    this.imageDescription = this.frontmatter["image-description"] ?? "";
    this.imageCredit = this.frontmatter["image-credit"] ?? "";
    this.imageSize = this.frontmatter["image-size"] ?? "";
    this.languageCode = this.frontmatter["language-code"] ?? "";
    this.slug = this.frontmatter["slug"] ?? "";
    this.extendedKicker = this.frontmatter["extended-kicker"] ?? "";
    this.modalExclude = this.frontmatter["modal-exclude"] ?? false;
    this.suppressDate = this.frontmatter["suppress-date"] ?? false;
    this.noIndex = this.frontmatter["no-index"] ?? "";
    this.overrideURL = this.frontmatter["url"] ?? "";
    this.aliases = this.frontmatter["aliases"] ?? [];
    this.layout = this.frontmatter["layout"] ?? "";

    // not a getter so it won't react to changes
    this.status = "pub";
    if (!this.lastPublished) {
      this.status = this.scheduleFor ? "sked" : "none";
    }
  }

  static getDate(data, prop) {
    let date = data[prop] ?? null;
    return date && new Date(date);
  }

  get isPublished() {
    return !!this.lastPublished;
  }

  get isFutureDated() {
    return this.publishedAt && this.publishedAt > new Date();
  }

  get statusVerbose() {
    return {
      pub: "published",
      sked: "scheduled to be published",
      none: "unpublished",
    }[this.status];
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
      let date = this.scheduleFor ? new Date(this.scheduleFor) : new Date();
      let year = date.getFullYear();
      let month = (date.getMonth() + 1).toString().padStart(2, "0");
      dir = `${dir}/${year}/${month}`;
    }
    return new URL(`/${dir}/${slug}/`, "https://www.spotlightpa.org").href;
  }

  getImagePreviewURL(options) {
    if (!this.image || this.image.match(/^http/)) {
      return "";
    }
    return imgproxyURL(this.image, options);
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

  get sharedViewRoute() {
    if (!this.arcID) {
      return "";
    }
    return { name: "article", params: { id: this.arcID } };
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
        published: this.publishedAt,
        "event-date": this.eventDate,
        kicker: this.kicker || this.mainTopic,
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
        "image-gravity": this.imageGravity,
        "image-description": this.imageDescription,
        "image-credit": this.imageCredit,
        "image-size": this.imageSize,
        "language-code": this.languageCode,
        slug: this.slug,
        "extended-kicker": this.extendedKicker,
        "modal-exclude": this.modalExclude,
        "suppress-date": this.suppressDate,
        "no-index": this.noIndex,
        url: this.overrideURL,
        aliases: this.aliases,
        layout: this.layout,
      },
      set_body: true,
      body: this.body,
      set_schedule_for: true,
      schedule_for: this.scheduleFor,
      url_path: "", // leave blank to prevent changes
      set_last_published: false,
    };
  }
}

function useAutocompletions() {
  let { listAllTopics, listAllSeries } = useClient();
  const autocomplete = reactive({
    topics: [],
    series: [],
  });

  listAllTopics().then(([data, err]) => {
    if (!err) {
      autocomplete.topics = data.topics || [];
    } else {
      // eslint-disable-next-line no-console
      console.warn(err);
    }
  });
  listAllSeries().then(([data, err]) => {
    if (!err) {
      autocomplete.series = data.series || [];
    } else {
      // eslint-disable-next-line no-console
      console.warn(err);
    }
  });
  return autocomplete;
}

export function usePage(id) {
  const {
    getPageWithContent,
    postPage,
    listImages,
    postRefreshPageFromArc,
    postRefreshPageFromMailchimp,
  } = useClient();
  const { apiState, exec } = makeState();

  const fetch = (id) => exec(() => getPageWithContent(id));
  const post = (page) => exec(() => postPage(page));

  const page = computed(() =>
    apiState.rawData ? reactive(new Page(apiState.rawData)) : null
  );

  watch(() => id.value, fetch, {
    immediate: true,
  });

  const { apiState: imageState, exec: execImage } = makeState();
  execImage(() => listImages());

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
    arcRefresh() {
      return exec(() => postRefreshPageFromArc(id.value));
    },
    mailchimpRefresh() {
      return exec(() => postRefreshPageFromMailchimp(id.value));
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
