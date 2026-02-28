<script setup>
import { ref, nextTick } from "vue";

defineProps({
  size: {
    type: String,
    default: "is-medium",
  },
});

const emit = defineEmits(["copied"]);

const textarea = ref(null);
const isFocused = ref(false);

function click() {
  if (!isFocused.value) {
    isFocused.value = true;
    select();
  }
}

function select() {
  let range = document.createRange();
  range.selectNodeContents(textarea.value);
  let selection = window.getSelection();
  selection.removeAllRanges();
  selection.addRange(range);
}

async function copy() {
  select();
  await nextTick();
  if (document.execCommand("copy")) {
    emit("copied", true);
    window.setTimeout(() => {
      emit("copied", false);
    }, 5000); // 5s
  }
}

defineExpose({ copy });
</script>

<template>
  <div
    ref="textarea"
    class="textarea"
    rows="bulmaoverride"
    :class="size"
    :contenteditable="true"
    @click="click"
    @blur="isFocused = false"
    @focus="select"
  >
    <slot></slot>
  </div>
</template>

<style scoped>
.pre-wrap {
  white-space: pre-wrap;
}
</style>
