<script>
import { reactive, computed, toRefs } from "@vue/composition-api";

import { useClient, makeState } from "@/api/hooks.js";

import { formatDate } from "@/utils/time-format.js";

import APILoader from "./APILoader.vue";
import CopyWithButton from "./CopyWithButton.vue";

export default {
  name: "ViewUploader",
  components: {
    APILoader,
    CopyWithButton,
  },
  props: ["page"],
  metaInfo: {
    title: "File Uploads",
  },
  setup() {
    let { listFiles, updateFile, uploadFile } = useClient();
    let { apiState, exec } = makeState();

    let state = reactive({
      files: computed(() => {
        return apiState.rawData?.files || [];
      }),
      isDragging: false,
      isUploading: false,
      uploadError: null,
    });

    let actions = {
      async fetch() {
        exec(listFiles);
      },
      updateDescription(file) {
        let description = window.prompt("Update description", file.description);
        if (description !== null && description !== file.description) {
          exec(() =>
            Promise.resolve()
              .then(() => updateFile(file.url, { description }))
              .then(listFiles)
          );
        }
      },
      async uploadFileInput(ev) {
        let { files } = ev.target;
        if (files.length !== 1) {
          state.uploadError = new Error("Can only upload one file at a time");
          return;
        }
        let [body] = files;
        state.isUploading = true;
        state.uploadError = null;
        [state.fileURL, state.uploadError] = await uploadFile(body);
        state.isUploading = false;
        await actions.fetch();
      },
      dropFile(ev) {
        state.isDragging = false;
        let { files = [] } = ev.dataTransfer;
        return actions.uploadFileInput({ target: { files } });
      },
    };

    actions.fetch();

    return {
      ...toRefs(apiState),
      ...toRefs(state),
      ...actions,

      formatDate,
    };
  },
};
</script>

<template>
  <div>
    <h1 class="title">Upload a file</h1>
    <div class="level">
      <div class="level-left">
        <div class="level-item">
          <div class="is-inline-block">
            <fieldset
              class="file"
              :class="
                isUploading ? 'is-warning' : isDragging ? 'is-success' : ''
              "
              :disabled="isUploading"
              @dragover.prevent="isDragging = true"
              @dragleave.prevent="isDragging = false"
              @drop.prevent="dropFile"
            >
              <label class="file-label">
                <input
                  type="file"
                  class="file-input"
                  @change="uploadFileInput"
                />

                <span class="file-cta" :disabled="isUploading">
                  <span class="file-icon">
                    <font-awesome-icon
                      :icon="['fas', 'sync-alt']"
                      :spin="isUploading"
                    />
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
    <div v-if="uploadError" class="message is-danger">
      <p class="message-header" v-text="uploadError.name"></p>
      <p class="message-body" v-text="uploadError.message"></p>
    </div>

    <h2 class="title has-margin-top">Existing files</h2>
    <APILoader :is-loading="isLoading" :reload="fetch" :error="error">
      <table class="table is-striped is-fullwidth">
        <thead>
          <tr>
            <th>URL</th>
            <th>Name</th>
            <th>Description</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="file of files" :key="file.id">
            <td>
              <CopyWithButton :value="file.url" label="URL" size="is-small" />
            </td>
            <td>
              <a :href="file.url" target="_blank" v-text="file.filename" />
            </td>
            <td>
              <p>
                <a @click="updateDescription(file)">
                  {{ file.description || "&lt;no description&gt;" }}
                </a>
              </p>
            </td>
            <td>
              {{ formatDate(file.created_at) }}
            </td>
          </tr>
        </tbody>
      </table>
    </APILoader>
  </div>
</template>
