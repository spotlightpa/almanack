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
  required: {
    type: Boolean,
    default: false,
  },
});

defineEmits(["update:modelValue"]);

const validationMessage = ref("");

const fieldProps = computed(() => ({
  label: props.label,
  help: props.help,
  labelClass: props.labelClass,
  required: props.required,
  validationMessage: validationMessage.value,
}));
</script>

<template>
  <BulmaField v-slot="{ idForLabel }" v-bind="fieldProps">
    <div class="is-flex is-align-items-center">
      <input
        :id="idForLabel"
        :value="modelValue"
        type="color"
        @input="$emit('update:modelValue', $event.target.value)"
      />
      <span class="ml-4 is-flex-grow-0">
        <input
          :value="modelValue"
          :required="required"
          :name="name"
          :placeholder="placeholder"
          type="text"
          class="input is-small"
          @input="$emit('update:modelValue', $event.target.value)"
        />
      </span>
      <BulmaPaste
        button-class="ml-2 button is-primary is-small has-text-weight-semibold"
        @paste="$emit('update:modelValue', $event)"
      ></BulmaPaste>
    </div>
  </BulmaField>
</template>
