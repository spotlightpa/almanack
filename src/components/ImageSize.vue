<script setup>
import { computed, ref } from "vue";
import imageSize from "@/utils/image-size.ts";
import { tryTo, post } from "@/api/client.ts";
import { postImageUpdate } from "@/api/endpoints.ts";
import { makeState } from "@/api/service-util.js";

const props = defineProps({
  path: String,
  width: { type: Number, default: 0 },
  height: { type: Number, default: 0 },
});

const emit = defineEmits(["update"]);

const open = ref(false);

const { apiStateRefs: saveState, exec } = makeState();
const { apiStateRefs: getDimensionsState, exec: getDimensionsExec } =
  makeState();

const hasDimensions = computed(() => props.width > 0 && props.height > 0);
const fetchedWidth = computed(
  () => getDimensionsState.rawData.value.width ?? 0
);
const fetchedHeight = computed(
  () => getDimensionsState.rawData.value.height ?? 0
);
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
  await getDimensionsExec(() => tryTo(imageSize(url.value)));
  await exec(() =>
    post(postImageUpdate, {
      path: props.path,
      set_width: true,
      width: fetchedWidth.value,
      set_height: true,
      height: fetchedHeight.value,
    })
  );
  if (!saveState.error.value) {
    emit("update");
  }
  open.value = true;
}
</script>

<template>
  <div>
    <LinkButton
      :label="hasDimensions ? `${width}\u00d7${height}` : 'Dimensions'"
      :class="{
        'is-loading':
          saveState.isLoadingThrottled.value ||
          getDimensionsState.isLoadingThrottled.value,
      }"
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
