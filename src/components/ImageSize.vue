<script setup>
import { computed, ref } from "vue";
import imageSize from "@/utils/image-size.ts";
import { post } from "@/api/client.ts";
import { postImageUpdate } from "@/api/endpoints.ts";
import { makeState } from "@/api/service-util.js";

const props = defineProps({
  path: String,
  width: { type: Number, default: 0 },
  height: { type: Number, default: 0 },
});

const emit = defineEmits(["update:width", "update:height"]);

const open = ref(false);
const fetchedWidth = ref(0);
const fetchedHeight = ref(0);

const { apiStateRefs: saveState, exec } = makeState();

const hasDimensions = computed(() => props.width > 0 && props.height > 0);
const displayWidth = computed(() =>
  hasDimensions.value ? props.width : fetchedWidth.value
);
const displayHeight = computed(() =>
  hasDimensions.value ? props.height : fetchedHeight.value
);

const url = computed(
  () => "/ssr/download-image?src=" + encodeURIComponent(props.path)
);

async function onclick() {
  if (hasDimensions.value) {
    open.value = true;
    return;
  }
  const size = await imageSize(url.value);
  fetchedWidth.value = size.width;
  fetchedHeight.value = size.height;
  await exec(() =>
    post(postImageUpdate, {
      path: props.path,
      set_width: true,
      width: size.width,
      set_height: true,
      height: size.height,
    })
  );
  if (!saveState.error.value) {
    emit("update:width", size.width);
    emit("update:height", size.height);
  }
  open.value = true;
}
</script>

<template>
  <div>
    <LinkButton
      :label="hasDimensions ? `${width}\u00d7${height}` : 'Dimensions'"
      :class="{ 'is-loading': saveState.isLoadingThrottled.value }"
      color="is-success"
      :icon="['fas', 'file-image']"
      @click.prevent="onclick"
    ></LinkButton>
    <BulmaModal v-model="open">
      <div class="box">
        <h3 class="title is-4">{{ path }}</h3>
        <ErrorSimple :error="saveState.error.value"></ErrorSimple>
        <template v-if="!saveState.error.value">
          <span class="label">Image width</span>
          <CopyWithButton
            :value="'' + displayWidth"
            label="Width"
          ></CopyWithButton>
          <span class="label">Image height</span>
          <CopyWithButton
            :value="'' + displayHeight"
            label="Height"
          ></CopyWithButton>
        </template>
      </div>
    </BulmaModal>
  </div>
</template>
