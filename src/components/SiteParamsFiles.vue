<script>
import { ref } from "@vue/composition-api";

export default {
  props: { files: Array, fileProps: Object },
  setup() {
    const inputText = ref("");
    return {
      inputText,
      addText() {
        let text = inputText.value;
        inputText.value = "";
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
      <BulmaPaste @paste="$emit('add', $event)" />
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
