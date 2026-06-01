<script setup>
import { watchAPI } from "@/api/service-util.js";
import { get, getPage } from "@/api/client-v2.js";
import { Page } from "@/api/spotlightpa-page.js";

const props = defineProps({
  filePath: String,
});

const { computedObj } = watchAPI(
  () => props.filePath,
  (path) => get(getPage, { by: "filepath", value: path, select: "-body" })
);

const name = computedObj((obj) => {
  if (!obj) return "";
  let page = new Page(obj);
  return page.internalID || page.title || "<no title>";
});
</script>

<template>
  <span class="tag is-medium spacer select-none">
    {{ name }}
    <button class="delete" @click="$emit('remove')"></button>
  </span>
</template>

<style scoped>
.plain-notification {
  max-width: 350px;
  border-radius: 4px;
  padding: 1.25rem 1.5rem 1.25rem 1.5rem;
}
</style>
