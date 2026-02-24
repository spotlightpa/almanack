<script setup>
import { ref, onMounted, nextTick } from "vue";

const emit = defineEmits(["mounted"]);
const el = ref(null);

onMounted(async () => {
  await nextTick();
  emit(
    "mounted",
    Array.from(el.value.children)
      .map((child) =>
        child.tagName === "RAW-HTML" ? child.getAttribute("block") : child.outerHTML
      )
      .join("\n\n")
  );
});
</script>

<template>
  <div ref="el" hidden>
    <slot></slot>
  </div>
</template>
