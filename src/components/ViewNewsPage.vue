<script>
import { computed, toRefs } from "vue";

import { usePage } from "@/api/spotlightpa-page.js";

import { formatDateTime } from "@/utils/time-format.js";

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
    />

    <h1 class="mb-2 is-spaced title">
      {{ title }}
    </h1>
    <h2 class="subtitle">
      <span class="tags">
        <TagStatus v-if="page" :status="page.status" />
        <router-link
          v-if="page && page.sharedAdminRoute"
          class="tag is-light has-text-weight-semibold"
          :to="page.sharedAdminRoute"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'sliders']" />
          </span>
          <span>Sharing admin</span>
        </router-link>
        <router-link
          v-if="page && page.sharedViewRoute"
          class="tag is-light has-text-weight-semibold"
          :to="page.sharedViewRoute"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'file-invoice']" />
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
            <font-awesome-icon :icon="['fas', 'link']" />
          </span>
          <span>Arc view</span>
        </a>
        <a
          v-if="page && page.status === 'pub' && page.link"
          :href="page.link"
          class="tag is-primary has-text-weight-semibold"
          target="_blank"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'link']" />
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
      />

      <div v-show="page.topics.includes('Events')">
        <BulmaDateTime
          v-model="page.eventDate"
          label="Event Date"
          help="If present, the events landing page will show this date for the event"
          icon="user-clock"
        />
        <BulmaFieldInput
          v-model="page.eventTitle"
          label="Name of Event"
          help="Shown in search results"
        />
        <BulmaFieldInput
          v-model="page.eventURL"
          label="Registration link"
          type="url"
          help="Shown in search results"
        />
      </div>

      <BulmaAutocompleteArray
        v-model="page.series"
        label="Series"
        :options="series"
        help="Series are limited-time collections, e.g. “Legislative privilege 2020”"
      />

      <BulmaFieldInput
        v-model="page.extendedKicker"
        placeholder="Top News"
        label="Homepage extended eyebrow (e.g. “Top News” if blank)"
      />

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
      />
      <BulmaCharLimit
        :warn="15"
        :max="20"
        :value="page.kicker"
        class="mt-1 mb-4"
      />

      <BulmaFieldInput
        id="hed"
        v-model="page.title"
        label="Hed"
        help="Hed on the page and the default value for link title, SEO title, and share titles"
        :required="true"
      />
      <BulmaCharLimit
        :warn="90"
        :max="100"
        :value="page.title"
        class="mt-1 mb-4"
      />

      <BulmaFieldInput
        v-model="page.linkTitle"
        label="Link to as"
        help="When linking to this page from the homepage or an article list, use this as the link title instead of the hed"
      />

      <BulmaFieldInput
        id="seo"
        v-model="page.titleTag"
        label="SEO Hed"
        help="If set, this is the title seen by search engines"
      />
      <BulmaCharLimit
        :warn="40"
        :max="55"
        :value="page.titleTag"
        class="mt-1 mb-4"
      />

      <BulmaFieldInput
        id="facebook"
        v-model="page.ogTitle"
        label="FaceBook Hed"
        help="If set, this overrides the SEO hed on Facebook"
      />
      <BulmaCharLimit
        :warn="60"
        :max="80"
        :value="page.ogTitle"
        class="mt-1 mb-4"
      />

      <BulmaFieldInput
        id="twitter"
        v-model="page.twitterTitle"
        label="Twitter Hed"
        help="If set, this overrides the SEO hed on Twitter"
      />
      <BulmaCharLimit
        :warn="60"
        :max="70"
        :value="page.twitterTitle"
        class="mt-1 mb-4"
      />

      <BulmaAutocompleteArray
        v-model="page.authors"
        label="Authors"
        help="Adds links to and from each listed author page"
        :options="[]"
      />

      <BulmaFieldInput
        v-model="page.byline"
        label="Byline"
        help="If present, overrides the byline created from authors list"
      />

      <BulmaTextarea
        id="description"
        v-model="page.summary"
        label="SEO Description"
        help="Shown in social share previews and search results"
      />
      <BulmaCharLimit
        :warn="135"
        :max="150"
        :value="page.summary"
        class="mt-1 mb-4"
      />

      <BulmaTextarea
        id="blurb"
        v-model="page.blurb"
        label="Blurb"
        help="Short summary to appear in article rivers"
      />
      <BulmaCharLimit
        :warn="190"
        :max="200"
        :value="page.blurb"
        class="mt-1 mb-4"
      />

      <PickerImages :images="images" @select-image="setImageProps($event)" />

      <BulmaField
        label="Photo ID"
        help="Image is shown in article rivers and on social media"
        v-slot="{ idForLabel }"
      >
        <div class="is-flex">
          <input :id="idForLabel" v-model="page.image" class="input" />
          <BulmaPaste @paste="page.image = $event" />
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
      />
      <BulmaCharLimit
        :warn="100"
        :max="120"
        :value="page.imageDescription"
        class="mt-1 mb-4"
      />

      <BulmaFieldInput v-model="page.imageCredit" label="Image credit" />

      <BulmaTextarea
        id="caption"
        v-model="page.imageCaption"
        label="Image Caption"
        help="If set, captions appear as an overlay on top of the image on the article page"
      />

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

      <BulmaFieldInput
        v-model="page.slug"
        label="URL keywords slug"
        :disabled="page.isPublished || null"
        :readonly="page.isPublished || null"
        :required="true"
      />
      <button
        class="block button is-small is-light has-text-weight-semibold"
        type="button"
        :disabled="page.isPublished || null"
        @click.prevent="deriveSlug"
      >
        Derive keywords from title
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

      <BulmaTextarea v-model="page.body" label="Content" :rows="8" />

      <BulmaField help="Remember to save pages after refreshing">
        <div class="buttons">
          <button
            v-if="page.arcID"
            class="block button is-warning is-small has-text-weight-semibold"
            :class="{ 'is-loading': isLoadingThrottled }"
            type="button"
            @click.prevent="refreshFromSource({ metadata: false })"
          >
            Refresh content from Arc
          </button>
          <button
            v-if="page.arcID"
            class="block button is-warning is-small has-text-weight-semibold"
            :class="{ 'is-loading': isLoadingThrottled }"
            type="button"
            @click.prevent="refreshFromSource({ metadata: true })"
          >
            Refresh content and metadata from Arc
          </button>
          <button
            v-if="page.isGDoc"
            class="block button is-warning is-small has-text-weight-semibold"
            :class="{ 'is-loading': isLoadingThrottled }"
            @click.prevent="refreshFromSource({ metadata: false })"
          >
            Refresh content from Google Docs
          </button>
        </div>
      </BulmaField>

      <details class="field">
        <summary class="has-text-weight-semibold">Advanced options</summary>

        <BulmaField v-slot="{ idForLabel }" label="Language">
          <div class="select is-fullwidth">
            <select :id="idForLabel" v-model="page.languageCode" class="select">
              <option value="">English</option>
              <option value="es">Spanish</option>
            </select>
          </div>
        </BulmaField>

        <BulmaField label="Evergreen content">
          <div>
            <label class="checkbox">
              <input v-model="page.suppressDate" type="checkbox" />
              Don't show date on page
            </label>
          </div>
        </BulmaField>

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

        <BulmaField v-slot="{ idForLabel }" label="Layout override">
          <input v-model="page.layout" class="input" :list="idForLabel" />
          <datalist :id="idForLabel">
            <option value="blank" />
            <option value="featured" />
          </datalist>
        </BulmaField>
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
      />

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

    <SpinnerProgress :is-loading="isLoadingThrottled" />
    <div class="my-5">
      <ErrorReloader :error="error" @reload="fetch(id)" />
    </div>
  </div>
</template>

<style scoped>
.border-thick {
  border: 2px solid #ccc;
}
</style>
