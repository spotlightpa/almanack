<script>
import { computed, toRefs } from "@vue/composition-api";

import { usePage } from "@/api/spotlightpa-page.js";

import { formatDateTime } from "@/utils/time-format.js";

export default {
  name: "ViewNewsPage",
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
    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        { name: 'Spotlight PA Pages', to: { name: 'news-pages' } },
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
          :datepicker="{ required: false }"
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
      <p v-if="page.isFutureDated" class="content has-text-warning is-small">
        Article publication date is in the future.
      </p>

      <BulmaAutocompleteArray
        v-model="page.topics"
        label="Topics"
        :options="topics"
        help="Topics are open-ended collections, e.g. “Events”, “Coronavirus”"
      />

      <BulmaField
        v-if="page.topics.includes('Events')"
        v-slot="{ idForLabel }"
        label="Event Date"
        help="If present, the events landing page will show this date for the event"
      >
        <b-datetimepicker
          :id="idForLabel"
          v-model="page.eventDate"
          icon="user-clock"
          :datetime-formatter="formatDateTime"
          locale="en-US"
        />
      </BulmaField>

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

      <BulmaFieldInput
        v-model="page.title"
        label="Title"
        help="Default value for title tag, link title, and share title"
        :required="true"
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

      <BulmaFieldInput
        v-model="page.image"
        label="Photo ID"
        help="Image is shown in article rivers and on social media"
      />
      <picture v-if="page.imagePreviewURL" class="has-ratio">
        <img :src="page.imagePreviewURL" class="is-3x4" width="200" />
      </picture>

      <PickerImages :images="images" @select-image="setImageProps($event)" />

      <BulmaFieldInput
        v-model="page.imageDescription"
        label="Image description"
      />
      <BulmaFieldInput v-model="page.imageCredit" label="Image credit" />

      <b-field label="Image size">
        <b-select v-model="page.imageSize" expanded>
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
        </b-select>
      </b-field>

      <BulmaFieldInput
        v-model="page.slug"
        label="URL keywords slug"
        :disabled="page.isPublished"
        :readonly="page.isPublished"
        :required="true"
      />
      <button
        class="block button is-small is-light has-text-weight-semibold"
        type="button"
        :disabled="page.isPublished"
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
        <BulmaField
          v-slot="{ idForLabel }"
          :label="
            page.status === 'sked'
              ? `Scheduled for ${formatDateTime(page.scheduleFor)}`
              : `Schedule for`
          "
        >
          <b-datetimepicker
            :id="idForLabel"
            v-model="page.scheduleFor"
            icon="user-clock"
            :datetime-formatter="formatDateTime"
            :inline="true"
            locale="en-US"
          />
          <p v-if="page.isFutureDated" class="mt-2 content is-small">
            <a
              href="#"
              class="has-text-info"
              @click.prevent="page.scheduleFor = page.publishedAt"
            >
              Schedule for publication date
            </a>
          </p>
        </BulmaField>
      </div>
      <div class="field">
        <div class="buttons">
          <button
            class="button is-success has-text-weight-semibold"
            :disabled="isLoading"
            type="button"
            @click="publishNow($refs.form)"
          >
            {{ page.status === "pub" ? "Update page" : "Publish now" }}
          </button>
          <button
            v-if="page.status !== 'pub'"
            class="button is-warning has-text-weight-semibold"
            :disabled="
              isLoading || !page.scheduleFor || page.scheduleFor < new Date()
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
    </form>

    <SpinnerProgress :is-loading="isLoadingThrottled" />
    <div class="my-5">
      <ErrorReloader :error="error" @reload="fetch(id)" />
    </div>
  </div>
</template>
