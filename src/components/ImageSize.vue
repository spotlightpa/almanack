<script setup>
import { computed, ref } from "vue";
import imageSize from "@/utils/image-size";

const props = defineProps({
  path: String,
});

const isLoading = ref(false);
const open = ref(false);
const error = ref(null);
const height = ref(0);
const width = ref(0);
const url = computed(
  () => "/ssr/download-image?src=" + encodeURIComponent(props.path)
);

async function onclick() {
  isLoading.value = true;
  try {
    let size = await imageSize(url.value);
    height.value = size.height;
    width.value = size.width;
  } catch (e) {
    error.value = e;
  } finally {
    open.value = true;
    isLoading.value = false;
  }
}
</script>

<template>
  <div>
    <LinkButton
      label="Dimensions"
      :class="{ 'is-loading': isLoading }"
      color="is-success"
      :icon="['fas', 'file-image']"
      @click.prevent="onclick"
    ></LinkButton>
    <BulmaModal v-model="open">
      <div class="box">
        <h3 class="title is-4">{{ path }}</h3>
        <div v-if="error" class="message is-danger">
          <div class="message-header">Could not load image.</div>
          <div class="message-body">Something went wrong.</div>
        </div>
        <template v-else>
          <span class="label">Image width</span>
          <CopyWithButton :value="'' + width" label="Width"></CopyWithButton>
          <span class="label">Image height</span>
          <CopyWithButton :value="'' + height" label="Height"></CopyWithButton>
        </template>
      </div>
    </BulmaModal>
  </div>
</template>
