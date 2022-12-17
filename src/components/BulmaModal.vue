<script setup>
import { computed } from "vue";
const props = defineProps({
  modelValue: Boolean,
});
const emit = defineEmits(["update:modelValue"]);
const open = computed({
  get: () => props.modelValue,
  set: (value) => emit("update:modelValue", value),
});
</script>

<template>
  <Teleport v-if="open" to="body">
    <div v-if="open" class="modal is-active">
      <div class="modal-background" @click="open = false"></div>
      <div class="modal-content" tabindex="-1" @keyup.esc="open = false">
        <slot />
      </div>
      <button
        class="modal-close is-large"
        aria-label="close"
        @click="open = false"
      ></button>
    </div>
  </Teleport>
</template>
