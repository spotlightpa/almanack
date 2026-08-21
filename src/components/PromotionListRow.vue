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

const name = ref("");
const description = ref("");
const link = ref("");
const width = ref(0);
const height = ref(0);
const items = ref([]);

function initValues() {
  name.value = props.modelValue.name;
  description.value = props.modelValue.description;
  link.value = props.modelValue.link ?? "";
  width.value = props.modelValue.width;
  height.value = props.modelValue.height;
  items.value = [...(props.modelValue.items ?? [])];
}

function toggle() {
  if (isOpen.value) {
    isOpen.value = false;
  } else {
    initValues();
    isOpen.value = true;
  }
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
  await deleteExec(() =>
    post(deletePromotion, { id: props.modelValue.id })
  );
  if (!deleteStateRefs.error.value) {
    emit("delete");
  }
}

async function save() {
  await exec(() =>
    post(postPromotion, {
      id: props.modelValue.id,
      name: name.value,
      description: description.value,
      link: link.value,
      width: width.value,
      height: height.value,
      items: items.value,
    })
  );
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
          {{ modelValue.items?.length ?? 0 }} image{{
            modelValue.items?.length !== 1 ? "s" : ""
          }}
          &middot;
          {{ new Date(modelValue.updated_at).toLocaleString() }}
        </p>
        <div class="mt-2 buttons">
          <button
            v-if="!isOpen"
            class="button is-light has-text-weight-semibold"
            type="button"
            @click="toggle"
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
      <BulmaFieldInput
        v-model="name"
        label="Name"
        placeholder="e.g. Rail sticky promo"
      />
      <BulmaFieldInput
        v-model="description"
        label="Description"
        placeholder="Short description"
      />
      <BulmaFieldInput
        v-model="link"
        label="Link URL"
        type="url"
        placeholder="https://"
      />
      <div class="is-flex mb-3" style="gap: 1rem">
        <BulmaFieldInput
          label="Width"
          inputmode="numeric"
          :model-value="width || ''"
          @update:model-value="width = +$event || 0"
        />
        <BulmaFieldInput
          label="Height"
          inputmode="numeric"
          :model-value="height || ''"
          @update:model-value="height = +$event || 0"
        />
      </div>
      <BulmaField
        label="Image URLs"
        help="One image URL per slot; one will be chosen randomly on each page load."
      >
        <SiteParamsFiles
          :files="items"
          :file-props="fileList"
          @add="items.push($event)"
          @remove="items.splice($event, 1)"
        />
      </BulmaField>
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
