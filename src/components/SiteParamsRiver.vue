<script>
export default {
  props: { params: Object, fileProps: Object },
};
</script>

<template>
  <div>
    <details class="mt-4">
      <summary class="title is-4">Homepage river</summary>
      <BulmaFieldInput
        v-model="params.riverLink"
        label="Homepage river promo link"
        type="url"
      />
      <BulmaField
        v-slot="{ idForLabel }"
        label="Homepage river promo image description"
        help="For blind readers and search engines"
      >
        <textarea
          :id="idForLabel"
          v-model="params.riverDescription"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>
      <div class="is-flex mb-2">
        <BulmaField v-slot="{ idForLabel }" label="Desktop Image Width">
          <input
            :id="idForLabel"
            v-model.number="params.riverDesktopWidth"
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
            v-model.number="params.riverDesktopHeight"
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
          :files="params.riverDesktopImages"
          :file-props="fileProps"
          @add="params.riverDesktopImages.push($event)"
          @remove="params.riverDesktopImages.splice($event, 1)"
        />
      </BulmaField>

      <div class="is-flex mb-2">
        <BulmaField v-slot="{ idForLabel }" label="Mobile Image Width">
          <input
            :id="idForLabel"
            v-model.number="params.riverMobileWidth"
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
            v-model.number="params.riverMobileHeight"
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
          :files="params.riverMobileImages"
          :file-props="fileProps"
          @add="params.riverMobileImages.push($event)"
          @remove="params.riverMobileImages.splice($event, 1)"
        />
      </BulmaField>
    </details>
  </div>
</template>
