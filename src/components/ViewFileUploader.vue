<script setup>
import { ref, computed, watch } from "vue";

import {
  get,
  post,
  listFiles,
  updateFile,
  uploadFile,
} from "@/api/client-v2.js";
import { makeState } from "@/api/service-util.js";

import { formatDate } from "@/utils/time-format.js";
import humanSize from "@/utils/human-size.js";

const props = defineProps({
  page: { type: String, default: "0" },
});

const { apiStateRefs, exec } = makeState();
const { rawData, isLoading, error } = apiStateRefs;

const isDragging = ref(false);
const isUploading = ref(false);
const uploadError = ref(null);

const files = computed(() => rawData.value?.files || []);
const nextPage = computed(() => {
  if (!rawData.value?.next_page) {
    return null;
  }
  return {
    name: "file-uploader",
    query: {
      page: "" + rawData.value.next_page,
    },
  };
});

async function fetch() {
  exec(() => get(listFiles, { page: props.page }));
}

function updateDescription(file) {
  let description = window.prompt("Update description", file.description);
  if (description !== null && description !== file.description) {
    exec(() =>
      post(updateFile, {
        url: file.url,
        description,
        set_description: true,
      }).then(() => get(listFiles, { page: props.page }))
    );
  }
}

async function uploadFileInput(ev) {
  let { files } = ev.target;
  isUploading.value = true;
  uploadError.value = null;

  for (let body of files) {
    [, uploadError.value] = await uploadFile(body);
    if (uploadError.value) {
      break;
    }
  }
  isUploading.value = false;
  await fetch();
}

function dropFile(ev) {
  isDragging.value = false;
  let { files = [] } = ev.dataTransfer;
  return uploadFileInput({ target: { files } });
}

watch(
  () => props.page,
  () => fetch(),
  { immediate: true }
);
</script>

<template>
  <MetaHead>
    <title>File Uploads • Spotlight PA</title>
  </MetaHead>
  <div>
    <h1 class="title">
      Upload a file
      <template v-if="page !== '0'">(overflow page {{ page }})</template>
    </h1>
    <div class="level">
      <div class="level-left">
        <div class="level-item">
          <div class="is-inline-block">
            <fieldset
              class="file"
              :class="
                isUploading ? 'is-warning' : isDragging ? 'is-success' : ''
              "
              :disabled="isUploading || null"
              @dragover.prevent="isDragging = true"
              @dragleave.prevent="isDragging = false"
              @drop.prevent="dropFile"
            >
              <label class="file-label">
                <input
                  type="file"
                  class="file-input"
                  multiple
                  @change="uploadFileInput"
                />

                <span class="file-cta" :disabled="isUploading || null">
                  <span class="file-icon">
                    <font-awesome-icon
                      :icon="['fas', 'sync-alt']"
                      :spin="isUploading"
                    ></font-awesome-icon>
                  </span>
                  <span
                    v-if="isUploading"
                    class=""
                    :class="
                      isUploading ? 'has-text-weight-semibold' : 'file-label'
                    "
                  >
                    Uploading…
                  </span>
                  <span v-else class="file-label"> Choose a file… </span>
                </span>
              </label>
            </fieldset>
          </div>
        </div>
        <div class="level-item">
          <span
            v-if="isUploading"
            class="tag is-success has-text-weight-semibold"
            >Uploading</span
          >
        </div>
      </div>
    </div>
    <ErrorSimple :error="uploadError"></ErrorSimple>

    <h2 class="title has-margin-top">Existing files</h2>
    <APILoader :is-loading="isLoading" :reload="fetch" :error="error">
      <table class="table is-striped is-narrow is-fullwidth">
        <tbody>
          <tr v-for="file of files" :key="file.id">
            <td style="vertical-align: middle">
              <a
                class="icon has-text-success"
                :href="file.url"
                target="_blank"
                :title="`Download ${file.filename}`"
              >
                <font-awesome-icon
                  :icon="['fas', 'file-download']"
                  size="2x"
                ></font-awesome-icon>
              </a>
            </td>
            <td>
              <p>
                <strong>Name: </strong>
                {{ file.filename }}
              </p>
              <p>
                <strong>Uploaded: </strong>
                {{ formatDate(file.created_at) }}
              </p>
              <p>
                <a
                  class="has-text-weight-semibold"
                  @click="updateDescription(file)"
                >
                  Description:
                </a>
                {{ file.description || "&lt;no description&gt;" }}
              </p>
              <p v-if="file.bytes">
                <strong>Size: </strong>
                {{ humanSize(file.bytes) }}
              </p>
              <p>
                <CopyWithButton
                  :value="file.url"
                  label="URL"
                  size="is-small"
                ></CopyWithButton>
              </p>
            </td>
          </tr>
        </tbody>
      </table>
      <div class="buttons mt-5">
        <router-link
          v-if="nextPage"
          :to="nextPage"
          class="button is-primary has-text-weight-semibold"
        >
          Show Older Files…
        </router-link>
      </div>
    </APILoader>
  </div>
</template>
