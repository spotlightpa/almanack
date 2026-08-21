<script setup>
import { ref } from "vue";

import { post, postPromotion } from "@/api/client-v2.js";
import { makeState } from "@/api/service-util.js";
import { useFileList } from "@/api/file-list.js";

const props = defineProps({
  modelValue: {
    type: Object,
    required: true,
  },
});

const emit = defineEmits(["update:modelValue"]);

const fileList = useFileList();

const isOpen = ref(false);

const name = ref("");
const description = ref("");
const width = ref(0);
const height = ref(0);
const items = ref([]);

function initValues() {
  name.value = props.modelValue.name;
  description.value = props.modelValue.description;
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

async function save() {
  await exec(() =>
    post(postPromotion, {
      id: props.modelValue.id,
      name: name.value,
      description: description.value,
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
          {{ modelValue.name }}
          <span class="has-text-grey is-size-7"> #{{ modelValue.id }}</span>
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
        <button
          v-if="!isOpen"
          class="mt-2 button is-light has-text-weight-semibold"
          type="button"
          @click="toggle"
        >
          Edit
        </button>
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
      <div class="is-flex mb-3" style="gap: 1rem">
        <BulmaField v-slot="{ idForLabel }" label="Width">
          <input
            :id="idForLabel"
            :value="width || ''"
            class="input"
            inputmode="numeric"
            @change="width = +$event.target.value || 0"
          />
        </BulmaField>
        <BulmaField v-slot="{ idForLabel }" label="Height">
          <input
            :id="idForLabel"
            :value="height || ''"
            @change="height = +$event.target.value || 0"
            class="input"
            inputmode="numeric"
          />
        </BulmaField>
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
