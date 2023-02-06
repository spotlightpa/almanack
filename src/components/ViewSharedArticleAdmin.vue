<script setup>
import { computed, ref, nextTick } from "vue";
import { useRouter } from "vue-router";

import { intcomma } from "journalize";

import {
  get,
  post,
  postPageCreate,
  getSharedArticle,
  postSharedArticle,
  postSharedArticleFromArc,
} from "@/api/client-v2.js";
import { watchAPI, makeState } from "@/api/service-util.js";
import SharedArticle from "@/api/shared-article.js";
import { formatDate, formatDateTime, tomorrow } from "@/utils/time-format.js";

const props = defineProps({
  id: String,
});

const { apiState, fetch, computedObj } = watchAPI(
  () => props.id,
  (id) => get(getSharedArticle, { id })
);

const showComposer = ref(false);
const composer = ref(null);
const isDirty = ref(false);
const status = ref(null);
const note = ref("");
const embargo = ref(null);

const article = computedObj((rawData) => {
  let a = new SharedArticle(rawData);
  isDirty.value = false;
  status.value = a._status;
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

  const { resolve } = useRouter();
  let { href } = resolve(a.detailsRoute);
  let noteText = !note.value ? "" : `\n\nPublication Notes:\n\n${note.value}`;
  let embargoText = !embargo.value
    ? ""
    : `\n\nEmbargoed until ${formatDateTime(embargo.value)}`;

  let segments = [
    `New ${a.slug}`,
    `https://almanack.data.spotlightpa.org${href}`,
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

const { exec: pageExec, apiStateRefs: pageState } = makeState();
const { isLoadingThrottled: pageLoading, error: pageError } = pageState;
async function createPage(kind) {
  await pageExec(() =>
    post(postPageCreate, { shared_article_id: props.id, page_kind: kind })
  );
  await fetch();
  await window.fetch("/api-background/images");
}

const { exec: arcExec, apiStateRefs: arcState } = makeState();
const { isLoadingThrottled: arcLoading, error: arcError } = arcState;
function refreshArc() {
  arcExec(() =>
    post(postSharedArticleFromArc, { arc_id: article.value.sourceID })
  );
}

const { exec: saveExec, apiStateRefs: saveState } = makeState();
const { isLoadingThrottled: saveLoading, error: saveError } = saveState;
async function save() {
  const obj = {
    id: +props.id,
    note: note.value,
    embargo_until: embargo.value,
    status: status.value,
  };
  await saveExec(() => post(postSharedArticle, obj));
  await fetch();
  isDirty.value = false;
}

const savingEnabled = computed(() => {
  if (!isDirty.value) return false;
  return true;
});

async function toggleComposer() {
  showComposer.value = !showComposer.value;
  if (!showComposer.value) {
    return;
  }
  await nextTick();
  composer.value.$el.scrollIntoView({ behavior: "smooth", block: "center" });
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
        <p class="label">Budget</p>
        <p class="mb-5 content">{{ article.arc.budgetLine }}</p>
        <div v-if="!article.pageID" class="mb-5">
          <div class="label">Import to Spotlight PA</div>
          <div class="buttons">
            <button
              class="button is-primary has-text-weight-semibold"
              :class="pageLoading ? 'is-loading' : ''"
              @click="createPage('news')"
            >
              As News article
            </button>
            <button
              class="button is-primary has-text-weight-semibold"
              :class="pageLoading ? 'is-loading' : ''"
              @click="createPage('statecollege')"
            >
              As State College article
            </button>
          </div>
        </div>

        <ErrorSimple :error="pageError" />

        <h3 class="label">Status</h3>
        <div class="buttons">
          <button
            class="button is-small has-text-weight-semibold"
            :class="statusClass('U')"
            type="button"
            @click="
              status = 'U';
              isDirty = true;
            "
          >
            Not Shared
          </button>
          <button
            class="button is-small has-text-weight-semibold"
            :class="statusClass('S')"
            type="button"
            @click="
              status = 'S';
              isDirty = true;
            "
          >
            Shared with partners
          </button>
        </div>

        <div class="mb-5">
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
        <ErrorSimple :error="saveError || arcError" class="mt-1" />

        <div class="mt-5 buttons">
          <button
            class="button is-success has-text-weight-semibold"
            :class="saveLoading ? 'is-loading' : ''"
            :disabled="!savingEnabled || null"
            @click="save()"
          >
            Save changes
          </button>
          <button
            class="button is-danger has-text-weight-semibold"
            :class="saveLoading ? 'is-loading' : ''"
            :disabled="!savingEnabled || null"
            @click="fetch"
          >
            Discard changes
          </button>
        </div>
        <button
          type="button"
          class="button is-small has-text-weight-semibold"
          :class="showComposer ? 'is-danger' : 'is-primary'"
          @click="toggleComposer()"
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
          ref="composer"
          class="mt-5"
          :initial-subject="`New Spotlight PA story ${article.slug}`"
          :initial-body="emailBody"
          @hide="showComposer = false"
        />

        <div
          v-if="showComposer && article._status !== 'S'"
          class="mt-5 message is-danger"
        >
          <p class="message-body">
            <strong>Warning:</strong> Article has not been shared with partners
            yet.
          </p>
        </div>
      </div>
    </article>
  </div>
</template>
