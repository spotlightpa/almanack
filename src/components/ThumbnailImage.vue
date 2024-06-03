<script setup>
defineProps({
  imageUrl: { type: String, default: "" },
  downloadUrl: { type: String, default: "" },
  description: { type: String, default: "" },
  credit: { type: String, default: "" },
  caption: { type: String, default: "" },
});
</script>

<template>
  <figure>
    <div class="has-margin-bottom">
      <a
        :href="downloadUrl"
        class="is-inline-flex max-256 has-background-grey-lighter"
        target="_blank"
        download
      >
        <img class="max-256" :src="imageUrl" width="256" height="192" />
      </a>
    </div>
    <figcaption>
      <p class="has-margin-bottom">
        <a
          :href="downloadUrl"
          class="button is-danger has-text-weight-semibold"
          target="_blank"
          download
        >
          <span class="icon">
            <font-awesome-icon
              :icon="['fas', 'file-download']"
            ></font-awesome-icon>
          </span>
          <span>Download image</span>
        </a>
      </p>
      <template v-if="description">
        <p class="has-margin-bottom-thin">
          <strong>Description (“alt” text):</strong>
        </p>
        <CopyWithButton
          :value="description"
          label="description"
        ></CopyWithButton>
      </template>
      <template v-if="credit">
        <p class="has-margin-bottom-thin">
          <strong>Credit:</strong>
        </p>
        <CopyWithButton :value="credit" label="credit"></CopyWithButton>
      </template>
      <template v-if="caption">
        <p class="has-margin-bottom-thin">
          <strong>Caption:</strong>
        </p>
        <CopyWithButton :value="caption" label="caption"></CopyWithButton>
      </template>
    </figcaption>
  </figure>
</template>

<style scoped>
.max-256 {
  width: auto;
  height: auto;
  max-height: 256px;
  max-width: 256px;
  min-height: 1rem;
  min-width: 1rem;
}
</style>
