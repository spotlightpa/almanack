<script setup>
import { ref, computed } from "vue";

import { uploadImage } from "@/api/client-v2.js";
import imgproxyURL from "@/api/imgproxy-url.js";

const acceptedTypes = [
  "image/jpeg",
  "image/png",
  "image/tiff",
  "image/webp",
  "image/avif",
  "image/heic",
];

const emit = defineEmits(["update-image-list"]);

const isUploading = ref(false);
const filename = ref("");
const error = ref(null);
const isDragging = ref(false);

const imageURL = computed(() => imgproxyURL(filename.value));

async function uploadFileInput(ev) {
  if (isUploading.value) {
    return;
  }
  let { files } = ev.target;

  for (let body of files) {
    if (!acceptedTypes.includes(body.type)) {
      error.value = new Error(
        "Only JPEG, PNG, WEBP, AVIF, HEIC, and TIFF are supported"
      );
      return;
    }
  }
  isUploading.value = true;
  error.value = null;

  for (let body of files) {
    [filename.value, error.value] = await uploadImage(body);
    if (error.value) {
      break;
    }
  }

  isUploading.value = false;
  emit("update-image-list");
}

function dropFile(ev) {
  isDragging.value = false;
  let { files = [] } = ev.dataTransfer;
  return uploadFileInput({ target: { files } });
}
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
                  :accept="acceptedTypes.join(',')"
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
    <ErrorSimple :error="error"></ErrorSimple>
    <CopyWithButton
      v-if="filename"
      :value="filename"
      label="image path"
    ></CopyWithButton>
    <picture v-if="imageURL && !isUploading" class="has-ratio">
      <img :src="imageURL" class="is-3x4" width="200" />
    </picture>
  </div>
</template>
