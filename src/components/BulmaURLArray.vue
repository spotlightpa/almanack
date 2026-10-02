<script setup lang="ts">
import { toAbs, toRel } from "@/utils/link.ts";

const props = defineProps<{
  label?: string;
  labelClass?: string;
  modelValue: string[];
  options?: string[];
  placeholder?: string;
  help?: string;
  validator?: (value: string) => boolean;
  required?: boolean;
  readonly?: boolean;
  relative?: boolean;
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: string[]): void;
}>();

function onUpdate(vals: string[]): void {
  if (props.relative) {
    emit("update:modelValue", vals.map(toRel));
  } else {
    emit("update:modelValue", vals.map(toAbs));
  }
}
</script>

<template>
  <BulmaAutocompleteArray
    :label="props.label"
    :label-class="props.labelClass"
    :model-value="props.modelValue"
    :options="props.options"
    :placeholder="props.placeholder"
    :help="props.help"
    :validator="props.validator"
    :required="props.required"
    :readonly="props.readonly"
    @update:model-value="onUpdate"
  />
</template>
