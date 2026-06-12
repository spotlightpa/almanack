<script>
import { computed, reactive, toRefs, watch } from "vue";

import { makeState } from "@/api/service-util.js";
import {
  get as clientGet,
  post as clientPost,
  getPage,
  listImages,
  postPage,
} from "@/api/client-v2.js";
import imgproxyURL from "@/api/imgproxy-url.js";
import { toRel } from "@/utils/link.js";
import maybeDate from "@/utils/maybe-date.js";

class TaxonomyPage {
  constructor(data) {
    this.id = data["id"] ?? "";
    this.body = data["body"] ?? "";
    this.frontmatter = data["frontmatter"] ?? {};
    this.filePath = data["file_path"] ?? "";
    this.urlPath = data["url_path"] ?? "";
    this.lastPublished = maybeDate(data, "last_published");
    this.publicationDate = maybeDate(this.frontmatter, "published");
    this.kicker = this.frontmatter["kicker"] ?? "";
    this.title = this.frontmatter["title"] ?? "";
    this.linkTitle = this.frontmatter["linktitle"] ?? "";
    this.titleTag = this.frontmatter["title-tag"] ?? "";
    this.summary = this.frontmatter["description"] ?? "";
    this.appImage = this.frontmatter["app-image"] ?? "";
    this.appImageGravity = this.frontmatter["app-image-gravity"] ?? "";
    this.appImageDescription = this.frontmatter["app-image-description"] ?? "";
    this.appImageCredit = this.frontmatter["app-image-credit"] ?? "";
    this.image = this.frontmatter["image"] ?? "";
    this.imageGravity = this.frontmatter["image-gravity"] ?? "";
    this.imageDescription = this.frontmatter["image-description"] ?? "";
    this.imageCredit = this.frontmatter["image-credit"] ?? "";
    this.imageSize = this.frontmatter["image-size"] ?? "";
    this.modalExclude = this.frontmatter["modal-exclude"] ?? false;
    this.suppressAds = this.frontmatter["suppress-ads"] ?? false;
    this.noIndex = this.frontmatter["no-index"] ?? null;
    this.overrideURL = this.frontmatter["url"] ?? "";
    this.aliases = this.frontmatter["aliases"] ?? [];
    this.layout = this.frontmatter["layout"] ?? "";
    this.hideDescription = this.frontmatter["hide-description"] ?? false;
    this.descriptionHed = this.frontmatter["description-hed"] ?? "";
    this.descriptionDek = this.frontmatter["description-dek"] ?? "";

    this.shouldUpdateURLPath = false;
  }

  get taxoName() {
    return this.filePath?.replace(
      /^content\/(topics|series)\/(.+)\/_index\.md$/,
      "$2"
    );
  }

  get taxoKind() {
    return this.filePath?.replace(
      /^content\/(topics|series)\/(.+)\/_index\.md$/,
      "$1"
    );
  }

  get taxoLink() {
    if (this.taxoKind === "topics") {
      return { name: "Topic Pages", to: { name: "topic-pages" } };
    }
    return {
      name: "Investigation Series Pages",
      to: { name: "series-pages" },
    };
  }

  get isPublished() {
    return !!this.lastPublished;
  }

  get isFutureDated() {
    return this.publicationDate && this.publicationDate > new Date();
  }

  get link() {
    if (this.urlPath) {
      return new URL(this.urlPath, "https://www.spotlightpa.org").href;
    }
    if (this.overrideURL) {
      return new URL(this.overrideURL, "https://www.spotlightpa.org").href;
    }
    return "";
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

  toJSON() {
    return {
      id: this.id,
      set_frontmatter: true,
      frontmatter: {
        // preserve unknown props
        ...this.frontmatter,
        // copy others
        published: this.publicationDate,
        kicker: this.kicker,
        title: this.title,
        linktitle: this.linkTitle,
        "title-tag": this.titleTag,
        description: this.summary,
        "app-image": this.appImage,
        "app-image-gravity": this.appImageGravity,
        "app-image-description": this.appImageDescription,
        "app-image-credit": this.appImageCredit,
        image: this.image,
        "image-gravity": this.imageGravity,
        "image-description": this.imageDescription,
        "image-credit": this.imageCredit,
        "image-size": this.imageSize,
        "modal-exclude": this.modalExclude,
        "suppress-ads": this.suppressAds,
        "no-index": this.noIndex,
        url: toRel(this.overrideURL),
        aliases: this.aliases,
        layout: this.layout,
        "hide-description": this.hideDescription,
        "description-hed": this.descriptionHed,
        "description-dek": this.descriptionDek,
      },
      set_body: true,
      body: this.body,
      set_schedule_for: false,
      url_path: this.shouldUpdateURLPath ? this.urlPath : "",
      set_last_published: false,
    };
  }
}

function usePage(id) {
  const { apiState, exec } = makeState();

  const fetch = (id) =>
    exec(() =>
      clientGet(getPage, { by: "id", value: id, refresh_content_store: true })
    );
  const post = (page) => exec(() => clientPost(postPage, page));

  const page = computed(() =>
    apiState.rawData ? reactive(new TaxonomyPage(apiState.rawData)) : null
  );

  watch(() => id.value, fetch, {
    immediate: true,
  });

  const { apiState: imageState, exec: execImage } = makeState();
  execImage(() => clientGet(listImages));

  return {
    ...toRefs(apiState),
    fetch,
    post,
    page,

    discardChanges() {
      if (window.confirm("Do you really want to discard all changes?")) {
        fetch(id.value);
      }
    },
    save(formEl) {
      if (!formEl.reportValidity()) {
        return;
      }
      return post(page.value);
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
    const title = computed(() => {
      if (!pageData.page.value) {
        return `Landing page ${id.value}`;
      }
      return pageData.page.value.taxoName || "Untitled";
    });
    const breadcrumbs = computed(() => {
      return pageData.page.value
        ? [
            { name: "Admin", to: { name: "admin" } },
            pageData.page.value.taxoLink,
            {
              name: title.value,
              to: { name: "taxonomy-page", params: { id: id.value } },
            },
          ]
        : [
            { name: "Admin", to: { name: "admin" } },
            {
              name: title.value,
              to: { name: "taxonomy-page", params: { id: id.value } },
            },
          ];
    });
    return {
      ...pageData,
      title,
      breadcrumbs,
    };
  },
};
</script>

<template>
  <div>
    <MetaHead>
      <title>{{ title }} • Spotlight PA Almanack</title>
    </MetaHead>
    <BulmaBreadcrumbs :links="breadcrumbs"></BulmaBreadcrumbs>

    <h1 class="mb-2 is-spaced title">
      {{ title }}
    </h1>

    <form v-if="page" ref="form">
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

      <BulmaFieldInput
        id="hed"
        v-model="page.title"
        label="Title"
        :help="`Title on the ${page.taxoKind} landing page`"
        :required="true"
      ></BulmaFieldInput>

      <BulmaFieldInput
        id="eyebrow"
        v-model="page.kicker"
        :label="`Eyebrow on the ${page.taxoKind} landing page`"
        :help="`This should be a short version of the title for the list of all ${page.taxoKind} page`"
        autocomplete="off"
      ></BulmaFieldInput>

      <BulmaFieldInput
        v-model="page.linkTitle"
        label="Landing page dek"
        :help="`Dek used on the list of all ${page.taxoKind} page`"
      ></BulmaFieldInput>

      <BulmaFieldInput
        id="seo"
        v-model="page.titleTag"
        label="SEO Hed"
        help="If set, this is the land page title seen by search engines"
      ></BulmaFieldInput>
      <BulmaCharLimit
        :warn="40"
        :max="55"
        :value="page.titleTag"
        class="mt-1 mb-4"
      ></BulmaCharLimit>

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

      <BulmaFieldCheckbox
        v-model="page.hideDescription"
        label="Hide description box"
      >
        Hide the black description box on the landing page
      </BulmaFieldCheckbox>
      <BulmaFieldInput
        v-model="page.descriptionHed"
        label="Description box hed"
        help='Hed used in the black description box on the landing page. Defaults to "About our TITLE coverage" if blank.'
      ></BulmaFieldInput>
      <BulmaTextarea
        v-model="page.descriptionDek"
        label="Description box dek"
        help="Dek used in the black description box on the landing page. Defaults to SEO Description if blank. Markdown okay."
      ></BulmaTextarea>

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
        v-model="page.body"
        label="Content"
        :rows="8"
      ></BulmaTextarea>

      <BulmaFieldCheckbox v-model="page.modalExclude" label="Hide pop-up ads">
        Don't trigger Wisepops and newsletter modal screens on this page
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
        help="Redirect these URLs to the page"
        :options="[]"
      ></BulmaAutocompleteArray>

      <BulmaField v-slot="{ idForLabel }" label="Layout override">
        <input v-model="page.layout" class="input" :list="idForLabel" />
        <datalist :id="idForLabel">
          <option value="blank"></option>
          <option value="featured"></option>
        </datalist>
      </BulmaField>

      <BulmaDateTime
        v-model="page.publicationDate"
        label="Publication Date"
        :icon="['fas', 'user-clock']"
        help="Page will be listed on the site under this date"
      >
        <p class="content is-small">
          <a
            href="#"
            class="has-text-info"
            @click.prevent="page.publicationDate = new Date()"
          >
            Set to now
          </a>
        </p>
      </BulmaDateTime>

      <p v-if="page.isFutureDated" class="content has-text-warning is-small">
        Publication date is in the future.
      </p>

      <div class="field">
        <div class="buttons">
          <button
            class="button is-success has-text-weight-semibold"
            :disabled="isLoading || null"
            type="button"
            @click="save($refs.form)"
          >
            Save page
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
