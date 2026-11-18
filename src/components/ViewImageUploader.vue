<script setup>
import { ref } from "vue";
import { useDebouncedRef, seconds } from "@/utils/wait.ts";

import { get, post, listImages, postImageUpdate } from "@/api/client.ts";
import { makeState, watchAPI } from "@/api/loader.ts";
import imgproxyURL from "@/api/imgproxy-url.js";

import humanSize from "@/utils/human-size.ts";

const props = defineProps({
  page: { type: String, default: "0" },
});

const rawQuery = ref("");
const query = useDebouncedRef(rawQuery, seconds(1));

const {
  apiState: listState,
  fetch,
  computedList,
  computedProp,
} = watchAPI(
  () => [props.page || 0, query.value],
  ([page, query]) => get(listImages, { page, query, show_unlicensed: "1" })
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
  width: rawImage.width ?? 0,
  height: rawImage.height ?? 0,
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
</script>

<template>
  <MetaHead>
    <title>Photo Uploads • Spotlight PA</title>
  </MetaHead>
  <div>
    <h1 class="mb-2 title">
      Upload an image
      <template v-if="page !== '0'">(overflow page {{ page }})</template>
    </h1>
    <p class="mb-5 hint">
      <a href="/ssr/redirect/image-guide" target="_blank">Image guide</a>.
    </p>
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
        <ImageUploadRow
          v-for="image of images"
          :key="image.id"
          :image="image"
          @update-description="updateDescription"
          @update-credit="updateCredit"
          @update-keywords="updateKeywords"
          @update-is-licensed="updateIsLicensed"
          @update-image-list="fetch"
        ></ImageUploadRow>
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
