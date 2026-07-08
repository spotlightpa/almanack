<script setup>
import BulmaAutocompleteArray from "./BulmaAutocompleteArray.vue";

const props = defineProps({
  label: String,
  labelClass: {
    type: String,
    default: "label",
  },
  modelValue: Array,
  options: Array,
  placeholder: String,
  help: String,
  validator: Function,
  required: {
    type: Boolean,
    default: false,
  },
  readonly: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["update:modelValue"]);

function ensureLeadingSlash(value) {
  return value.startsWith("/") ? value : "/" + value;
}

function onUpdate(vals) {
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
