<script>
import { ref, computed } from "@vue/composition-api";

import { useAuth, useService } from "@/api/hooks.js";
import imgproxyURL from "@/api/imgproxy-url.js";

import CopyWithButton from "./CopyWithButton.vue";

export default {
  name: "ViewUploader",
  components: {
    CopyWithButton,
  },
  setup() {
    let { isSpotlightPAUser } = useAuth();
    let { uploadFile } = useService();

    let isUploading = ref(false);
    let filename = ref("");
    let error = ref(null);

    async function uploadFileInput(ev) {
      if (isUploading.value) {
        return;
      }
      let [body] = ev.target.files;
      isUploading.value = true;
      error.value = null;
      [filename.value, error.value] = await uploadFile(body);
      isUploading.value = false;
    }

    let isDragging = ref(false);

    return {
      isSpotlightPAUser,
      isUploading,
      uploadFileInput,
      filename,
      error,

      isDragging,
      dropFile(e) {
        isDragging.value = false;
        let { files = [] } = e.dataTransfer;
        if (files.length !== 1) {
          error.value = new Error("Can only upload one file at a time");
          return;
        }
        if (files[0].type !== "image/jpeg") {
          error.value = new Error("Only JPEG is supported");
          return;
        }
        uploadFileInput({ target: { files } });
      },

      imageURL: computed(() => imgproxyURL(filename.value)),
    };
  },
};
</script>

<template>
  <div>
    <div v-if="!isSpotlightPAUser" class="message is-danger">
      <p class="message-header">Not Authorized</p>

      <p class="message-body">
        You do not have permission to use this page.
        <strong
          ><router-link :to="{ name: 'home' }">Go home</router-link>?</strong
        >
      </p>
    </div>
    <div v-else>
      <h1 class="title">Upload an image</h1>
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
  </div>
</template>
