<script>
export default {
  props: { params: Object, fileProps: Object },
};
</script>

<template>
  <div>
    <details class="mt-4">
      <summary class="title is-4">Support Us Box</summary>
      <BulmaField label="Support Us Box">
        <div>
          <label class="checkbox">
            <input v-model="params.supportActive" type="checkbox" />
            Show support box in footer of articles
          </label>
        </div>
      </BulmaField>
      <template v-if="params.supportActive">
        <BulmaFieldInput
          v-model="params.supportLink"
          label="Support box link"
          type="url"
        />
        <BulmaFieldInput
          v-model="params.supportHed"
          label="Support Us Box hed"
        />
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
        <p class="label">Some common colors</p>
        <p class="content">
          White is #ffffff. Black is #000000. Our blue is #009edb. Our light
          blue is #99d9f1. Our periwinkle is #e5f6ff. Our dark blue is #22416e.
          Our CTA orange is #ff6c36. Our green is #78bc20. Our yellow is
          #ffcb05. Our goldenrod is #fff1bd. Our beige is #f4f1ee.
        </p>
        <BulmaFieldColor
          v-model="params.supportHedColor"
          label="Support Hed Color"
        />
        <BulmaFieldColor
          v-model="params.supportTextColor"
          label="Support Text Color"
        />
        <BulmaFieldColor
          v-model="params.supportBgColor"
          label="Support Background Color"
        />
        <BulmaFieldColor
          v-model="params.supportButtonBgColor"
          label="Support Button Background Color"
        />
        <BulmaFieldColor
          v-model="params.supportButtonTextColor"
          label="Support Button Text Color"
        />

        <BulmaField label="Support Box Preview">
          <div class="support-wrapper">
            <div
              class="support-content"
              :style="{ 'background-color': params.supportBgColor }"
            >
              <h2
                class="support-header"
                :style="{ color: params.supportHedColor }"
                v-text="params.supportHed"
              ></h2>
              <p
                v-for="(p, i) of params.supportHTML.split('\n')"
                :key="i"
                class="support-p"
                :style="{ color: params.supportTextColor }"
                v-html="p"
              ></p>
              <div class="mt-6 has-text-centered">
                <span
                  class="button has-text-weight-semibold is-uppercase"
                  :style="{
                    color: params.supportButtonTextColor,
                    backgroundColor: params.supportButtonBgColor,
                    borderColor: params.supportButtonBgColor,
                  }"
                  v-text="params.supportCTA"
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
    font-family: system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI",
      Roboto, "Helvetica Neue", Arial, "Noto Sans", "Apple Color Emoji",
      "Segoe UI Emoji", "Segoe UI Symbol", "Noto Color Emoji", sans-serif;
  }
}

.support-content {
  background-color: rgba(255, 241, 189, 1);
  border-radius: 0.25rem;
  box-shadow: 0 1px 3px 0 rgba(0, 0, 0, 0.1), 0 1px 2px 0 rgba(0, 0, 0, 0.06);
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
