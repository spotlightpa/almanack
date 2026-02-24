<script setup>
import { ref } from "vue";

defineProps({
  value: String,
  label: {
    type: String,
    default: "text",
  },
  size: {
    type: String,
    default: "",
  },
});

const copier = ref(null);
const copied = ref(false);
</script>

<template>
  <div class="has-margin-bottom">
    <div class="field">
      <div class="control">
        <CopyTextarea
          ref="copier"
          :size="size + ' pre-wrap'"
          @copied="copied = $event"
          >{{ value }}</CopyTextarea
        >
      </div>
    </div>
    <div class="field">
      <div class="buttons">
        <button
          type="button"
          class="button is-primary has-text-weight-semibold"
          :class="size"
          title="Copy"
          @click="copier.copy()"
        >
          <span class="icon">
            <font-awesome-icon :icon="['far', 'copy']"></font-awesome-icon>
          </span>
          <span> Copy {{ label }} </span>
        </button>
        <transition name="fade">
          <div
            v-if="copied"
            class="tag is-rounded is-success is-light has-text-weight-semibold"
          >
            Copied
          </div>
        </transition>
      </div>
    </div>
  </div>
</template>

<style scoped>
.fade {
  transition: all 0.5s ease;
}
</style>
