<script setup>
import { get, post, listImages, postImageUpdate } from "@/api/client-v2.js";
import { makeState, watchAPI } from "@/api/service-util.js";
import imgproxyURL from "@/api/imgproxy-url.js";

import { formatDate } from "@/utils/time-format.js";

const props = defineProps({
  page: { type: String, default: "0" },
});

const {
  apiState: listState,
  fetch,
  computedList,
  computedProp,
} = watchAPI(
  () => props.page || 0,
  (page) => get(listImages, { page })
);

const { apiStateRefs: saveState, exec } = makeState();

const toImageObj = (rawImage) => ({
  id: rawImage.id,
  path: rawImage.path,
  url: imgproxyURL(rawImage.path, {
    width: 256,
    height: (256 * 9) / 16,
    extension: "webp",
  }),
  description: rawImage.description,
  credit: rawImage.credit,
  keywords: rawImage.keywords,
  srcURL: rawImage.src_url,
  date: new Date(rawImage.created_at),
  downloadURL: "/ssr/download-image?src=" + encodeURIComponent(rawImage.path),
});

const images = computedList("images", (obj) => toImageObj(obj));
const nextPage = computedProp("next_page", (page) => ({
  name: "image-uploader",
  query: {
    page,
  },
}));

const showReload = computedProp("waiting_for_upload", () => {
  // Trigger image upload queue if not triggered already
  window.fetch("/api-background/images").catch(() => {});
  return true;
});

function updateObj(
  path,
  { credit = "", description = "", keywords = "" } = {}
) {
  return {
    path,
    credit,
    set_credit: !!credit,
    description,
    set_description: !!description,
    keywords,
    set_keywords: !!keywords,
  };
}

async function doUpdate(image, opt) {
  return exec(async () => {
    let [, err] = await post(postImageUpdate, updateObj(image.path, opt));
    if (err) return [null, err];
    await fetch();
    return [null, null];
  });
}

function updateDescription(image) {
  let description = window.prompt("Update description", image.description);
  if (description !== null && description !== image.description) {
    doUpdate(image, { description });
  }
}

function updateCredit(image) {
  let credit = window.prompt("Update credit", image.credit);
  if (credit !== null && credit !== image.credit) {
    doUpdate(image, { credit });
  }
}

function updateKeywords(image) {
  let keywords = window.prompt("Update keywords", image.keywords);
  if (keywords !== null && keywords !== image.keywords) {
    doUpdate(image, { keywords });
  }
}
</script>

<template>
  <MetaHead>
    <title>Photo Uploads • Spotlight PA</title>
  </MetaHead>
  <div>
    <h1 class="title">
      Upload an image
      <template v-if="page !== '0'">(overflow page {{ page }})</template>
    </h1>

    <ImageUploader @update-image-list="fetch" />

    <div v-if="showReload" class="mt-5 message is-warning">
      <p class="message-header">Waiting for background uploads</p>
      <p class="message-body">
        <button
          class="button is-warning has-text-weight-semibold"
          @click="fetch"
        >
          Reload now
        </button>
      </p>
    </div>

    <SpinnerProgress :is-loading="listState.isLoadingThrottled.value" />
    <ErrorSimple :error="saveState.error.value" />
    <ErrorReloader :error="listState.error.value" @reload="fetch" />
    <h2 class="title has-margin-top">Existing Images</h2>

    <table class="table is-striped is-fullwidth">
      <tbody>
        <tr v-for="image of images" :key="image.id">
          <td>
            <div class="max-128">
              <a :href="image.downloadURL" class="is-flex">
                <picture
                  class="image is-3x4 has-background-grey-lighter has-margin-bottom"
                >
                  <img
                    :src="image.url"
                    class="border-thick"
                    width="256"
                    height="192"
                    loading="lazy"
                  />
                </picture>
              </a>
            </div>
            <div class="mt-4">
              <LinkHref
                label="Original"
                color="is-success"
                :icon="['fas', 'file-download']"
                :href="image.downloadURL"
              />
              <ImageSize class="mt-1" :path="image.path" />
            </div>
          </td>
          <td>
            <p>
              <a
                class="has-text-weight-semibold"
                @click="updateDescription(image)"
              >
                Description:
              </a>
              {{ image.description }}
            </p>

            <p>
              <a class="has-text-weight-semibold" @click="updateCredit(image)">
                Credit:
              </a>
              {{ image.credit }}
            </p>

            <p>
              <a
                class="has-text-weight-semibold"
                @click="updateKeywords(image)"
              >
                Keywords:
              </a>
              {{ image.keywords }}
            </p>
            <p>
              <strong>Date:</strong>
              {{ formatDate(image.date) }}
            </p>
            <p class="has-margin-top-thin">
              <CopyWithButton
                :value="image.path"
                label="path"
                size="is-small"
              />
            </p>
          </td>
          <td></td>
        </tr>
      </tbody>
    </table>
    <div class="buttons mt-5">
      <router-link
        v-if="nextPage"
        :to="nextPage"
        class="button is-primary has-text-weight-semibold"
      >
        Show Older Images…
      </router-link>
    </div>
  </div>
</template>

<style scoped>
.max-128 {
  height: 72px;
  width: 128px;
  min-height: 1rem;
  min-width: 1rem;
}
.border-thick {
  border: 2px solid #ccc;
}
</style>
