<script>
import { computed, toRefs } from "@vue/composition-api";
import { usePage } from "../api/spotlightpa-page.js";

import { formatDateTime } from "@/utils/time-format.js";

import BulmaAutocompleteArray from "./BulmaAutocompleteArray.vue";
import BulmaField from "./BulmaField.vue";
import BulmaFieldInput from "./BulmaFieldInput.vue";
import CopyWithButton from "./CopyWithButton.vue";

export default {
  name: "ViewNewsPage",
  components: {
    BulmaAutocompleteArray,
    BulmaField,
    BulmaFieldInput,
    CopyWithButton,
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
    const { id } = toRefs(props);
    const pageData = usePage(id);

    return {
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
    <nav class="breadcrumb has-succeeds-separator" aria-label="breadcrumbs">
      <ul>
        <li>
          <router-link :to="{ name: 'admin' }">Admin</router-link>
        </li>
        <li>
          <router-link exact :to="{ name: 'news-pages' }">
            Spotlight PA Pages
          </router-link>
        </li>
        <li class="is-active">
          <router-link exact :to="{ name: 'news-page', params: { id } }">
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
        <div class="textarea preview-frame">
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

      <CopyWithButton v-if="page.link" :value="page.link" label="Page URL" />

      <div v-if="page.isPublished && page.link" class="buttons">
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

      <BulmaField v-slot="{ idForLabel }" label="Page content">
        <textarea
          :id="idForLabel"
          v-model="page.body"
          class="textarea"
          rows="8"
        ></textarea>
      </BulmaField>

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
.preview-frame {
  height: 300px;
  overflow-y: auto;
}
</style>