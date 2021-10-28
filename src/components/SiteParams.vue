<script>
import sanitizeText from "@/utils/sanitize-text.js";

export default {
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
    <details class="mt-4">
      <summary class="title is-4">Banner</summary>
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
            <input
              :id="idForLabel"
              v-model="params.bannerBgColor"
              type="color"
            />
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
    </details>
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
    <details class="mt-4">
      <summary class="title is-4">Sticky</summary>
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
    </details>
    <details class="mt-4">
      <summary class="title is-4">Homepage sidebar top</summary>
      <BulmaFieldInput
        v-model="params.sidebarTopLink"
        label="Homepage sidebar top promo link"
        type="url"
      />
      <BulmaField
        v-slot="{ idForLabel }"
        label="Homepage siderbar top promo image description"
        help="For blind readers and search engines"
      >
        <textarea
          :id="idForLabel"
          v-model="params.sidebarTopDescription"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>
      <div class="is-flex mb-2">
        <BulmaField v-slot="{ idForLabel }" label="Image Width">
          <input
            :id="idForLabel"
            v-model.number="params.sidebarTopWidth"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
        <BulmaField v-slot="{ idForLabel }" class="ml-2" label="Image Height">
          <input
            :id="idForLabel"
            v-model.number="params.sidebarTopHeight"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
      </div>
      <BulmaField
        label="Images"
        help="If multiple images are provided, each page load will select one randomly"
      >
        <SiteParamsFiles
          :files="params.sidebarTopImages"
          :file-props="fileProps"
          @add="params.sidebarTopImages.push($event)"
          @remove="params.sidebarTopImages.splice($event, 1)"
        />
      </BulmaField>
    </details>
    <details class="mt-4">
      <summary class="title is-4">Homepage sidebar sticky</summary>
      <BulmaFieldInput
        v-model="params.sidebarStickyLink"
        label="Homepage sidebar sticky bottom promo link"
        type="url"
      />
      <BulmaField
        v-slot="{ idForLabel }"
        label="Homepage siderbar sticky bottom promo image description"
        help="For blind readers and search engines"
      >
        <textarea
          :id="idForLabel"
          v-model="params.sidebarStickyDescription"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>
      <div class="is-flex mb-2">
        <BulmaField v-slot="{ idForLabel }" label="Image Width">
          <input
            :id="idForLabel"
            v-model.number="params.sidebarStickyWidth"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
        <BulmaField v-slot="{ idForLabel }" class="ml-2" label="Image Height">
          <input
            :id="idForLabel"
            v-model.number="params.sidebarStickyHeight"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
      </div>
      <BulmaField
        label="Images"
        help="If multiple images are provided, each page load will select one randomly"
      >
        <SiteParamsFiles
          :files="params.sidebarStickyImages"
          :file-props="fileProps"
          @add="params.sidebarStickyImages.push($event)"
          @remove="params.sidebarStickyImages.splice($event, 1)"
        />
      </BulmaField>
    </details>
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
    <details class="mt-4">
      <summary class="title is-4">Support Us Box</summary>
      <BulmaFieldInput v-model="params.supportHed" label="Support Us Box hed" />
      <BulmaField
        v-slot="{ idForLabel }"
        label="Support Us Box text"
        help="Supports bold and italics tags"
      >
        <textarea
          :id="idForLabel"
          v-model="params.supportText"
          class="textarea"
          rows="2"
        ></textarea>
      </BulmaField>
      <BulmaFieldInput
        v-model="params.supportCTA"
        label="Support Us Box call to action"
      />

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
    </details>
  </div>
</template>
