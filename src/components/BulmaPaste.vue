<script setup>
defineProps({
  buttonClass: {
    type: String,
    default: "ml-2 button is-primary has-text-weight-semibold",
  },
});

const emit = defineEmits(["paste"]);

const canPaste = !!navigator.clipboard.readText;

async function pasteText() {
  let text = await navigator.clipboard.readText().catch(() => "");
  if (!text) {
    alert("Could not paste");
    return;
  }
  emit("paste", text);
}
</script>

<template>
  <div>
    <button
      v-if="canPaste"
      type="button"
      :class="buttonClass"
      @click="pasteText"
    >
      <font-awesome-icon :icon="['fas', 'paste']"></font-awesome-icon>
      <span class="ml-1">Paste</span>
    </button>
  </div>
</template>
