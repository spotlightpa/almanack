<script setup lang="ts">
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
}>();

const emit = defineEmits<{
  (e: "update:modelValue", value: string[]): void;
}>();

function ensureLeadingSlash(value: string): string {
  return value.startsWith("/") ? value : "/" + value;
}

function onUpdate(vals: string[]): void {
  emit("update:modelValue", vals.map(ensureLeadingSlash));
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
