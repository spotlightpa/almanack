<script setup>
import { computed, ref } from "vue";

import imgproxyURL from "@/api/imgproxy-url.js";

function ratio(width, wr, hr) {
  return Math.floor((width * hr) / wr);
}

const path = ref("");
const width = ref(600);
const height = ref(ratio(600, 16, 9));
const gravity = ref("sm");
const extension = ref("jpeg");
const quality = ref(0);

const badPath = computed(() => path.value.startsWith("http"));

const url = computed(
  () =>
    !badPath.value &&
    imgproxyURL(path.value, {
      width: width.value,
      height: height.value,
      gravity: gravity.value,
      extension: extension.value,
      quality: quality.value,
    })
);

function doRatio(wr, hr) {
  height.value = ratio(width.value, wr, hr);
}
</script>

<template>
  <BulmaField
    label="Photo ID"
    help="Almanack photo path for image to resize"
    v-slot="{ idForLabel }"
  >
    <div class="is-flex">
      <input :id="idForLabel" v-model="path" class="input" />
      <BulmaPaste @paste="path = $event"></BulmaPaste>
    </div>
  </BulmaField>
  <p class="help is-danger mb-5" v-if="badPath">
    Invalid path. Path must not be a URL!
  </p>
  <div class="field is-horizontal">
    <div class="field-label is-normal">
      <label class="label">Width</label>
    </div>
    <div class="field-body">
      <div class="field">
        <p class="control is-expanded">
          <input class="input is-small" type="number" v-model="width" />
        </p>
      </div>
    </div>
    <div class="ml-5 field-label is-normal">
      <label class="label">Height</label>
    </div>
    <div class="field-body">
      <div class="field">
        <p class="control is-expanded">
          <input class="input is-small" type="number" v-model="height" />
        </p>
      </div>
    </div>
    <div class="ml-5 field-label is-normal">
      <label class="label">Quality</label>
    </div>
    <div class="field-body">
      <div class="field">
        <p class="control is-expanded">
          <input class="input is-small" type="number" v-model="quality" />
        </p>
      </div>
    </div>
  </div>
  <div class="buttons">
    <button
      class="button is-primary has-text-weight-semibold"
      @click="doRatio(16, 9)"
    >
      16 x 9
    </button>
    <button
      class="button is-primary has-text-weight-semibold"
      @click="doRatio(3, 2)"
    >
      3 x 2
    </button>
    <button
      class="button is-primary has-text-weight-semibold"
      @click="doRatio(4, 3)"
    >
      4 x 3
    </button>
    <button
      class="button is-primary has-text-weight-semibold"
      @click="doRatio(1, 1)"
    >
      Square
    </button>
    <button
      class="button is-primary has-text-weight-semibold"
      @click="doRatio(4, 5)"
    >
      4 x 5
    </button>
  </div>

  <div v-if="url">
    <span class="label">Right click to save as</span>
    <div class="is-flex" v-if="width <= 400">
      <div>
        <p>
          <a :href="url" download="" target="_blank">
            <img
              :src="url"
              :width="width"
              :height="height"
              class="border-thick"
            />
          </a>
        </p>
        <p class="has-text-centered">Standard 1x</p>
      </div>
      <div class="ml-5">
        <p>
          <a :href="url" download="" target="_blank">
            <img
              :src="url"
              :width="width / 2"
              :height="height / 2"
              class="border-thick"
            />
          </a>
        </p>
        <p class="has-text-centered">Retina 2x</p>
      </div>
    </div>
    <div v-else>
      <p>
        <a :href="url" download="" target="_blank">
          <img
            :src="url"
            :width="width"
            :height="height"
            class="border-thick"
          />
        </a>
      </p>
    </div>
  </div>

  <BulmaField label="Image focus">
    <div class="control is-expanded">
      <span class="select is-fullwidth">
        <select v-model="gravity">
          <option
            v-for="[val, desc] in [
              ['', 'Auto'],
              ['we', 'Left'],
              ['no', 'Top'],
              ['ea', 'Right'],
              ['so', 'Bottom'],
              ['ce', 'Center'],
            ]"
            :key="val"
            :value="val"
          >
            {{ desc }}
          </option>
        </select>
      </span>
    </div>
  </BulmaField>
  <BulmaField label="File format">
    <div class="control is-expanded">
      <span class="select is-fullwidth">
        <select v-model="extension">
          <option
            v-for="[val, desc] in [
              ['jpeg', 'JPEG'],
              ['png', 'PNG'],
              ['webp', 'WebP'],
              ['tiff', 'TIFF'],
              ['avif', 'AVIF'],
              ['heic', 'HEIC'],
            ]"
            :key="val"
            :value="val"
          >
            {{ desc }}
          </option>
        </select>
      </span>
    </div>
  </BulmaField>
  <div class="mt-5" v-if="url">
    <CopyWithButton
      :value="url"
      label="resized image URL"
      size="is-small"
    ></CopyWithButton>
  </div>
</template>

<style scoped>
.border-thick {
  border: 2px solid #ccc;
}
</style>
