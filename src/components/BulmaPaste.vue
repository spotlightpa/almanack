<script>
const canPaste = !!navigator.clipboard.readText;

export default {
  props: {
    buttonClass: {
      type: String,
      default: "ml-2 button is-primary has-text-weight-semibold",
    },
  },
  setup(_, { emit }) {
    return {
      canPaste,
      async pasteText() {
        let text = await navigator.clipboard.readText().catch(() => "");
        if (!text) {
          alert("Could not paste");
          return;
        }
        emit("paste", text);
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
      :class="buttonClass"
      @click="pasteText"
    >
      <font-awesome-icon :icon="['fas', 'paste']" />
      <span class="ml-1">Paste</span>
    </button>
  </div>
</template>
