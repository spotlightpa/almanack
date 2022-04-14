<script>
export default {
  props: { params: Object, fileProps: Object },
};
</script>

<template>
  <div>
    <details class="mt-4">
      <summary class="title is-4">Top Promo</summary>
      <BulmaField label="Top promo">
        <div>
          <label class="checkbox">
            <input v-model="params.promoActive" type="checkbox" />
            Top promo is the native ad-like slot at top of the page
          </label>
        </div>
      </BulmaField>
      <template v-if="params.promoActive">
        <BulmaFieldInput
          v-model="params.promoLink"
          label="Top promo link"
          type="url"
        />
        <BulmaField v-slot="{ idForLabel }" label="Top promo kind">
          <div class="select is-fullwidth">
            <select :id="idForLabel" v-model="params.promoType" class="select">
              <option value="image">Image</option>
              <option value="text">Text</option>
            </select>
          </div>
        </BulmaField>

        <BulmaField
          v-if="params.promoType === 'text'"
          v-slot="{ idForLabel }"
          label="Top promo text"
          help="Text will appear between navbar and page content"
        >
          <textarea
            :id="idForLabel"
            v-model="params.promoText"
            class="textarea"
            rows="2"
          ></textarea>
        </BulmaField>

        <template v-if="params.promoType === 'image'">
          <BulmaField
            v-slot="{ idForLabel }"
            label="Top promo image description"
            help="For blind readers and search engines"
          >
            <textarea
              :id="idForLabel"
              v-model="params.promoImageDescription"
              class="textarea"
              rows="2"
            ></textarea>
          </BulmaField>
          <div class="is-flex mb-2">
            <BulmaField v-slot="{ idForLabel }" label="Desktop Image Width">
              <input
                :id="idForLabel"
                v-model.number="params.promoDesktopWidth"
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
                v-model.number="params.promoDesktopHeight"
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
              :files="params.promoDesktopImages"
              :file-props="fileProps"
              @add="params.promoDesktopImages.push($event)"
              @remove="params.promoDesktopImages.splice($event, 1)"
            />
          </BulmaField>

          <div class="is-flex mb-2">
            <BulmaField v-slot="{ idForLabel }" label="Mobile Image Width">
              <input
                :id="idForLabel"
                v-model.number="params.promoMobileWidth"
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
                v-model.number="params.promoMobileHeight"
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
              :files="params.promoMobileImages"
              :file-props="fileProps"
              @add="params.promoMobileImages.push($event)"
              @remove="params.promoMobileImages.splice($event, 1)"
            />
          </BulmaField>
        </template>
      </template>
    </details>
  </div>
</template>
