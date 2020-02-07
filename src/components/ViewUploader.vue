<script>
import { ref, computed } from "@vue/composition-api";

import { useAuth } from "@/api/hooks.js";

import CopyWithButton from "./CopyWithButton.vue";

export default {
  name: "ViewUploader",
  components: {
    CopyWithButton,
  },
  setup() {
    let { isSpotlightPAUser } = useAuth();
    let isUploading = ref(false);
    let filename = ref("");
    let error = ref(null);
    async function uploadFile(ev) {
      if (isUploading.value) {
        return;
      }
      let [body] = ev.target.files;
      isUploading.value = true;
      error.value = null;
      try {
        let data = await fetch("/api/get-signed-upload", {
          method: "POST",
        }).then(rsp => rsp.json());
        let postURL = data["signed-url"];
        filename.value = data.filename;

        await fetch(postURL, { method: "PUT", body });
      } catch (e) {
        error.value = e;
      } finally {
        isUploading.value = false;
      }
    }

    return {
      isSpotlightPAUser,
      isUploading,
      uploadFile,
      filename,
      error,

      imageURL: computed(() => {
        if (!filename.value) {
          return "";
        }
        let baseURL = "https://images.data.spotlightpa.org";
        let signature = "insecure";
        let resizing_type = "auto";
        let width = 400;
        let height = 300;
        let gravity = "sm";
        let enlarge = "1";
        let quality = "75";
        let encoded_source_url = btoa(filename.value);
        let extension = "jpeg";

        return `${baseURL}/${signature}/rs:${resizing_type}:${width}:${height}/g:${gravity}/el:${enlarge}/q:${quality}/${encoded_source_url}.${extension}`;
      }),
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
                :class="{ 'is-warning': isUploading }"
                :disabled="isUploading"
              >
                <label class="file-label">
                  <input
                    type="file"
                    accept="image/jpeg"
                    class="file-input"
                    @change="uploadFile"
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
