<script setup>
import useProps from "@/utils/use-props.js";
import { toRel, toAbs } from "@/utils/link.js";

const props = defineProps({
  params: Object,
  fileProps: Object,
});

const [data, saveData] = useProps(props.params.data, {
  featuredPromoLink: ["featured-promo-link", toAbs, toRel],
  featuredPromoImageDescription: ["featured-promo-image-description"],
  featuredPromoBgColor: ["featured-promo-bg-color"],
  featuredPromoWidth: ["featured-promo-width"],
  featuredPromoHeight: ["featured-promo-height"],
  featuredPromoImages: ["featured-promo-images"],
});

defineExpose({ saveData });
</script>

<template>
  <details class="mt-4">
    <summary class="title is-4">Featured layout promo</summary>
    <p class="mb-4 content">
      Featured layout promo is a large house ad in articles using the featured
      layout.
    </p>
    <BulmaFieldInput
      v-model="data.featuredPromoLink.value"
      label="Featured promo link"
      type="url"
      help="Defaults to https://www.spotlightpa.org/donate/"
    />

    <BulmaTextarea
      v-model="data.featuredPromoImageDescription.value"
      label="Featured layout promo image description"
      help="For blind readers and search engines"
    />

    <BulmaFieldColor
      v-model="data.featuredPromoBgColor.value"
      label="Featured layout promo background color"
    />
    <div class="is-flex mb-2">
      <BulmaField v-slot="{ idForLabel }" label="Image width">
        <input
          :id="idForLabel"
          v-model.number="data.featuredPromoWidth.value"
          class="input"
          inputmode="numeric"
        />
      </BulmaField>
      <BulmaField v-slot="{ idForLabel }" class="ml-2" label="Image height">
        <input
          :id="idForLabel"
          v-model.number="data.featuredPromoHeight.value"
          class="input"
          inputmode="numeric"
        />
      </BulmaField>
    </div>

    <BulmaField
      label="Images"
      help="If multiple images are provided, each page load will select one randomly"
    >
      <SiteParamsFiles
        :files="data.featuredPromoImages.value"
        :file-props="fileProps"
        @add="data.featuredPromoImages.value.push($event)"
        @remove="data.featuredPromoImages.value.splice($event, 1)"
      />
    </BulmaField>
  </details>
</template>
