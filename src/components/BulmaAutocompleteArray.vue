<script setup>
import { ref, computed } from "vue";
import newID from "@/utils/new-id.js";

const props = defineProps({
  label: String,
  labelClass: {
    type: String,
    default: "label",
  },
  modelValue: Array,
  options: Array,
  placeholder: String,
  help: String,
  validator: Function,
  required: {
    type: Boolean,
    default: false,
  },
  readonly: {
    type: Boolean,
    default: false,
  },
});

const emit = defineEmits(["update:modelValue"]);

let labelIDCounter = newID();
const idForDatalist = `BulmaAutocompleteArray-${labelIDCounter}`;

const currentInput = ref("");
const hasFocus = ref(false);
const dragoverIndex = ref(null);
const dragoverFrom = ref(null);

const fieldProps = computed(() => ({
  label: props.label,
  help: props.help,
  labelClass: props.labelClass,
  required: props.required,
}));

function detectChange(value) {
  let oldValue = currentInput.value;
  currentInput.value = value;
  // assume big growth is from datalist entry
  if (currentInput.value.length - oldValue.length > 1) {
    push();
  }
}

function push() {
  let input = currentInput.value.trim();
  if (!input) {
    return;
  }
  let vals = [...props.modelValue];
  vals.push(input);
  emit("update:modelValue", vals);
  currentInput.value = "";
}

function paste(event) {
  currentInput.value = event?.clipboardData?.getData?.("text") ?? "";
}

function remove(i) {
  let vals = [...props.modelValue];
  vals.splice(i, 1);
  emit("update:modelValue", vals);
}

function dragover(i) {
  if (
    // abort if some random other thing is being dragged on it
    dragoverFrom.value === null ||
    // performance? avoid triggering reactivity system
    dragoverIndex.value === i
  ) {
    return;
  }
  dragoverIndex.value = i;
}

function dragleave() {
  // performance? avoid triggering reactivity system
  if (dragoverIndex.value === null) {
    return;
  }
  dragoverIndex.value = null;
}

function drop(to) {
  let from = dragoverFrom.value;
  if (from === null || from === to) {
    return;
  }
  let vals = [...props.modelValue];
  vals[from] = props.modelValue[to];
  vals[to] = props.modelValue[from];
  emit("update:modelValue", vals);
}
</script>

<template>
  <BulmaField v-slot="{ idForLabel }" v-bind="fieldProps">
    <div class="field is-grouped is-grouped-multiline">
      <div v-for="(v, i) of modelValue" :key="i" class="control">
        <div
          class="tags has-addons"
          :draggable="true"
          @dragstart="dragoverFrom = i"
          @dragover.prevent.stop="dragover(i)"
          @dragleave="dragleave()"
          @drop.prevent="drop(i)"
          @dragend="
            dragoverFrom = null;
            dragoverIndex = null;
          "
        >
          <span
            :class="{
              'is-link': dragoverIndex === i && dragoverFrom !== i,
              grabbable: dragoverFrom === null,
              grabbed: dragoverFrom !== null,
            }"
            class="tag is-small"
            v-text="v"
          ></span>
          <a
            :class="{ 'is-link': dragoverIndex === i && dragoverFrom !== i }"
            class="tag is-delete"
            @click="remove(i)"
          ></a>
        </div>
      </div>
    </div>

    <input
      :id="idForLabel"
      :value="currentInput"
      class="input"
      :placeholder="placeholder"
      :required="required"
      :readonly="readonly"
      :list="idForDatalist"
      @input="detectChange($event.currentTarget.value)"
      @paste.prevent="paste"
      @keydown.enter="push"
      @keydown.esc="currentInput = ''"
      @focusin="hasFocus = true"
      @focusout="
        push();
        hasFocus = false;
      "
    />
    <datalist v-show="hasFocus" :id="idForDatalist">
      <option v-for="opt of options" :key="opt" :value="opt"></option>
    </datalist>
  </BulmaField>
</template>

<style scoped>
.grabbable {
  cursor: grab;
  user-select: none;
}
.grabbed {
  cursor: grabbing;
}
</style>
