<script setup>
import useProps from "@/utils/use-props.js";
import { toRel, toAbs } from "@/utils/link.js";

const props = defineProps({ params: Object, fileProps: Object });

const [d, saveData] = useProps(props.params.data, {
  active: ["ad-headwater-active"],
  imageDescription: ["ad-headwater-image-description"],
  images: ["ad-headwater-images"],
  link: ["ad-headwater-link", toAbs, toRel],
});

defineExpose({ saveData });
</script>

<template>
  <details class="mt-4">
    <summary class="title is-4">Homepage headwater ad</summary>
    <BulmaField
      label="Headwater ad is 970x90 rectangle before the homepage river"
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
