<script>
export default {
  props: { params: Object, fileProps: Object },
};
</script>

<template>
  <div>
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
