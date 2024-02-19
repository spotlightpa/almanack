<script>
import useProps from "@/utils/use-props.js";
import { toRel, toAbs } from "@/utils/link.js";

export default {
  props: { params: Object, fileProps: Object },
  setup(props) {
    let [data, saveData] = useProps(props.params.data, {
      riverDescription: ["river-promo-description"],
      riverLink: ["river-promo-link", toAbs, toRel],
      riverDesktopImages: ["river-promo-desktop-images"],
      riverDesktopWidth: ["river-promo-desktop-width"],
      riverDesktopHeight: ["river-promo-desktop-height"],
      riverMobileImages: ["river-promo-mobile-images"],
      riverMobileWidth: ["river-promo-mobile-width"],
      riverMobileHeight: ["river-promo-mobile-height"],
    });
    return {
      ...data,
      saveData,
    };
  },
};
</script>

<template>
  <details class="mt-4">
    <summary class="title is-4">Homepage river</summary>
    <BulmaFieldInput
      v-model="riverLink"
      label="Homepage river promo link"
      type="url"
    />
    <BulmaTextarea
      v-model="riverDescription"
      label="Homepage river promo image description"
      help="For blind readers and search engines"
    />
    <div class="is-flex mb-2">
      <BulmaField v-slot="{ idForLabel }" label="Desktop Image Width">
        <input
          :id="idForLabel"
          v-model.number="riverDesktopWidth"
          class="input"
          inputmode="numeric"
        />
      </BulmaField>
      <BulmaField
        v-slot="{ idForLabel }"
        class="ml-2"
        label="Desktop Image Height"
      >
        <input
          :id="idForLabel"
          v-model.number="riverDesktopHeight"
          class="input"
          inputmode="numeric"
        />
      </BulmaField>
    </div>
    <BulmaField
      label="Desktop Images"
      help="If multiple images are provided, each page load will select one randomly"
    >
      <SiteParamsFiles
        :files="riverDesktopImages"
        :file-props="fileProps"
        @add="riverDesktopImages.push($event)"
        @remove="riverDesktopImages.splice($event, 1)"
      />
    </BulmaField>

    <div class="is-flex mb-2">
      <BulmaField v-slot="{ idForLabel }" label="Mobile Image Width">
        <input
          :id="idForLabel"
          v-model.number="riverMobileWidth"
          class="input"
          inputmode="numeric"
        />
      </BulmaField>
      <BulmaField
        v-slot="{ idForLabel }"
        class="ml-2"
        label="Mobile Image Height"
      >
        <input
          :id="idForLabel"
          v-model.number="riverMobileHeight"
          class="input"
          inputmode="numeric"
        />
      </BulmaField>
    </div>

    <BulmaField
      label="Mobile Images"
      help="If multiple images are provided, each page load will select one randomly"
    >
      <SiteParamsFiles
        :files="riverMobileImages"
        :file-props="fileProps"
        @add="riverMobileImages.push($event)"
        @remove="riverMobileImages.splice($event, 1)"
      />
    </BulmaField>
  </details>
</template>
