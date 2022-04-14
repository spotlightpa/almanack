<script>
import { ref } from "@vue/composition-api";

const canPaste = !!navigator.clipboard.readText;

export default {
  props: { files: Array, fileProps: Object },
  setup() {
    const inputText = ref("");

    return {
      canPaste,
      inputText,
      addText() {
        let text = inputText.value;
        inputText.value = "";
        this.$emit("add", text);
      },
      async pasteText() {
        let text = await navigator.clipboard.readText().catch(() => "");
        if (!text) {
          alert("Could not paste");
          return;
        }
        this.$emit("add", text);
      },
    };
  },
};
</script>

<template>
  <div>
    <div v-for="(file, i) of files" :key="i" class="is-flex pb-1">
      <div>
        <img :src="file" :alt="file" :title="file" class="thumb" />
      </div>
      <input
        class="ml-2 input"
        :value="file"
        readonly
        @input="$emit('change', [$event.currentTarget.value, i])"
      />
      <button
        type="button"
        class="ml-2 button is-danger has-text-weight-semibold"
        @click="$emit('remove', i)"
      >
        Remove
      </button>
    </div>
    <div class="is-flex">
      <input
        v-model="inputText"
        type="url"
        class="input"
        placeholder="https://example.com/image.png"
      />
      <button
        type="button"
        class="ml-2 button is-success has-text-weight-semibold"
        @click="addText"
      >
        Add
      </button>
      <button
        v-if="canPaste"
        type="button"
        class="ml-2 button is-primary has-text-weight-semibold"
        @click="pasteText"
      >
        Paste
      </button>
    </div>
    <PickerFiles
      :files="fileProps.files.value"
      @select-file="$emit('add', $event.url)"
    />
  </div>
</template>

<style scoped>
.thumb {
  object-fit: scale-down;
  width: 8rem;
  height: 2rem;
}
</style>
