<script>
let labelIDCounter = 0;

export default {
  emits: ["update:modelValue"],
  props: {
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
  },
  data() {
    labelIDCounter++;
    return {
      currentInput: "",
      hasFocus: false,
      dragoverIndex: null,
      dragoverFrom: null,
      idForDatalist: `BulmaAutocompleteArray-${labelIDCounter}`,
    };
  },
  computed: {
    fieldProps() {
      return {
        label: this.label,
        help: this.help,
        labelClass: this.labelClass,
        required: this.required,
      };
    },
  },
  watch: {},
  methods: {
    detectChange(value) {
      let oldValue = this.currentInput;
      this.currentInput = value;
      // assume big growth is from datalist entry
      if (this.currentInput.length - oldValue.length > 1) {
        this.push();
      }
    },
    push() {
      let input = this.currentInput.trim();
      if (!input) {
        return;
      }
      let vals = [...this.modelValue];
      vals.push(input);
      this.$emit("update:modelValue", vals);
      this.currentInput = "";
    },
    paste(event) {
      this.currentInput = event?.clipboardData?.getData?.("text") ?? "";
    },
    remove(i) {
      let vals = [...this.modelValue];
      vals.splice(i, 1);
      this.$emit("update:modelValue", vals);
    },
    dragover(i) {
      if (
        // abort if some random other thing is being dragged on it
        this.dragoverFrom === null ||
        // performance? avoid triggering reactivity system
        this.dragoverIndex === i
      ) {
        return;
      }
      this.dragoverIndex = i;
    },
    dragleave() {
      // performance? avoid triggering reactivity system
      if (this.dragoverIndex === null) {
        return;
      }
      this.dragoverIndex = null;
    },
    drop(to) {
      let from = this.dragoverFrom;
      if (from === null || from === to) {
        return;
      }
      let vals = [...this.modelValue];
      vals[from] = this.modelValue[to];
      vals[to] = this.modelValue[from];
      this.$emit("update:modelValue", vals);
    },
  },
};
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
