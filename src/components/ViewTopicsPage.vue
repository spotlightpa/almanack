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

class Page {
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
    this.blurb = this.frontmatter["blurb"] ?? "";
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
    this.modalExclude = this.frontmatter["modal-exclude"] ?? false;
    this.suppressAds = this.frontmatter["suppress-ads"] ?? false;
    this.noIndex = this.frontmatter["no-index"] ?? null;
    this.overrideURL = this.frontmatter["url"] ?? "";
    this.aliases = this.frontmatter["aliases"] ?? [];
    this.layout = this.frontmatter["layout"] ?? "";

    this.shouldUpdateURLPath = false;
  }

  get topicName() {
    return (
      this.filePath.replace(/^content\/topics\/(.+)\/_index\.md$/, "$1") ||
      this.filePath
    );
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
        blurb: this.blurb,
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
        "modal-exclude": this.modalExclude,
        "suppress-ads": this.suppressAds,
        "no-index": this.noIndex,
        url: toRel(this.overrideURL),
        aliases: this.aliases,
        layout: this.layout,
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
    apiState.rawData ? reactive(new Page(apiState.rawData)) : null
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
    return {
      ...pageData,
      title: computed(() => {
        if (!pageData.page.value) {
          return `Topic ${id.value}`;
        }
        return pageData.page.value.topicName || "Untitled";
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
        { name: 'Topic Pages', to: { name: 'topic-pages' } },
        { name: title, to: { name: 'topic-page', params: { id } } },
      ]"
    ></BulmaBreadcrumbs>

    <h1 class="mb-2 is-spaced title">
      {{ title }}
    </h1>
    <div v-if="page && page.link" class="mb-4">
      <a
        :href="page.link"
        class="tag is-primary has-text-weight-semibold"
        target="_blank"
      >
        <span class="icon is-size-6">
          <font-awesome-icon :icon="['fas', 'link']"></font-awesome-icon>
        </span>
        <span>Live URL</span>
      </a>
    </div>

    <form v-if="page" ref="form">
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

      <BulmaFieldInput
        id="eyebrow"
        v-model="page.kicker"
        label="Eyebrow"
        help="Small text appearing above the page hed"
        autocomplete="off"
      ></BulmaFieldInput>
      <BulmaCharLimit
        :warn="15"
        :max="20"
        :value="page.kicker"
        class="mt-1 mb-4"
      ></BulmaCharLimit>

      <BulmaFieldInput
        id="hed"
        v-model="page.title"
        label="Hed"
        help="Hed on the page and the default value for link title and SEO title"
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

      <details class="field">
        <summary class="has-text-weight-semibold">Advanced options</summary>

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
      </details>

      <BulmaWarnings
        :values="[
          [page.kicker.length < 1, '#eyebrow', 'Eyebrow is unset'],
          [page.kicker.length > 20, '#eyebrow', 'Eyebrow is long'],
          [page.title.length < 1, '#hed', 'Hed is unset'],
          [page.title.length > 100, '#hed', 'Hed is long'],
          [page.titleTag.length < 1, '#seo', 'SEO hed is unset'],
          [page.titleTag.length > 55, '#seo', 'SEO hed is long'],
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
