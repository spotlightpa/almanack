<script>
export default {
  props: { params: Object, fileProps: Object },
};
</script>

<template>
  <div>
    <details class="mt-4">
      <summary class="title is-4">Sticky popup</summary>
      <BulmaField label="Sticky" help="Pop up is a bottom-right sticky slider">
        <div>
          <label class="checkbox">
            <input v-model="params.stickyActive" type="checkbox" />
            Show corner sticky to all visitors
          </label>
        </div>
      </BulmaField>
      <template v-if="params.stickyActive">
        <BulmaFieldInput
          v-model="params.stickyLink"
          label="Sticky link"
          type="url"
        />
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
    </details>
  </div>
</template>
