<script>
const canPaste = !!navigator.clipboard.readText;

export default {
  setup() {
    return {
      canPaste,
      async pasteText() {
        let text = await navigator.clipboard.readText().catch(() => "");
        if (!text) {
          alert("Could not paste");
          return;
        }
        this.$emit("paste", text);
      },
    };
  },
};
</script>

<template>
  <div>
    <button
      v-if="canPaste"
      type="button"
      class="ml-2 button is-primary has-text-weight-semibold"
      @click="pasteText"
    >
      <font-awesome-icon :icon="['fas', 'paste']" />
      <span class="ml-1">Paste</span>
    </button>
  </div>
</template>
