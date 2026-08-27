<script setup lang="ts">
import { type Promotion } from "@/api/promotion.ts";

const props = defineProps<{
  promo: Promotion;
  fileList: object;
}>();

const {
  name,
  description,
  width,
  height,
  link,
  bannerLabel,
  bannerLabelLink,
  imageDescription,
  imageUrls,
} = props.promo;
</script>

<template>
  <BulmaFieldInput
    v-model="name"
    label="Name"
  />
  <BulmaTextarea
    v-model="description"
    label="Description"
    :rows="2"
  />
  <div class="is-flex mb-3" style="gap: 1rem">
    <BulmaFieldInput
      label="Width"
      inputmode="numeric"
      :model-value="width || ''"
      @update:model-value="width = +$event || 0"
    />
    <BulmaFieldInput
      label="Height"
      inputmode="numeric"
      :model-value="height || ''"
      @update:model-value="height = +$event || 0"
    />
  </div>
  <BulmaFieldInput
    v-model="link"
    label="Link URL"
    type="url"
    placeholder="https://www.spotlightpa.org/donate/"
  />
  <BulmaFieldInput
    v-model="bannerLabel"
    label="Banner label"
    placeholder="Sponsored by Acme"
    help="Text accompanying a banner specifying the sponsor or presenter"
  />
  <BulmaFieldInput
    v-model="bannerLabelLink"
    label="Banner label link"
    type="url"
    placeholder="https://www.spotlightpa.org/support/"
    help="Link that clicking the ad label will go to"
  />
  <BulmaTextarea
    v-model="imageDescription"
    label="Image description (alt text)"
    help="For blind readers and search engines"
  />
  <BulmaField
    label="Images"
    help="If multiple images are provided for the same promotion, each page load will select one randomly"
  >
    <SiteParamsFiles
      :files="imageUrls"
      :file-props="fileList"
      @add="imageUrls.push($event)"
      @remove="imageUrls.splice($event, 1)"
    />
  </BulmaField>
</template>
