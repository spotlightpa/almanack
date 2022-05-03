<script>
import useData from "@/utils/use-data.js";
import { toRel, toAbs } from "@/utils/link.js";

export default {
  props: { params: Object, fileProps: Object },
  setup(props, { emit }) {
    return {
      ...useData(emit, props.params.data, {
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
      }),
    };
  },
};
</script>

<template>
  <div>
    <details class="mt-4">
      <summary class="title is-4">Top Promo</summary>
      <BulmaField label="Top promo">
        <div>
          <label class="checkbox">
            <input v-model="promoActive" type="checkbox" />
            Top promo is the native ad-like slot at top of the page
          </label>
        </div>
      </BulmaField>
      <template v-if="promoActive">
        <BulmaFieldInput
          v-model="promoLink"
          label="Top promo link"
          type="url"
        />
        <BulmaField v-slot="{ idForLabel }" label="Top promo kind">
          <div class="select is-fullwidth">
            <select :id="idForLabel" v-model="promoType" class="select">
              <option value="image">Image</option>
              <option value="text">Text</option>
            </select>
          </div>
        </BulmaField>

        <BulmaField
          v-if="promoType === 'text'"
          v-slot="{ idForLabel }"
          label="Top promo text"
          help="Text will appear between navbar and page content"
        >
          <textarea
            :id="idForLabel"
            v-model="promoText"
            class="textarea"
            rows="2"
          ></textarea>
        </BulmaField>

        <template v-if="promoType === 'image'">
          <BulmaField
            v-slot="{ idForLabel }"
            label="Top promo image description"
            help="For blind readers and search engines"
          >
            <textarea
              :id="idForLabel"
              v-model="promoImageDescription"
              class="textarea"
              rows="2"
            ></textarea>
          </BulmaField>
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
            />
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
            />
          </BulmaField>
        </template>
      </template>
    </details>
  </div>
</template>
