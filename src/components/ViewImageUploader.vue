<script setup>
import { ref, watch } from "vue";

import { get, post, listImages, postImageUpdate } from "@/api/client-v2.js";
import { makeState, watchAPI } from "@/api/service-util.js";
import imgproxyURL from "@/api/imgproxy-url.js";

import { formatDate } from "@/utils/time-format.js";
import humanSize from "@/utils/human-size.js";

const props = defineProps({
  page: { type: String, default: "0" },
});

const query = ref("");

const {
  apiState: listState,
  fetch,
  computedList,
  computedProp,
} = watchAPI(
  () => [props.page || 0, query.value],
  ([page, query]) => get(listImages, { page, query })
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
  size: rawImage.bytes ? humanSize(rawImage.bytes) : "",
  srcURL: rawImage.src_url,
  isLicensed: rawImage.is_licensed,
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
  {
    credit = "",
    description = "",
    keywords = "",
    set_is_licensed = false,
    is_licensed = false,
  } = {}
) {
  return {
    path,
    credit,
    set_credit: !!credit,
    description,
    set_description: !!description,
    keywords,
    set_keywords: !!keywords,
    set_is_licensed,
    is_licensed,
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

function updateIsLicensed(image) {
  if (
    !window.confirm(`Set to ${image.isLicensed ? "unlicensed" : "licensed"}?`)
  ) {
    return;
  }
  doUpdate(image, { set_is_licensed: true, is_licensed: !image.isLicensed });
}

const rawQuery = ref("");
let timeout = null;
watch(rawQuery, (val) => {
  window.clearTimeout(timeout);
  timeout = window.setTimeout(() => {
    query.value = val;
  }, 1000);
});
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

    <ImageUploader @update-image-list="fetch"></ImageUploader>

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

    <h2 class="mt-5 title">Resize an image</h2>
    <details>
      <summary>Almanack image resizer</summary>
      <ImageResize />
    </details>

    <h2 class="mt-5 title">Existing Images</h2>

    <BulmaFieldInput
      v-model="rawQuery"
      class="mb-5"
      type="search"
      label="Filter by search terms"
      placeholder="Pennsylvania state capitol in Harrisburg"
    ></BulmaFieldInput>

    <SpinnerProgress
      :is-loading="listState.isLoadingThrottled.value"
    ></SpinnerProgress>
    <ErrorSimple :error="saveState.error.value"></ErrorSimple>
    <ErrorReloader
      :error="listState.error.value"
      @reload="fetch"
    ></ErrorReloader>

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
                :label="image.size ? `Original (${image.size})` : 'Original'"
                color="is-success"
                :icon="['fas', 'file-download']"
                :href="image.downloadURL"
              ></LinkHref>
              <ImageSize class="mt-1" :path="image.path"></ImageSize>
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
              <a @click="updateIsLicensed(image)">
                <span class="has-text-weight-semibold">License: </span>
                <span v-if="image.isLicensed">
                  <span class="has-text-black">Reuse permitted</span>
                  <span class="icon is-size-6 has-text-success">
                    <font-awesome-icon
                      :icon="['fas', 'check-circle']"
                    ></font-awesome-icon>
                  </span>
                </span>
                <span v-else>
                  <span class="has-text-black">Reuse not permitted</span>
                  <span class="icon is-size-6 has-text-danger">
                    <font-awesome-icon
                      :icon="['fas', 'circle-exclamation']"
                    ></font-awesome-icon>
                  </span>
                </span>
              </a>
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
              ></CopyWithButton>
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
