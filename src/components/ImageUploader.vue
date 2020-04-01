<script>
import { ref, computed } from "@vue/composition-api";

import { useClient } from "@/api/hooks.js";
import imgproxyURL from "@/api/imgproxy-url.js";

import CopyWithButton from "./CopyWithButton.vue";

export default {
  name: "ImageUploader",
  components: {
    CopyWithButton,
  },
  setup(props, { emit }) {
    let { uploadFile } = useClient();

    let isUploading = ref(false);
    let filename = ref("");
    let error = ref(null);

    async function uploadFileInput(ev) {
      if (isUploading.value) {
        return;
      }
      let { files } = ev.target;
      if (files.length !== 1) {
        error.value = new Error("Can only upload one file at a time");
        return;
      }
      if (files[0].type !== "image/jpeg") {
        error.value = new Error("Only JPEG is supported");
        return;
      }

      let [body] = files;
      isUploading.value = true;
      error.value = null;
      [filename.value, error.value] = await uploadFile(body);
      isUploading.value = false;
      emit("update-image-list");
    }

    let isDragging = ref(false);

    return {
      isUploading,
      uploadFileInput,
      filename,
      error,

      isDragging,
      dropFile(e) {
        isDragging.value = false;
        let { files = [] } = e.dataTransfer;
        uploadFileInput({ target: { files } });
      },

      imageURL: computed(() => imgproxyURL(filename.value)),
    };
  },
};
</script>

<template>
  <div>
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
                  accept="image/jpeg"
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
                  <span v-else class="file-label">
                    Choose a file…
                  </span>
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
    <div v-if="error" class="message is-danger">
      <p class="message-header" v-text="error.name"></p>
      <p class="message-body" v-text="error.message"></p>
    </div>
    <CopyWithButton
      v-else-if="filename"
      :value="filename"
      label="image path"
    ></CopyWithButton>
    <picture v-if="imageURL && !isUploading" class="has-ratio">
      <img :src="imageURL" class="is-3x4" width="200" />
    </picture>
  </div>
</template>
