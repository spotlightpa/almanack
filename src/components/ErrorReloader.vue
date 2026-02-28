<script setup>
import { computed } from "vue";

const props = defineProps({
  error: [Error, String],
});

defineEmits(["reload"]);

const name = computed(() => props.error?.name ?? props.error ?? "Error");
const message = computed(() => props.error?.message ?? props.error ?? "");
const details = computed(() => Object.entries(props.error?.details ?? {}));
</script>

<template>
  <div v-if="error" class="message is-danger">
    <div class="message-header">{{ name }}</div>
    <div class="message-body">
      <p class="content">{{ message }}</p>
      <details>
        <p v-for="[key, vals] of details" :key="key" class="content">
          <strong v-text="key"></strong>:
          <span v-for="(val, i) of vals" :key="i" v-text="val"></span>
        </p>
      </details>
      <div class="mt-4 buttons">
        <button
          class="button is-danger has-text-weight-semibold"
          @click="$emit('reload')"
        >
          Reload?
        </button>
      </div>
    </div>
  </div>
</template>
