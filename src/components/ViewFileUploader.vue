<script>
import { reactive, computed, toRefs } from "@vue/composition-api";

import { useClient, makeState } from "@/api/hooks.js";

import { formatDate } from "@/utils/time-format.js";

export default {
  name: "ViewUploader",
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
        state.isUploading = true;
        state.uploadError = null;

        for (let body of files) {
          [state.fileURL, state.uploadError] = await uploadFile(body);
          if (state.uploadError) {
            break;
          }
        }
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
                  multiple
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
    <ErrorSimple :error="uploadError" />

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
                <font-awesome-icon :icon="['fas', 'file-download']" size="2x" />
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
              <p>
                <CopyWithButton :value="file.url" label="URL" size="is-small" />
              </p>
            </td>
          </tr>
        </tbody>
      </table>
    </APILoader>
  </div>
</template>
