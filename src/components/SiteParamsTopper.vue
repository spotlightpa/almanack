<script>
export default {
  props: { params: Object, fileProps: Object },
};
</script>

<template>
  <div>
    <details class="mt-4">
      <summary class="title is-4">Topper</summary>
      <BulmaField
        label="Topper"
        help="Topper is a full width promo at the top of the page under the navbar"
      >
        <div>
          <label class="checkbox">
            <input v-model="params.topperActive" type="checkbox" />
            Show topper
          </label>
        </div>
      </BulmaField>
      <template v-if="params.topperActive">
        <BulmaFieldInput
          v-model="params.topperLink"
          label="Topper link"
          type="url"
        />

        <BulmaField v-slot="{ idForLabel }" label="Topper Background Color">
          <div class="is-flex is-align-items-center">
            <input
              :id="idForLabel"
              v-model="params.topperBgColor"
              type="color"
            />
            <span class="ml-4 is-flex-grow-0">
              <input
                v-model="params.topperBgColor"
                type="text"
                class="input is-small"
              />
            </span>
          </div>
        </BulmaField>

        <BulmaField
          v-slot="{ idForLabel }"
          label="Navbar Divider Color"
          help="If banner is turned off, this will separate the topper from the navbar. Our orange is #ff6c36. Our yellow is #ffcb05. Our dark blue is #009edb. Our light blue is #99d9f1."
        >
          <div class="is-flex is-align-items-center">
            <input
              :id="idForLabel"
              v-model="params.topperDividerColor"
              type="color"
            />
            <span class="ml-4 is-flex-grow-0">
              <input
                v-model="params.topperDividerColor"
                type="text"
                class="input is-small"
              />
            </span>
          </div>
        </BulmaField>

        <BulmaField
          v-slot="{ idForLabel }"
          label="Topper image description"
          help="For blind readers and search engines"
        >
          <textarea
            :id="idForLabel"
            v-model="params.topperImageDescription"
            class="textarea"
            rows="2"
          ></textarea>
        </BulmaField>
        <div class="is-flex mb-2">
          <BulmaField v-slot="{ idForLabel }" label="Desktop Image Width">
            <input
              :id="idForLabel"
              v-model.number="params.topperDesktopWidth"
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
              v-model.number="params.topperDesktopHeight"
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
            :files="params.topperDesktopImages"
            :file-props="fileProps"
            @add="params.topperDesktopImages.push($event)"
            @remove="params.topperDesktopImages.splice($event, 1)"
          />
        </BulmaField>

        <div class="is-flex mb-2">
          <BulmaField v-slot="{ idForLabel }" label="Mobile Image Width">
            <input
              :id="idForLabel"
              v-model.number="params.topperMobileWidth"
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
              v-model.number="params.topperMobileHeight"
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
            :files="params.topperMobileImages"
            :file-props="fileProps"
            @add="params.topperMobileImages.push($event)"
            @remove="params.topperMobileImages.splice($event, 1)"
          />
        </BulmaField>
      </template>
    </details>
  </div>
</template>
