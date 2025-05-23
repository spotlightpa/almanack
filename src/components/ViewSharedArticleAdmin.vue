<script setup>
import { computed, ref, nextTick } from "vue";
import { useRouter } from "vue-router";

import { intcomma } from "journalize";

import {
  get,
  post,
  postPageCreate,
  listImages,
  getSharedArticle,
  postSharedArticle,
  postSharedArticleFromGDocs,
} from "@/api/client-v2.js";
import { processGDocsDoc } from "@/api/gdocs.js";
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
const publicationDate = ref(null);
const internalID = ref("");
const byline = ref("");
const budget = ref("");
const hed = ref("");
const description = ref("");
const blurb = ref("");
const ledeImage = ref("");
const ledeImageCredit = ref("");
const ledeImageDescription = ref("");
const ledeImageCaption = ref("");

const article = computedObj((rawData) => {
  let a = new SharedArticle(rawData);
  isDirty.value = false;
  status.value = a._status;
  note.value = a.note;
  embargo.value = a.embargoUntil;
  publicationDate.value = a.publicationDate;
  internalID.value = a.internalID;
  byline.value = a.byline;
  budget.value = a.budget;
  hed.value = a.hed;
  description.value = a.description;
  blurb.value = a.blurb;
  ledeImage.value = a.ledeImage;
  ledeImageCredit.value = a.ledeImageCredit;
  ledeImageDescription.value = a.ledeImageDescription;
  ledeImageCaption.value = a.ledeImageCaption;
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
    `New ${a.internalID}`,
    `https://almanack.data.spotlightpa.org${href}`,
    `Planned for ${formatDate(a.publicationDate)}`,
    embargoText,
    noteText,
    `Budget:`,
    budget.value,
    a.arc
      ? `Word count: ${intcomma(a.arc.actualWordCount)}
Lines: ${intcomma(a.arc.actualLineCount)}
Column inches: ${intcomma(a.arc.actualInchCount)}`
      : ``,
  ];
  return segments
    .map((text) => text.trim())
    .filter((text) => !!text)
    .join("\n\n");
});

const { exec: pageExec, apiStateRefs: pageState } = makeState();
const { isLoadingThrottled: pageLoading, error: pageError } = pageState;
async function createPage(kind) {
  let prompt = "Create Spotlight PA page?";
  if (kind === "news") {
    prompt = "Create Spotlight PA /news/ page?";
  }
  if (kind === "statecollege") {
    prompt = "Create Spotlight PA State College page?";
  }
  if (kind === "berks") {
    prompt = "Create Berks County page?";
  }
  if (!window.confirm(prompt)) return;
  await pageExec(() =>
    post(postPageCreate, { shared_article_id: props.id, page_kind: kind })
  );
  await fetch();
  await window.fetch("/api-background/images");
}
const pageCreateDisabled = computed(() => {
  if (isDirty.value || publicationDate.value === null || !internalID.value) {
    return true;
  }
  return null;
});

const { apiStateRefs: gdocsState, exec: gdocsExec } = makeState();
const { isLoadingThrottled: gdocsLoading, error: gdocsError } = gdocsState;
async function refreshGDocs({ refreshMetadata = false } = {}) {
  await gdocsExec(async () => {
    if (isDirty.value) {
      return;
    }
    let [, err] = await processGDocsDoc(article.value.sourceID);
    if (err) {
      return [null, err];
    }
    return await post(postSharedArticleFromGDocs, {
      external_gdocs_id: article.value.sourceID,
      force_update: true,
      refresh_metadata: refreshMetadata,
    });
  });
  if (gdocsState.error.value) {
    return;
  }
  await fetch();
}

const { exec: saveExec, apiStateRefs: saveState } = makeState();
const { isLoadingThrottled: saveLoading, error: saveError } = saveState;
async function save() {
  const obj = {
    id: +props.id,
    note: note.value,
    embargo_until: embargo.value,
    status: status.value,
    publication_date: publicationDate.value,
    internal_id: internalID.value,
    byline: byline.value,
    budget: budget.value,
    hed: hed.value,
    description: description.value,
    blurb: blurb.value,
    lede_image: ledeImage.value,
    lede_image_credit: ledeImageCredit.value,
    lede_image_description: ledeImageDescription.value,
    lede_image_caption: ledeImageCaption.value,
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

const { computedList: imageList } = watchAPI(
  () => 0,
  (page) => get(listImages, { page })
);
const images = imageList("images", (obj) => obj);
function setImageProps(image) {
  ledeImage.value = image.path;
  ledeImageDescription.value = image.description;
  ledeImageCredit.value = image.credit;
  ledeImageCaption.value = "";
  isDirty.value = true;
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
    ></BulmaBreadcrumbs>

    <SpinnerProgress :is-loading="apiState.isLoading.value"></SpinnerProgress>
    <ErrorReloader
      :error="apiState.error.value"
      @reload="fetch"
    ></ErrorReloader>

    <article v-if="article">
      <MetaHead>
        <title>{{ article.internalID }} Admin • Spotlight PA</title>
      </MetaHead>

      <GDocsDocWarnings :article="article"></GDocsDocWarnings>

      <div class="message is-primary">
        <div class="message-header">
          <p>
            <font-awesome-icon :icon="['far', 'newspaper']"></font-awesome-icon>
            {{ article.internalID }}
          </p>
          <span class="tags">
            <TagLink :to="article.detailsRoute" :icon="['fas', 'file-invoice']">
              Partner view
            </TagLink>
            <TagLink
              v-if="article.isArc"
              :href="article.arc.arcURL"
              :icon="['fas', 'link']"
            >
              Arc
            </TagLink>
            <TagLink
              v-if="article.isGDoc"
              :href="article.gdocsURL"
              :icon="['fas', 'link']"
            >
              Google Docs
            </TagLink>
            <TagLink
              v-if="article.pageRoute"
              :to="article.pageRoute"
              :icon="['fas', 'user-clock']"
            >
              Spotlight admin
            </TagLink>
          </span>
        </div>
        <div class="message-body">
          <template v-if="article.isGDoc">
            <BulmaFieldInput
              label="Slug"
              :model-value="internalID"
              help="Short internal ID for stories, such as SPLSTORY12."
              @update:modelValue="
                isDirty = true;
                internalID = $event;
              "
            ></BulmaFieldInput>

            <BulmaDateTime
              :model-value="publicationDate"
              label="Planned publication date"
              @update:modelValue="
                publicationDate = $event;
                isDirty = true;
              "
            >
              <a
                @click="
                  publicationDate = new Date();
                  isDirty = true;
                "
                >Set to now</a
              >.
              <a
                @click="
                  publicationDate = tomorrow();
                  isDirty = true;
                "
                >Set to tomorrow</a
              >.
            </BulmaDateTime>

            <BulmaTextarea
              :model-value="budget"
              label="Budget"
              help="Description of the story for partners and editors."
              @update:modelValue="
                isDirty = true;
                budget = $event;
              "
            ></BulmaTextarea>

            <BulmaFieldInput
              label="Suggested hed"
              :model-value="hed"
              help=""
              @update:modelValue="
                isDirty = true;
                hed = $event;
              "
            ></BulmaFieldInput>
            <template v-if="article.isGDoc">
              <BulmaTextarea
                label="SEO description"
                :model-value="description"
                help="Only used for Spotlight PA Pages. Not shown to partners."
                @update:modelValue="
                  isDirty = true;
                  description = $event;
                "
              ></BulmaTextarea>
              <BulmaTextarea
                label="Suggested blurb"
                :model-value="blurb"
                help="Short description shown to partners."
                @update:modelValue="
                  isDirty = true;
                  blurb = $event;
                "
              ></BulmaTextarea>
            </template>
            <BulmaFieldInput
              v-else
              label="Suggested description"
              :model-value="description"
              help=""
              @update:modelValue="
                isDirty = true;
                description = $event;
              "
            ></BulmaFieldInput>
            <BulmaFieldInput
              label="Byline"
              :model-value="byline"
              help="Omit “by”; include “of Spotlight PA”"
              @update:modelValue="
                isDirty = true;
                byline = $event;
              "
            ></BulmaFieldInput>

            <BulmaField label="Lede Image" v-slot="{ idForLabel }">
              <div class="is-flex">
                <input
                  :id="idForLabel"
                  :value="ledeImage"
                  class="input"
                  @change="
                    isDirty = true;
                    ledeImage = $event.target.value;
                  "
                />
                <BulmaPaste
                  @paste="
                    isDirty = true;
                    ledeImage = $event;
                  "
                ></BulmaPaste>
              </div>
            </BulmaField>

            <PickerImages
              :images="images"
              @select-image="setImageProps($event)"
            ></PickerImages>

            <BulmaTextarea
              :model-value="ledeImageDescription"
              label="Lede image description"
              help="A description of the image for visually impaired readers (“alt” text)"
              @update:modelValue="
                isDirty = true;
                ledeImageDescription = $event;
              "
            ></BulmaTextarea>

            <BulmaFieldInput
              :model-value="ledeImageCredit"
              label="Lede image credit"
              @update:modelValue="
                isDirty = true;
                ledeImageCredit = $event;
              "
            ></BulmaFieldInput>

            <BulmaTextarea
              :model-value="ledeImageCaption"
              label="Lede image caption"
              help="An optional caption underneath the image"
              @update:modelValue="
                isDirty = true;
                ledeImageCaption = $event;
              "
            ></BulmaTextarea>
          </template>

          <p v-if="article.isArc" class="label">Budget</p>
          <p v-if="article.isArc" class="mb-5 content">{{ article.budget }}</p>

          <h3 class="label">Sharing status</h3>
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
            ></BulmaDateTime>
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
          ></BulmaTextarea>

          <ErrorSimple :error="saveError" class="mt-1"></ErrorSimple>
          <div v-if="article.isGDoc" class="buttons">
            <button
              class="button is-warning has-text-weight-semibold"
              type="button"
              :class="gdocsLoading && 'is-loading'"
              :disabled="isDirty || null"
              @click="refreshGDocs({ refreshMetadata: false })"
            >
              Refresh content from Google Docs
            </button>
            <button
              class="button is-warning has-text-weight-semibold"
              type="button"
              :class="gdocsLoading && 'is-loading'"
              :disabled="isDirty || null"
              @click="refreshGDocs({ refreshMetadata: true })"
            >
              Refresh content and metadata
            </button>
          </div>
          <ErrorSimple
            :error="saveError || gdocsError"
            class="mt-1"
          ></ErrorSimple>

          <GDocsDocWarnings class="mt-5" :article="article"></GDocsDocWarnings>

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

          <div v-if="!article.pageID" class="mb-5">
            <div class="label">Import to Spotlight PA</div>
            <p v-if="!publicationDate" class="mb-2 help is-danger">
              Must set publication date first.
            </p>
            <p v-if="!internalID" class="mb-2 help is-danger">
              Must set slug first.
            </p>
            <div class="buttons">
              <button
                class="button is-primary has-text-weight-semibold"
                :class="pageLoading ? 'is-loading' : ''"
                :disabled="pageCreateDisabled"
                @click="createPage('news')"
              >
                As News article
              </button>
              <button
                class="button is-primary has-text-weight-semibold"
                :class="pageLoading ? 'is-loading' : ''"
                :disabled="pageCreateDisabled"
                @click="createPage('statecollege')"
              >
                As State College article
              </button>
              <button
                class="button is-primary has-text-weight-semibold"
                :class="pageLoading ? 'is-loading' : ''"
                :disabled="pageCreateDisabled"
                @click="createPage('berks')"
              >
                As Berks County article
              </button>
            </div>
          </div>
          <div v-else class="mb-5">
            <router-link
              :to="article.pageRoute"
              class="button is-primary has-text-weight-semibold"
            >
              <span class="icon">
                <font-awesome-icon
                  :icon="['fas', 'user-clock']"
                ></font-awesome-icon>
              </span>
              <span>Spotlight PA Page Admin</span>
            </router-link>
          </div>
          <ErrorSimple :error="pageError"></ErrorSimple>

          <button
            type="button"
            class="button is-small has-text-weight-semibold"
            :class="showComposer ? 'is-danger' : 'is-primary'"
            @click="toggleComposer()"
          >
            <span class="icon">
              <font-awesome-icon
                :icon="['fas', 'paper-plane']"
              ></font-awesome-icon>
            </span>
            <span
              v-text="!showComposer ? 'Compose Message' : 'Discard Message'"
            ></span>
          </button>

          <EmailComposer
            v-if="showComposer"
            ref="composer"
            class="mt-5"
            :initial-subject="`New Spotlight PA story ${article.internalID}`"
            :initial-body="emailBody"
            @hide="showComposer = false"
          ></EmailComposer>

          <div
            v-if="showComposer && article._status !== 'S'"
            class="mt-5 message is-danger"
          >
            <p class="message-body">
              <strong>Warning:</strong> Article has not been shared with
              partners yet.
            </p>
          </div>
        </div>
      </div>
      <div v-if="article.isGDoc">
        <h2 class="title is-5">Article Preview</h2>
        <div class="textarea" rows="whatever">
          <h1 class="title">{{ hed }}</h1>
          <h2 class="subtitle is-3">{{ blurb }}</h2>
          <h2 v-if="byline" class="subtitle is-5">By {{ byline }}</h2>
          <div class="content" v-html="article.gdocs.rich_text"></div>
        </div>
      </div>
    </article>
  </div>
</template>
