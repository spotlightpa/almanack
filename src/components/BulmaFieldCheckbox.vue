<script setup>
import { computed } from "vue";

const props = defineProps({
  label: String,
  labelClass: {
    type: String,
    default: "label",
  },
  modelValue: Boolean,
  placeholder: String,
  help: String,
  name: String,
});

const emit = defineEmits(["update:modelValue"]);

const fieldProps = computed(() => ({
  label: props.label,
  help: props.help,
  labelClass: props.labelClass,
}));

const checkboxValue = computed({
  get: () => props.modelValue,
  set: (val) => emit("update:modelValue", val),
});
</script>

<template>
  <BulmaField v-slot="{ idForLabel }" v-bind="fieldProps">
    <div>
      <label class="checkbox">
        <input :id="idForLabel" v-model="checkboxValue" type="checkbox" />
        <slot></slot>
      </label>
    </div>
  </BulmaField>
</template>
