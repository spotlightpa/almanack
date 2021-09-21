<script>
import sanitizeText from "@/utils/sanitize-text.js";

import BulmaField from "./BulmaField.vue";
import BulmaFieldInput from "./BulmaFieldInput.vue";
import SiteParamsFiles from "./SiteParamsFiles.vue";

export default {
  components: {
    BulmaField,
    BulmaFieldInput,
    SiteParamsFiles,
  },
  props: { params: Object, fileProps: Object },
  setup() {
    return {
      sanitizeText,
    };
  },
};
</script>

<template>
  <div>
    <BulmaField label="Banner">
      <div>
        <label class="checkbox">
          <input v-model="params.bannerActive" type="checkbox" />
          Show sitewide alert banner
        </label>
      </div>
    </BulmaField>
    <template v-if="params.bannerActive">
      <BulmaField
        v-slot="{ idForLabel }"
        label="Banner Text"
        help="Supports bold and italics tags"
      >
        <textarea
          :id="idForLabel"
          v-model="params.bannerText"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>
      <BulmaFieldInput
        v-model="params.bannerLink"
        label="Banner link"
        type="url"
      />
      <BulmaField v-slot="{ idForLabel }" label="Banner Text Color">
        <div class="is-flex is-align-items-center">
          <input
            :id="idForLabel"
            v-model="params.bannerTextColor"
            type="color"
          />
          <span class="ml-4 is-flex-grow-0">
            <input
              v-model="params.bannerTextColor"
              type="text"
              class="input is-small"
            />
          </span>
        </div>
      </BulmaField>
      <BulmaField
        v-slot="{ idForLabel }"
        label="Banner Background Color"
        help="Our orange is #ff6c36. Our yellow is #ffcb05. Our dark blue is #009edb. Our light blue is #99d9f1."
      >
        <div class="is-flex is-align-items-center">
          <input :id="idForLabel" v-model="params.bannerBgColor" type="color" />
          <span class="ml-4 is-flex-grow-0">
            <input
              v-model="params.bannerBgColor"
              type="text"
              class="input is-small"
            />
          </span>
        </div>
      </BulmaField>
      <BulmaField label="Banner Preview">
        <div
          class="has-radius-padding"
          :style="{ 'background-color': params.bannerBgColor }"
        >
          <a :href="params.bannerLink" target="_blank">
            <div
              class="is-size-3-fullscreen is-size-4 has-text-centered"
              :style="{ color: params.bannerTextColor }"
              v-html="params.bannerHTML"
            ></div>
          </a>
        </div>
      </BulmaField>
    </template>

    <BulmaField label="Promo">
      <div>
        <label class="checkbox">
          <input v-model="params.promoActive" type="checkbox" />
          Promo is the native ad-like slot at top of the page
        </label>
      </div>
    </BulmaField>
    <template v-if="params.promoActive">
      <BulmaFieldInput
        v-model="params.promoLink"
        label="Promo link"
        type="url"
      />
      <BulmaField v-slot="{ idForLabel }" label="Promo kind">
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
        label="Promo text"
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
          label="Promo image description"
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
    <BulmaField label="Sticky" help="Pop up is a bottom-right sticky slider">
      <div>
        <label class="checkbox">
          <input v-model="params.stickyActive" type="checkbox" />
          Show corner sticky to all visitors
        </label>
      </div>
    </BulmaField>
    <template v-if="params.stickyActive">
      <BulmaField
        v-slot="{ idForLabel }"
        label="Sticky image description"
        help="For blind readers and search engines"
      >
        <textarea
          :id="idForLabel"
          v-model="params.stickyImageDescription"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>
      <BulmaField
        label="Sticky images"
        help="If multiple images are provided, each page load will select one randomly"
      >
        <SiteParamsFiles
          :files="params.stickyImages"
          :file-props="fileProps"
          @add="params.stickyImages.push($event)"
          @remove="params.stickyImages.splice($event, 1)"
        />
      </BulmaField>
    </template>
    <BulmaField
      label="Newsletter"
      help="Pop up is full screen newsletter takeover"
    >
      <div>
        <label class="checkbox">
          <input v-model="params.newsletterActive" type="checkbox" />
          Show newsletter pop up
        </label>
      </div>
    </BulmaField>
  </div>
</template>
