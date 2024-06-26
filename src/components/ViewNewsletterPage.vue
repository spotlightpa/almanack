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
      ...pageData,
      formatDateTime,
      title: computed(() => {
        if (!pageData.page.value) {
          return `Newsletter Page ${id.value}`;
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
      <title>{{ title }} • Spotlight PA</title>
    </MetaHead>
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: 'Newsletter Pages', to: { name: 'newsletters' } },
        { name: title, to: { name: 'newsletter-page', params: { id } } },
      ]"
    ></BulmaBreadcrumbs>

    <h1 class="mb-2 is-spaced title">
      {{ title }}
    </h1>
    <h2 class="subtitle">
      <span class="tags">
        <TagStatus v-if="page" :status="page.status"></TagStatus>
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
      </div>

      <BulmaField
        label="Content"
        help="If content did not load correctly, try refreshing then save"
      >
        <iframe
          ref="iframe"
          :src="`/ssr/page/${page.id}`"
          class="textarea preview-frame"
        ></iframe>
        <button
          v-if="page.sourceType === 'mailchimp' && page.sourceID"
          class="mt-2 block button is-warning has-text-weight-semibold"
          :class="{ 'is-loading': isLoadingThrottled }"
          type="button"
          title="Remember to update published pages after refreshing"
          @click.prevent="
            refreshFromSource().then(() =>
              $refs.iframe.contentDocument.location.reload(true)
            )
          "
        >
          Refresh content from Mailchimp
        </button>
      </BulmaField>

      <details class="field">
        <summary class="has-text-weight-semibold">Advanced options</summary>

        <BulmaAutocompleteArray
          id="topics"
          v-model="page.topics"
          label="Topics"
          :options="topics"
          help="Topics are open-ended collections, e.g. “Events”, “Coronavirus”"
        ></BulmaAutocompleteArray>

        <BulmaAutocompleteArray
          v-model="page.series"
          label="Series"
          :options="series"
          help="Series are limited-time collections, e.g. “Legislative privilege 2020”"
        ></BulmaAutocompleteArray>

        <BulmaFieldCheckbox v-model="page.isPinned" label="Pin article">
          Pin article to the top of topic and series landing pages
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

        <BulmaFieldCheckbox
          v-model="page.modalExclude"
          label="Hide newsletters pop-up"
        >
          Don't show newsletters modal screen on this page
        </BulmaFieldCheckbox>

        <BulmaFieldCheckbox v-model="page.noIndex" label="No index">
          Hide page from Google search results
        </BulmaFieldCheckbox>

        <BulmaFieldInput
          v-model="page.overrideURL"
          label="Override URL"
        ></BulmaFieldInput>

        <BulmaAutocompleteArray
          v-model="page.aliases"
          label="URL Aliases"
          help="Redirect these URLs to the story"
          :options="[]"
        ></BulmaAutocompleteArray>

        <BulmaFieldInput
          v-model="page.layout"
          label="Layout override"
        ></BulmaFieldInput>
      </details>

      <BulmaWarnings
        :values="[
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
        <details>
          <summary class="has-text-weight-semibold">
            {{
              page.status === "sked"
                ? `Scheduled for ${formatDateTime(page.scheduleFor)}`
                : "Schedule for"
            }}
          </summary>
          <BulmaDateTime v-model="page.scheduleFor" icon="user-clock">
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
        </details>
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
.preview-frame {
  height: 300px;
  overflow-y: auto;
}
</style>
