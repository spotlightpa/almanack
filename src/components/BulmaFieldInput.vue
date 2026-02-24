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
  inputmode: String,
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
  type: {
    type: String,
    default: "text",
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
    <input
      :id="idForLabel"
      class="input"
      :autocomplete="autocomplete"
      :autofocus="autofocus"
      :inputmode="inputmode"
      :maxlength="maxLength"
      :minlength="minLength"
      :name="name"
      :placeholder="placeholder"
      :required="required"
      :type="type"
      :value="modelValue"
      :readonly="readonly"
      @blur="updateValidationMessage"
      @invalid="updateValidationMessage"
      @input="updateValue"
      @focusout="$emit('focusout', $event)"
    />
  </BulmaField>
</template>
