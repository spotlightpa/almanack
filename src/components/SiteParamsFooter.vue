<script>
export default {
  props: { params: Object, fileProps: Object },
};
</script>

<template>
  <div>
    <details class="mt-4">
      <summary class="title is-4">Footer promo</summary>
      <BulmaFieldInput
        v-model="params.footerLink"
        label="Footer promo link"
        type="url"
      />

      <BulmaField
        v-slot="{ idForLabel }"
        label="Footer promo image description"
        help="For blind readers and search engines"
      >
        <textarea
          :id="idForLabel"
          v-model="params.footerDescription"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>
      <div class="is-flex mb-2">
        <BulmaField v-slot="{ idForLabel }" label="Desktop Image Width">
          <input
            :id="idForLabel"
            v-model.number="params.footerDesktopWidth"
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
            v-model.number="params.footerDesktopHeight"
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
          :files="params.footerDesktopImages"
          :file-props="fileProps"
          @add="params.footerDesktopImages.push($event)"
          @remove="params.footerDesktopImages.splice($event, 1)"
        />
      </BulmaField>

      <div class="is-flex mb-2">
        <BulmaField v-slot="{ idForLabel }" label="Mobile Image Width">
          <input
            :id="idForLabel"
            v-model.number="params.footerMobileWidth"
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
            v-model.number="params.footerMobileHeight"
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
          :files="params.footerMobileImages"
          :file-props="fileProps"
          @add="params.footerMobileImages.push($event)"
          @remove="params.footerMobileImages.splice($event, 1)"
        />
      </BulmaField>
    </details>
  </div>
</template>
