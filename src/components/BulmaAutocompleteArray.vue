<script>
import BulmaField from "./BulmaField.vue";

let labelIDCounter = 0;

export default {
  name: "BulmaAutocompleteArray",
  components: {
    BulmaField,
  },
  props: {
    label: String,
    labelClass: {
      type: String,
      default: "label",
    },
    value: Array,
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
        validationMessage: this.validationMessage,
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
      let vals = [...this.value];
      vals.push(input);
      this.$emit("input", vals);
      this.currentInput = "";
    },
    paste(event) {
      this.currentInput = event?.clipboardData?.getData?.("text") ?? "";
    },
    remove(i) {
      let vals = [...this.value];
      vals.splice(i, 1);
      this.$emit("input", vals);
    },
  },
};
</script>

<template>
  <BulmaField ref="field" v-slot="{ idForLabel }" v-bind="fieldProps">
    <div class="field is-grouped is-grouped-multiline">
      <div v-for="(v, i) of value" :key="i" class="control">
        <div class="tags has-addons">
          <span class="tag is-small" v-text="v" />
          <a class="tag is-delete" @click="remove(i)" />
        </div>
      </div>
    </div>

    <input
      :id="idForLabel"
      ref="input"
      :value="currentInput"
      class="input"
      :placeholder="placeholder"
      :required="required"
      :readonly="readonly"
      :list="idForDatalist"
      @input="detectChange($event.currentTarget.value)"
      @paste.prevent="paste"
      @keydown.enter="push"
      @focusin="hasFocus = true"
      @focusout="
        push();
        hasFocus = false;
      "
    />
    <datalist v-show="hasFocus" :id="idForDatalist">
      <option v-for="opt of options" :key="opt" :value="opt" />
    </datalist>
  </BulmaField>
</template>
