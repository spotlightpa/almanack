<script setup>
import { ref } from "vue";

import { post, postPromotion, deletePromotion } from "@/api/client-v2.js";
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

// modelValue is already a Promotion; use it directly as the edit form.
const promo = props.modelValue;

function initValues() {
  promo.init(promo.toJSON());
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
  if (!confirm(`Delete "${promo.name.value || "this promotion"}"?`)) {
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
          <span v-if="promo.name.value">{{ promo.name.value }}</span>
          <span v-else class="has-text-grey is-italic">&lt;untitled&gt;</span>
        </p>
        <p v-if="promo.description.value" class="is-size-7 has-text-grey">
          {{ promo.description.value }}
        </p>
        <p class="is-size-7 has-text-grey">
          <template v-if="promo.width.value || promo.height.value">
            {{ promo.width.value }}×{{ promo.height.value }}px &middot;
          </template>
          {{ promo.imageUrls.value.length }} image{{
            promo.imageUrls.value.length !== 1 ? "s" : ""
          }}
          &middot;
          {{ new Date(promo.updatedAt.value).toLocaleString() }}
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
