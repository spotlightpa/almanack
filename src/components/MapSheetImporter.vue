<script setup>
import { ref } from "vue";
import { useRouter } from "vue-router";

import { post, postMapSheet } from "@/api/client.ts";
import { makeState } from "@/api/loader.ts";
import { prepareMapSheet } from "@/api/map-sheet.js";

const router = useRouter();

const mapSheetImportURL = ref("");
const { apiStateRefs, exec } = makeState();
const { isLoading, error, rawData } = apiStateRefs;

async function importMapSheetURL(url) {
  await exec(async () => {
    let [info, err] = await prepareMapSheet(url);
    if (err) {
      return [null, err];
    }
    return await post(postMapSheet, { id: info.sheet_id });
  });
  if (error.value) {
    return;
  }

  router.push({
    name: "news-page",
    params: {
      id: "" + rawData.value,
    },
  });
}
</script>

<template>
  <label for="map-sheet-importer" class="mt-4 label">
    Import a custom template from Google Sheets
  </label>
  <form
    class="field is-grouped"
    @submit.prevent="importMapSheetURL(mapSheetImportURL)"
  >
    <div class="control is-expanded">
      <input
        id="map-sheet-importer"
        v-model="mapSheetImportURL"
        class="input"
        placeholder="https://docs.google.com/spreadsheets/d/abc123/edit"
      />
    </div>
    <div class="control">
      <BulmaPaste
        @paste="
          mapSheetImportURL = $event;
          importMapSheetURL(mapSheetImportURL);
        "
      ></BulmaPaste>
    </div>
    <div class="control">
      <button
        class="button is-success has-text-weight-semibold"
        :class="isLoading && 'is-loading'"
        :disabled="!mapSheetImportURL || null"
      >
        Import
      </button>
    </div>
  </form>
  <div class="field">
    <p class="help">
      Sheet must be shared with Spotlight PA.
      <a
        href="https://drive.google.com/drive/folders/1IfzFMfX596bvkD5ArtfyE5PgWu-ECzrC"
        target="_blank"
        >Templates folder</a
      >.
    </p>
  </div>
  <div v-if="error" class="field">
    <ErrorSimple :error="error"></ErrorSimple>
  </div>
</template>
