<script>
export default {
  props: { params: Object, fileProps: Object },
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
        <BulmaFieldColor
          v-model="params.bannerTextColor"
          label="Banner Text Color"
        />
        <BulmaFieldColor
          v-model="params.bannerBgColor"
          label="Banner Background Color"
          help="Our orange is #ff6c36. Our yellow is #ffcb05. Our dark blue is #009edb. Our light blue is #99d9f1."
        />
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
  </div>
</template>
