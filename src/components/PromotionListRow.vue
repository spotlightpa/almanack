<script setup>
import { ref, watch } from "vue";

import { post, postPromotion, deletePromotion } from "@/api/client.ts";
import { makePromotion } from "@/api/promotion.ts";
import { formatDateTime } from "@/utils/time-format.ts";
import { makeState } from "@/api/service-util.js";
import { useFileList } from "@/api/file-list.js";

const props = defineProps({
  modelValue: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(["update:modelValue", "delete"]);

const fileList = useFileList();

const isOpen = ref(false);

const promo = makePromotion();

watch(
  () => props.modelValue,
  (val) => promo.init(val),
  { immediate: true }
);

function initValues() {
  promo.init(props.modelValue);
}

function toggle() {
  isOpen.value = !isOpen.value;
}

const { exec, apiStateRefs } = makeState();
const { isLoadingThrottled, error } = apiStateRefs;

const { exec: deleteExec, apiStateRefs: deleteStateRefs } = makeState();
const { isLoadingThrottled: deleteLoadingThrottled, error: deleteError } =
  deleteStateRefs;

async function remove() {
  if (!confirm(`Delete "${props.modelValue.name || "this promotion"}"?`)) {
    return;
  }
  await deleteExec(() => post(deletePromotion, { id: promo.id }));
  if (!deleteStateRefs.error.value) {
    emit("delete");
  }
}

async function save() {
  await exec(() => post(postPromotion, promo.toJSON()));
  if (!apiStateRefs.error.value) {
    emit("update:modelValue", apiStateRefs.rawData.value);
    isOpen.value = false;
  }
}
</script>

<template>
  <div>
    <div class="is-flex is-justify-content-space-between is-align-items-start">
      <div>
        <p class="has-text-weight-semibold">
          <span v-if="modelValue.name">{{ modelValue.name }}</span>
          <span v-else class="has-text-grey is-italic">&lt;untitled&gt;</span>
        </p>
        <p v-if="modelValue.description" class="is-size-7 has-text-grey">
          {{ modelValue.description }}
        </p>
        <p class="is-size-7 has-text-grey">
          <template v-if="modelValue.width || modelValue.height">
            {{ modelValue.width }}×{{ modelValue.height }}px &middot;
          </template>
          {{ modelValue.image_urls?.length ?? 0 }} image{{
            modelValue.image_urls?.length !== 1 ? "s" : ""
          }}
          &middot;
          {{ formatDateTime(promo.updatedAt.value) }}
        </p>
        <div class="mt-2 buttons">
          <button
            class="button is-light has-text-weight-semibold"
            type="button"
            @click="toggle"
            :disabled="isOpen"
          >
            Edit
          </button>
          <button
            class="button is-danger is-light has-text-weight-semibold"
            :class="{ 'is-loading': deleteLoadingThrottled }"
            :disabled="deleteLoadingThrottled || null"
            type="button"
            @click="remove"
          >
            Delete
          </button>
        </div>
        <ErrorSimple :error="deleteError" />
      </div>
    </div>

    <div v-if="isOpen" class="mt-3">
      <PromotionForm :promo="promo" :file-list="fileList" />
      <ErrorSimple :error="error" />
      <div class="buttons">
        <button
          class="button is-success has-text-weight-semibold"
          :class="{ 'is-loading': isLoadingThrottled }"
          type="button"
          @click="save"
        >
          Save
        </button>
        <button
          class="button is-light has-text-weight-semibold"
          :disabled="isLoadingThrottled || null"
          type="button"
          @click="initValues"
        >
          Discard Changes
        </button>
        <button
          class="button is-light has-text-weight-semibold"
          type="button"
          @click="toggle"
        >
          Close
        </button>
      </div>
    </div>
  </div>
</template>
