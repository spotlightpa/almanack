<script>
import { computed } from "vue";

import useProps from "@/utils/use-props.js";
import { toRel, toAbs } from "@/utils/link.js";
import sanitizeText from "@/utils/sanitize-text.js";

function cleanText(text) {
  return text
    .split("\n")
    .map((p) => sanitizeText(p))
    .join("\n");
}

export default {
  props: { params: Object, fileProps: Object },
  setup(props) {
    let [data, saveData] = useProps(props.params.data, {
      supportActive: ["support-active"],
      supportLink: ["support-link", toAbs, toRel],
      supportHed: ["support-hed"],
      supportText: ["support-text", undefined, cleanText],
      supportCTA: ["support-cta"],
      supportHedColor: ["support-hed-color"],
      supportTextColor: ["support-text-color"],
      supportBgColor: ["support-bg-color"],
      supportButtonBgColor: ["support-button-bg-color"],
      supportButtonTextColor: ["support-button-text-color"],
    });
    return {
      ...data,
      saveData,
      paragraphs: computed(() => {
        return data.supportText.value.split("\n");
      }),
    };
  },
};
</script>

<template>
  <div>
    <details class="mt-4">
      <summary class="title is-4">Support us box</summary>
      <BulmaField label="Support Us Box">
        <div>
          <label class="checkbox">
            <input v-model="supportActive" type="checkbox" />
            Show support box in footer of articles
          </label>
        </div>
      </BulmaField>
      <template v-if="supportActive">
        <BulmaFieldInput
          v-model="supportLink"
          label="Support box link"
          type="url"
        />
        <BulmaFieldInput v-model="supportHed" label="Support Us Box hed" />
        <BulmaTextarea
          v-model="supportText"
          label="Support Us Box text"
          help="Supports bold and italics tags"
        />
        <BulmaFieldInput
          v-model="supportCTA"
          label="Support Us Box call to action"
        />
        <p class="label">Some common colors</p>
        <p class="content">
          White is #ffffff. Black is #000000. Our blue is #009edb. Our light
          blue is #99d9f1. Our periwinkle is #e5f6ff. Our dark blue is #22416e.
          Our CTA orange is #ff6c36. Our green is #78bc20. Our yellow is
          #ffcb05. Our goldenrod is #fff1bd. Our beige is #f4f1ee.
        </p>
        <BulmaFieldColor v-model="supportHedColor" label="Support Hed Color" />
        <BulmaFieldColor
          v-model="supportTextColor"
          label="Support Text Color"
        />
        <BulmaFieldColor
          v-model="supportBgColor"
          label="Support Background Color"
        />
        <BulmaFieldColor
          v-model="supportButtonBgColor"
          label="Support Button Background Color"
        />
        <BulmaFieldColor
          v-model="supportButtonTextColor"
          label="Support Button Text Color"
        />

        <BulmaField label="Support Box Preview">
          <div class="support-wrapper">
            <div
              class="support-content"
              :style="{ 'background-color': supportBgColor }"
            >
              <h2
                class="support-header"
                :style="{ color: supportHedColor }"
                v-text="supportHed"
              ></h2>
              <p
                v-for="(p, i) of paragraphs"
                :key="i"
                class="support-p"
                :style="{ color: supportTextColor }"
                v-html="p"
              ></p>
              <div class="mt-6 has-text-centered">
                <span
                  class="button has-text-weight-semibold is-uppercase"
                  :style="{
                    color: supportButtonTextColor,
                    backgroundColor: supportButtonBgColor,
                    borderColor: supportButtonBgColor,
                  }"
                  v-text="supportCTA"
                >
                </span>
              </div>
            </div>
          </div>
        </BulmaField>
      </template>
    </details>
  </div>
</template>

<style>
@media (min-width: 850px) {
  .support-wrapper {
    font-size: 1.2rem;
    font-family:
      system-ui,
      -apple-system,
      BlinkMacSystemFont,
      "Segoe UI",
      Roboto,
      "Helvetica Neue",
      Arial,
      "Noto Sans",
      "Apple Color Emoji",
      "Segoe UI Emoji",
      "Segoe UI Symbol",
      "Noto Color Emoji",
      sans-serif;
  }
}

.support-content {
  background-color: rgba(255, 241, 189, 1);
  border-radius: 0.25rem;
  box-shadow:
    0 1px 3px 0 rgba(0, 0, 0, 0.1),
    0 1px 2px 0 rgba(0, 0, 0, 0.06);
  color: #000;
  margin-left: auto;
  margin-right: auto;
  max-width: 65ch;
  padding-bottom: 2rem;
  padding-left: 1rem;
  padding-right: 1rem;
  padding-top: 2rem;
}

@media (min-width: 640px) {
  .support-content {
    padding-left: 2rem;
    padding-right: 2rem;
    padding-top: 3rem;
    padding-bottom: 3rem;
  }
}

.support-header {
  font-weight: 700;
  font-size: 1.3rem;
}

@media (min-width: 640px) {
  .support-header {
    font-size: 1.4rem;
  }
}

.support-p {
  line-height: 1.375;
  font-weight: 400;
  margin-top: 1.5rem;
}

.support-p strong,
.support-p b,
.support-p i,
.support-p em {
  color: inherit;
}
</style>
