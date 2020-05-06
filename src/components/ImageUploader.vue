<script>
import { reactive, toRefs, computed } from "@vue/composition-api";

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
        if (files.length !== 1) {
          state.error = new Error("Can only upload one file at a time");
          return;
        }
        let [body] = files;
        if (!["image/jpeg", "image/png"].includes(body.type)) {
          state.error = new Error("Only JPEG and PNG are supported");
          return;
        }

        state.isUploading = true;
        state.error = null;
        [state.filename, state.error] = await uploadFile(body);
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
              :disabled="isUploading"
              @dragover.prevent="isDragging = true"
              @dragleave.prevent="isDragging = false"
              @drop.prevent="dropFile"
            >
              <label class="file-label">
                <input
                  type="file"
                  accept="image/jpeg,image/png"
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
