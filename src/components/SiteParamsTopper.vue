<script>
import useProps from "@/utils/use-props.js";
import { toRel, toAbs } from "@/utils/link.js";

export default {
  props: { params: Object, fileProps: Object },
  setup(props) {
    let [data, saveData] = useProps(props.params.data, {
      topperActive: ["topper-active"],
      topperBgColor: ["topper-bg-color"],
      topperDividerColor: ["topper-divider-color"],
      topperLink: ["topper-link", toAbs, toRel],
      topperImageDescription: ["topper-image-description"],
      topperDesktopHeight: ["topper-desktop-height"],
      topperDesktopWidth: ["topper-desktop-width"],
      topperDesktopImages: ["topper-desktop-images"],
      topperMobileHeight: ["topper-mobile-height"],
      topperMobileWidth: ["topper-mobile-width"],
      topperMobileImages: ["topper-mobile-images"],
    });
    return {
      ...data,
      saveData,
    };
  },
};
</script>

<template>
  <div>
    <details class="mt-4">
      <summary class="title is-4">Topper promo</summary>
      <BulmaField
        label="Topper"
        help="Topper is a full width promo at the top of the page under the navbar"
      >
        <div>
          <label class="checkbox">
            <input v-model="topperActive" type="checkbox" />
            Show topper
          </label>
        </div>
      </BulmaField>
      <template v-if="topperActive">
        <BulmaFieldInput v-model="topperLink" label="Topper link" type="url" />

        <BulmaFieldColor
          v-model="topperBgColor"
          label="Topper Background Color"
        />

        <BulmaFieldColor
          v-model="topperDividerColor"
          label="Navbar Divider Color"
          help="If banner is turned off, this will separate the topper from the navbar. Our orange is #ff6c36. Our yellow is #ffcb05. Our dark blue is #009edb. Our light blue is #99d9f1."
        />

        <BulmaTextarea
          v-model="topperImageDescription"
          label="Topper image description"
          help="For blind readers and search engines"
        />
        <div class="is-flex mb-2">
          <BulmaField v-slot="{ idForLabel }" label="Desktop Image Width">
            <input
              :id="idForLabel"
              v-model.number="topperDesktopWidth"
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
              v-model.number="topperDesktopHeight"
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
            :files="topperDesktopImages"
            :file-props="fileProps"
            @add="topperDesktopImages.push($event)"
            @remove="topperDesktopImages.splice($event, 1)"
          />
        </BulmaField>

        <div class="is-flex mb-2">
          <BulmaField v-slot="{ idForLabel }" label="Mobile Image Width">
            <input
              :id="idForLabel"
              v-model.number="topperMobileWidth"
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
              v-model.number="topperMobileHeight"
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
            :files="topperMobileImages"
            :file-props="fileProps"
            @add="topperMobileImages.push($event)"
            @remove="topperMobileImages.splice($event, 1)"
          />
        </BulmaField>
      </template>
    </details>
  </div>
</template>
