<script setup>
import { ref, computed } from "vue";

const props = defineProps({
  label: String,
  labelClass: {
    type: String,
    default: "label",
  },
  modelValue: String,
  placeholder: String,
  help: String,
  validator: Function,
  name: String,
  minLength: {
    type: Number,
    default: null,
  },
  maxLength: {
    type: Number,
    default: null,
  },
  required: {
    type: Boolean,
    default: false,
  },
  autofocus: {
    type: Boolean,
    default: false,
  },
  readonly: {
    type: Boolean,
    default: false,
  },
  autocomplete: {
    type: String,
    default: null,
  },
  rows: {
    type: Number,
    default: 2,
  },
  cols: {
    type: Number,
    default: null,
  },
});

const emit = defineEmits(["update:modelValue", "focusout"]);

const validationMessage = ref("");

const fieldProps = computed(() => ({
  label: props.label,
  help: props.help,
  labelClass: props.labelClass,
  required: props.required,
  validationMessage: validationMessage.value,
}));

function updateValidationMessage(ev) {
  validationMessage.value = ev.target.validationMessage;
}

function updateValue(ev) {
  let newVal = ev.target.value;
  if (props.maxLength && newVal.length > props.maxLength) {
    newVal = newVal.slice(0, props.maxLength);
  }
  updateValidationMessage(ev);
  emit("update:modelValue", newVal);
}
</script>

<template>
  <BulmaField v-slot="{ idForLabel }" v-bind="fieldProps">
    <textarea
      :id="idForLabel"
      class="textarea"
      :autocomplete="autocomplete"
      :autofocus="autofocus"
      :maxlength="maxLength"
      :minlength="minLength"
      :name="name"
      :placeholder="placeholder"
      :required="required || null"
      :rows="rows"
      :cols="cols"
      :value="modelValue"
      :readonly="readonly || null"
      @blur="updateValidationMessage"
      @invalid="updateValidationMessage"
      @input="updateValue"
      @focusout="$emit('focusout', $event)"
    ></textarea>
  </BulmaField>
</template>
