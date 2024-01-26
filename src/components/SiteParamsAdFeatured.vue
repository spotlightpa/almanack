<script setup>
import useData from "@/utils/use-data.js";
import { toRel, toAbs } from "@/utils/link.js";

const props = defineProps({ params: Object, fileProps: Object });

const d = useData(() => props.params.data, {
  active: ["ad-featured-active"],
  imageDescription: ["ad-featured-image-description"],
  images: ["ad-featured-images"],
  link: ["ad-featured-link", toAbs, toRel],
});
</script>

<template>
  <details class="mt-4">
    <summary class="title is-4">Homepage featured ad</summary>
    <BulmaField
      label="Featured ad is 250x300 square near the top of the homepage"
    >
      <div>
        <label class="checkbox">
          <input v-model="d.active.value" type="checkbox" />
          Show featured ad on homepage
        </label>
      </div>
    </BulmaField>
    <template v-if="d.active.value">
      <BulmaFieldInput v-model="d.link.value" label="Ad link URL" type="url" />
      <BulmaTextarea
        v-model="d.imageDescription.value"
        label="Image description (alt text)"
        help="For blind readers and search engines"
      />
      <BulmaField
        label="Images"
        help="If multiple images are provided, each page load will select one randomly"
      >
        <SiteParamsFiles
          :files="d.images.value"
          :file-props="fileProps"
          @add="d.images.value.push($event)"
          @remove="d.images.value.splice($event, 1)"
        />
      </BulmaField>
    </template>
  </details>
</template>
