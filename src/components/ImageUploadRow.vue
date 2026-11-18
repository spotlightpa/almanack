<script setup>
import { formatDate } from "@/utils/time-format.ts";

defineProps({
  image: { type: Object, required: true },
});

defineEmits([
  "update-description",
  "update-credit",
  "update-keywords",
  "update-is-licensed",
  "update-image-list",
]);
</script>

<template>
  <tr>
    <td>
      <div class="max-128">
        <a :href="image.downloadURL" class="is-flex">
          <picture
            class="image is-3x4 has-background-grey-lighter has-margin-bottom"
          >
            <img
              :src="image.url"
              class="border-thick"
              width="256"
              height="192"
              loading="lazy"
            />
          </picture>
        </a>
      </div>
      <div class="mt-4">
        <LinkHref
          :label="image.size ? `Original (${image.size})` : 'Original'"
          color="is-success"
          :icon="['fas', 'file-download']"
          :href="image.downloadURL"
        ></LinkHref>
        <ImageSize
          class="mt-1"
          :path="image.path"
          :width="image.width"
          :height="image.height"
          @update="$emit('update-image-list')"
        ></ImageSize>
      </div>
    </td>
    <td>
      <p>
        <a
          class="has-text-weight-semibold"
          @click="$emit('update-description', image)"
        >
          Description:
        </a>
        {{ image.description }}
      </p>

      <p>
        <a
          class="has-text-weight-semibold"
          @click="$emit('update-credit', image)"
        >
          Credit:
        </a>
        {{ image.credit }}
      </p>

      <p>
        <a
          class="has-text-weight-semibold"
          @click="$emit('update-keywords', image)"
        >
          Keywords:
        </a>
        {{ image.keywords }}
      </p>
      <p>
        <a @click="$emit('update-is-licensed', image)">
          <span class="has-text-weight-semibold">License: </span>
          <span v-if="image.isLicensed">
            <span class="has-text-black">Reuse permitted</span>
            <span class="icon is-size-6 has-text-success">
              <font-awesome-icon
                :icon="['fas', 'check-circle']"
              ></font-awesome-icon>
            </span>
          </span>
          <span v-else>
            <span class="has-text-black">Reuse not permitted</span>
            <span class="icon is-size-6 has-text-danger">
              <font-awesome-icon
                :icon="['fas', 'circle-exclamation']"
              ></font-awesome-icon>
            </span>
          </span>
        </a>
      </p>
      <p>
        <strong>Date:</strong>
        {{ formatDate(image.date) }}
      </p>
      <p class="has-margin-top-thin">
        <CopyWithButton
          :value="image.path"
          label="path"
          size="is-small"
        ></CopyWithButton>
      </p>
    </td>
    <td></td>
  </tr>
</template>

<style scoped>
.max-128 {
  height: 72px;
  width: 128px;
  min-height: 1rem;
  min-width: 1rem;
}
.border-thick {
  border: 2px solid #ccc;
}
</style>
