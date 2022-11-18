<script>
export default {
  name: "BulmaFieldInput",
  props: {
    label: String,
    labelClass: {
      type: String,
      default: "label",
    },
    modelValue: String,
    placeholder: String,
    help: String,
    validator: Function,
    name: String,
    minLength: {
      type: Number,
      default: null,
    },
    maxLength: {
      type: Number,
      default: null,
    },
    required: {
      type: Boolean,
      default: false,
    },
    autofocus: {
      type: Boolean,
      default: false,
    },
    readonly: {
      type: Boolean,
      default: false,
    },
    autocomplete: {
      type: String,
      default: null,
    },
    rows: {
      type: Number,
      default: 2,
    },
    cols: {
      type: Number,
      default: null,
    },
  },
  data() {
    return {
      validationMessage: "",
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
  methods: {
    updateValidationMessage(ev) {
      this.validationMessage = ev.target.validationMessage;
    },
    updateValue(ev) {
      let newVal = ev.target.value;
      if (this.maxLength && newVal.length > this.maxLength) {
        newVal = newVal.slice(0, this.maxLength);
      }
      this.updateValidationMessage(ev);
      this.$emit("update:modelValue", newVal);
    },
  },
};
</script>

<template>
  <BulmaField v-slot="{ idForLabel }" v-bind="fieldProps">
    <textarea
      :id="idForLabel"
      class="textarea"
      :autocomplete="autocomplete"
      :autofocus="autofocus"
      :maxlength="maxLength"
      :minlength="minLength"
      :name="name"
      :placeholder="placeholder"
      :required="required || null"
      :rows="rows"
      :cols="cols"
      :value="modelValue"
      :readonly="readonly || null"
      @blur="updateValidationMessage"
      @invalid="updateValidationMessage"
      @input="updateValue"
      @focusout="$emit('focusout', $event)"
    />
  </BulmaField>
</template>
