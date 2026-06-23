<script setup>
import { computed } from "vue";

import { apAgreed } from "@/api/ap-agreed.js";

const props = defineProps({
  imageUrl: { type: String, default: "" },
  downloadUrl: { type: String, default: "" },
  description: { type: String, default: "" },
  credit: { type: String, default: "" },
  caption: { type: String, default: "" },
});

const isAP = computed(() => /\bAP|Associated Press\b/.test(props.credit));
const canDownload = computed(() => !isAP.value || apAgreed.value);
</script>

<template>
  <figure>
    <div class="has-margin-bottom">
      <component
        :is="canDownload ? 'a' : 'span'"
        :href="downloadUrl"
        class="is-inline-flex max-256 has-background-grey-lighter"
        target="_blank"
        download
      >
        <img class="max-256" :src="imageUrl" width="256" height="192" />
      </component>
    </div>
    <figcaption>
      <template v-if="isAP">
        <label
          class="checkbox has-margin-bottom is-flex is-align-items-center"
          style="gap: 0.5rem"
        >
          <input type="checkbox" v-model="apAgreed" />
          I affirm that my organization has permission to use AP images
        </label>
      </template>
      <p class="has-margin-bottom">
        <component
          :is="canDownload ? 'a' : 'button'"
          :href="downloadUrl"
          class="button is-danger has-text-weight-semibold"
          target="_blank"
          :download="canDownload || null"
          :type="canDownload ? null : 'button'"
          :disabled="!canDownload || null"
        >
          <span class="icon">
            <font-awesome-icon
              :icon="['fas', 'file-download']"
            ></font-awesome-icon>
          </span>
          <span>Download image</span>
        </component>
      </p>
      <template v-if="description">
        <p class="has-margin-bottom-thin">
          <strong>Description (“alt” text):</strong>
        </p>
        <CopyWithButton
          :value="description"
          label="description"
        ></CopyWithButton>
      </template>
      <template v-if="credit">
        <p class="has-margin-bottom-thin">
          <strong>Credit:</strong>
        </p>
        <CopyWithButton :value="credit" label="credit"></CopyWithButton>
      </template>
      <template v-if="caption">
        <p class="has-margin-bottom-thin">
          <strong>Caption:</strong>
        </p>
        <CopyWithButton :value="caption" label="caption"></CopyWithButton>
      </template>
    </figcaption>
  </figure>
</template>

<style scoped>
.max-256 {
  width: auto;
  height: auto;
  max-height: 256px;
  max-width: 256px;
  min-height: 1rem;
  min-width: 1rem;
}
</style>
