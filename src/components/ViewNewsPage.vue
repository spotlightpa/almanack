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

    <h1 class="mb-2 is-spaced title is-flex-desktop">
      {{ title }}
    </h1>
    <h2 class="subtitle">
      <span class="tags">
        <TagStatus v-if="page" :status="page.status" />
        <router-link
          v-if="page && page.sharedViewRoute"
          class="tag is-light has-text-weight-semibold"
          :to="page.sharedViewRoute"
        >
          <span class="icon is-size-6">
            <font-awesome-icon :icon="['fas', 'link']" />
          </span>
          <span>External Editor view</span>
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
        v-model="page.publishedAt"
        label="Publication Date"
        :icon="['fas', 'user-clock']"
        help="Page will be listed on the site under this date"
      >
        <p class="content is-small">
          <a
            href="#"
            class="has-text-info"
            @click.prevent="page.publishedAt = new Date()"
          >
            Set to now
          </a>
        </p>
      </BulmaDateTime>

      <p v-if="page.isFutureDated" class="content has-text-warning is-small">
        Article publication date is in the future.
      </p>

      <BulmaAutocompleteArray
        v-model="page.topics"
        label="Topics"
        :options="topics"
        help="Topics are open-ended collections, e.g. “Events”, “Coronavirus”"
      />

      <BulmaDateTime
        v-if="page.topics.includes('Events')"
        v-model="page.eventDate"
        label="Event Date"
        help="If present, the events landing page will show this date for the event"
        icon="user-clock"
      />

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
        v-model="page.title"
        label="Hed"
        help="Default value for title tag, link title, and share title"
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

      <BulmaFieldInput
        v-model="page.imageDescription"
        label="Image description"
      />
      <BulmaFieldInput v-model="page.imageCredit" label="Image credit" />

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

      <BulmaField v-slot="{ idForLabel }" label="Page content">
        <textarea
          :id="idForLabel"
          v-model="page.body"
          class="textarea"
          rows="8"
        ></textarea>
      </BulmaField>

      <button
        v-if="page.arcID"
        class="block button is-warning has-text-weight-semibold"
        :class="{ 'is-loading': isLoadingThrottled }"
        type="button"
        title="Remember to update published pages after refreshing"
        @click.prevent="arcRefresh"
      >
        Refresh content from Arc
      </button>

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
              @click.prevent="page.scheduleFor = page.publishedAt"
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
