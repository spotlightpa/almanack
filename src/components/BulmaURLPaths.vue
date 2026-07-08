<script>
import BulmaAutocompleteArray from "./BulmaAutocompleteArray.vue";

function ensureLeadingSlash(value) {
  if (!value.startsWith("/")) {
    return "/" + value;
  }
  return value;
}

export default {
  components: { BulmaAutocompleteArray },
  emits: ["update:modelValue"],
  props: {
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
  },
  methods: {
    onUpdate(vals) {
      this.$emit("update:modelValue", vals.map(ensureLeadingSlash));
    },
  },
};
</script>

<template>
  <BulmaAutocompleteArray
    :label="label"
    :label-class="labelClass"
    :model-value="modelValue"
    :options="options"
    :placeholder="placeholder"
    :help="help"
    :validator="validator"
    :required="required"
    :readonly="readonly"
    @update:model-value="onUpdate"
  />
</template>
