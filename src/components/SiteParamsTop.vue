<script>
import useProps from "@/utils/use-props.js";
import { toRel, toAbs } from "@/utils/link.js";

export default {
  props: { params: Object, fileProps: Object },
  setup(props) {
    let [data, saveData] = useProps(props.params.data, {
      promoActive: ["promo-active"],
      promoType: ["promo-type"],
      promoImageDescription: ["promo-image-description"],
      promoDesktopImages: ["promo-desktop-images"],
      promoDesktopWidth: ["promo-desktop-width"],
      promoDesktopHeight: ["promo-desktop-height"],
      promoMobileImages: ["promo-mobile-images"],
      promoMobileWidth: ["promo-mobile-width"],
      promoMobileHeight: ["promo-mobile-height"],
      promoLink: ["promo-link", toAbs, toRel],
      promoText: ["promo-text"],
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
    <summary class="title is-4">Top promo</summary>
    <BulmaField label="Top promo">
      <div>
        <label class="checkbox">
          <input v-model="promoActive" type="checkbox" />
          Top promo is the native ad-like slot at top of the page
        </label>
      </div>
    </BulmaField>
    <div v-show="promoActive">
      <BulmaFieldInput
        v-model="promoLink"
        label="Top promo link"
        type="url"
      ></BulmaFieldInput>
      <BulmaField v-slot="{ idForLabel }" label="Top promo kind">
        <div class="select is-fullwidth">
          <select :id="idForLabel" v-model="promoType" class="select">
            <option value="image">Image</option>
            <option value="text">Text</option>
          </select>
        </div>
      </BulmaField>

      <BulmaTextarea
        v-show="promoType === 'text'"
        v-model="promoText"
        label="Top promo text"
        help="Text will appear between navbar and page content"
      ></BulmaTextarea>

      <div v-show="promoType === 'image'">
        <BulmaTextarea
          v-model="promoImageDescription"
          label="Top promo image description"
          help="For blind readers and search engines"
        ></BulmaTextarea>
        <div class="is-flex mb-2">
          <BulmaField v-slot="{ idForLabel }" label="Desktop Image Width">
            <input
              :id="idForLabel"
              v-model.number="promoDesktopWidth"
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
              v-model.number="promoDesktopHeight"
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
            :files="promoDesktopImages"
            :file-props="fileProps"
            @add="promoDesktopImages.push($event)"
            @remove="promoDesktopImages.splice($event, 1)"
          ></SiteParamsFiles>
        </BulmaField>

        <div class="is-flex mb-2">
          <BulmaField v-slot="{ idForLabel }" label="Mobile Image Width">
            <input
              :id="idForLabel"
              v-model.number="promoMobileWidth"
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
              v-model.number="promoMobileHeight"
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
            :files="promoMobileImages"
            :file-props="fileProps"
            @add="promoMobileImages.push($event)"
            @remove="promoMobileImages.splice($event, 1)"
          ></SiteParamsFiles>
        </BulmaField>
      </div>
    </div>
  </details>
</template>
