<script setup>
import useProps from "@/utils/use-props.js";
import { toRel, toAbs } from "@/utils/link.js";

const props = defineProps({ params: Object, fileProps: Object });

const [d, saveData] = useProps(props.params.data, {
  active: ["ad-rail-active"],
  imageDescription: ["ad-rail-image-description"],
  images: ["ad-rail-images"],
  link: ["ad-rail-link", toAbs, toRel],
});

defineExpose({ saveData });
</script>

<template>
  <details class="mt-4">
    <summary class="title is-4">Homepage rail ad</summary>
    <BulmaField
      label="Rail ad is 600x300 rectangle at the top of the sidebar right rail"
    >
      <div>
        <label class="checkbox">
          <input v-model="d.active.value" type="checkbox" />
          Show featured ad on homepage
        </label>
      </div>
    </BulmaField>
    <div v-show="d.active.value">
      <BulmaFieldInput
        v-model="d.link.value"
        label="Ad link URL"
        type="url"
      ></BulmaFieldInput>
      <BulmaTextarea
        v-model="d.imageDescription.value"
        label="Image description (alt text)"
        help="For blind readers and search engines"
      ></BulmaTextarea>
      <BulmaField
        label="Images"
        help="If multiple images are provided, each page load will select one randomly"
      >
        <SiteParamsFiles
          :files="d.images.value"
          :file-props="fileProps"
          @add="d.images.value.push($event)"
          @remove="d.images.value.splice($event, 1)"
        ></SiteParamsFiles>
      </BulmaField>
    </div>
  </details>
</template>
