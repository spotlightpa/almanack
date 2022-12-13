<script>
import { reactive, toRefs, computed } from "vue";

import { useClient } from "@/api/hooks.js";
import imgproxyURL from "@/api/imgproxy-url.js";

export default {
  name: "ImageUploader",
  setup(props, { emit }) {
    let { uploadImage } = useClient();

    let state = reactive({
      isUploading: false,
      filename: "",
      error: null,
      isDragging: false,

      imageURL: computed(() => imgproxyURL(state.filename)),
    });

    let actions = {
      async uploadFileInput(ev) {
        if (state.isUploading) {
          return;
        }
        let { files } = ev.target;

        for (let body of files) {
          if (
            ![
              "image/jpeg",
              "image/png",
              "image/tiff",
              "image/webp",
              "image/avif",
              "image/heic",
            ].includes(body.type)
          ) {
            state.error = new Error(
              "Only JPEG, PNG, WEBP, AVIF, HEIC, and TIFF are supported"
            );
            return;
          }
        }
        state.isUploading = true;
        state.error = null;

        for (let body of files) {
          [state.filename, state.error] = await uploadImage(body);
          if (state.error) {
            break;
          }
        }

        state.isUploading = false;
        emit("update-image-list");
      },
      dropFile(ev) {
        state.isDragging = false;
        let { files = [] } = ev.dataTransfer;
        return actions.uploadFileInput({ target: { files } });
      },
    };

    return {
      ...toRefs(state),
      ...actions,
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
              :disabled="isUploading || null"
              @dragover.prevent="isDragging = true"
              @dragleave.prevent="isDragging = false"
              @drop.prevent="dropFile"
            >
              <label class="file-label">
                <input
                  type="file"
                  accept="image/jpeg,image/png,image/tiff,image/webp,image/avif,image/heic"
                  class="file-input"
                  multiple
                  @change="uploadFileInput"
                />

                <span class="file-cta" :disabled="isUploading || null">
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
    <ErrorSimple :error="error" />
    <CopyWithButton v-if="filename" :value="filename" label="image path" />
    <picture v-if="imageURL && !isUploading" class="has-ratio">
      <img :src="imageURL" class="is-3x4" width="200" />
    </picture>
  </div>
</template>
