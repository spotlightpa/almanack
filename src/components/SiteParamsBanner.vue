<script>
import useData from "@/utils/use-data.js";
import sanitizeText from "@/utils/sanitize-text.js";
import { toRel, toAbs } from "@/utils/link.js";

export default {
  props: { params: Object, fileProps: Object },
  setup(props, { emit }) {
    return {
      ...useData(emit, props.params.data, {
        bannerActive: ["banner-active"],
        bannerText: ["banner", undefined, sanitizeText],
        bannerLink: ["banner-link", (v) => toAbs(v), (v) => toRel(v)],
        bannerBgColor: ["banner-bg-color"],
        bannerTextColor: ["banner-text-color"],
      }),
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
            <input v-model="bannerActive" type="checkbox" />
            Show sitewide alert banner
          </label>
        </div>
      </BulmaField>
      <template v-if="bannerActive">
        <BulmaField
          v-slot="{ idForLabel }"
          label="Banner Text"
          help="Supports bold and italics tags"
        >
          <textarea
            :id="idForLabel"
            v-model="bannerText"
            class="textarea"
            rows="2"
          ></textarea>
        </BulmaField>

        <BulmaFieldInput v-model="bannerLink" label="Banner link" type="url" />
        <BulmaFieldColor v-model="bannerTextColor" label="Banner Text Color" />
        <BulmaFieldColor
          v-model="bannerBgColor"
          label="Banner Background Color"
          help="Our orange is #ff6c36. Our yellow is #ffcb05. Our dark blue is #009edb. Our light blue is #99d9f1."
        />
        <BulmaField label="Banner Preview">
          <div
            class="has-radius-padding"
            :style="{ 'background-color': bannerBgColor }"
          >
            <a :href="params.data['banner-link']" target="_blank">
              <div
                class="is-size-3-fullscreen is-size-4 has-text-centered"
                :style="{ color: bannerTextColor }"
                v-html="params.data['banner']"
              ></div>
            </a>
          </div>
        </BulmaField>
      </template>
    </details>
  </div>
</template>
