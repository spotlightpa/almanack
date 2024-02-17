<script>
import useProps from "@/utils/use-props.js";
import { toRel, toAbs } from "@/utils/link.js";

export default {
  props: { params: Object, fileProps: Object },
  setup(props) {
    let [data, saveData] = useProps(props.params.data, {
      footerDescription: ["footer-promo-description"],
      footerLink: ["footer-promo-link", toAbs, toRel],
      footerDesktopImages: ["footer-promo-desktop-images"],
      footerDesktopWidth: ["footer-promo-desktop-width"],
      footerDesktopHeight: ["footer-promo-desktop-height"],
      footerMobileImages: ["footer-promo-mobile-images"],
      footerMobileWidth: ["footer-promo-mobile-width"],
      footerMobileHeight: ["footer-promo-mobile-height"],
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
    <summary class="title is-4">Footer promo</summary>
    <BulmaFieldInput
      v-model="footerLink"
      label="Footer promo link"
      type="url"
    />

    <BulmaTextarea
      v-model="footerDescription"
      label="Footer promo image description"
      help="For blind readers and search engines"
    />
    <div class="is-flex mb-2">
      <BulmaField v-slot="{ idForLabel }" label="Desktop Image Width">
        <input
          :id="idForLabel"
          v-model.number="footerDesktopWidth"
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
          v-model.number="footerDesktopHeight"
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
        :files="footerDesktopImages"
        :file-props="fileProps"
        @add="footerDesktopImages.push($event)"
        @remove="footerDesktopImages.splice($event, 1)"
      />
    </BulmaField>

    <div class="is-flex mb-2">
      <BulmaField v-slot="{ idForLabel }" label="Mobile Image Width">
        <input
          :id="idForLabel"
          v-model.number="footerMobileWidth"
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
          v-model.number="footerMobileHeight"
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
        :files="footerMobileImages"
        :file-props="fileProps"
        @add="footerMobileImages.push($event)"
        @remove="footerMobileImages.splice($event, 1)"
      />
    </BulmaField>
  </details>
</template>
