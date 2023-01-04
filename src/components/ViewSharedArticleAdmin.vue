<script setup>
import { computed, ref } from "vue";

import { intcomma } from "journalize";

import {
  get,
  post,
  getSharedArticle,
  postSharedArticleFromArc,
} from "@/api/client-v2.js";
import { watchAPI, makeState } from "@/api/service-util.js";
import SharedArticle from "@/api/shared-article.js";
import { formatDate, formatDateTime, tomorrow } from "@/utils/time-format.js";

const props = defineProps({
  id: String,
});

const { apiState, fetch, computer } = watchAPI(
  () => props.id,
  (id) => get(getSharedArticle, { id })
);

const showComposer = ref(false);
const isDirty = ref(false);
const status = ref(null);
const note = ref("");
const embargo = ref(null);

const article = computer((rawData) => {
  if (!rawData) return null;
  let a = new SharedArticle(rawData);
  isDirty.value = false;
  status.value = a.status;
  note.value = a.note;
  embargo.value = a.embargoUntil;
  return a;
});

function statusClass(val) {
  return {
    "is-white": status.value !== val,
    "is-primary": status.value === val,
  };
}

const emailBody = computed(() => {
  let a = article.value;
  let noteText = !note.value ? "" : `\n\nPublication Notes:\n\n${note.value}`;
  let embargoText =
    status.value !== "embargo" || !embargo.value
      ? ""
      : `\n\nEmbargoed until ${formatDateTime(embargo.value)}`;

  let segments = [
    `New ${a.slug}`,
    `https://almanack.data.spotlightpa.org/shared-articles/${a.id}`,
    `Planned for ${formatDate(a.arc.plannedDate)}`,
    embargoText,
    noteText,
    `Budget:`,
    a.arc.budgetLine,
    `
Word count planned: ${intcomma(a.arc.plannedWordCount)}
Word count actual: ${intcomma(a.arc.actualWordCount)}
Lines: ${a.arc.actualLineCount}
Column inches: ${a.arc.actualInchCount}`,
  ];
  return segments
    .map((text) => text.trim())
    .filter((text) => !!text)
    .join("\n\n");
});

const { exec: arcExec, apiStateRefs: arcState } = makeState();
const { isLoadingThrottled: arcLoading, error: arcError } = arcState;
function refreshArc() {
  arcExec(() =>
    post(postSharedArticleFromArc, { arc_id: article.value.sourceID })
  );
}
</script>

<template>
  <div>
    <MetaHead>
      <title>Shared Article Admin • Spotlight PA</title>
    </MetaHead>

    <BulmaBreadcrumbs
      :links="[
        { name: 'Admin', to: { name: 'admin' } },
        {
          name: 'Shared Article',
          to: { name: 'shared-article-admin', params: { id } },
        },
      ]"
    />

    <SpinnerProgress :is-loading="apiState.isLoading.value" />
    <ErrorReloader :error="apiState.error.value" @reload="fetch" />

    <article v-if="article" class="message is-primary">
      <MetaHead>
        <title>{{ article.slug }} Admin • Spotlight PA</title>
      </MetaHead>

      <div class="message-header">
        <p>
          <font-awesome-icon :icon="['far', 'newspaper']" /> {{ article.slug }}
        </p>
        <span class="tags">
          <router-link v-if="article.id" class="tag" :to="article.detailsRoute">
            <span class="icon">
              <font-awesome-icon :icon="['fas', 'file-invoice']" />
            </span>
            <span>Partner view</span>
          </router-link>
          <a
            v-if="article.isArc"
            class="tag"
            :href="article.arc.arcURL"
            target="_blank"
          >
            <span class="icon">
              <font-awesome-icon :icon="['fas', 'link']" />
            </span>
            <span>Arc</span>
          </a>
          <router-link
            v-if="article.pageRoute"
            class="tag is-light"
            :to="article.pageRoute"
          >
            <span class="icon">
              <font-awesome-icon :icon="['fas', 'user-clock']" />
            </span>
            <span>Spotlight admin</span>
          </router-link>
        </span>
      </div>
      <div class="message-body">
        <div v-if="!article.pageID" class="mb-5">
          <div class="label">Import to Spotlight PA</div>
          <div class="buttons">
            <button class="button is-primary has-text-weight-semibold">
              As News article
            </button>
            <button
              class="button is-primary has-text-weight-semibold"
              @click="close"
            >
              As State College article
            </button>
          </div>
        </div>

        <h3 class="label">Status</h3>
        <div class="buttons">
          <button
            class="button is-small has-text-weight-semibold"
            :class="statusClass('imported')"
            type="button"
            @click="
              status = 'imported';
              isDirty = true;
            "
          >
            Imported
          </button>
          <button
            class="button is-small has-text-weight-semibold"
            :class="statusClass('preview')"
            type="button"
            @click="
              status = 'preview';
              isDirty = true;
            "
          >
            Preview
          </button>
          <button
            class="button is-small has-text-weight-semibold"
            :class="statusClass('embargo')"
            type="button"
            @click="
              status = 'embargo';
              isDirty = true;
            "
          >
            Embargo
          </button>
          <button
            class="button is-small has-text-weight-semibold"
            :class="statusClass('released')"
            type="button"
            @click="
              status = 'released';
              isDirty = true;
            "
          >
            Release
          </button>
        </div>

        <div v-if="status === 'embargo'" class="mb-5">
          <BulmaDateTime
            :model-value="embargo"
            label="Embargo time"
            help="List the latest time that an article will be under embargo for partners."
            @update:modelValue="
              embargo = $event;
              isDirty = true;
            "
          />
          <a
            @click="
              embargo = tomorrow();
              isDirty = true;
            "
          >
            Set for tomorrow
          </a>
        </div>

        <BulmaTextarea
          :model-value="note"
          label="Note"
          help="Additional clarifications and instructions for partners"
          @update:modelValue="
            isDirty = true;
            note = $event;
          "
        />

        <button
          class="button is-warning has-text-weight-semibold"
          type="button"
          :class="arcLoading && 'is-loading'"
          @click="refreshArc"
        >
          Refresh from Arc
        </button>
        <ErrorSimple :error="arcError" class="mt-1" />

        <div class="mt-5 buttons">
          <button
            class="button is-success has-text-weight-semibold"
            :disabled="!isDirty || null"
          >
            Save changes
          </button>
          <button
            class="button is-danger has-text-weight-semibold"
            :disabled="!isDirty || null"
            @click="close"
          >
            Discard changes
          </button>
        </div>
        <button
          type="button"
          class="button is-small has-text-weight-semibold"
          :class="showComposer ? 'is-danger' : 'is-primary'"
          @click="showComposer = !showComposer"
        >
          <span class="icon">
            <font-awesome-icon :icon="['fas', 'paper-plane']" />
          </span>
          <span
            v-text="!showComposer ? 'Compose Message' : 'Discard Message'"
          />
        </button>

        <EmailComposer
          v-if="showComposer"
          class="mt-5"
          :initial-subject="`New Spotlight PA story ${article.slug}`"
          :initial-body="emailBody"
          @hide="showComposer = false"
        />
      </div>
    </article>
  </div>
</template>
